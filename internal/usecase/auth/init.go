package auth

import (
	"context"
	"time"

	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/internal/repository/auth"
)

type AuthUsecase interface {
	Login(ctx context.Context, email, password string) (string, string, *base.User, time.Time, error) // access_token, refresh_token, user, expires_at
	GetUserByID(ctx context.Context, id int32) (*base.User, error)
	GetUserByEmail(ctx context.Context, email string) (*base.User, error)
	CreateUser(ctx context.Context, user *base.User) (*base.User, error)
	UpdateUser(ctx context.Context, id int32, updates map[string]interface{}) (*base.User, error)
	DeleteUser(ctx context.Context, id int32) error
	ListUsers(ctx context.Context, role int32, statusFilter int32, limit, offset int) ([]*base.User, int, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, time.Time, error) // new_access_token, new_refresh_token, expires_at
	GetProfile(ctx context.Context) (*base.User, error)
}

// InitAuthUsecase initializes the auth usecase
func InitAuthUsecase(repo auth.AuthRepository, config *config.Main) AuthUsecase {
	return NewAuthUsecase(repo, config)
}