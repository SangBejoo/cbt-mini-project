package mata_pelajaran

import (
	"cbt-test-mini-project/internal/entity"
)

// MataPelajaranUsecase defines the interface for MataPelajaran usecase operations
type MataPelajaranUsecase interface {
	CreateMataPelajaran(nama string) (*entity.MataPelajaran, error)
	GetMataPelajaran(id int) (*entity.MataPelajaran, error)
	UpdateMataPelajaran(id int, nama string) (*entity.MataPelajaran, error)
	DeleteMataPelajaran(id int) error
	ListMataPelajaran(page, pageSize int) ([]entity.MataPelajaran, *entity.PaginationResponse, error)
}