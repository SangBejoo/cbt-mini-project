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

func lmsUserIDValue(id *int64) int64 {
	if id == nil {
		return 0
	}
	return *id
}

func normalizeRoleForDB(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin":
		return "admin"
	case "student", "siswa", "teacher", "superadmin":
		return "student"
	default:
		return "student"
	}
}

func protoRoleToDB(role base.UserRole) string {
	if role == base.UserRole_ADMIN {
		return "admin"
	}
	return "student"
}

func dbRoleToProto(role string) base.UserRole {
	if strings.EqualFold(role, "admin") {
		return base.UserRole_ADMIN
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

	query := `SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at, lms_user_id FROM users WHERE email = $1 AND is_active = $2 AND role = 'admin'`
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
		Id:           int32(userEntity.ID),
		Email:        userEntity.Email,
		Nama:         userEntity.Nama,
		Role:         role,
		IsActive:     userEntity.IsActive,
		LmsUserId:    lmsUserIDValue(userEntity.LmsUserID),
		
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
		Id:           int32(userEntity.ID),
		Email:        userEntity.Email,
		Nama:         userEntity.Nama,
		Role:         role,
		IsActive:     userEntity.IsActive,
		LmsUserId:    lmsUserIDValue(userEntity.LmsUserID),
		
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
		Id:           int32(userEntity.ID),
		Email:        userEntity.Email,
		Nama:         userEntity.Nama,
		Role:         role,
		IsActive:     userEntity.IsActive,
		LmsUserId:    lmsUserIDValue(userEntity.LmsUserID),
		
	}

	return user, nil
}

// CreateUser creates a new user
func (r *authRepositoryImpl) CreateUser(ctx context.Context, user *base.User) (*base.User, error) {
	// Convert proto to entity
	roleStr := protoRoleToDB(user.Role)

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

	query := `INSERT INTO users (email, password_hash, full_name, role, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err = r.db.QueryRowContext(ctx, query, userEntity.Email, userEntity.PasswordHash, userEntity.Nama, userEntity.Role, userEntity.IsActive, userEntity.CreatedAt, userEntity.UpdatedAt).Scan(&userEntity.ID)
	if err != nil {
		return nil, err
	}

	// Convert back to proto
	user.Id = int32(userEntity.ID)
	

	return user, nil
}

// UpdateUser updates a user
func (r *authRepositoryImpl) UpdateUser(ctx context.Context, id int32, updates map[string]interface{}) (*base.User, error) {
	// First get current user
	var userEntity entity.User
	query := `SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at, lms_user_id FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&userEntity.ID, &userEntity.Email, &userEntity.PasswordHash, &userEntity.Nama, &userEntity.Role, &userEntity.IsActive, &userEntity.CreatedAt, &userEntity.UpdatedAt, &userEntity.LmsUserID)
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
		userEntity.Role = normalizeRoleForDB(role)
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		userEntity.IsActive = isActive
	}
	userEntity.UpdatedAt = time.Now()

	// Update
	updateQuery := `UPDATE users SET email = $1, full_name = $2, role = $3, is_active = $4, updated_at = $5 WHERE id = $6`
	_, err = r.db.ExecContext(ctx, updateQuery, userEntity.Email, userEntity.Nama, userEntity.Role, userEntity.IsActive, userEntity.UpdatedAt, userEntity.ID)
	if err != nil {
		return nil, err
	}

	role := dbRoleToProto(userEntity.Role)

	user := &base.User{
		Id:           int32(userEntity.ID),
		Email:        userEntity.Email,
		Nama:         userEntity.Nama,
		Role:         role,
		IsActive:     userEntity.IsActive,
		LmsUserId:    lmsUserIDValue(userEntity.LmsUserID),
		
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
			Id:           int32(userEntity.ID),
			Email:        userEntity.Email,
			Nama:         userEntity.Nama,
			Role:         role,
			IsActive:     userEntity.IsActive,
			CreatedAt:    timestamppb.New(userEntity.CreatedAt),
			UpdatedAt:    timestamppb.New(userEntity.UpdatedAt),
			LmsUserId:    lmsUserIDValue(userEntity.LmsUserID),
			
		}
	}

	return users, int(total), nil
}

