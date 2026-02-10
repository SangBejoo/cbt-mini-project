package entity

import "time"

// ClassStudent represents student enrollment in a class (synced from LMS)
type ClassStudent struct {
	ID         int       `json:"id" db:"id"`
	LMSClassID int64     `json:"lms_class_id" db:"lms_class_id"`
	LMSUserID  int64     `json:"lms_user_id" db:"lms_user_id"`
	JoinedAt   time.Time `json:"joined_at" db:"joined_at"`
}
