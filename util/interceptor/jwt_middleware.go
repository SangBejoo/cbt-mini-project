package interceptor

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/init/config"
	authRepo "cbt-test-mini-project/internal/repository/auth"
	"context"
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// JWTClaims represents the JWT claims
type JWTClaims struct {
	LMSUserID int64  `json:"lms_user_id,omitempty"`
	UserID    int32  `json:"user_id,omitempty"`
	Email     string `json:"email"`
	FullName  string `json:"full_name,omitempty"`
	Nama      string `json:"nama,omitempty"`
	RoleName  string `json:"role_name,omitempty"`
	Type      string `json:"type,omitempty"`
	jwt.RegisteredClaims
}

// JWTMiddleware validates JWT tokens for protected endpoints
type JWTMiddleware struct {
	config   *config.Main
	authRepo authRepo.AuthRepository
}

// NewJWTMiddleware creates a new JWT middleware
func NewJWTMiddleware(config *config.Main, authRepo authRepo.AuthRepository) *JWTMiddleware {
	return &JWTMiddleware{config: config, authRepo: authRepo}
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

		// Validate LMS token and map/provision local user context
		user, err := m.validateAndResolveUser(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		// Add user info to context
		ctx = AddUserToContext(ctx, user)

		return handler(ctx, req)
	}
}

// shouldSkipAuth determines if authentication should be skipped for the method
func (m *JWTMiddleware) shouldSkipAuth(method string) bool {
	skipMethods := []string{
		"/base.Base/HealthCheck",
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

func (m *JWTMiddleware) validateAndResolveUser(ctx context.Context, tokenString string) (*base.User, error) {
	claims, err := m.validateLMSToken(tokenString)
	if err != nil {
		return nil, errors.New("invalid LMS access token")
	}

	lmsUserID := claims.LMSUserID
	if lmsUserID == 0 {
		lmsUserID = int64(claims.UserID)
	}
	if lmsUserID == 0 {
		return nil, errors.New("token missing lms_user_id/user_id")
	}

	name := strings.TrimSpace(claims.FullName)
	if name == "" {
		name = strings.TrimSpace(claims.Nama)
	}
	if name == "" {
		name = "LMS User"
	}

	role := base.UserRole_SISWA
	normalizedRole := strings.ToLower(strings.TrimSpace(claims.RoleName))
	if normalizedRole == "admin" || normalizedRole == "school_admin" || normalizedRole == "superadmin" {
		role = base.UserRole_ADMIN
	}

	user, syncErr := m.authRepo.FindOrCreateByLMSID(ctx, lmsUserID, claims.Email, name, int32(role))
	if syncErr != nil {
		return nil, errors.New("failed to provision local user from token")
	}
	if !user.IsActive {
		return nil, errors.New("user is inactive")
	}
	return user, nil
}

func (m *JWTMiddleware) validateLMSToken(tokenString string) (*JWTClaims, error) {
	secret := m.config.JWT.LMSTokenSecret
	if secret == "" {
		secret = m.config.JWT.Secret
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims := &JWTClaims{}
	if lmsUserIDRaw, has := mapClaims["lms_user_id"]; has {
		switch typed := lmsUserIDRaw.(type) {
		case float64:
			claims.LMSUserID = int64(typed)
		case int64:
			claims.LMSUserID = typed
		}
	}
	if claims.LMSUserID == 0 {
		if userIDRaw, has := mapClaims["user_id"]; has {
			switch typed := userIDRaw.(type) {
			case float64:
				claims.UserID = int32(typed)
			case int64:
				claims.UserID = int32(typed)
			}
		}
	}
	if email, has := mapClaims["email"].(string); has {
		claims.Email = email
	}
	if roleName, has := mapClaims["role"].(string); has {
		claims.RoleName = roleName
	}
	if fullName, has := mapClaims["full_name"].(string); has {
		claims.FullName = fullName
	}
	if nama, has := mapClaims["nama"].(string); has {
		claims.Nama = nama
	}
	if typ, has := mapClaims["type"].(string); has {
		claims.Type = typ
	}
	if sub, has := mapClaims["sub"].(string); has {
		claims.Subject = sub
	}
	if iss, has := mapClaims["iss"].(string); has {
		claims.Issuer = iss
	}
	if audClaim, has := mapClaims["aud"]; has {
		switch typed := audClaim.(type) {
		case string:
			claims.Audience = []string{typed}
		case []interface{}:
			aud := make([]string, 0, len(typed))
			for _, item := range typed {
				if v, ok := item.(string); ok {
					aud = append(aud, v)
				}
			}
			claims.Audience = aud
		}
	}

	isAccessType := claims.Type == "access" || claims.Subject == "access"
	if !isAccessType {
		return nil, errors.New("invalid token type")
	}

	if m.config.JWT.LMSIssuer != "" && claims.Issuer != m.config.JWT.LMSIssuer {
		return nil, errors.New("invalid token issuer")
	}

	if m.config.JWT.LMSAudience != "" {
		audOk := false
		for _, aud := range claims.Audience {
			if aud == m.config.JWT.LMSAudience {
				audOk = true
				break
			}
		}
		if !audOk {
			return nil, errors.New("invalid token audience")
		}
	}

	return claims, nil
}

// AddUserToContext adds user information to the context
func AddUserToContext(ctx context.Context, user *base.User) context.Context {
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
