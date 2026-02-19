package entity

import (
	"encoding/json"
	"time"
)

type JawabanSiswa struct {
	ID int `json:"id" gorm:"primaryKey;autoIncrement"`

	// Unique Index penting buat fitur UPSERT
	IDTestSessionSoal int             `json:"id_test_session_soal" gorm:"not null;unique"`
	TestSessionSoal   TestSessionSoal `json:"-" gorm:"foreignKey:IDTestSessionSoal;constraint:OnDelete:CASCADE"`

	// Multiple-choice answer (for MULTIPLE_CHOICE questions)
	JawabanDipilih *JawabanOption `json:"jawaban_dipilih" gorm:"type:char(1)"`

	// Question type for routing
	QuestionType QuestionType `json:"question_type" gorm:"type:enum('multiple_choice','drag_drop','essay','multiple_choices_complex');default:'multiple_choice'"`

	JawabanDipilihComplex *string `json:"jawaban_dipilih_complex,omitempty" gorm:"column:jawaban_dipilih_complex;type:json"`

	// Drag-drop answer (for DRAG_DROP questions) - stored as JSON
	JawabanDragDrop *string  `json:"jawaban_drag_drop,omitempty" gorm:"type:json"`
	JawabanEssay    *string  `json:"jawaban_essay,omitempty" gorm:"column:jawaban_essay;type:text"`
	NilaiEssay      *float64 `json:"nilai_essay,omitempty" gorm:"column:nilai_essay;type:decimal(5,2)"`
	FeedbackTeacher *string  `json:"feedback_teacher,omitempty" gorm:"column:feedback_teacher;type:text"`

	IsCorrect   bool      `json:"is_correct" gorm:"not null"`
	DijawabPada time.Time `json:"dijawab_pada" gorm:"autoCreateTime:milli"` // Create once, don't update
}

func (JawabanSiswa) TableName() string { return "jawaban_siswa" }

// GetDragDropAnswer parses the JSON drag-drop answer
func (j *JawabanSiswa) GetDragDropAnswer() map[int]int {
	if j.JawabanDragDrop == nil {
		return nil
	}
	var answer map[int]int
	json.Unmarshal([]byte(*j.JawabanDragDrop), &answer)
	return answer
}

// SetDragDropAnswer serializes the drag-drop answer to JSON
func (j *JawabanSiswa) SetDragDropAnswer(answer map[int]int) error {
	data, err := json.Marshal(answer)
	if err != nil {
		return err
	}
	str := string(data)
	j.JawabanDragDrop = &str
	return nil
}

// JawabanDetail for test results (multiple-choice)
type JawabanDetail struct {
	NomorUrut      int            `json:"nomor_urut"`
	QuestionType   QuestionType   `json:"question_type"`
	Pertanyaan     string         `json:"pertanyaan"`
	OpsiA          string         `json:"opsi_a,omitempty"`
	OpsiB          string         `json:"opsi_b,omitempty"`
	OpsiC          string         `json:"opsi_c,omitempty"`
	OpsiD          string         `json:"opsi_d,omitempty"`
	JawabanDipilih *JawabanOption `json:"jawaban_dipilih,omitempty"`
	JawabanBenar   JawabanOption  `json:"jawaban_benar,omitempty"`
	IsCorrect      bool           `json:"is_correct"`
	IsAnswered     bool           `json:"is_answered"`
	Pembahasan     *string        `json:"pembahasan,omitempty"`
	Gambar         []SoalGambar   `json:"gambar"`
	// Drag Drop Fields
	DragType          *DragDropType `json:"drag_type,omitempty"`
	DragItems         []DragItem    `json:"items,omitempty"`
	DragSlots         []DragSlot    `json:"slots,omitempty"`
	UserDragAnswer    map[int]int   `json:"user_drag_answer,omitempty"`
	CorrectDragAnswer map[int]int   `json:"correct_drag_answer,omitempty"`
	JawabanEssay      *string       `json:"jawaban_essay,omitempty"`
	NilaiEssay        *float64      `json:"nilai_essay,omitempty"`
	FeedbackTeacher   *string       `json:"feedback_teacher,omitempty"`
	JawabanDipilihComplex []JawabanOption `json:"jawaban_dipilih_complex,omitempty"`
	JawabanBenarComplex []JawabanOption `json:"jawaban_benar_complex,omitempty"`
}

func (j *JawabanSiswa) GetJawabanDipilihComplex() []JawabanOption {
	if j.JawabanDipilihComplex == nil {
		return nil
	}
	var raw []string
	if err := json.Unmarshal([]byte(*j.JawabanDipilihComplex), &raw); err != nil {
		return nil
	}
	answers := make([]JawabanOption, 0, len(raw))
	for _, option := range raw {
		answers = append(answers, JawabanOption(option))
	}
	return answers
}

func (j *JawabanSiswa) SetJawabanDipilihComplex(options []JawabanOption) error {
	if len(options) == 0 {
		j.JawabanDipilihComplex = nil
		return nil
	}
	raw := make([]string, 0, len(options))
	for _, option := range options {
		raw = append(raw, string(option))
	}
	bytes, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	encoded := string(bytes)
	j.JawabanDipilihComplex = &encoded
	return nil
}
