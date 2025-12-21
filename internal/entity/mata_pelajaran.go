package entity

// MataPelajaran represents the mata_pelajaran table
type MataPelajaran struct {
	ID       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Nama     string `json:"nama" gorm:"unique;not null;type:varchar(50)"`
	IsActive bool   `json:"is_active" gorm:"default:true"`
}

// Paksa nama tabel jadi 'mata_pelajaran' (bukan mata_pelajarans)
func (MataPelajaran) TableName() string { return "mata_pelajaran" }

// Materi represents the materi table
type Materi struct {
	ID                  int           `json:"id" gorm:"primaryKey;autoIncrement"`
	IDMataPelajaran     int           `json:"id_mata_pelajaran" gorm:"not null"`
	MataPelajaran       MataPelajaran `json:"mata_pelajaran" gorm:"foreignKey:IDMataPelajaran"`
	IDTingkat           int           `json:"id_tingkat" gorm:"not null"`
	Tingkat             Tingkat       `json:"tingkat" gorm:"foreignKey:IDTingkat"`
	Nama                string        `json:"nama" gorm:"not null;type:varchar(100)"`
	IsActive            bool          `json:"is_active" gorm:"default:true"`
	DefaultDurasiMenit  int           `json:"default_durasi_menit" gorm:"default:60"`
	DefaultJumlahSoal   int           `json:"default_jumlah_soal" gorm:"default:20"`
}

func (Materi) TableName() string { return "materi" }
