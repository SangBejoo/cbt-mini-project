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
	ID           int          `json:"id" gorm:"primaryKey;autoIncrement"`
	IDMateri     int          `json:"id_materi" gorm:"not null"`
	Materi       Materi       `json:"materi" gorm:"foreignKey:IDMateri"`
	IDTingkat    int          `json:"id_tingkat" gorm:"not null"`
	Tingkat      Tingkat      `json:"tingkat" gorm:"foreignKey:IDTingkat"`
	Pertanyaan   string       `json:"pertanyaan" gorm:"type:text;not null"`
	OpsiA        string       `json:"opsi_a" gorm:"not null"`
	OpsiB        string       `json:"opsi_b" gorm:"not null"`
	OpsiC        string       `json:"opsi_c" gorm:"not null"`
	OpsiD        string       `json:"opsi_d" gorm:"not null"`
	JawabanBenar JawabanOption `json:"-" gorm:"type:char(1);not null"`
	Pembahasan   *string      `json:"pembahasan,omitempty" gorm:"type:text"`
	IsActive     bool         `json:"is_active" gorm:"default:true"`
	Gambar       []SoalGambar `json:"gambar" gorm:"foreignKey:IDSoal;references:ID;constraint:OnDelete:CASCADE"`
}

func (Soal) TableName() string { return "soal" }

// SoalForStudent represents a question for students (without correct answer)
type SoalForStudent struct {
	ID             int          `json:"id"`
	NomorUrut      int          `json:"nomor_urut"`
	Pertanyaan     string       `json:"pertanyaan"`
	OpsiA          string       `json:"opsi_a"`
	OpsiB          string       `json:"opsi_b"`
	OpsiC          string       `json:"opsi_c"`
	OpsiD          string       `json:"opsi_d"`
	JawabanDipilih *JawabanOption `json:"jawaban_dipilih"`
	IsAnswered     bool         `json:"is_answered"`
	Materi         Materi       `json:"materi"`
	Gambar         []SoalGambar `json:"gambar"`
}

// QuestionForStudent represents a unified question for students (multiple choice or drag-drop)
type QuestionForStudent struct {
	NomorUrut    int         `json:"nomor_urut"`
	QuestionType QuestionType `json:"question_type"`
	Materi       Materi      `json:"materi"`
	IsAnswered   bool        `json:"is_answered"`

	// Multiple choice fields
	MCID             *int          `json:"mc_id,omitempty"`
	MCPertanyaan     *string       `json:"mc_pertanyaan,omitempty"`
	MCOpsiA          *string       `json:"mc_opsi_a,omitempty"`
	MCOpsiB          *string       `json:"mc_opsi_b,omitempty"`
	MCOpsiC          *string       `json:"mc_opsi_c,omitempty"`
	MCOpsiD          *string       `json:"mc_opsi_d,omitempty"`
	MCJawabanDipilih *JawabanOption `json:"mc_jawaban_dipilih,omitempty"`
	MCGambar         []SoalGambar  `json:"mc_gambar,omitempty"`

	// Drag-drop fields
	DDID          *int                  `json:"dd_id,omitempty"`
	DDPertanyaan  *string               `json:"dd_pertanyaan,omitempty"`
	DDDDragType   *DragDropType         `json:"dd_drag_type,omitempty"`
	DDItems       []DragItem            `json:"dd_items,omitempty"`
	DDSlots       []DragSlot            `json:"dd_slots,omitempty"`
	DDUserAnswer  map[int]int           `json:"dd_user_answer,omitempty"`
}
