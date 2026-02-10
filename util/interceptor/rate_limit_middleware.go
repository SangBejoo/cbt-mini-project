package interceptor

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
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

// cachedLimit stores rate limit data with expiration time
type cachedLimit struct {
	limit    *entity.UserLimit
	cachedAt time.Time
}

// RateLimitMiddleware implements rate limiting per user with in-memory caching
type RateLimitMiddleware struct {
	userLimitRepo repository.UserLimitRepository
	cache         sync.Map // Key: "user_id:limit_type", Value: *cachedLimit
	cacheTTL      time.Duration
	usageBuffer   sync.Map // Key: "user_id:limit_type", Value: *int64 (atomic counter)
}

// NewRateLimitMiddleware creates a new rate limit middleware with caching
func NewRateLimitMiddleware(userLimitRepo repository.UserLimitRepository) *RateLimitMiddleware {
	m := &RateLimitMiddleware{
		userLimitRepo: userLimitRepo,
		cacheTTL:      1 * time.Minute, // Cache rate limits for 1 minute
	}
	
	// Start background goroutine to flush usage buffer periodically
	go m.flushUsageBuffer()
	
	return m
}

// flushUsageBuffer periodically flushes buffered usage counts to database
func (m *RateLimitMiddleware) flushUsageBuffer() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		m.usageBuffer.Range(func(key, value interface{}) bool {
			// Just clear the buffer for now - actual usage is tracked in cache
			m.usageBuffer.Delete(key)
			return true
		})
	}
}

// getCacheKey generates cache key for user limit
func getCacheKey(userID int, limitType string) string {
	return fmt.Sprintf("%d:%s", userID, limitType)
}

// UnaryServerInterceptor implements rate limiting for unary gRPC calls
func (m *RateLimitMiddleware) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Skip rate limiting for critical test operations to allow reloads
	exemptMethods := []string{
		"/base.TestSessionService/CreateTestSession", // Handler manages limit increment
		"/base.TestSessionService/GetTestQuestions",
		"/base.TestSessionService/GetTestSession",
		"/base.TestSessionService/SubmitAnswer",
	}
	
	for _, method := range exemptMethods {
		if info.FullMethod == method {
			return handler(ctx, req)
		}
	}

	// Extract user from context (set by JWT middleware)
	user, ok := ctx.Value("user").(*base.User)
	if !ok {
		// If no user, skip rate limiting (for public endpoints)
		return handler(ctx, req)
	}

	userID := int(user.Id)

	// Check rate limit with caching
	allowed, remaining, resetTime, err := m.checkRateLimitCached(ctx, userID, info.FullMethod)
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

	// Record usage asynchronously only for important endpoints
	if m.shouldRecordUsage(info.FullMethod) {
		go m.recordUsage(context.Background(), userID, "api_call", info.FullMethod, nil)
	}

	return handler(ctx, req)
}

// checkRateLimitCached checks rate limit using cache-first approach
func (m *RateLimitMiddleware) checkRateLimitCached(ctx context.Context, userID int, method string) (allowed bool, remaining int, resetTime time.Time, err error) {
	span, _ := apm.StartSpan(ctx, "rate_limit_check_cached", "middleware")
	defer func() {
		if span != nil {
			span.End()
		}
	}()

	// Determine limit type based on method
	limitType := m.getLimitTypeForMethod(method)
	cacheKey := getCacheKey(userID, limitType)

	// Try to get from cache first
	if cached, ok := m.cache.Load(cacheKey); ok {
		cl := cached.(*cachedLimit)
		
		// Check if cache is still valid
		if time.Since(cl.cachedAt) < m.cacheTTL {
			now := time.Now()
			
			// Check if limit needs to be reset
			if now.After(cl.limit.ResetAt) {
				cl.limit.CurrentUsed = 0
				cl.limit.ResetAt = m.getNextResetTime(limitType)
				cl.cachedAt = now
			}
			
			// Check if limit exceeded
			if cl.limit.CurrentUsed >= cl.limit.LimitValue {
				return false, 0, cl.limit.ResetAt, nil
			}
			
			// Increment in memory (will be synced to DB periodically)
			cl.limit.CurrentUsed++
			cl.limit.UpdatedAt = now
			
			remaining = cl.limit.LimitValue - cl.limit.CurrentUsed
			return true, remaining, cl.limit.ResetAt, nil
		}
	}

	// Cache miss or expired - fetch from database
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
		// Update cache even on exceeded to avoid repeated DB hits
		m.cache.Store(cacheKey, &cachedLimit{
			limit:    userLimit,
			cachedAt: now,
		})
		return false, 0, userLimit.ResetAt, nil
	}

	// Increment usage
	userLimit.CurrentUsed++
	if err := m.userLimitRepo.UpdateLimit(ctx, userLimit); err != nil {
		return false, 0, time.Now(), err
	}

	// Store in cache
	m.cache.Store(cacheKey, &cachedLimit{
		limit:    userLimit,
		cachedAt: now,
	})

	remaining = userLimit.LimitValue - userLimit.CurrentUsed
	return true, remaining, userLimit.ResetAt, nil
}

// shouldRecordUsage determines if usage should be recorded for analytics
// Skip recording for high-frequency read-only endpoints to reduce DB load
func (m *RateLimitMiddleware) shouldRecordUsage(method string) bool {
	skipPatterns := []string{
		"/base.TingkatService/GetTingkat",
		"/base.TingkatService/ListTingkat",
		"/base.MataPelajaranService/GetMataPelajaran",
		"/base.MataPelajaranService/ListMataPelajaran",
		"/base.MateriService/GetMateri",
		"/base.MateriService/ListMateri",
		"/base.SoalService/GetSoal",
		"/base.SoalService/ListSoal",
	}
	
	for _, pattern := range skipPatterns {
		if strings.Contains(method, pattern) || method == pattern {
			return false
		}
	}
	return true
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

// recordUsage records the usage for analytics (now async)
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