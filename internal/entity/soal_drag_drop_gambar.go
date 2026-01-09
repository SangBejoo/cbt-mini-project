package entity

import "time"

// SoalDragDropGambar represents image metadata for a drag-drop question
type SoalDragDropGambar struct {
	ID             int           `json:"id" gorm:"primaryKey;autoIncrement"`
	IDSoalDragDrop int           `json:"id_soal_drag_drop" gorm:"not null;index"`
	SoalDragDrop   SoalDragDrop  `json:"-" gorm:"foreignKey:IDSoalDragDrop;references:ID"`
	NamaFile       string        `json:"nama_file" gorm:"type:varchar(255);not null"`
	FilePath       string        `json:"file_path" gorm:"type:varchar(500);not null"`
	FileSize       int           `json:"file_size" gorm:"not null"`
	MimeType       string        `json:"mime_type" gorm:"type:varchar(50);not null"`
	Urutan         int           `json:"urutan" gorm:"type:tinyint;default:1;not null"`
	Keterangan     *string       `json:"keterangan" gorm:"type:varchar(255)"`
	CloudId        *string       `json:"cloud_id" gorm:"type:varchar(255)"`
	PublicId       *string       `json:"public_id" gorm:"type:varchar(500)"`
	CreatedAt      time.Time     `json:"created_at" gorm:"autoCreateTime"`
}

func (SoalDragDropGambar) TableName() string { return "soal_drag_drop_gambar" }
