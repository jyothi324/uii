package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	controlPlaneDB   *gorm.DB
	controlPlaneOnce sync.Once
	controlPlaneErr  error
)

// InitControlPlaneDB connects to the shared SaaS billing control-plane database
// (tulip_control by default). Safe to call when billing env is not configured —
// GetControlPlaneDB will return an error and billing handlers respond with 503.
func InitControlPlaneDB() error {
	controlPlaneOnce.Do(func() {
		dbName := os.Getenv("CONTROL_PLANE_DB_NAME")
		if dbName == "" {
			dbName = "tulip_control"
		}

		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		sslMode := os.Getenv("DB_SSL_MODE")
		if sslMode == "" {
			sslMode = "disable"
		}

		if host == "" || user == "" {
			controlPlaneErr = fmt.Errorf("control-plane DB: DB_HOST/DB_USER not configured")
			log.Printf("[billing] %v — billing endpoints will return 503", controlPlaneErr)
			return
		}

		if err := ensureDatabaseExists(host, port, user, password, sslMode, dbName); err != nil {
			controlPlaneErr = fmt.Errorf("ensure control-plane db: %w", err)
			log.Printf("[billing] %v", controlPlaneErr)
			return
		}

		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			host, user, password, dbName, port, sslMode,
		)
		gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger:                                   logger.Default.LogMode(getGormLogLevel()),
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			controlPlaneErr = err
			log.Printf("[billing] failed to open control-plane DB: %v", err)
			return
		}

		sqlDB, err := gdb.DB()
		if err != nil {
			controlPlaneErr = err
			return
		}
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(time.Hour)

		controlPlaneDB = gdb
		log.Printf("[billing] control-plane DB connected (%s)", dbName)
	})
	return controlPlaneErr
}

func ensureDatabaseExists(host, port, user, password, sslMode, dbName string) error {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		host, port, user, password, sslMode,
	)
	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	var exists bool
	if err := sqlDB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		dbName,
	).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}

	// CREATE DATABASE cannot run in a prepared statement / transaction with params.
	_, err = sqlDB.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, dbName))
	return err
}

// GetControlPlaneDB returns the shared control-plane gorm DB.
func GetControlPlaneDB() (*gorm.DB, error) {
	if controlPlaneDB == nil {
		if err := InitControlPlaneDB(); err != nil {
			return nil, err
		}
	}
	if controlPlaneDB == nil {
		if controlPlaneErr != nil {
			return nil, controlPlaneErr
		}
		return nil, fmt.Errorf("control-plane DB not initialized")
	}
	return controlPlaneDB, nil
}
