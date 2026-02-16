package auth

import (
	"context"

	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/usecase/auth"

	"google.golang.org/protobuf/types/known/emptypb"
)

// authHandler implements the AuthServiceServer.
type authHandler struct {
	base.UnimplementedAuthServiceServer
	usecase auth.AuthUsecase
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(usecase auth.AuthUsecase) base.AuthServiceServer {
	return &authHandler{usecase: usecase}
}

// GetProfile handles getting the current user's profile.
func (h *authHandler) GetProfile(ctx context.Context, req *emptypb.Empty) (*base.UserResponse, error) {
	user, err := h.usecase.GetProfile(ctx)
	if err != nil {
		return &base.UserResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &base.UserResponse{
		User:    user,
		Success: true,
		Message: "Profile retrieved successfully",
	}, nil
}
