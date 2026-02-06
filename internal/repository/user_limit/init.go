package user_limit

import (
	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/internal/repository"
	"database/sql"
)

// Init initializes the user limit repository
func Init(db *sql.DB, cfg *config.Main) repository.UserLimitRepository {
	return repository.NewUserLimitRepository(db, cfg)
}