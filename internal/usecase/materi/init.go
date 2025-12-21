package materi

import (
	"cbt-test-mini-project/internal/entity"
)

// MateriUsecase defines the interface for Materi usecase operations
type MateriUsecase interface {
	CreateMateri(idMataPelajaran int, nama string, idTingkat int, isActive bool, defaultDurasiMenit, defaultJumlahSoal int) (*entity.Materi, error)
	GetMateri(id int) (*entity.Materi, error)
	UpdateMateri(id, idMataPelajaran int, nama string, idTingkat int, isActive bool, defaultDurasiMenit, defaultJumlahSoal int) (*entity.Materi, error)
	DeleteMateri(id int) error
	ListMateri(idMataPelajaran, idTingkat int, page, pageSize int) ([]entity.Materi, *entity.PaginationResponse, error)
}