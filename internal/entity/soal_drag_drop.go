package entity

import "time"

// DragDropType defines the type of drag-drop question
type DragDropType string

const (
	DragTypeOrdering DragDropType = "ordering"
	DragTypeMatching DragDropType = "matching"
)

// QuestionType defines the type of question (MC or Drag-Drop)
type QuestionType string

const (
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
	QuestionTypeDragDrop       QuestionType = "drag_drop"
)

// SoalDragDrop represents a drag-and-drop question
type SoalDragDrop struct {
	ID             int            `json:"id" gorm:"primaryKey;autoIncrement"`
	IDMateri       int            `json:"id_materi" gorm:"not null"`
	Materi         Materi         `json:"materi" gorm:"foreignKey:IDMateri"`
	IDTingkat      int            `json:"id_tingkat" gorm:"not null"`
	Tingkat        Tingkat        `json:"tingkat" gorm:"foreignKey:IDTingkat"`
	Pertanyaan     string         `json:"pertanyaan" gorm:"type:text;not null"`
	DragType       DragDropType   `json:"drag_type" gorm:"type:enum('ordering','matching');not null"`
	Pembahasan     *string        `json:"pembahasan,omitempty" gorm:"type:text"`
	IsActive       bool           `json:"is_active" gorm:"default:true"`
	Items          []DragItem     `json:"items" gorm:"foreignKey:IDSoalDragDrop;constraint:OnDelete:CASCADE"`
	Slots          []DragSlot           `json:"slots" gorm:"foreignKey:IDSoalDragDrop;constraint:OnDelete:CASCADE"`
	Gambar         []SoalDragDropGambar `json:"gambar" gorm:"foreignKey:IDSoalDragDrop;constraint:OnDelete:CASCADE"`
	CorrectAnswers []DragCorrectAnswer `json:"-" gorm:"-"` // Loaded separately for security
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

func (SoalDragDrop) TableName() string { return "soal_drag_drop" }

// DragItem represents a draggable element
type DragItem struct {
	ID             int       `json:"id" gorm:"primaryKey;autoIncrement"`
	IDSoalDragDrop int       `json:"id_soal_drag_drop" gorm:"not null"`
	Label          string    `json:"label" gorm:"size:255;not null"`
	ImageURL       *string   `json:"image_url,omitempty" gorm:"size:500"`
	Urutan         int       `json:"urutan" gorm:"default:1"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (DragItem) TableName() string { return "drag_item" }

// DragSlot represents a drop zone
type DragSlot struct {
	ID             int       `json:"id" gorm:"primaryKey;autoIncrement"`
	IDSoalDragDrop int       `json:"id_soal_drag_drop" gorm:"not null"`
	Label          string    `json:"label" gorm:"size:255;not null"`
	ImageURL       *string   `json:"image_url,omitempty" gorm:"size:500"`
	Urutan         int       `json:"urutan" gorm:"default:1"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (DragSlot) TableName() string { return "drag_slot" }

// DragCorrectAnswer represents the correct item-to-slot mapping
type DragCorrectAnswer struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	IDDragItem int       `json:"id_drag_item" gorm:"not null"`
	IDDragSlot int       `json:"id_drag_slot" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (DragCorrectAnswer) TableName() string { return "drag_correct_answer" }

// SoalDragDropForStudent represents a drag-drop question for students (no correct answers)
type SoalDragDropForStudent struct {
	ID             int          `json:"id"`
	NomorUrut      int          `json:"nomor_urut"`
	Pertanyaan     string       `json:"pertanyaan"`
	DragType       DragDropType `json:"drag_type"`
	Items          []DragItem   `json:"items"`
	Slots          []DragSlot   `json:"slots"`
	Materi         Materi       `json:"materi"`
	UserAnswer     map[int]int  `json:"user_answer,omitempty"` // item_id -> slot_id
	IsAnswered     bool         `json:"is_answered"`
}

// DragDropJawabanDetail for test results
type DragDropJawabanDetail struct {
	NomorUrut      int          `json:"nomor_urut"`
	Pertanyaan     string       `json:"pertanyaan"`
	DragType       DragDropType `json:"drag_type"`
	Items          []DragItem   `json:"items"`
	Slots          []DragSlot   `json:"slots"`
	UserAnswer     map[int]int  `json:"user_answer"`     // item_id -> slot_id
	CorrectAnswer  map[int]int  `json:"correct_answer"`  // item_id -> slot_id
	IsCorrect      bool         `json:"is_correct"`
	IsAnswered     bool         `json:"is_answered"`
	Pembahasan     *string      `json:"pembahasan,omitempty"`
}
