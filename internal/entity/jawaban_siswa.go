package entity

import "time"

type JawabanSiswa struct {
	ID int `json:"id" gorm:"primaryKey;autoIncrement"`

	// Unique Index penting buat fitur UPSERT
	IDTestSessionSoal int             `json:"id_test_session_soal" gorm:"not null;unique"`
	TestSessionSoal   TestSessionSoal `json:"-" gorm:"foreignKey:IDTestSessionSoal;constraint:OnDelete:CASCADE"`

	JawabanDipilih *JawabanOption `json:"jawaban_dipilih" gorm:"type:char(1)"`
	IsCorrect      bool           `json:"is_correct" gorm:"not null"`
	DijawabPada    time.Time      `json:"dijawab_pada" gorm:"autoCreateTime:milli"` // Create once, don't update
}

func (JawabanSiswa) TableName() string { return "jawaban_siswa" }

// JawabanDetail for test results
type JawabanDetail struct {
	NomorUrut      int            `json:"nomor_urut"`
	Pertanyaan     string         `json:"pertanyaan"`
	OpsiA          string         `json:"opsi_a"`
	OpsiB          string         `json:"opsi_b"`
	OpsiC          string         `json:"opsi_c"`
	OpsiD          string         `json:"opsi_d"`
	JawabanDipilih *JawabanOption `json:"jawaban_dipilih"`
	JawabanBenar   JawabanOption  `json:"jawaban_benar"`
	IsCorrect      bool           `json:"is_correct"`
}
