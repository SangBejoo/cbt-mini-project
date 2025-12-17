package infra

import (
	"fmt"
	"log"

	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/init/infra/db"

	"gorm.io/gorm"
)

type Repository struct {
	GormDB *gorm.DB
}

func (r *Repository) Close() error {
	if r != nil && r.GormDB != nil {
		sqlDB, _ := r.GormDB.DB()
		if err := sqlDB.Close(); err != nil {
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
		GormDB: dbConn,
	}
}
