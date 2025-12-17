package tingkat

import (
	"cbt-test-mini-project/internal/entity"

	"gorm.io/gorm"
)

// tingkatRepositoryImpl implements TingkatRepository
type tingkatRepositoryImpl struct {
	db *gorm.DB
}

// NewTingkatRepository creates a new TingkatRepository instance
func NewTingkatRepository(db *gorm.DB) TingkatRepository {
	return &tingkatRepositoryImpl{db: db}
}

// Create a new tingkat
func (r *tingkatRepositoryImpl) Create(t *entity.Tingkat) error {
	return r.db.Create(t).Error
}

// Get by ID
func (r *tingkatRepositoryImpl) GetByID(id int) (*entity.Tingkat, error) {
	var t entity.Tingkat
	err := r.db.First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Update existing
func (r *tingkatRepositoryImpl) Update(t *entity.Tingkat) error {
	return r.db.Save(t).Error
}

// Delete by ID
func (r *tingkatRepositoryImpl) Delete(id int) error {
	return r.db.Delete(&entity.Tingkat{}, id).Error
}

// List all
func (r *tingkatRepositoryImpl) List(limit, offset int) ([]entity.Tingkat, int, error) {
	var tingkats []entity.Tingkat
	var total int64

	query := r.db.Model(&entity.Tingkat{})

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Limit(limit).Offset(offset).Find(&tingkats).Error; err != nil {
		return nil, 0, err
	}

	return tingkats, int(total), nil
}