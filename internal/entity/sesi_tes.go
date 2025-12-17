package entity

import "time"

// TestStatus defines valid test session statuses
type TestStatus string

const (
	TestStatusOngoing   TestStatus = "ongoing"
	TestStatusCompleted TestStatus = "completed"
	TestStatusTimeout   TestStatus = "timeout"
)

// TestSession represents the test_session table
type TestSession struct {
	ID              int           `json:"id" gorm:"primaryKey;autoIncrement"`
	SessionToken    string        `json:"session_token" gorm:"unique;not null;size:64"`
	NamaPeserta     string        `json:"nama_peserta" gorm:"not null;size:100"`
	IDTingkat       int           `json:"id_tingkat" gorm:"not null"`
	Tingkat         Tingkat       `json:"tingkat" gorm:"foreignKey:IDTingkat"`
	IDMataPelajaran int           `json:"id_mata_pelajaran" gorm:"not null"`
	MataPelajaran   MataPelajaran `json:"mata_pelajaran" gorm:"foreignKey:IDMataPelajaran"`

	// Ganti default:CURRENT_TIMESTAMP dengan autoCreateTime biar dihandle GORM
	WaktuMulai   time.Time  `json:"waktu_mulai" gorm:"autoCreateTime"`
	WaktuSelesai *time.Time `json:"waktu_selesai"`

	DurasiMenit int        `json:"durasi_menit" gorm:"not null"`
	NilaiAkhir  *float64   `json:"nilai_akhir" gorm:"type:decimal(5,2)"`
	JumlahBenar *int       `json:"jumlah_benar"`
	TotalSoal   *int       `json:"total_soal"`
	Status      TestStatus `json:"status" gorm:"type:enum('ongoing','completed','timeout');default:'ongoing'"`
}

func (TestSession) TableName() string { return "test_session" }

// BatasWaktu calculates deadline from WaktuMulai + DurasiMenit
func (ts TestSession) BatasWaktu() time.Time {
	return ts.WaktuMulai.Add(time.Duration(ts.DurasiMenit) * time.Minute)
}

// TestSessionSoal represents the test_session_soal table
type TestSessionSoal struct {
	ID int `json:"id" gorm:"primaryKey;autoIncrement"`

	// Composite Unique Index: Satu sesi tidak boleh punya dua soal dengan nomor urut sama
	IDTestSession int         `json:"id_test_session" gorm:"not null;index:idx_session_urut,unique"`
	TestSession   TestSession `json:"-" gorm:"foreignKey:IDTestSession;constraint:OnDelete:CASCADE"`

	IDSoal int  `json:"id_soal" gorm:"not null"`
	Soal   Soal `json:"soal" gorm:"foreignKey:IDSoal"`

	NomorUrut int `json:"nomor_urut" gorm:"not null;index:idx_session_urut,unique"`
}

func (TestSessionSoal) TableName() string { return "test_session_soal" }
