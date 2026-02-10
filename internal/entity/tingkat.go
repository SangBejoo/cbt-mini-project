package entity

// Tingkat represents the tingkat table
type Tingkat struct {
	ID         int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Nama       string `json:"nama" gorm:"unique;not null;type:varchar(50)"`
	IsActive   bool   `json:"is_active" gorm:"default:true"`
	LmsLevelID *int64 `json:"lms_level_id" gorm:"column:lms_level_id"`
}

func (Tingkat) TableName() string { return "tingkat" }