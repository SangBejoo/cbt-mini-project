package materi

import (
	"cbt-test-mini-project/internal/entity"

	"gorm.io/gorm"
)

// materiRepositoryImpl implements MateriRepository
type materiRepositoryImpl struct {
	db *gorm.DB
}

// NewMateriRepository creates a new MateriRepository instance
func NewMateriRepository(db *gorm.DB) MateriRepository {
	return &materiRepositoryImpl{db: db}
}

// Create a new materi
func (r *materiRepositoryImpl) Create(materi *entity.Materi) error {
	return r.db.Create(materi).Error
}

// Get by ID
func (r *materiRepositoryImpl) GetByID(id int) (*entity.Materi, error) {
	var materi entity.Materi
	err := r.db.Preload("MataPelajaran").First(&materi, id).Error
	if err != nil {
		return nil, err
	}
	return &materi, nil
}

// Update existing
func (r *materiRepositoryImpl) Update(materi *entity.Materi) error {
	return r.db.Save(materi).Error
}

// Delete by ID
func (r *materiRepositoryImpl) Delete(id int) error {
	return r.db.Delete(&entity.Materi{}, id).Error
}

// List with filters
func (r *materiRepositoryImpl) List(idMataPelajaran, tingkatan *int, limit, offset int) ([]entity.Materi, int, error) {
	var materis []entity.Materi
	var total int64

	query := r.db.Model(&entity.Materi{}).Preload("MataPelajaran")

	if idMataPelajaran != nil {
		query = query.Where("id_mata_pelajaran = ?", *idMataPelajaran)
	}
	if tingkatan != nil {
		query = query.Where("tingkatan = ?", *tingkatan)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Limit(limit).Offset(offset).Find(&materis).Error; err != nil {
		return nil, 0, err
	}

	return materis, int(total), nil
}

// Get by mata pelajaran ID
func (r *materiRepositoryImpl) GetByMataPelajaranID(idMataPelajaran int) ([]entity.Materi, error) {
	var materis []entity.Materi
	err := r.db.Preload("MataPelajaran").Where("id_mata_pelajaran = ?", idMataPelajaran).Find(&materis).Error
	return materis, err
}