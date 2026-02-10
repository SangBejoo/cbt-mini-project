package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/internal/entity"
)

// userLimitRepository implements UserLimitRepository
type userLimitRepository struct {
	db     *sql.DB
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
func NewUserLimitRepository(db *sql.DB, config *config.Main) UserLimitRepository {
	return &userLimitRepository{
		db:     db,
		config: config,
	}
}

// GetOrCreateLimit gets an existing limit or creates a new one with default values
func (r *userLimitRepository) GetOrCreateLimit(ctx context.Context, userID int, limitType string) (*entity.UserLimit, error) {
	var limit entity.UserLimit

	query := `SELECT id, user_id, limit_type, limit_value, current_used, reset_at, created_at, updated_at FROM user_limits WHERE user_id = $1 AND limit_type = $2`

	err := r.db.QueryRowContext(ctx, query, userID, limitType).Scan(&limit.ID, &limit.UserID, &limit.LimitType, &limit.LimitValue, &limit.CurrentUsed, &limit.ResetAt, &limit.CreatedAt, &limit.UpdatedAt)

	if err == nil {
		return &limit, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new limit
	defaultLimit := r.getDefaultLimit(limitType)
	resetAt := r.getNextResetTime(limitType)
	now := time.Now()

	insertQuery := `INSERT INTO user_limits (user_id, limit_type, limit_value, current_used, reset_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, user_id, limit_type, limit_value, current_used, reset_at, created_at, updated_at`

	err = r.db.QueryRowContext(ctx, insertQuery, userID, limitType, defaultLimit, 0, resetAt, now, now).Scan(&limit.ID, &limit.UserID, &limit.LimitType, &limit.LimitValue, &limit.CurrentUsed, &limit.ResetAt, &limit.CreatedAt, &limit.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &limit, nil
}

// UpdateLimit updates an existing user limit
func (r *userLimitRepository) UpdateLimit(ctx context.Context, limit *entity.UserLimit) error {
	limit.UpdatedAt = time.Now()
	query := `UPDATE user_limits SET limit_value = $1, current_used = $2, reset_at = $3, updated_at = $4 WHERE id = $5`
	_, err := r.db.ExecContext(ctx, query, limit.LimitValue, limit.CurrentUsed, limit.ResetAt, limit.UpdatedAt, limit.ID)
	return err
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
	query := `
		UPDATE user_limits
		SET current_used = current_used + 1, updated_at = $1
		WHERE user_id = $2 AND limit_type = $3 AND current_used < limit_value
	`
	result, err := r.db.ExecContext(ctx, query, time.Now(), userID, limitType)
	if err != nil {
		fmt.Printf("=== REPO: Atomic increment failed: %v ===\n", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		fmt.Printf("=== REPO: Atomic increment: no rows affected (limit exceeded) ===\n")
		return sql.ErrNoRows // Indicates limit exceeded
	}

	fmt.Printf("=== REPO: Atomic increment success, rows affected: %d ===\n", rowsAffected)

	// Record usage for analytics - DISABLED FOR PERFORMANCE
	// usage := &entity.UserLimitUsage{
	// 	UserID:     userID,
	// 	LimitType:  limitType,
	// 	Action:     "increment",
	// 	ResourceID: resourceID,
	// 	CreatedAt:  time.Now(),
	// }

	// Optimization: Disable detailed usage logging to prevent DB write bottlenecks on every request
	// if err := r.RecordUsage(ctx, usage); err != nil {
	// 	// Log error but don't fail the operation
	// 	// Note: in a real app, you might want to log this
	// }

	return nil
}

// GetLimitsByUser gets all limits for a user
func (r *userLimitRepository) GetLimitsByUser(ctx context.Context, userID int) ([]*entity.UserLimit, error) {
	query := `SELECT id, user_id, limit_type, limit_value, current_used, reset_at, created_at, updated_at FROM user_limits WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var limits []*entity.UserLimit
	for rows.Next() {
		var limit entity.UserLimit
		err := rows.Scan(&limit.ID, &limit.UserID, &limit.LimitType, &limit.LimitValue, &limit.CurrentUsed, &limit.ResetAt, &limit.CreatedAt, &limit.UpdatedAt)
		if err != nil {
			return nil, err
		}
		limits = append(limits, &limit)
	}
	return limits, rows.Err()
}

// RecordUsage records a usage event
func (r *userLimitRepository) RecordUsage(ctx context.Context, usage *entity.UserLimitUsage) error {
	usage.CreatedAt = time.Now()
	query := `INSERT INTO user_limit_usage (user_id, limit_type, action, resource_id, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, usage.UserID, usage.LimitType, usage.Action, usage.ResourceID, usage.CreatedAt)
	return err
}

// GetUsageHistory gets usage history for a user and limit type
func (r *userLimitRepository) GetUsageHistory(ctx context.Context, userID int, limitType string, since time.Time) ([]*entity.UserLimitUsage, error) {
	query := `SELECT id, user_id, limit_type, action, resource_id, created_at FROM user_limit_usage WHERE user_id = $1 AND limit_type = $2 AND created_at >= $3 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userID, limitType, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usages []*entity.UserLimitUsage
	for rows.Next() {
		var usage entity.UserLimitUsage
		err := rows.Scan(&usage.ID, &usage.UserID, &usage.LimitType, &usage.Action, &usage.ResourceID, &usage.CreatedAt)
		if err != nil {
			return nil, err
		}
		usages = append(usages, &usage)
	}
	return usages, rows.Err()
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