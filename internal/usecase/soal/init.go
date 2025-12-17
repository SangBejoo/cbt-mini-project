package soal

import (
	"cbt-test-mini-project/internal/entity"
)

// SoalUsecase defines the interface for Soal usecase operations
type SoalUsecase interface {
	CreateSoal(idMateri int, pertanyaan, opsiA, opsiB, opsiC, opsiD string, jawabanBenar entity.JawabanOption) (*entity.Soal, error)
	GetSoal(id int) (*entity.Soal, error)
	UpdateSoal(id, idMateri int, pertanyaan, opsiA, opsiB, opsiC, opsiD string, jawabanBenar entity.JawabanOption) (*entity.Soal, error)
	DeleteSoal(id int) error
	ListSoal(idMateri, tingkatan, idMataPelajaran int, page, pageSize int) ([]entity.Soal, *entity.PaginationResponse, error)
}