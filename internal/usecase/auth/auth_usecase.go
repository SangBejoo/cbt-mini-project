package auth

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/internal/repository/auth"
	"cbt-test-mini-project/util/interceptor"
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// authUsecaseImpl implements AuthUsecase
type authUsecaseImpl struct {
	repo   auth.AuthRepository
	config *config.Main
}

// NewAuthUsecase creates a new auth usecase
func NewAuthUsecase(repo auth.AuthRepository, config *config.Main) AuthUsecase {
	return &authUsecaseImpl{repo: repo, config: config}
}

// Login authenticates a user
func (u *authUsecaseImpl) Login(ctx context.Context, email, password string) (string, string, *base.User, time.Time, error) {
	if email == "" || password == "" {
		return "", "", nil, time.Time{}, errors.New("email and password are required")
	}

	user, err := u.repo.Login(ctx, email, password)
	if err != nil {
		return "", "", nil, time.Time{}, err
	}

	// Generate access token
	accessToken, accessExpiresAt, err := u.generateAccessToken(user)
	if err != nil {
		return "", "", nil, time.Time{}, err
	}

	// Generate refresh token
	refreshToken, _, err := u.generateRefreshToken(user)
	if err != nil {
		return "", "", nil, time.Time{}, err
	}

	return accessToken, refreshToken, user, accessExpiresAt, nil
}

// GetUserByID retrieves a user by ID
func (u *authUsecaseImpl) GetUserByID(ctx context.Context, id int32) (*base.User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user ID")
	}

	return u.repo.GetUserByID(ctx, id)
}

// GetUserByEmail retrieves a user by email
func (u *authUsecaseImpl) GetUserByEmail(ctx context.Context, email string) (*base.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	return u.repo.GetUserByEmail(ctx, email)
}

// CreateUser creates a new user
func (u *authUsecaseImpl) CreateUser(ctx context.Context, user *base.User) (*base.User, error) {
	if user == nil {
		return nil, errors.New("user data is required")
	}
	if user.Email == "" || user.Nama == "" {
		return nil, errors.New("email and nama are required")
	}

	// Check if user already exists
	existing, err := u.repo.GetUserByEmail(ctx, user.Email)
	if err == nil && existing != nil {
		return nil, errors.New("user with this email already exists")
	}

	return u.repo.CreateUser(ctx, user)
}

// UpdateUser updates a user
func (u *authUsecaseImpl) UpdateUser(ctx context.Context, id int32, updates map[string]interface{}) (*base.User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user ID")
	}
	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	// Validate email uniqueness if email is being updated
	if email, ok := updates["email"]; ok {
		emailStr, ok := email.(string)
		if !ok {
			return nil, errors.New("invalid email format")
		}
		existing, err := u.repo.GetUserByEmail(ctx, emailStr)
		if err == nil && existing != nil && existing.Id != id {
			return nil, errors.New("email already in use")
		}
	}

	return u.repo.UpdateUser(ctx, id, updates)
}

// DeleteUser deletes a user
func (u *authUsecaseImpl) DeleteUser(ctx context.Context, id int32) error {
	if id <= 0 {
		return errors.New("invalid user ID")
	}

	return u.repo.DeleteUser(ctx, id)
}

// ListUsers lists users with filters
func (u *authUsecaseImpl) ListUsers(ctx context.Context, role int32, statusFilter int32, limit, offset int) ([]*base.User, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return u.repo.ListUsers(ctx, role, statusFilter, limit, offset)
}

// RefreshToken validates refresh token and generates new tokens
func (u *authUsecaseImpl) RefreshToken(ctx context.Context, refreshToken string) (string, string, time.Time, error) {
	// Parse and validate refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(u.config.JWT.Secret), nil
	})

	if err != nil {
		return "", "", time.Time{}, errors.New("invalid refresh token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check token type
		if tokenType, ok := claims["type"].(string); !ok || tokenType != "refresh" {
			return "", "", time.Time{}, errors.New("invalid token type")
		}

		// Get user ID from claims
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return "", "", time.Time{}, errors.New("invalid user ID in token")
		}
		userID := int32(userIDFloat)

		// Get user from database
		user, err := u.repo.GetUserByID(ctx, userID)
		if err != nil {
			return "", "", time.Time{}, errors.New("user not found")
		}

		// Generate new tokens
		accessToken, accessExpiresAt, err := u.generateAccessToken(user)
		if err != nil {
			return "", "", time.Time{}, err
		}

		refreshToken, _, err := u.generateRefreshToken(user)
		if err != nil {
			return "", "", time.Time{}, err
		}

		return accessToken, refreshToken, accessExpiresAt, nil
	}

	return "", "", time.Time{}, errors.New("invalid refresh token")
}

// generateAccessToken generates an access JWT token for the user
func (u *authUsecaseImpl) generateAccessToken(user *base.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(u.config.JWT.AccessTokenTTL) * time.Minute)

	claims := jwt.MapClaims{
		"user_id": user.Id,
		"email":   user.Email,
		"role":    user.Role,
		"type":    "access",
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(u.config.JWT.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// generateRefreshToken generates a refresh JWT token for the user
func (u *authUsecaseImpl) generateRefreshToken(user *base.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(u.config.JWT.RefreshTokenTTL) * 24 * time.Hour)

	claims := jwt.MapClaims{
		"user_id": user.Id,
		"email":   user.Email,
		"type":    "refresh",
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(u.config.JWT.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// GetProfile gets the current authenticated user's profile
func (u *authUsecaseImpl) GetProfile(ctx context.Context) (*base.User, error) {
	// Get user from JWT context (set by JWT middleware)
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Optionally fetch fresh user data from database if needed
	// For now, return the user from context which should be sufficient
	return user, nil
}