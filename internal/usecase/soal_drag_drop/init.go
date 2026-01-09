package soal_drag_drop

import (
	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/internal/entity"
	repository "cbt-test-mini-project/internal/repository/soal_drag_drop"
)

// Usecase interface for soal_drag_drop business logic
type Usecase interface {
	Create(req *CreateRequest) (*entity.SoalDragDrop, error)
	GetByID(id int) (*entity.SoalDragDrop, error)
	GetByIDWithCorrectAnswers(id int) (*entity.SoalDragDrop, []entity.DragCorrectAnswer, error)
	Update(id int, req *UpdateRequest) (*entity.SoalDragDrop, error)
	Delete(id int) error
	List(idMateri, idTingkat int, page, pageSize int) ([]entity.SoalDragDrop, int64, error)
	GetActiveByMateri(idMateri int, limit int) ([]entity.SoalDragDrop, error)
	CheckDragDropAnswer(soalID int, userAnswer map[int]int) (bool, error)
	CountByMateri(idMateri int) (int64, error)
}

// CreateRequest for creating a new drag-drop question
type CreateRequest struct {
	IDMateri       int
	IDTingkat      int
	Pertanyaan     string
	DragType       entity.DragDropType
	Pembahasan     *string
	Items          []ItemRequest
	Slots          []SlotRequest
	CorrectAnswers []CorrectAnswerRequest
}

// ItemRequest for creating a drag item
type ItemRequest struct {
	Label    string
	ImageURL *string
	Urutan   int
}

// SlotRequest for creating a drag slot
type SlotRequest struct {
	Label    string
	ImageURL *string
	Urutan   int
}

// CorrectAnswerRequest for creating a correct answer mapping
type CorrectAnswerRequest struct {
	ItemUrutan int // References by urutan (frontend doesn't know IDs yet)
	SlotUrutan int
}

// UpdateRequest for updating a drag-drop question
type UpdateRequest struct {
	IDMateri       int
	IDTingkat      int
	Pertanyaan     string
	DragType       entity.DragDropType
	Pembahasan     *string
	IsActive       bool
	Items          []ItemRequest
	Slots          []SlotRequest
	CorrectAnswers []CorrectAnswerRequest
}

type usecase struct {
	repo   repository.Repository
	config *config.Main
}

// NewUsecase creates a new soal_drag_drop usecase
func NewUsecase(repo repository.Repository, config *config.Main) Usecase {
	return &usecase{
		repo:   repo,
		config: config,
	}
}
