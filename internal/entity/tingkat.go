package entity

// Tingkat represents the tingkat table
type Tingkat struct {
	ID   int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Nama string `json:"nama" gorm:"unique;not null;type:varchar(50)"`
}

func (Tingkat) TableName() string { return "tingkat" }