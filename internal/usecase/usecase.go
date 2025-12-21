package usecase

import (
	"context"
	"fmt"
	"time"

	"go.elastic.co/apm"

	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/repository"
)

// UserLimitUsecase defines the interface for user limit business logic
type UserLimitUsecase interface {
	CheckLimit(ctx context.Context, userID int, limitType string) (*entity.UserLimit, error)
	IncrementUsage(ctx context.Context, userID int, limitType string, resourceID *int) error
	GetUserLimits(ctx context.Context, userID int) ([]*entity.UserLimit, error)
	GetUsageHistory(ctx context.Context, userID int, limitType string, days int) ([]*entity.UserLimitUsage, error)
	ResetLimit(ctx context.Context, userID int, limitType string) error
}

// userLimitUsecase implements UserLimitUsecase
type userLimitUsecase struct {
	userLimitRepo repository.UserLimitRepository
}

// NewUserLimitUsecase creates a new user limit usecase
func NewUserLimitUsecase(userLimitRepo repository.UserLimitRepository) UserLimitUsecase {
	return &userLimitUsecase{
		userLimitRepo: userLimitRepo,
	}
}

// CheckLimit checks if a user is within their limit for a specific type
func (u *userLimitUsecase) CheckLimit(ctx context.Context, userID int, limitType string) (*entity.UserLimit, error) {
	span, ctx := apm.StartSpan(ctx, "check_limit", "usecase")
	defer span.End()

	limit, err := u.userLimitRepo.GetOrCreateLimit(ctx, userID, limitType)
	if err != nil {
		return nil, fmt.Errorf("failed to get limit: %w", err)
	}

	// Check if limit needs to be reset
	now := time.Now()
	if now.After(limit.ResetAt) {
		limit.CurrentUsed = 0
		limit.ResetAt = u.getNextResetTime(limitType)
		if err := u.userLimitRepo.UpdateLimit(ctx, limit); err != nil {
			return nil, fmt.Errorf("failed to reset limit: %w", err)
		}
	}

	return limit, nil
}

// IncrementUsage increments the usage counter for a limit type
func (u *userLimitUsecase) IncrementUsage(ctx context.Context, userID int, limitType string, resourceID *int) error {
	span, ctx := apm.StartSpan(ctx, "increment_usage", "usecase")
	defer span.End()

	// Check current limit
	limit, err := u.CheckLimit(ctx, userID, limitType)
	if err != nil {
		return err
	}

	// Check if limit would be exceeded
	if limit.CurrentUsed >= limit.LimitValue {
		return fmt.Errorf("limit exceeded for %s: %d/%d", limitType, limit.CurrentUsed, limit.LimitValue)
	}

	// Increment usage
	limit.CurrentUsed++
	if err := u.userLimitRepo.UpdateLimit(ctx, limit); err != nil {
		return fmt.Errorf("failed to update limit: %w", err)
	}

	// Record usage for analytics
	usage := &entity.UserLimitUsage{
		UserID:     userID,
		LimitType:  limitType,
		Action:     "increment",
		ResourceID: resourceID,
		CreatedAt:  time.Now(),
	}

	if err := u.userLimitRepo.RecordUsage(ctx, usage); err != nil {
		// Log error but don't fail the operation
		apm.CaptureError(ctx, err).Send()
	}

	return nil
}

// GetUserLimits gets all limits for a user
func (u *userLimitUsecase) GetUserLimits(ctx context.Context, userID int) ([]*entity.UserLimit, error) {
	span, ctx := apm.StartSpan(ctx, "get_user_limits", "usecase")
	defer span.End()

	limits, err := u.userLimitRepo.GetLimitsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user limits: %w", err)
	}

	// Update any expired limits
	now := time.Now()
	for _, limit := range limits {
		if now.After(limit.ResetAt) {
			limit.CurrentUsed = 0
			limit.ResetAt = u.getNextResetTime(limit.LimitType)
			if err := u.userLimitRepo.UpdateLimit(ctx, limit); err != nil {
				apm.CaptureError(ctx, err).Send()
			}
		}
	}

	return limits, nil
}

// GetUsageHistory gets usage history for a user and limit type
func (u *userLimitUsecase) GetUsageHistory(ctx context.Context, userID int, limitType string, days int) ([]*entity.UserLimitUsage, error) {
	span, ctx := apm.StartSpan(ctx, "get_usage_history", "usecase")
	defer span.End()

	since := time.Now().AddDate(0, 0, -days)
	usages, err := u.userLimitRepo.GetUsageHistory(ctx, userID, limitType, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage history: %w", err)
	}

	return usages, nil
}

// ResetLimit manually resets a user's limit (admin function)
func (u *userLimitUsecase) ResetLimit(ctx context.Context, userID int, limitType string) error {
	span, ctx := apm.StartSpan(ctx, "reset_limit", "usecase")
	defer span.End()

	limit, err := u.userLimitRepo.GetOrCreateLimit(ctx, userID, limitType)
	if err != nil {
		return fmt.Errorf("failed to get limit: %w", err)
	}

	limit.CurrentUsed = 0
	limit.ResetAt = u.getNextResetTime(limitType)

	if err := u.userLimitRepo.UpdateLimit(ctx, limit); err != nil {
		return fmt.Errorf("failed to reset limit: %w", err)
	}

	return nil
}

// getNextResetTime calculates the next reset time for a limit type
func (u *userLimitUsecase) getNextResetTime(limitType string) time.Time {
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