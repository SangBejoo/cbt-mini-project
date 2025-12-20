package interceptor

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.elastic.co/apm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/repository"
)

// RateLimitMiddleware implements rate limiting per user
type RateLimitMiddleware struct {
	userLimitRepo repository.UserLimitRepository
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(userLimitRepo repository.UserLimitRepository) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		userLimitRepo: userLimitRepo,
	}
}

// UnaryServerInterceptor implements rate limiting for unary gRPC calls
func (m *RateLimitMiddleware) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Extract user from context (set by JWT middleware)
	user, ok := ctx.Value("user").(*base.User)
	if !ok {
		// If no user, skip rate limiting (for public endpoints)
		return handler(ctx, req)
	}

	userID := int(user.Id)

	// Check rate limit
	allowed, remaining, resetTime, err := m.checkRateLimit(ctx, userID, info.FullMethod)
	if err != nil {
		slog.Error("Rate limit check failed", "error", err, "user_id", userID)
		// Allow request on error to avoid blocking users
		return handler(ctx, req)
	}

	if !allowed {
		// Return rate limit exceeded error
		return nil, status.Error(codes.ResourceExhausted, fmt.Sprintf("Rate limit exceeded. Try again in %d seconds", int(time.Until(resetTime).Seconds())))
	}

	// Add rate limit headers to response metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		newMD := metadata.MD{}
		for k, v := range md {
			newMD[k] = v
		}
		newMD["x-ratelimit-remaining"] = []string{fmt.Sprintf("%d", remaining)}
		newMD["x-ratelimit-reset"] = []string{fmt.Sprintf("%d", resetTime.Unix())}
		ctx = metadata.NewIncomingContext(ctx, newMD)
	}

	// Record usage
	go m.recordUsage(ctx, userID, "api_call", info.FullMethod, nil)

	return handler(ctx, req)
}

// checkRateLimit checks if the user is within rate limits
func (m *RateLimitMiddleware) checkRateLimit(ctx context.Context, userID int, method string) (allowed bool, remaining int, resetTime time.Time, err error) {
	span, _ := apm.StartSpan(ctx, "rate_limit_check", "middleware")
	defer span.End()

	// Determine limit type based on method
	limitType := m.getLimitTypeForMethod(method)

	// Get or create user limit
	userLimit, err := m.userLimitRepo.GetOrCreateLimit(ctx, userID, limitType)
	if err != nil {
		return false, 0, time.Now(), err
	}

	// Check if limit needs to be reset
	now := time.Now()
	if now.After(userLimit.ResetAt) {
		// Reset the counter
		userLimit.CurrentUsed = 0
		userLimit.ResetAt = m.getNextResetTime(limitType)
		if err := m.userLimitRepo.UpdateLimit(ctx, userLimit); err != nil {
			return false, 0, time.Now(), err
		}
	}

	// Check if limit exceeded
	if userLimit.CurrentUsed >= userLimit.LimitValue {
		return false, 0, userLimit.ResetAt, nil
	}

	// Increment usage
	userLimit.CurrentUsed++
	if err := m.userLimitRepo.UpdateLimit(ctx, userLimit); err != nil {
		return false, 0, time.Now(), err
	}

	remaining = userLimit.LimitValue - userLimit.CurrentUsed
	return true, remaining, userLimit.ResetAt, nil
}

// getLimitTypeForMethod determines the limit type based on gRPC method
func (m *RateLimitMiddleware) getLimitTypeForMethod(method string) string {
	// Extract service and method name
	parts := strings.Split(method, "/")
	if len(parts) < 3 {
		return entity.LimitTypeAPIRequestsPerHour
	}

	service := parts[1]
	methodName := parts[2]

	// Define limits based on service and method
	switch service {
	case "base.TestSessionService":
		switch methodName {
		case "CreateTestSession":
			return entity.LimitTypeTestSessionsPerDay
		default:
			return entity.LimitTypeAPIRequestsPerHour
		}
	case "base.SoalService":
		switch methodName {
		case "CreateSoal":
			return entity.LimitTypeQuestionsPerDay
		default:
			return entity.LimitTypeAPIRequestsPerHour
		}
	default:
		return entity.LimitTypeAPIRequestsPerHour
	}
}

// getNextResetTime calculates the next reset time based on limit type
func (m *RateLimitMiddleware) getNextResetTime(limitType string) time.Time {
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

// recordUsage records the usage for analytics
func (m *RateLimitMiddleware) recordUsage(ctx context.Context, userID int, action, _ string, resourceID *int) {
	usage := &entity.UserLimitUsage{
		UserID:     userID,
		LimitType:  "api_requests_per_hour", // Could be more specific
		Action:     action,
		ResourceID: resourceID,
	}

	if err := m.userLimitRepo.RecordUsage(ctx, usage); err != nil {
		slog.Error("Failed to record usage", "error", err, "user_id", userID, "action", action)
	}
}