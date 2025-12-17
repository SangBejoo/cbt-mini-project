package materi

import (
	"cbt-test-mini-project/internal/entity"
)

// MateriRepository defines the interface for Materi (material) repository operations
type MateriRepository interface {
	// Create a new materi
	Create(materi *entity.Materi) error

	// Get by ID
	GetByID(id int) (*entity.Materi, error)

	// Update existing
	Update(materi *entity.Materi) error

	// Delete by ID
	Delete(id int) error

	// List with filters
	List(idMataPelajaran, tingkatan *int, limit, offset int) ([]entity.Materi, int, error)

	// Get by mata pelajaran ID
	GetByMataPelajaranID(idMataPelajaran int) ([]entity.Materi, error)
}