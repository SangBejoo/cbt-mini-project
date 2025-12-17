package entity

// JawabanOption defines valid answer options
type JawabanOption string

const (
	JawabanA JawabanOption = "A"
	JawabanB JawabanOption = "B"
	JawabanC JawabanOption = "C"
	JawabanD JawabanOption = "D"
)

type Soal struct {
	ID         int    `json:"id" gorm:"primaryKey;autoIncrement"`
	IDMateri   int    `json:"id_materi" gorm:"not null"`
	Materi     Materi `json:"materi" gorm:"foreignKey:IDMateri"`
	IDTingkat  int    `json:"id_tingkat" gorm:"not null"`
	Tingkat    Tingkat `json:"tingkat" gorm:"foreignKey:IDTingkat"`
	Pertanyaan string `json:"pertanyaan" gorm:"type:text;not null"`
	OpsiA      string `json:"opsi_a" gorm:"not null"`
	OpsiB      string `json:"opsi_b" gorm:"not null"`
	OpsiC      string `json:"opsi_c" gorm:"not null"`
	OpsiD      string `json:"opsi_d" gorm:"not null"`
	// Hati-hati! json:"-" supaya kunci jawaban tidak bocor di API umum
	JawabanBenar JawabanOption `json:"-" gorm:"type:char(1);not null"`
}

func (Soal) TableName() string { return "soal" }

// SoalForStudent represents a question for students (without correct answer)
type SoalForStudent struct {
	ID             int            `json:"id"`
	NomorUrut      int            `json:"nomor_urut"`
	Pertanyaan     string         `json:"pertanyaan"`
	OpsiA          string         `json:"opsi_a"`
	OpsiB          string         `json:"opsi_b"`
	OpsiC          string         `json:"opsi_c"`
	OpsiD          string         `json:"opsi_d"`
	JawabanDipilih *JawabanOption `json:"jawaban_dipilih"`
	IsAnswered     bool           `json:"is_answered"`
}
