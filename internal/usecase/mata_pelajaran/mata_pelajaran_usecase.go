package mata_pelajaran

import (
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/repository/mata_pelajaran"
	"errors"
)

// mataPelajaranUsecaseImpl implements MataPelajaranUsecase
type mataPelajaranUsecaseImpl struct {
	repo mata_pelajaran.MataPelajaranRepository
}

// NewMataPelajaranUsecase creates a new MataPelajaranUsecase instance
func NewMataPelajaranUsecase(repo mata_pelajaran.MataPelajaranRepository) MataPelajaranUsecase {
	return &mataPelajaranUsecaseImpl{repo: repo}
}

// CreateMataPelajaran creates a new mata pelajaran
func (u *mataPelajaranUsecaseImpl) CreateMataPelajaran(nama string) (*entity.MataPelajaran, error) {
	if nama == "" {
		return nil, errors.New("nama cannot be empty")
	}

	// Check if already exists
	existing, _ := u.repo.GetByName(nama)
	if existing != nil {
		return nil, errors.New("mata pelajaran with this name already exists")
	}

	mp := &entity.MataPelajaran{Nama: nama}
	err := u.repo.Create(mp)
	if err != nil {
		return nil, err
	}
	return mp, nil
}

// GetMataPelajaran gets by ID
func (u *mataPelajaranUsecaseImpl) GetMataPelajaran(id int) (*entity.MataPelajaran, error) {
	return u.repo.GetByID(id)
}

// UpdateMataPelajaran updates existing
func (u *mataPelajaranUsecaseImpl) UpdateMataPelajaran(id int, nama string) (*entity.MataPelajaran, error) {
	if nama == "" {
		return nil, errors.New("nama cannot be empty")
	}

	mp, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	mp.Nama = nama
	err = u.repo.Update(mp)
	if err != nil {
		return nil, err
	}
	return mp, nil
}

// DeleteMataPelajaran deletes by ID
func (u *mataPelajaranUsecaseImpl) DeleteMataPelajaran(id int) error {
	_, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	return u.repo.Delete(id)
}

// ListMataPelajaran lists with pagination
func (u *mataPelajaranUsecaseImpl) ListMataPelajaran(page, pageSize int) ([]entity.MataPelajaran, *entity.PaginationResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	mps, total, err := u.repo.List(pageSize, offset)
	if err != nil {
		return nil, nil, err
	}

	totalPages := (total + pageSize - 1) / pageSize
	pagination := &entity.PaginationResponse{
		TotalCount:  total,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
	}

	return mps, pagination, nil
}