package test_soal

import (
	"cbt-test-mini-project/internal/entity"

	"gorm.io/gorm"
)

// soalRepositoryImpl implements SoalRepository
type soalRepositoryImpl struct {
	db *gorm.DB
}

// NewSoalRepository creates a new SoalRepository instance
func NewSoalRepository(db *gorm.DB) SoalRepository {
	return &soalRepositoryImpl{db: db}
}

// Create a new soal
func (r *soalRepositoryImpl) Create(soal *entity.Soal) error {
	return r.db.Create(soal).Error
}

// Get soal by ID
func (r *soalRepositoryImpl) GetByID(id int) (*entity.Soal, error) {
	var soal entity.Soal
	err := r.db.Preload("Materi").Preload("Materi.MataPelajaran").Preload("Materi.Tingkat").First(&soal, id).Error
	if err != nil {
		return nil, err
	}
	return &soal, nil
}

// Update existing soal
func (r *soalRepositoryImpl) Update(soal *entity.Soal) error {
	return r.db.Save(soal).Error
}

// Delete soal by ID
func (r *soalRepositoryImpl) Delete(id int) error {
	return r.db.Delete(&entity.Soal{}, id).Error
}

// List soal with filters
func (r *soalRepositoryImpl) List(idMateri, tingkatan, idMataPelajaran *int, limit, offset int) ([]entity.Soal, int, error) {
	var soals []entity.Soal
	var total int64

	query := r.db.Model(&entity.Soal{}).Preload("Materi").Preload("Materi.MataPelajaran").Preload("Materi.Tingkat")

	if idMateri != nil {
		query = query.Where("id_materi = ?", *idMateri)
	}
	if tingkatan != nil {
		query = query.Joins("JOIN materi ON soal.id_materi = materi.id").Where("materi.tingkatan = ?", *tingkatan)
	}
	if idMataPelajaran != nil {
		query = query.Joins("JOIN materi ON soal.id_materi = materi.id").
			Joins("JOIN mata_pelajaran ON materi.id_mata_pelajaran = mata_pelajaran.id").
			Where("mata_pelajaran.id = ?", *idMataPelajaran)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Limit(limit).Offset(offset).Find(&soals).Error; err != nil {
		return nil, 0, err
	}

	return soals, int(total), nil
}

// Get soal by materi ID
func (r *soalRepositoryImpl) GetByMateriID(idMateri int) ([]entity.Soal, error) {
	var soals []entity.Soal
	err := r.db.Preload("Materi").Preload("Materi.MataPelajaran").Preload("Materi.Tingkat").Where("id_materi = ?", idMateri).Find(&soals).Error
	return soals, err
}