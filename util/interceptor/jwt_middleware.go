package interceptor

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/init/config"
	"context"
	"errors"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// JWTClaims represents the JWT claims
type JWTClaims struct {
	UserID int32  `json:"user_id"`
	Email  string `json:"email"`
	Role   int32  `json:"role"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

// JWTMiddleware validates JWT tokens for protected endpoints
type JWTMiddleware struct {
	config *config.Main
}

// NewJWTMiddleware creates a new JWT middleware
func NewJWTMiddleware(config *config.Main) *JWTMiddleware {
	return &JWTMiddleware{config: config}
}

// UnaryServerInterceptor returns a gRPC unary server interceptor for JWT validation
func (m *JWTMiddleware) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip authentication for login, refresh token, and health check
		if m.shouldSkipAuth(info.FullMethod) {
			return handler(ctx, req)
		}

		// Extract token from metadata
		token, err := ExtractTokenFromContext(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		// Validate token
		claims, err := ValidateToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		// Add user info to context
		ctx = AddUserToContext(ctx, claims)

		return handler(ctx, req)
	}
}

// shouldSkipAuth determines if authentication should be skipped for the method
func (m *JWTMiddleware) shouldSkipAuth(method string) bool {
	skipMethods := []string{
		"/base.Base/HealthCheck",
		"/base.AuthService/Login",
		"/base.AuthService/RefreshToken",
		"/base.AuthService/CreateUser",
	}

	for _, skip := range skipMethods {
		if method == skip {
			return true
		}
	}
	return false
}

// ExtractTokenFromContext extracts JWT token from gRPC metadata
func ExtractTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("missing metadata")
	}

	authHeader, exists := md["authorization"]
	if !exists || len(authHeader) == 0 {
		return "", errors.New("missing authorization header")
	}

	// Extract token from "Bearer <token>"
	parts := strings.SplitN(authHeader[0], " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// ValidateToken validates the JWT token and returns claims
func ValidateToken(tokenString string) (*JWTClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-super-secret-jwt-key-change-this-in-production"
	}
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// Check if it's an access token
		if claims.Type != "access" {
			return nil, errors.New("invalid token type")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// AddUserToContext adds user information to the context
func AddUserToContext(ctx context.Context, claims *JWTClaims) context.Context {
	user := &base.User{
		Id:    claims.UserID,
		Email: claims.Email,
		Role:  base.UserRole(claims.Role),
	}

	return context.WithValue(ctx, "user", user)
}

// GetUserFromContext extracts user from context
func GetUserFromContext(ctx context.Context) (*base.User, error) {
	user, ok := ctx.Value("user").(*base.User)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}