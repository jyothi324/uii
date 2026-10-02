package db

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// AITenantDBManager mirrors TenantDBManager but connects using a separate,
// read-only Postgres role (see db/readonly_role.sql). It's the real
// safety backstop for the AI chatbot feature: even if the app-level SQL
// guardrail (services/sql_guardrail.go) is somehow bypassed, a query
// executed through this manager physically cannot mutate data.
//
// Unlike TenantDBManager, this never calls ensureDatabaseExists — a
// read-only role has no business creating databases, and the tenant DB is
// always already provisioned by the time an AI query runs against it.
type AITenantDBManager struct {
	connections map[string]*GormDB
	mu          sync.RWMutex
	config      *DBConfig
}

var (
	aiTenantManager *AITenantDBManager
	aiManagerOnce   sync.Once
)

// InitAITenantManager wires up the AI connection pool. Safe to call even
// before db/readonly_role.sql has been run in a given environment — see
// GetAIDBConfig, which falls back to the app's normal read-write
// credentials until AI_DB_USER/AI_DB_PASSWORD are configured, so the
// feature degrades to "app-level guardrail only" rather than failing
// outright during rollout.
func InitAITenantManager(config *DBConfig) *AITenantDBManager {
	aiManagerOnce.Do(func() {
		aiTenantManager = &AITenantDBManager{
			connections: make(map[string]*GormDB),
			config:      config,
		}
	})
	return aiTenantManager
}

func GetAITenantManager() *AITenantDBManager {
	return aiTenantManager
}

func (m *AITenantDBManager) GetTenantDB(tenant string) (Database, error) {
	m.mu.RLock()
	db, exists := m.connections[tenant]
	m.mu.RUnlock()
	if exists && db != nil {
		sqlDB, err := db.DB.DB()
		if err == nil {
			if err := sqlDB.Ping(); err == nil {
				return db, nil
			}
		}
		m.mu.Lock()
		delete(m.connections, tenant)
		m.mu.Unlock()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if db, exists := m.connections[tenant]; exists {
		return db, nil
	}
	db, err := m.createTenantConnection(tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI read-only connection for tenant %s: %w", tenant, err)
	}
	m.connections[tenant] = db
	return db, nil
}

func (m *AITenantDBManager) createTenantConnection(tenant string) (*GormDB, error) {
	var dbName string
	if tenant == "default" {
		dbName = "tenant_app"
	} else {
		dbName = fmt.Sprintf("tenant_%s", tenant)
	}
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		m.config.Host,
		m.config.Port,
		m.config.User,
		m.config.Password,
		dbName,
		m.config.SSLMode,
	)
	logLevel := getGormLogLevel()
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database %s as AI read-only role: %w", dbName, err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	// Small dedicated pool — this connection is only used for occasional
	// chatbot queries, not general app traffic.
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(90 * time.Minute)
	return &GormDB{DB: gormDB}, nil
}

func (m *AITenantDBManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for tenant, db := range m.connections {
		if sqlDB, err := db.DB.DB(); err == nil {
			sqlDB.Close()
		}
		log.Printf("✅ Closed AI read-only connection for tenant: %s", tenant)
	}
	m.connections = make(map[string]*GormDB)
}

// GetAIDBConfig reads AI_DB_USER/AI_DB_PASSWORD for the dedicated
// read-only role. If they're unset — i.e. db/readonly_role.sql hasn't
// been run in this environment yet — it falls back to the same
// credentials the main app uses (DB_USER/DB_PASSWORD), so the feature
// still works end-to-end during rollout, backed only by the app-level
// guardrail until the role is provisioned. This fallback is intentional
// and logged loudly so it isn't mistaken for the real safety posture.
func GetAIDBConfig() *DBConfig {
	base := GetDBConfig()

	aiUser := os.Getenv("AI_DB_USER")
	aiPassword := os.Getenv("AI_DB_PASSWORD")
	if aiUser == "" || aiPassword == "" {
		log.Printf("⚠️  AI_DB_USER/AI_DB_PASSWORD not set — AI chatbot queries will run under the " +
			"normal read-write DB role. Run db/readonly_role.sql and set these env vars to enable " +
			"the read-only safety backstop.")
		return base
	}

	return &DBConfig{
		Host:            base.Host,
		Port:            base.Port,
		User:            aiUser,
		Password:        aiPassword,
		SSLMode:         base.SSLMode,
		MaxOpenConns:    5,
		MaxIdleConns:    5,
		ConnMaxLifetime: 90 * time.Minute,
	}
}
