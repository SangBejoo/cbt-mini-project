package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/internal/entity"
)

// userLimitRepository implements UserLimitRepository
type userLimitRepository struct {
	db     *gorm.DB
	config *config.Main
}

// UserLimitRepository defines the interface for user limit operations
type UserLimitRepository interface {
	GetOrCreateLimit(ctx context.Context, userID int, limitType string) (*entity.UserLimit, error)
	UpdateLimit(ctx context.Context, limit *entity.UserLimit) error
	IncrementUsageAtomic(ctx context.Context, userID int, limitType string, resourceID *int) error
	GetLimitsByUser(ctx context.Context, userID int) ([]*entity.UserLimit, error)
	RecordUsage(ctx context.Context, usage *entity.UserLimitUsage) error
	GetUsageHistory(ctx context.Context, userID int, limitType string, since time.Time) ([]*entity.UserLimitUsage, error)
}

// NewUserLimitRepository creates a new user limit repository
func NewUserLimitRepository(db *gorm.DB, config *config.Main) UserLimitRepository {
	return &userLimitRepository{
		db:     db,
		config: config,
	}
}

// GetOrCreateLimit gets an existing limit or creates a new one with default values
func (r *userLimitRepository) GetOrCreateLimit(ctx context.Context, userID int, limitType string) (*entity.UserLimit, error) {
	var limit entity.UserLimit

	// Try to find existing limit
	err := r.db.WithContext(ctx).Where("user_id = ? AND limit_type = ?", userID, limitType).First(&limit).Error
	if err == nil {
		return &limit, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create new limit with default values
	defaultLimit := r.getDefaultLimit(limitType)
	limit = entity.UserLimit{
		UserID:      userID,
		LimitType:   limitType,
		LimitValue:  defaultLimit,
		CurrentUsed: 0,
		ResetAt:     r.getNextResetTime(limitType),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := r.db.WithContext(ctx).Create(&limit).Error; err != nil {
		return nil, err
	}

	return &limit, nil
}

// UpdateLimit updates an existing user limit
func (r *userLimitRepository) UpdateLimit(ctx context.Context, limit *entity.UserLimit) error {
	limit.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(limit).Error
}

// IncrementUsageAtomic atomically increments usage if under limit
func (r *userLimitRepository) IncrementUsageAtomic(ctx context.Context, userID int, limitType string, resourceID *int) error {
	// First, ensure the limit exists
	_, err := r.GetOrCreateLimit(ctx, userID, limitType)
	if err != nil {
		return err
	}

	fmt.Printf("=== REPO: Executing atomic increment for user %d, type %s ===\n", userID, limitType)
	// Atomic increment: update only if current_used < limit_value
	result := r.db.WithContext(ctx).Model(&entity.UserLimit{}).
		Where("user_id = ? AND limit_type = ? AND current_used < limit_value", userID, limitType).
		Updates(map[string]interface{}{
			"current_used": gorm.Expr("current_used + 1"),
			"updated_at":   time.Now(),
		})

	if result.Error != nil {
		fmt.Printf("=== REPO: Atomic increment failed: %v ===\n", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		fmt.Printf("=== REPO: Atomic increment: no rows affected (limit exceeded) ===\n")
		return gorm.ErrRecordNotFound // Indicates limit exceeded
	}

	fmt.Printf("=== REPO: Atomic increment success, rows affected: %d ===\n", result.RowsAffected)

	// Record usage for analytics
	usage := &entity.UserLimitUsage{
		UserID:     userID,
		LimitType:  limitType,
		Action:     "increment",
		ResourceID: resourceID,
		CreatedAt:  time.Now(),
	}

	if err := r.RecordUsage(ctx, usage); err != nil {
		// Log error but don't fail the operation
		// Note: in a real app, you might want to log this
	}

	return nil
}

// GetLimitsByUser gets all limits for a user
func (r *userLimitRepository) GetLimitsByUser(ctx context.Context, userID int) ([]*entity.UserLimit, error) {
	var limits []*entity.UserLimit
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&limits).Error
	return limits, err
}

// RecordUsage records a usage event
func (r *userLimitRepository) RecordUsage(ctx context.Context, usage *entity.UserLimitUsage) error {
	usage.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(usage).Error
}

// GetUsageHistory gets usage history for a user and limit type
func (r *userLimitRepository) GetUsageHistory(ctx context.Context, userID int, limitType string, since time.Time) ([]*entity.UserLimitUsage, error) {
	var usages []*entity.UserLimitUsage
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND limit_type = ? AND created_at >= ?", userID, limitType, since).
		Order("created_at DESC").
		Find(&usages).Error
	return usages, err
}

// getDefaultLimit returns the default limit value for a limit type
func (r *userLimitRepository) getDefaultLimit(limitType string) int {
	switch limitType {
	case entity.LimitTypeAPIRequestsPerHour:
		return r.config.RateLimit.APIRequestsPerHour
	case entity.LimitTypeAPIRequestsPerDay:
		return r.config.RateLimit.APIRequestsPerDay
	case entity.LimitTypeTestSessionsPerDay:
		return r.config.RateLimit.TestSessionsPerDay
	case entity.LimitTypeTestSessionsPerWeek:
		return r.config.RateLimit.TestSessionsPerWeek
	case entity.LimitTypeQuestionsPerDay:
		return r.config.RateLimit.QuestionsPerDay
	default:
		return 100 // Default fallback
	}
}

// getNextResetTime calculates the next reset time for a limit type
func (r *userLimitRepository) getNextResetTime(limitType string) time.Time {
	now := time.Now()
	switch limitType {
	case entity.LimitTypeAPIRequestsPerHour:
		return now.Add(time.Hour)
	case entity.LimitTypeAPIRequestsPerDay, entity.LimitTypeTestSessionsPerDay, entity.LimitTypeQuestionsPerDay:
		return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	case entity.LimitTypeTestSessionsPerWeek:
		// Reset every Monday
		daysUntilMonday := (8 - int(now.Weekday())) % 7
		if daysUntilMonday == 0 {
			daysUntilMonday = 7
		}
		return time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 0, 0, 0, 0, now.Location())
	default:
		return now.Add(time.Hour)
	}
}