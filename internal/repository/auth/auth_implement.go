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
	roleStr := protoRoleToDB(base.UserRole(role))
	resolvedEmail := strings.ToLower(strings.TrimSpace(email))
	if resolvedEmail == "" {
		resolvedEmail = fmt.Sprintf("lms_user_%d@cbt.local", lmsID)
	}
	resolvedName := strings.TrimSpace(name)
	if resolvedName == "" {
		resolvedName = "LMS User"
	}

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

	now := time.Now()
	if err == nil {
		if userEntity.LmsUserID == nil || *userEntity.LmsUserID != lmsID || userEntity.Nama != resolvedName || normalizeRoleForDB(userEntity.Role) != roleStr || !userEntity.IsActive {
			_, updateErr := r.db.ExecContext(
				ctx,
				`UPDATE users SET lms_user_id = $1, email = $2, full_name = $3, role = $4, is_active = $5, updated_at = $6 WHERE id = $7`,
				lmsID,
				resolvedEmail,
				resolvedName,
				roleStr,
				true,
				now,
				userEntity.ID,
			)
			if updateErr != nil {
				return nil, updateErr
			}
			userEntity.Email = resolvedEmail
			userEntity.Nama = resolvedName
			userEntity.Role = roleStr
			userEntity.IsActive = true
			userEntity.UpdatedAt = now
			lmsIDVal := lmsID
			userEntity.LmsUserID = &lmsIDVal
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
		_, updateErr := r.db.ExecContext(
			ctx,
			`UPDATE users SET email = $1, full_name = $2, role = $3, is_active = $4, updated_at = $5 WHERE id = $6`,
			resolvedEmail,
			resolvedName,
			roleStr,
			true,
			now,
			userEntity.ID,
		)
		if updateErr != nil {
			if isUsersEmailUniqueViolation(updateErr) {
				var canonical entity.User
				canonicalErr := r.db.QueryRowContext(ctx, fetchByEmailQuery, resolvedEmail).Scan(
					&canonical.ID,
					&canonical.Email,
					&canonical.PasswordHash,
					&canonical.Nama,
					&canonical.Role,
					&canonical.IsActive,
					&canonical.CreatedAt,
					&canonical.UpdatedAt,
					&canonical.LmsUserID,
				)
				if canonicalErr == nil {
					_, relinkErr := r.db.ExecContext(
						ctx,
						`UPDATE users SET lms_user_id = $1, full_name = $2, role = $3, is_active = $4, updated_at = $5 WHERE id = $6`,
						lmsID,
						resolvedName,
						roleStr,
						true,
						now,
						canonical.ID,
					)
					if relinkErr == nil {
						lmsIDVal := lmsID
						canonical.LmsUserID = &lmsIDVal
						canonical.Nama = resolvedName
						canonical.Role = roleStr
						canonical.IsActive = true
						canonical.UpdatedAt = now
						return toProto(canonical), nil
					}
				}
			}
			return nil, updateErr
		}

		userEntity.Email = resolvedEmail
		userEntity.Nama = resolvedName
		userEntity.Role = roleStr
		userEntity.IsActive = true
		userEntity.UpdatedAt = now
		lmsIDVal := lmsID
		userEntity.LmsUserID = &lmsIDVal

		return toProto(userEntity), nil
	}

	randomPassword := fmt.Sprintf("lms_%d", lmsID)
	hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)
	if hashErr != nil {
		return nil, hashErr
	}

	createQuery := `
		INSERT INTO users (email, password_hash, full_name, role, is_active, lms_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (email) DO UPDATE SET lms_user_id = EXCLUDED.lms_user_id, full_name = EXCLUDED.full_name, role = EXCLUDED.role, is_active = EXCLUDED.is_active, updated_at = EXCLUDED.updated_at
		RETURNING id, created_at, updated_at`

	err = r.db.QueryRowContext(ctx, createQuery, resolvedEmail, string(hashedPassword), resolvedName, roleStr, true, lmsID, now, now).Scan(
		&userEntity.ID,
		&userEntity.CreatedAt,
		&userEntity.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	userEntity.Email = resolvedEmail
	userEntity.Nama = resolvedName
	userEntity.Role = roleStr
	userEntity.IsActive = true
	userEntity.LmsUserID = &lmsID

	return toProto(userEntity), nil
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

