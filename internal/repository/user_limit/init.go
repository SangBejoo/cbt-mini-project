package user_limit

import (
	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/internal/repository"

	"gorm.io/gorm"
)

// Init initializes the user limit repository
func Init(db *gorm.DB, cfg *config.Main) repository.UserLimitRepository {
	return repository.NewUserLimitRepository(db, cfg)
}