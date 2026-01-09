package soal_drag_drop

import (
	"cbt-test-mini-project/internal/entity"

	"gorm.io/gorm"
)

// Repository interface for soal_drag_drop operations
type Repository interface {
	Create(soal *entity.SoalDragDrop, items []entity.DragItem, slots []entity.DragSlot, correctAnswers []entity.DragCorrectAnswer) error
	GetByID(id int) (*entity.SoalDragDrop, error)
	GetByIDWithCorrectAnswers(id int) (*entity.SoalDragDrop, []entity.DragCorrectAnswer, error)
	Update(soal *entity.SoalDragDrop, items []entity.DragItem, slots []entity.DragSlot, correctAnswers []entity.DragCorrectAnswer) error
	Delete(id int) error
	List(idMateri, idTingkat int, page, pageSize int) ([]entity.SoalDragDrop, int64, error)
	GetActiveByMateri(idMateri int, limit int) ([]entity.SoalDragDrop, error)
	GetCorrectAnswersBySoalID(soalID int) ([]entity.DragCorrectAnswer, error)
	CountByMateri(idMateri int) (int64, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new soal_drag_drop repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}
