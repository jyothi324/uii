package db

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type TenantDBManager struct {
	connections map[string]*GormDB
	mu          sync.RWMutex
	config      *DBConfig
}

type DBConfig struct {
	Host                string
	Port                string
	User                string
	Password            string
	SSLMode             string
	MaxOpenConns        int
	MaxIdleConns        int
	ConnMaxLifetime     time.Duration
	MaxWSConns          int
	WorkerPoolSize      int
	BatchUpdateSize     int
	BatchUpdateInterval time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	ShutdownTimeout     time.Duration
}

var (
	tenantManager *TenantDBManager
	managerOnce   sync.Once
)

func InitTenantManager(config *DBConfig) *TenantDBManager {
	managerOnce.Do(func() {
		tenantManager = &TenantDBManager{
			connections: make(map[string]*GormDB),
			config:      config,
		}
	})
	return tenantManager
}
func GetTenantManager() *TenantDBManager {
	return tenantManager
}

func getGormLogLevel() logger.LogLevel {
	lvl := strings.ToLower(strings.TrimSpace(os.Getenv("GORM_LOG_LEVEL")))
	switch lvl {
	case "silent":
		return logger.Silent
	case "info":
		return logger.Info
	case "warn", "warning":
		return logger.Warn
	case "error", "":
		return logger.Error
	default:
		return logger.Error
	}
}

func (m *TenantDBManager) GetTenantDB(tenant string) (Database, error) {
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
		return nil, fmt.Errorf("failed to create connection for tenant %s: %w", tenant, err)
	}
	m.connections[tenant] = db
	return db, nil
}
func (m *TenantDBManager) createTenantConnection(tenant string) (*GormDB, error) {
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
	if tenant != "default" {
		if err := m.ensureDatabaseExists(dbName); err != nil {
			return nil, fmt.Errorf("failed to ensure database exists: %w", err)
		}
	}
	logLevel := getGormLogLevel()
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true, // Ignore "record not found"
			Colorful:                  true,
		},
	)
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database %s: %w", dbName, err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(m.config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(m.config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(m.config.ConnMaxLifetime)
	db := &GormDB{DB: gormDB}
	// Migrate(db)
	return db, nil
}
func (m *TenantDBManager) ensureDatabaseExists(dbName string) error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		m.config.Host,
		m.config.Port,
		m.config.User,
		m.config.Password,
		m.config.SSLMode,
	)
	loglevel := getGormLogLevel()
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  loglevel, // Only real errors
			IgnoreRecordNotFoundError: true,     // Ignore "record not found"
			Colorful:                  true,
		},
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	var exists bool
	err = db.Raw("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = ?)", dbName).Scan(&exists).Error
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("failed to create database %s: %w", dbName, err)
	} else {
	}
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
	return nil
}
func (m *TenantDBManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for tenant, db := range m.connections {
		if sqlDB, err := db.DB.DB(); err == nil {
			sqlDB.Close()
		}
		log.Printf("✅ Closed database connection for tenant: %s", tenant)
	}
	m.connections = make(map[string]*GormDB)
}
func (m *TenantDBManager) GetActiveTenantsCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connections)
}
func (m *TenantDBManager) GetActiveTenants() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenants := make([]string, 0, len(m.connections))
	for tenant := range m.connections {
		tenants = append(tenants, tenant)
	}
	return tenants
}
func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		var i int
		if _, err := fmt.Sscanf(v, "%d", &i); err == nil {
			return i
		}
	}
	return def
}
func GetDBConfig() *DBConfig {
	sslMode := os.Getenv("DB_SSL_MODE")
	if sslMode == "" {
		sslMode = "disable"
	}
	return &DBConfig{
		Host:            os.Getenv("DB_HOST"),
		Port:            os.Getenv("DB_PORT"),
		User:            os.Getenv("DB_USER"),
		Password:        os.Getenv("DB_PASSWORD"),
		SSLMode:         sslMode,
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 90 * time.Minute,
	}
}
