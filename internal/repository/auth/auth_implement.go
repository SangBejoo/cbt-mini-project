package auth

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// authRepositoryImpl implements AuthRepository
type authRepositoryImpl struct {
	db *sql.DB
}

// NewAuthRepository creates a new auth repository
func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepositoryImpl{db: db}
}

// Login authenticates a user
func (r *authRepositoryImpl) Login(ctx context.Context, email, password string) (*base.User, error) {
	var userEntity entity.User

	query := `SELECT id, email, password_hash, nama, role, is_active, created_at, updated_at FROM users WHERE email = $1 AND is_active = $2`
	err := r.db.QueryRowContext(ctx, query, email, true).Scan(&userEntity.ID, &userEntity.Email, &userEntity.PasswordHash, &userEntity.Nama, &userEntity.Role, &userEntity.IsActive, &userEntity.CreatedAt, &userEntity.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(userEntity.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Convert entity to proto
	var role base.UserRole
	switch userEntity.Role {
	case "siswa":
		role = base.UserRole_SISWA
	case "admin":
		role = base.UserRole_ADMIN
	default:
		role = base.UserRole_ROLE_INVALID
	}

	user := &base.User{
		Id:           int32(userEntity.ID),
		Email:        userEntity.Email,
		Nama:         userEntity.Nama,
		Role:         role,
		IsActive:     userEntity.IsActive,
		PasswordHash: userEntity.PasswordHash,
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *authRepositoryImpl) GetUserByID(ctx context.Context, id int32) (*base.User, error) {
	var userEntity entity.User
	query := `SELECT id, email, password_hash, nama, role, is_active, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&userEntity.ID, &userEntity.Email, &userEntity.PasswordHash, &userEntity.Nama, &userEntity.Role, &userEntity.IsActive, &userEntity.CreatedAt, &userEntity.UpdatedAt)
	if err != nil {
		return nil, err
	}

	var role base.UserRole
	switch userEntity.Role {
	case "siswa":
		role = base.UserRole_SISWA
	case "admin":
		role = base.UserRole_ADMIN
	default:
		role = base.UserRole_ROLE_INVALID
	}

	user := &base.User{
		Id:           int32(userEntity.ID),
		Email:        userEntity.Email,
		Nama:         userEntity.Nama,
		Role:         role,
		IsActive:     userEntity.IsActive,
		PasswordHash: userEntity.PasswordHash,
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *authRepositoryImpl) GetUserByEmail(ctx context.Context, email string) (*base.User, error) {
	var userEntity entity.User
	query := `SELECT id, email, password_hash, nama, role, is_active, created_at, updated_at FROM users WHERE email = $1`
	err := r.db.QueryRowContext(ctx, query, email).Scan(&userEntity.ID, &userEntity.Email, &userEntity.PasswordHash, &userEntity.Nama, &userEntity.Role, &userEntity.IsActive, &userEntity.CreatedAt, &userEntity.UpdatedAt)
	if err != nil {
		return nil, err
	}

	var role base.UserRole
	switch userEntity.Role {
	case "siswa":
		role = base.UserRole_SISWA
	case "admin":
		role = base.UserRole_ADMIN
	default:
		role = base.UserRole_ROLE_INVALID
	}

	user := &base.User{
		Id:           int32(userEntity.ID),
		Email:        userEntity.Email,
		Nama:         userEntity.Nama,
		Role:         role,
		IsActive:     userEntity.IsActive,
		PasswordHash: userEntity.PasswordHash,
	}

	return user, nil
}

// CreateUser creates a new user
func (r *authRepositoryImpl) CreateUser(ctx context.Context, user *base.User) (*base.User, error) {
	// Convert proto to entity
	var roleStr string
	switch user.Role {
	case base.UserRole_SISWA:
		roleStr = "siswa"
	case base.UserRole_ADMIN:
		roleStr = "admin"
	default:
		roleStr = "siswa"
	}

	userEntity := entity.User{
		Email:    user.Email,
		Nama:     user.Nama,
		Role:     roleStr,
		IsActive: user.IsActive,
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Email), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	userEntity.PasswordHash = string(hashedPassword)
	userEntity.CreatedAt = time.Now()
	userEntity.UpdatedAt = time.Now()

	query := `INSERT INTO users (email, password_hash, nama, role, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err = r.db.QueryRowContext(ctx, query, userEntity.Email, userEntity.PasswordHash, userEntity.Nama, userEntity.Role, userEntity.IsActive, userEntity.CreatedAt, userEntity.UpdatedAt).Scan(&userEntity.ID)
	if err != nil {
		return nil, err
	}

	// Convert back to proto
	user.Id = int32(userEntity.ID)
	user.PasswordHash = userEntity.PasswordHash

	return user, nil
}

// UpdateUser updates a user
func (r *authRepositoryImpl) UpdateUser(ctx context.Context, id int32, updates map[string]interface{}) (*base.User, error) {
	// First get current user
	var userEntity entity.User
	query := `SELECT id, email, password_hash, nama, role, is_active, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&userEntity.ID, &userEntity.Email, &userEntity.PasswordHash, &userEntity.Nama, &userEntity.Role, &userEntity.IsActive, &userEntity.CreatedAt, &userEntity.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if email, ok := updates["email"].(string); ok {
		userEntity.Email = email
	}
	if nama, ok := updates["nama"].(string); ok {
		userEntity.Nama = nama
	}
	if role, ok := updates["role"].(string); ok {
		userEntity.Role = role
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		userEntity.IsActive = isActive
	}
	userEntity.UpdatedAt = time.Now()

	// Update
	updateQuery := `UPDATE users SET email = $1, nama = $2, role = $3, is_active = $4, updated_at = $5 WHERE id = $6`
	_, err = r.db.ExecContext(ctx, updateQuery, userEntity.Email, userEntity.Nama, userEntity.Role, userEntity.IsActive, userEntity.UpdatedAt, userEntity.ID)
	if err != nil {
		return nil, err
	}

	var role base.UserRole
	switch userEntity.Role {
	case "siswa":
		role = base.UserRole_SISWA
	case "admin":
		role = base.UserRole_ADMIN
	default:
		role = base.UserRole_ROLE_INVALID
	}

	user := &base.User{
		Id:           int32(userEntity.ID),
		Email:        userEntity.Email,
		Nama:         userEntity.Nama,
		Role:         role,
		IsActive:     userEntity.IsActive,
		PasswordHash: userEntity.PasswordHash,
	}

	return user, nil
}

// CheckUserHasTestSessions checks if user has any test sessions
func (r *authRepositoryImpl) CheckUserHasTestSessions(ctx context.Context, id int32) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM test_session WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&count)
	return count > 0, err
}

// DeleteUser deletes a user (hard delete)
func (r *authRepositoryImpl) DeleteUser(ctx context.Context, id int32) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListUsers lists users with filters
func (r *authRepositoryImpl) ListUsers(ctx context.Context, role int32, statusFilter int32, limit, offset int) ([]*base.User, int, error) {
	var conditions []string
	var args []interface{}

	if role != 0 {
		var roleStr string
		switch base.UserRole(role) {
		case base.UserRole_SISWA:
			roleStr = "siswa"
		case base.UserRole_ADMIN:
			roleStr = "admin"
		}
		conditions = append(conditions, "role = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, roleStr)
	}

	// Filter by active status if specified
	// 0=all, 1=active only, 2=inactive only
	switch statusFilter {
	case 1:
		conditions = append(conditions, "is_active = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, true)
	case 2:
		conditions = append(conditions, "is_active = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, false)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM users " + whereClause
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	selectQuery := "SELECT id, email, password_hash, nama, role, is_active, created_at, updated_at FROM users " + whereClause + " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1) + " OFFSET $" + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var userEntities []entity.User
	for rows.Next() {
		var userEntity entity.User
		err := rows.Scan(&userEntity.ID, &userEntity.Email, &userEntity.PasswordHash, &userEntity.Nama, &userEntity.Role, &userEntity.IsActive, &userEntity.CreatedAt, &userEntity.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		userEntities = append(userEntities, userEntity)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	// Convert to proto
	users := make([]*base.User, len(userEntities))
	for i, userEntity := range userEntities {
		var role base.UserRole
		switch userEntity.Role {
		case "siswa":
			role = base.UserRole_SISWA
		case "admin":
			role = base.UserRole_ADMIN
		default:
			role = base.UserRole_ROLE_INVALID
		}

		users[i] = &base.User{
			Id:           int32(userEntity.ID),
			Email:        userEntity.Email,
			Nama:         userEntity.Nama,
			Role:         role,
			IsActive:     userEntity.IsActive,
			CreatedAt:    timestamppb.New(userEntity.CreatedAt),
			UpdatedAt:    timestamppb.New(userEntity.UpdatedAt),
			PasswordHash: userEntity.PasswordHash,
		}
	}

	return users, int(total), nil
}