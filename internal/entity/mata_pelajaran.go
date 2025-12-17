package entity

// MataPelajaran represents the mata_pelajaran table
type MataPelajaran struct {
	ID   int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Nama string `json:"nama" gorm:"unique;not null;type:varchar(50)"`
}

// Paksa nama tabel jadi 'mata_pelajaran' (bukan mata_pelajarans)
func (MataPelajaran) TableName() string { return "mata_pelajaran" }

// Materi represents the materi table
type Materi struct {
	ID              int           `json:"id" gorm:"primaryKey;autoIncrement"`
	IDMataPelajaran int           `json:"id_mata_pelajaran" gorm:"not null"`
	MataPelajaran   MataPelajaran `json:"mata_pelajaran" gorm:"foreignKey:IDMataPelajaran"`
	Nama            string        `json:"nama" gorm:"not null;type:varchar(100)"`
	Tingkatan       int           `json:"tingkatan" gorm:"not null"`
}

func (Materi) TableName() string { return "materi" }
