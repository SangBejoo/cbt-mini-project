package entity

import "time"

// Class represents the classes table (synced from LMS)
type Class struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	LMSClassID  int64     `json:"lms_class_id" gorm:"uniqueIndex;not null"`
	LMSSchoolID int64     `json:"lms_school_id" gorm:"not null;index"`
	Name        string    `json:"name" gorm:"size:255;not null"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Class) TableName() string { return "classes" }
