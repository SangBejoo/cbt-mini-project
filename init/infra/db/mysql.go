package db

import (
	"cbt-test-mini-project/init/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func OpenSQL(cfgMain config.Main) (db *gorm.DB, err error) {
	cfg := cfgMain.Database
	db, err = gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