// FindOrCreateByLMSID finds a user by LMS ID or creates one if not found
func (r *authRepositoryImpl) FindOrCreateByLMSID(ctx context.Context, lmsID int64, email, name string, role int32) (*base.User, error) {
	// First try to find existing user by LMS ID
	var userEntity entity.User
	findQuery := `SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at, lms_user_id FROM users WHERE lms_user_id = $1`
	err := r.db.QueryRowContext(ctx, findQuery, lmsID).Scan(&userEntity.ID, &userEntity.Email, &userEntity.PasswordHash, &userEntity.Nama, &userEntity.Role, &userEntity.IsActive, &userEntity.CreatedAt, &userEntity.UpdatedAt, &userEntity.LmsUserID)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err == sql.ErrNoRows {
		// Fallback: try to find existing user by email (may have been created without lms_user_id)
		fallbackQuery := `SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at, COALESCE(lms_user_id, 0) FROM users WHERE email = $1`
		var existingLmsID int64
		scanErr := r.db.QueryRowContext(ctx, fallbackQuery, email).Scan(
			&userEntity.ID, &userEntity.Email, &userEntity.PasswordHash,
			&userEntity.Nama, &userEntity.Role, &userEntity.IsActive,
			&userEntity.CreatedAt, &userEntity.UpdatedAt, &existingLmsID,
		)
		if scanErr == nil {
			// Found by email — adopt this user by setting lms_user_id
			if existingLmsID == 0 {
				_, adoptErr := r.db.ExecContext(ctx,
					`UPDATE users SET lms_user_id = $1, updated_at = $2 WHERE id = $3`,
					lmsID, time.Now(), userEntity.ID,
				)
				if adoptErr != nil {
					return nil, fmt.Errorf("failed to link existing user to LMS ID: %w", adoptErr)
				}
			}
			lmsIDVal := lmsID
			userEntity.LmsUserID = &lmsIDVal
			err = nil // clear the ErrNoRows so we fall through to the JIT-update block below
		}
	}

	if err == nil {
		// User found, JIT-update stale fields from claims.
		roleStr := protoRoleToDB(base.UserRole(role))

		trimmedEmail := strings.TrimSpace(email)
		trimmedName := strings.TrimSpace(name)
		if trimmedEmail == "" {
			trimmedEmail = userEntity.Email
		}
		if trimmedName == "" {
			trimmedName = userEntity.Nama
		}

		if userEntity.Email != trimmedEmail || userEntity.Nama != trimmedName || normalizeRoleForDB(userEntity.Role) != roleStr {
			_, updateErr := r.db.ExecContext(
				ctx,
				`UPDATE users SET email = $1, full_name = $2, role = $3, updated_at = $4 WHERE id = $5`,
				trimmedEmail,
				trimmedName,
				roleStr,
				time.Now(),
				userEntity.ID,
			)
			if updateErr != nil {
				return nil, updateErr
			}
			userEntity.Email = trimmedEmail
			userEntity.Nama = trimmedName
			userEntity.Role = roleStr
		}

		protoRole := dbRoleToProto(userEntity.Role)

		return &base.User{
			Id:           int32(userEntity.ID),
			Email:        userEntity.Email,
			Nama:         userEntity.Nama,
			Role:         protoRole,
			IsActive:     userEntity.IsActive,
			CreatedAt:    timestamppb.New(userEntity.CreatedAt),
			UpdatedAt:    timestamppb.New(userEntity.UpdatedAt),
			LmsUserId:    lmsUserIDValue(userEntity.LmsUserID),
			
		}, nil
	}

	// User not found by lms_user_id or email, create new one
	roleStr := protoRoleToDB(base.UserRole(role))

	// Generate random hash for LMS users (non-admin local password login is disabled)
	randomPassword := fmt.Sprintf("lms_%d", lmsID)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	createQuery := `
		INSERT INTO users (email, password_hash, full_name, role, is_active, lms_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (email) DO UPDATE SET lms_user_id = EXCLUDED.lms_user_id, full_name = EXCLUDED.full_name, updated_at = EXCLUDED.updated_at
		RETURNING id`
	now := time.Now()
	err = r.db.QueryRowContext(ctx, createQuery, email, string(hashedPassword), name, roleStr, true, lmsID, now, now).Scan(&userEntity.ID)
	if err != nil {
		return nil, err
	}

	// Set the entity fields for return
	userEntity.Email = email
	userEntity.Nama = name
	userEntity.Role = roleStr
	userEntity.IsActive = true
	userEntity.CreatedAt = now
	userEntity.UpdatedAt = now
	userEntity.LmsUserID = &lmsID

	protoRole := dbRoleToProto(roleStr)

	return &base.User{
		Id:           int32(userEntity.ID),
		Email:        email,
		Nama:         name,
		Role:         protoRole,
		IsActive:     true,
		CreatedAt:    timestamppb.New(now),
		UpdatedAt:    timestamppb.New(now),
		LmsUserId:    lmsID,
		
	}, nil
}

// GetLMSUserIDByLocalID retrieves the LMS User ID mapped to a local User ID
func (r *authRepositoryImpl) GetLMSUserIDByLocalID(ctx context.Context, id int32) (int64, error) {
	var lmsUserID sql.NullInt64
	query := `SELECT lms_user_id FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&lmsUserID)
	if err != nil {
		return 0, err
	}
	if !lmsUserID.Valid {
		return 0, fmt.Errorf("user %d has no LMS ID link", id)
	}
	return lmsUserID.Int64, nil
}

