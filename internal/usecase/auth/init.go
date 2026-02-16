package auth

import (
	"context"

	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/internal/repository/auth"
)

type AuthUsecase interface {
	GetProfile(ctx context.Context) (*base.User, error)
}

// InitAuthUsecase initializes the auth usecase
func InitAuthUsecase(repo auth.AuthRepository, config *config.Main) AuthUsecase {
	return NewAuthUsecase(repo, config)
}