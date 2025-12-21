package entity

import "time"

// UserLimit represents user-specific limits and quotas
type UserLimit struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      int       `json:"user_id" gorm:"not null;uniqueIndex:idx_user_limit_type"`
	User        User      `json:"user" gorm:"foreignKey:UserID"`
	LimitType   string    `json:"limit_type" gorm:"not null;uniqueIndex:idx_user_limit_type;size:50"` // e.g., "test_sessions_per_day", "api_requests_per_hour"
	LimitValue  int       `json:"limit_value" gorm:"not null"`
	CurrentUsed int       `json:"current_used" gorm:"not null;default:0"`
	ResetAt     time.Time `json:"reset_at" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (UserLimit) TableName() string {
	return "user_limits"
}

// Predefined limit types
const (
	LimitTypeTestSessionsPerDay    = "test_sessions_per_day"
	LimitTypeTestSessionsPerWeek   = "test_sessions_per_week"
	LimitTypeAPIRequestsPerHour    = "api_requests_per_hour"
	LimitTypeAPIRequestsPerDay     = "api_requests_per_day"
	LimitTypeQuestionsPerDay       = "questions_per_day"
)

// UserLimitUsage tracks individual usage for detailed analytics
type UserLimitUsage struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID     int       `json:"user_id" gorm:"not null;index"`
	User       User      `json:"user" gorm:"foreignKey:UserID"`
	LimitType  string    `json:"limit_type" gorm:"not null;size:50"`
	Action     string    `json:"action" gorm:"not null;size:100"` // e.g., "create_test_session", "api_call"
	ResourceID *int      `json:"resource_id"` // ID of the resource (test_session_id, question_id, etc.)
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name for GORM
func (UserLimitUsage) TableName() string {
	return "user_limit_usage"
}