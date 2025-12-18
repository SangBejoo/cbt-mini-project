package entity

import "time"

// SoalGambar represents image metadata for a question
type SoalGambar struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	IDSoal    int       `json:"id_soal" gorm:"not null;index"`
	Soal      Soal      `json:"-" gorm:"foreignKey:IDSoal;references:ID"`
	NamaFile  string    `json:"nama_file" gorm:"type:varchar(255);not null"`
	FilePath  string    `json:"file_path" gorm:"type:varchar(500);not null"`
	FileSize  int       `json:"file_size" gorm:"not null"`
	MimeType  string    `json:"mime_type" gorm:"type:varchar(50);not null"`
	Urutan    int       `json:"urutan" gorm:"type:tinyint;default:1;not null"`
	Keterangan *string  `json:"keterangan" gorm:"type:varchar(255)"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (SoalGambar) TableName() string { return "soal_gambar" }
