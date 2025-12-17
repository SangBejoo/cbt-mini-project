package tingkat

import (
	"cbt-test-mini-project/internal/entity"
)

// TingkatUsecase defines the interface for Tingkat (level) usecase operations
type TingkatUsecase interface {
	// Create a new tingkat
	CreateTingkat(nama string) (*entity.Tingkat, error)

	// Get tingkat by ID
	GetTingkat(id int) (*entity.Tingkat, error)

	// Update tingkat
	UpdateTingkat(id int, nama string) (*entity.Tingkat, error)

	// Delete tingkat
	DeleteTingkat(id int) error

	// List tingkat
	ListTingkat(page, pageSize int) ([]entity.Tingkat, *entity.PaginationResponse, error)
}