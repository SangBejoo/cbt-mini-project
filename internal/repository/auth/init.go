package auth

import (
	base "cbt-test-mini-project/gen/proto"
	"context"
	"database/sql"
)

// AuthRepository defines the interface for auth repository
type AuthRepository interface {
	Login(ctx context.Context, email, password string) (*base.User, error)
	GetUserByID(ctx context.Context, id int32) (*base.User, error)
	GetUserByEmail(ctx context.Context, email string) (*base.User, error)
	CreateUser(ctx context.Context, user *base.User) (*base.User, error)
	UpdateUser(ctx context.Context, id int32, updates map[string]interface{}) (*base.User, error)
	DeleteUser(ctx context.Context, id int32) error
	ListUsers(ctx context.Context, role int32, statusFilter int32, limit, offset int) ([]*base.User, int, error)
	CheckUserHasTestSessions(ctx context.Context, id int32) (bool, error)

	// LMS sync methods
	FindOrCreateByLMSID(ctx context.Context, lmsID int64, email, name string, role int32) (*base.User, error)
}

// InitAuthRepository initializes the auth repository
func InitAuthRepository(db *sql.DB) AuthRepository {
	return NewAuthRepository(db)
}