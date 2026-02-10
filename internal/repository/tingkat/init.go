package tingkat

import (
	"cbt-test-mini-project/internal/entity"
)

// TingkatRepository defines the interface for Tingkat (level) repository operations
type TingkatRepository interface {
	// Create a new tingkat
	Create(t *entity.Tingkat) error

	// Get by ID
	GetByID(id int) (*entity.Tingkat, error)

	// Update existing
	Update(t *entity.Tingkat) error

	// Delete by ID
	Delete(id int) error

	// List all
	List(limit, offset int) ([]entity.Tingkat, int, error)

	// LMS sync methods
	UpsertByLMSID(lmsID int64, name string) error
	DeleteByLMSID(lmsID int64) error
}