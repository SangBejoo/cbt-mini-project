package mata_pelajaran

import (
	"cbt-test-mini-project/internal/entity"
)

// MataPelajaranRepository defines the interface for MataPelajaran (subject) repository operations
type MataPelajaranRepository interface {
	// Create a new mata pelajaran
	Create(mp *entity.MataPelajaran) error

	// Get by ID
	GetByID(id int) (*entity.MataPelajaran, error)

	// Update existing
	Update(mp *entity.MataPelajaran) error

	// Delete by ID
	Delete(id int) error

	// List all
	List(limit, offset int) ([]entity.MataPelajaran, int, error)

	// Get by name
	GetByName(name string) (*entity.MataPelajaran, error)
}