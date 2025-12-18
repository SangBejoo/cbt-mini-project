package soal

import (
	"cbt-test-mini-project/internal/entity"
)

// SoalUsecase defines the interface for Soal usecase operations
type SoalUsecase interface {
	CreateSoal(idMateri, idTingkat int, pertanyaan, opsiA, opsiB, opsiC, opsiD string, jawabanBenar entity.JawabanOption, imageFilesBytes [][]byte) (*entity.Soal, error)
	GetSoal(id int) (*entity.Soal, error)
	UpdateSoal(id, idMateri, idTingkat int, pertanyaan, opsiA, opsiB, opsiC, opsiD string, jawabanBenar entity.JawabanOption, imageFilesBytes [][]byte) (*entity.Soal, error)
	DeleteSoal(id int) error
	ListSoal(idMateri, tingkatan, idMataPelajaran int, page, pageSize int) ([]entity.Soal, *entity.PaginationResponse, error)
	UploadImageToSoal(idSoal int, imageBytes []byte, namaFile string, urutan int, keterangan *string) (*entity.SoalGambar, error)
	DeleteImageFromSoal(idGambar int) error
	UpdateImageInSoal(idGambar int, urutan int, keterangan *string) error
}