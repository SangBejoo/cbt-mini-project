package auth

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/usecase/auth"
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// authHandler implements the AuthServiceServer
type authHandler struct {
	base.UnimplementedAuthServiceServer
	usecase auth.AuthUsecase
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(usecase auth.AuthUsecase) base.AuthServiceServer {
	return &authHandler{usecase: usecase}
}

// Login handles user login
func (h *authHandler) Login(ctx context.Context, req *base.LoginRequest) (*base.LoginResponse, error) {
	accessToken, refreshToken, user, expiresAt, err := h.usecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		return &base.LoginResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &base.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresAt:    timestamppb.New(expiresAt),
		Success:      true,
		Message:      "Login successful",
	}, nil
}

// RefreshToken handles token refresh
func (h *authHandler) RefreshToken(ctx context.Context, req *base.RefreshTokenRequest) (*base.RefreshTokenResponse, error) {
	accessToken, refreshToken, expiresAt, err := h.usecase.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return &base.RefreshTokenResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &base.RefreshTokenResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    timestamppb.New(expiresAt),
		Success:      true,
		Message:      "Token refreshed successfully",
	}, nil
}

// GetUser handles getting user by ID
func (h *authHandler) GetUser(ctx context.Context, req *base.GetUserRequest) (*base.UserResponse, error) {
	user, err := h.usecase.GetUserByID(ctx, req.Id)
	if err != nil {
		return &base.UserResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &base.UserResponse{
		User:    user,
		Success: true,
		Message: "User retrieved successfully",
	}, nil
}

// CreateUser handles creating a new user
func (h *authHandler) CreateUser(ctx context.Context, req *base.CreateUserRequest) (*base.UserResponse, error) {
	user := &base.User{
		Email:    req.Email,
		Nama:     req.Nama,
		Role:     req.Role,
		IsActive: true,
	}
	user, err := h.usecase.CreateUser(ctx, user)
	if err != nil {
		return &base.UserResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &base.UserResponse{
		User:    user,
		Success: true,
		Message: "User created successfully",
	}, nil
}

// UpdateUser handles updating a user
func (h *authHandler) UpdateUser(ctx context.Context, req *base.UpdateUserRequest) (*base.UserResponse, error) {
	updates := make(map[string]interface{})
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Nama != "" {
		updates["nama"] = req.Nama
	}
	updates["role"] = req.Role
	updates["is_active"] = req.IsActive

	user, err := h.usecase.UpdateUser(ctx, req.Id, updates)
	if err != nil {
		return &base.UserResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &base.UserResponse{
		User:    user,
		Success: true,
		Message: "User updated successfully",
	}, nil
}

// DeleteUser handles deleting a user
func (h *authHandler) DeleteUser(ctx context.Context, req *base.DeleteUserRequest) (*base.MessageStatusResponse, error) {
	err := h.usecase.DeleteUser(ctx, req.Id)
	if err != nil {
		return &base.MessageStatusResponse{
			Message: err.Error(),
			Status:  "error",
		}, nil
	}

	return &base.MessageStatusResponse{
		Message: "User deleted successfully",
		Status:  "success",
	}, nil
}

// ListUsers handles listing users
func (h *authHandler) ListUsers(ctx context.Context, req *base.ListUsersRequest) (*base.ListUsersResponse, error) {
	page := 1
	pageSize := 10
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = int(req.Pagination.Page)
		}
		if req.Pagination.PageSize > 0 {
			pageSize = int(req.Pagination.PageSize)
		}
	}

	users, total, err := h.usecase.ListUsers(ctx, int32(req.Role), int32(req.StatusFilter), pageSize, (page-1)*pageSize)
	if err != nil {
		return &base.ListUsersResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &base.ListUsersResponse{
		Users: users,
		Pagination: &base.PaginationResponse{
			TotalCount:  int32(total),
			TotalPages:  int32((total + pageSize - 1) / pageSize),
			CurrentPage: int32(page),
			PageSize:    int32(pageSize),
		},
		Success: true,
		Message: "Users retrieved successfully",
		Total:   int32(total),
	}, nil
}

// GetProfile handles getting the current user's profile
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