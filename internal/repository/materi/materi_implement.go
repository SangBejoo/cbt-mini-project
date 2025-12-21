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
	err := r.db.Where("is_active = ?", true).Preload("MataPelajaran").Preload("Tingkat").First(&materi, id).Error
	if err != nil {
		return nil, err
	}
	return &materi, nil
}

// Update existing
func (r *materiRepositoryImpl) Update(materi *entity.Materi) error {
	return r.db.Save(materi).Error
}

// Delete by ID (soft delete)
func (r *materiRepositoryImpl) Delete(id int) error {
	return r.db.Model(&entity.Materi{}).Where("id = ?", id).Update("is_active", false).Error
}

// List with filters
func (r *materiRepositoryImpl) List(idMataPelajaran, idTingkat *int, limit, offset int) ([]entity.Materi, int, error) {
	var materis []entity.Materi
	var total int64

	query := r.db.Model(&entity.Materi{}).Where("is_active = ?", true).Preload("MataPelajaran").Preload("Tingkat")

	if idMataPelajaran != nil {
		query = query.Where("id_mata_pelajaran = ?", *idMataPelajaran)
	}
	if idTingkat != nil {
		query = query.Where("id_tingkat = ?", *idTingkat)
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
	err := r.db.Preload("MataPelajaran").Preload("Tingkat").Where("id_mata_pelajaran = ? AND is_active = ?", idMataPelajaran, true).Find(&materis).Error
	return materis, err
}