package db

import (
	"cbt-test-mini-project/init/config"
	"database/sql"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func OpenSQL(cfgMain config.Main) (db *gorm.DB, err error) {
	cfg := cfgMain.Database
	db, err = gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	configureConnectionPool(sqlDB, cfg)

	return db, nil
}

func configureConnectionPool(sqlDB *sql.DB, cfg config.Database) {
	// Set maximum number of open connections to the database
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	// Set maximum number of idle connections in the pool
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	// Set maximum amount of time a connection may be reused
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)
}
