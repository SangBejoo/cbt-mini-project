package db

import (
	"database/sql"

	"cbt-test-mini-project/init/config"

	_ "github.com/lib/pq" // or any other driver you want to use
)

func OpenSQL(cfgMain config.Main) (db *sql.DB, err error) {
	cfg := cfgMain.Database
	db, err = sql.Open(cfg.DriverName, cfg.DSN)
	if err != nil {
		return nil, err
	}
	return db, nil
}
