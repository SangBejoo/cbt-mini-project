package auth

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// authRepositoryImpl implements AuthRepository
type authRepositoryImpl struct {
	db *sql.DB
}

func lmsUserIDValue(id *int64) int64 {
	if id == nil {
		return 0
	}
	return *id
}

func normalizeRoleForDB(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "superadmin", "admin":
		return "superadmin"
	case "teacher", "guru":
		return "teacher"
	case "student", "siswa":
		return "student"
	default:
		return "student"
	}
}

func protoRoleToDB(role base.UserRole) string {
	switch int32(role) {
	case int32(base.UserRole_ADMIN):
		return "superadmin"
	case 3:
		return "teacher"
	case 4:
		return "superadmin"
	default:
		return "student"
	}
}

func dbRoleToProto(role string) base.UserRole {
	normalized := normalizeRoleForDB(role)
	if normalized == "superadmin" {
		return base.UserRole_ADMIN
	}
	if normalized == "teacher" {
		return 3
	}
	return base.UserRole_SISWA
}

// NewAuthRepository creates a new auth repository
func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepositoryImpl{db: db}
}

// Login authenticates a user
func (r *authRepositoryImpl) Login(ctx context.Context, email, password string) (*base.User, error) {
	var userEntity entity.User

	query := `SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at, lms_user_id FROM users WHERE email = $1 AND is_active = $2 AND role = 'superadmin'`
	err := r.db.QueryRowContext(ctx, query, email, true).Scan(&userEntity.ID, &userEntity.Email, &userEntity.PasswordHash, &userEntity.Nama, &userEntity.Role, &userEntity.IsActive, &userEntity.CreatedAt, &userEntity.UpdatedAt, &userEntity.LmsUserID)
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
	role := dbRoleToProto(userEntity.Role)

	user := &base.User{
		Id:        int32(userEntity.ID),
		Email:     userEntity.Email,
		Nama:      userEntity.Nama,
		Role:      role,
		IsActive:  userEntity.IsActive,
		LmsUserId: lmsUserIDValue(userEntity.LmsUserID),
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *authRepositoryImpl) GetUserByID(ctx context.Context, id int32) (*base.User, error) {
	var userEntity entity.User
	query := `SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at, lms_user_id FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&userEntity.ID, &userEntity.Email, &userEntity.PasswordHash, &userEntity.Nama, &userEntity.Role, &userEntity.IsActive, &userEntity.CreatedAt, &userEntity.UpdatedAt, &userEntity.LmsUserID)
	if err != nil {
		return nil, err
	}

	role := dbRoleToProto(userEntity.Role)

	user := &base.User{
		Id:        int32(userEntity.ID),
		Email:     userEntity.Email,
		Nama:      userEntity.Nama,
		Role:      role,
		IsActive:  userEntity.IsActive,
		LmsUserId: lmsUserIDValue(userEntity.LmsUserID),
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *authRepositoryImpl) GetUserByEmail(ctx context.Context, email string) (*base.User, error) {
	var userEntity entity.User
	query := `SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at, lms_user_id FROM users WHERE email = $1`
	err := r.db.QueryRowContext(ctx, query, email).Scan(&userEntity.ID, &userEntity.Email, &userEntity.PasswordHash, &userEntity.Nama, &userEntity.Role, &userEntity.IsActive, &userEntity.CreatedAt, &userEntity.UpdatedAt, &userEntity.LmsUserID)
	if err != nil {
		return nil, err
	}

	role := dbRoleToProto(userEntity.Role)

	user := &base.User{
		Id:        int32(userEntity.ID),
		Email:     userEntity.Email,
		Nama:      userEntity.Nama,
		Role:      role,
		IsActive:  userEntity.IsActive,
		LmsUserId: lmsUserIDValue(userEntity.LmsUserID),
	}

	return user, nil
}

// CreateUser creates a new user
func (r *authRepositoryImpl) CreateUser(ctx context.Context, user *base.User) (*base.User, error) {
	_ = ctx
	_ = user
	return nil, errors.New("user creation is managed by LMS service")
}

// UpdateUser updates a user
func (r *authRepositoryImpl) UpdateUser(ctx context.Context, id int32, updates map[string]interface{}) (*base.User, error) {
	_ = ctx
	_ = id
	_ = updates
	return nil, errors.New("user updates are managed by LMS service")
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
	_ = ctx
	_ = id
	return errors.New("user deletion is managed by LMS service")
}

// ListUsers lists users with filters
func (r *authRepositoryImpl) ListUsers(ctx context.Context, role int32, statusFilter int32, limit, offset int) ([]*base.User, int, error) {
	var conditions []string
	var args []interface{}

	if role != 0 {
		roleStr := protoRoleToDB(base.UserRole(role))
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
	selectQuery := "SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at, lms_user_id FROM users " + whereClause + " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1) + " OFFSET $" + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var userEntities []entity.User
	for rows.Next() {
		var userEntity entity.User
		err := rows.Scan(&userEntity.ID, &userEntity.Email, &userEntity.PasswordHash, &userEntity.Nama, &userEntity.Role, &userEntity.IsActive, &userEntity.CreatedAt, &userEntity.UpdatedAt, &userEntity.LmsUserID)
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
		role := dbRoleToProto(userEntity.Role)

		users[i] = &base.User{
			Id:        int32(userEntity.ID),
			Email:     userEntity.Email,
			Nama:      userEntity.Nama,
			Role:      role,
			IsActive:  userEntity.IsActive,
			CreatedAt: timestamppb.New(userEntity.CreatedAt),
			UpdatedAt: timestamppb.New(userEntity.UpdatedAt),
			LmsUserId: lmsUserIDValue(userEntity.LmsUserID),
		}
	}

	return users, int(total), nil
}

// FindOrCreateByLMSID finds a user by LMS ID or creates one if not found
func (r *authRepositoryImpl) FindOrCreateByLMSID(ctx context.Context, lmsID int64, email, name string, role int32) (*base.User, error) {
	_ = role
	resolvedEmail := strings.ToLower(strings.TrimSpace(email))
	_ = strings.TrimSpace(name)

	toProto := func(u entity.User) *base.User {
		return &base.User{
			Id:        int32(u.ID),
			Email:     u.Email,
			Nama:      u.Nama,
			Role:      dbRoleToProto(u.Role),
			IsActive:  u.IsActive,
			CreatedAt: timestamppb.New(u.CreatedAt),
			UpdatedAt: timestamppb.New(u.UpdatedAt),
			LmsUserId: lmsUserIDValue(u.LmsUserID),
		}
	}

	fetchByEmailQuery := `
		SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at, lms_user_id
		FROM users
		WHERE lower(email) = lower($1)
		ORDER BY updated_at DESC, id DESC
		LIMIT 1`

	fetchByLMSIDQuery := `
		SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at, lms_user_id
		FROM users
		WHERE lms_user_id = $1
		ORDER BY updated_at DESC, id DESC
		LIMIT 1`

	var userEntity entity.User
	err := r.db.QueryRowContext(ctx, fetchByEmailQuery, resolvedEmail).Scan(
		&userEntity.ID,
		&userEntity.Email,
		&userEntity.PasswordHash,
		&userEntity.Nama,
		&userEntity.Role,
		&userEntity.IsActive,
		&userEntity.CreatedAt,
		&userEntity.UpdatedAt,
		&userEntity.LmsUserID,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err == nil {
		if !userEntity.IsActive {
			return nil, fmt.Errorf("lms user %d is inactive", lmsID)
		}
		return toProto(userEntity), nil
	}

	err = r.db.QueryRowContext(ctx, fetchByLMSIDQuery, lmsID).Scan(
		&userEntity.ID,
		&userEntity.Email,
		&userEntity.PasswordHash,
		&userEntity.Nama,
		&userEntity.Role,
		&userEntity.IsActive,
		&userEntity.CreatedAt,
		&userEntity.UpdatedAt,
		&userEntity.LmsUserID,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err == nil {
		if userEntity.ID != int(lmsID) {
			return nil, fmt.Errorf("token user mismatch: email %s maps to lms id %d", resolvedEmail, userEntity.ID)
		}
		if !userEntity.IsActive {
			return nil, fmt.Errorf("lms user %d is inactive", lmsID)
		}
		return toProto(userEntity), nil
	}

	return nil, fmt.Errorf("lms user %d not found in canonical users table", lmsID)
}

func isUsersEmailUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "users_email_key") || (strings.Contains(errMsg, "duplicate key") && strings.Contains(errMsg, "email"))
}

// GetLMSUserIDByLocalID retrieves the LMS User ID mapped to a local User ID
func (r *authRepositoryImpl) GetLMSUserIDByLocalID(ctx context.Context, id int32) (int64, error) {
	var lmsUserID int64
	query := `SELECT lms_user_id FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&lmsUserID)
	if err != nil {
		return 0, err
	}
	if lmsUserID == 0 {
		return 0, fmt.Errorf("user %d has no LMS ID link", id)
	}
	return lmsUserID, nil
}
