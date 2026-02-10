package entity

import "time"

// User represents the users table
type User struct {
	ID           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Email        string    `json:"email" gorm:"unique;not null;size:100"`
	PasswordHash string    `json:"password_hash" gorm:"not null;size:255"`
	Nama         string    `json:"nama" gorm:"not null;size:100"`
	Role         string    `json:"role" gorm:"not null;type:enum('siswa','admin');default:'siswa'"`
	IsActive     bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	LmsUserID    *int64    `json:"lms_user_id" gorm:"column:lms_user_id"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}