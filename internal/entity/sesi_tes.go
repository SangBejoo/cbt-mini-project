package entity

import "time"

// TestSession represents the test_session table
type TestSession struct {
	ID              int           `json:"id" gorm:"primaryKey;autoIncrement"`
	SessionToken    string        `json:"session_token" gorm:"unique;not null"`
	NamaPeserta     string        `json:"nama_peserta" gorm:"not null"`
	Tingkatan       int           `json:"tingkatan" gorm:"not null"`
	IDMataPelajaran int           `json:"id_mata_pelajaran" gorm:"not null"`
	MataPelajaran   MataPelajaran `json:"mata_pelajaran" gorm:"foreignKey:IDMataPelajaran"`
	WaktuMulai      time.Time     `json:"waktu_mulai" gorm:"not null;default:CURRENT_TIMESTAMP"`
	WaktuSelesai    *time.Time    `json:"waktu_selesai"`
	DurasiMenit     int           `json:"durasi_menit" gorm:"not null"`
	NilaiAkhir      *float64      `json:"nilai_akhir" gorm:"type:decimal(5,2)"`
	JumlahBenar     *int          `json:"jumlah_benar"`
	TotalSoal       *int          `json:"total_soal"`
	Status          string        `json:"status" gorm:"type:enum('ongoing','completed');default:'ongoing'"`
}

// TestSessionSoal represents the test_session_soal table
type TestSessionSoal struct {
	ID            int         `json:"id" gorm:"primaryKey;autoIncrement"`
	IDTestSession int         `json:"id_test_session" gorm:"not null"`
	TestSession   TestSession `json:"test_session" gorm:"foreignKey:IDTestSession;constraint:OnDelete:CASCADE"`
	IDSoal        int         `json:"id_soal" gorm:"not null"`
	Soal          Soal        `json:"soal" gorm:"foreignKey:IDSoal"`
	NomorUrut     int         `json:"nomor_urut" gorm:"not null"`
}
type JawabanSiswa struct {
	ID                int             `json:"id" gorm:"primaryKey;autoIncrement"`
	IDTestSessionSoal int             `json:"id_test_session_soal" gorm:"not null"`
	TestSessionSoal   TestSessionSoal `json:"test_session_soal" gorm:"foreignKey:IDTestSessionSoal;constraint:OnDelete:CASCADE"`
	JawabanDipilih    *string         `json:"jawaban_dipilih" gorm:"type:char(1)"`
	IsCorrect         bool            `json:"is_correct" gorm:"not null"`
	DijawabPada       time.Time       `json:"dijawab_pada" gorm:"default:CURRENT_TIMESTAMP"`
}

// JawabanDetail for test results
type JawabanDetail struct {
	NomorUrut      int     `json:"nomor_urut"`
	Pertanyaan     string  `json:"pertanyaan"`
	OpsiA          string  `json:"opsi_a"`
	OpsiB          string  `json:"opsi_b"`
	OpsiC          string  `json:"opsi_c"`
	OpsiD          string  `json:"opsi_d"`
	JawabanDipilih *string `json:"jawaban_dipilih"`
	JawabanBenar   string  `json:"jawaban_benar"`
	IsCorrect      bool    `json:"is_correct"`
}
