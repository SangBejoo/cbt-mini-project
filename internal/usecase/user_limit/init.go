package user_limit

import (
	"cbt-test-mini-project/internal/repository"
	"cbt-test-mini-project/internal/usecase"
)

// Init initializes the user limit usecase
func Init(userLimitRepo repository.UserLimitRepository) usecase.UserLimitUsecase {
	return usecase.NewUserLimitUsecase(userLimitRepo)
}