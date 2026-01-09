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
	err := r.db.Preload("Materi").Preload("Materi.MataPelajaran").Preload("Materi.Tingkat").Preload("Gambar", func(db *gorm.DB) *gorm.DB { return db.Order("urutan ASC") }).First(&soal, id).Error
	if err != nil {
		return nil, err
	}
	return &soal, nil
}

// Update existing soal
func (r *soalRepositoryImpl) Update(soal *entity.Soal) error {
	return r.db.Save(soal).Error
}

// Delete soal by ID (soft delete)
func (r *soalRepositoryImpl) Delete(id int) error {
	return r.db.Model(&entity.Soal{}).Where("id = ?", id).Update("is_active", false).Error
}

// List soal with filters
func (r *soalRepositoryImpl) List(idMateri, tingkatan, idMataPelajaran *int, limit, offset int) ([]entity.Soal, int, error) {
	var soals []entity.Soal
	var total int64

	// Build base query for count (without preloads for performance)
	countQuery := r.db.Model(&entity.Soal{})

	if idMateri != nil {
		countQuery = countQuery.Where("id_materi = ?", *idMateri)
	}
	if tingkatan != nil {
		countQuery = countQuery.Joins("JOIN materi ON soal.id_materi = materi.id").Where("materi.tingkatan = ?", *tingkatan)
	}
	if idMataPelajaran != nil {
		countQuery = countQuery.Joins("JOIN materi ON soal.id_materi = materi.id").
			Joins("JOIN mata_pelajaran ON materi.id_mata_pelajaran = mata_pelajaran.id").
			Where("mata_pelajaran.id = ?", *idMataPelajaran)
	}

	// Count total
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Build query for data with preloads
	query := r.db.Model(&entity.Soal{}).Preload("Materi").Preload("Materi.MataPelajaran").Preload("Materi.Tingkat").Preload("Gambar", func(db *gorm.DB) *gorm.DB { return db.Order("urutan ASC") })

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

	// Get paginated results
	if err := query.Limit(limit).Offset(offset).Find(&soals).Error; err != nil {
		return nil, 0, err
	}

	return soals, int(total), nil
}

// Get soal by materi ID
func (r *soalRepositoryImpl) GetByMateriID(idMateri int) ([]entity.Soal, error) {
	var soals []entity.Soal
	err := r.db.Preload("Materi").Preload("Materi.MataPelajaran").Preload("Materi.Tingkat").Preload("Gambar", func(db *gorm.DB) *gorm.DB { return db.Order("urutan ASC") }).Where("id_materi = ?", idMateri).Find(&soals).Error
	return soals, err
}

// CreateGambar creates a new soal gambar
func (r *soalRepositoryImpl) CreateGambar(gambar *entity.SoalGambar) error {
	return r.db.Create(gambar).Error
}

// GetGambarByID gets gambar by ID
func (r *soalRepositoryImpl) GetGambarByID(id int) (*entity.SoalGambar, error) {
	var gambar entity.SoalGambar
	err := r.db.First(&gambar, id).Error
	if err != nil {
		return nil, err
	}
	return &gambar, nil
}

// UpdateGambar updates gambar urutan and keterangan
func (r *soalRepositoryImpl) UpdateGambar(id int, urutan int, keterangan *string) error {
	return r.db.Model(&entity.SoalGambar{}).Where("id = ?", id).Updates(map[string]interface{}{
		"urutan":     urutan,
		"keterangan": keterangan,
	}).Error
}

// DeleteGambar deletes gambar by ID
func (r *soalRepositoryImpl) DeleteGambar(id int) error {
	return r.db.Delete(&entity.SoalGambar{}, id).Error
}

// GetQuestionCountsByTopic returns the count of questions per topic (both MC and drag-drop)
func (r *soalRepositoryImpl) GetQuestionCountsByTopic() (map[int]int, error) {
	counts := make(map[int]int)
	
	// Count multiple-choice questions
	var mcResults []struct {
		IdMateri int
		Count    int
	}
	err := r.db.Model(&entity.Soal{}).
		Select("id_materi, count(*) as count").
		Where("is_active = ?", true).
		Group("id_materi").
		Scan(&mcResults).Error
	if err != nil {
		return nil, err
	}
	for _, result := range mcResults {
		counts[result.IdMateri] = result.Count
	}
	
	// Count drag-drop questions
	var ddResults []struct {
		IdMateri int
		Count    int
	}
	err = r.db.Table("soal_drag_drop").
		Select("id_materi, count(*) as count").
		Where("is_active = ?", true).
		Group("id_materi").
		Scan(&ddResults).Error
	if err != nil {
		return nil, err
	}
	for _, result := range ddResults {
		counts[result.IdMateri] += result.Count // Add to existing count
	}
	
	return counts, nil
}