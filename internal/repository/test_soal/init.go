package test_soal

import (
	"cbt-test-mini-project/internal/entity"
)

// SoalRepository defines the interface for Soal (question) repository operations
type SoalRepository interface {
	// Create a new soal
	Create(soal *entity.Soal) error

	// Get soal by ID
	GetByID(id int) (*entity.Soal, error)

	// Update existing soal
	Update(soal *entity.Soal) error

	// Delete soal by ID
	Delete(id int) error

	// List soal with filters
	List(idMateri, tingkatan, idMataPelajaran *int, limit, offset int) ([]entity.Soal, int, error)

	// Get soal by materi ID
	GetByMateriID(idMateri int) ([]entity.Soal, error)

	// Image operations
	CreateGambar(gambar *entity.SoalGambar) error
	GetGambarByID(id int) (*entity.SoalGambar, error)
	UpdateGambar(id int, urutan int, keterangan *string) error
	DeleteGambar(id int) error
}