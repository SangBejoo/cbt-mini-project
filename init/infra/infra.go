package infra

import (
	"database/sql"
	"fmt"
	"log"

	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/init/infra/db"
)

type Repository struct {
	DB *sql.DB
}

func (r *Repository) Close() error {
	if r != nil && r.DB != nil {
		if err := r.DB.Close(); err != nil {
			return err
		}
	}
	return nil
}

func LoadRepository(cfg config.Main) *Repository {
	dbConn, err := db.OpenSQL(cfg)
	fmt.Println("err", err)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return &Repository{
		DB: dbConn,
	}
}
