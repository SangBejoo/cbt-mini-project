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
	err := r.db.Preload("MataPelajaran").Preload("Tingkat").First(&materi, id).Error
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

	// Optimize query with indexes and eager loading
	query := r.db.Model(&entity.Materi{}).
		Select("materi.*, mata_pelajaran.nama as mata_pelajaran_nama, tingkat.nama as tingkat_nama").
		Joins("LEFT JOIN mata_pelajaran ON materi.id_mata_pelajaran = mata_pelajaran.id").
		Joins("LEFT JOIN tingkat ON materi.id_tingkat = tingkat.id")

	if idMataPelajaran != nil {
		query = query.Where("materi.id_mata_pelajaran = ?", *idMataPelajaran)
	}
	if idTingkat != nil {
		query = query.Where("materi.id_tingkat = ?", *idTingkat)
	}

	// Count total with optimized query
	countQuery := r.db.Model(&entity.Materi{}).
		Joins("LEFT JOIN mata_pelajaran ON materi.id_mata_pelajaran = mata_pelajaran.id").
		Joins("LEFT JOIN tingkat ON materi.id_tingkat = tingkat.id")
	
	if idMataPelajaran != nil {
		countQuery = countQuery.Where("materi.id_mata_pelajaran = ?", *idMataPelajaran)
	}
	if idTingkat != nil {
		countQuery = countQuery.Where("materi.id_tingkat = ?", *idTingkat)
	}
	
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with preloads
	if err := query.Preload("MataPelajaran").
		Preload("Tingkat").
		Order("materi.id DESC").
		Limit(limit).Offset(offset).
		Find(&materis).Error; err != nil {
		return nil, 0, err
	}

	return materis, int(total), nil
}

// Get by mata pelajaran ID
func (r *materiRepositoryImpl) GetByMataPelajaranID(idMataPelajaran int) ([]entity.Materi, error) {
	var materis []entity.Materi
	err := r.db.Preload("MataPelajaran").Preload("Tingkat").Where("id_mata_pelajaran = ?", idMataPelajaran).Find(&materis).Error
	return materis, err
}