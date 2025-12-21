package mata_pelajaran

import (
	"cbt-test-mini-project/internal/entity"

	"gorm.io/gorm"
)

// mataPelajaranRepositoryImpl implements MataPelajaranRepository
type mataPelajaranRepositoryImpl struct {
	db *gorm.DB
}

// NewMataPelajaranRepository creates a new MataPelajaranRepository instance
func NewMataPelajaranRepository(db *gorm.DB) MataPelajaranRepository {
	return &mataPelajaranRepositoryImpl{db: db}
}

// Create a new mata pelajaran
func (r *mataPelajaranRepositoryImpl) Create(mp *entity.MataPelajaran) error {
	return r.db.Create(mp).Error
}

// Get by ID
func (r *mataPelajaranRepositoryImpl) GetByID(id int) (*entity.MataPelajaran, error) {
	var mp entity.MataPelajaran
	err := r.db.Where("is_active = ?", true).First(&mp, id).Error
	if err != nil {
		return nil, err
	}
	return &mp, nil
}

// Update existing
func (r *mataPelajaranRepositoryImpl) Update(mp *entity.MataPelajaran) error {
	return r.db.Save(mp).Error
}

// Delete by ID (soft delete)
func (r *mataPelajaranRepositoryImpl) Delete(id int) error {
	return r.db.Model(&entity.MataPelajaran{}).Where("id = ?", id).Update("is_active", false).Error
}

// List all
func (r *mataPelajaranRepositoryImpl) List(limit, offset int) ([]entity.MataPelajaran, int, error) {
	var mps []entity.MataPelajaran
	var total int64

	query := r.db.Model(&entity.MataPelajaran{}).Where("is_active = ?", true)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Limit(limit).Offset(offset).Find(&mps).Error; err != nil {
		return nil, 0, err
	}

	return mps, int(total), nil
}

// Get by name
func (r *mataPelajaranRepositoryImpl) GetByName(name string) (*entity.MataPelajaran, error) {
	var mp entity.MataPelajaran
	err := r.db.Where("nama = ? AND is_active = ?", name, true).First(&mp).Error
	if err != nil {
		return nil, err
	}
	return &mp, nil
}