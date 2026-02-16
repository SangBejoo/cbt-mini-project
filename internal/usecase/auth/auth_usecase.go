package auth

import (
	"context"

	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/internal/repository/auth"
	"cbt-test-mini-project/util/interceptor"
)

// authUsecaseImpl implements AuthUsecase.
type authUsecaseImpl struct{}

// NewAuthUsecase creates a new auth usecase.
// Parameters are retained for dependency wiring compatibility.
func NewAuthUsecase(_ auth.AuthRepository, _ *config.Main) AuthUsecase {
	return &authUsecaseImpl{}
}

// GetProfile gets the current authenticated user's profile.
func (u *authUsecaseImpl) GetProfile(ctx context.Context) (*base.User, error) {
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}
