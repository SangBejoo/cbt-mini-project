package auth

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// authRepositoryImpl implements AuthRepository
type authRepositoryImpl struct {
	db *gorm.DB
}

// NewAuthRepository creates a new auth repository
func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepositoryImpl{db: db}
}

// Login authenticates a user
func (r *authRepositoryImpl) Login(ctx context.Context, email, password string) (*base.User, error) {
	var userEntity entity.User

	if err := r.db.Where("email = ? AND is_active = ?", email, true).First(&userEntity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
	if err := r.db.First(&userEntity, id).Error; err != nil {
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
	if err := r.db.Where("email = ?", email).First(&userEntity).Error; err != nil {
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

	if err := r.db.Create(&userEntity).Error; err != nil {
		return nil, err
	}

	// Convert back to proto
	user.Id = int32(userEntity.ID)
	user.PasswordHash = userEntity.PasswordHash

	return user, nil
}

// UpdateUser updates a user
func (r *authRepositoryImpl) UpdateUser(ctx context.Context, id int32, updates map[string]interface{}) (*base.User, error) {
	var userEntity entity.User
	if err := r.db.First(&userEntity, id).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(&userEntity).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Reload user
	if err := r.db.First(&userEntity, id).Error; err != nil {
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

// DeleteUser deletes a user (hard delete)
func (r *authRepositoryImpl) DeleteUser(ctx context.Context, id int32) error {
	return r.db.Unscoped().Delete(&entity.User{}, id).Error
}

// ListUsers lists users with filters
func (r *authRepositoryImpl) ListUsers(ctx context.Context, role int32, statusFilter int32, limit, offset int) ([]*base.User, int, error) {
	var userEntities []entity.User
	var total int64

	query := r.db.Model(&entity.User{})

	if role != 0 {
		var roleStr string
		switch base.UserRole(role) {
		case base.UserRole_SISWA:
			roleStr = "siswa"
		case base.UserRole_ADMIN:
			roleStr = "admin"
		}
		query = query.Where("role = ?", roleStr)
	}

	// Filter by active status if specified
	// 0=all, 1=active only, 2=inactive only
	switch statusFilter {
	case 1:
		query = query.Where("is_active = ?", true)
	case 2:
		query = query.Where("is_active = ?", false)
	// case 0: show all users (no filter)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&userEntities).Error; err != nil {
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