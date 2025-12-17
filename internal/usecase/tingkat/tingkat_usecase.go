package tingkat

import (
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/repository/tingkat"
	"errors"
)

// tingkatUsecaseImpl implements TingkatUsecase
type tingkatUsecaseImpl struct {
	repo tingkat.TingkatRepository
}

// NewTingkatUsecase creates a new TingkatUsecase instance
func NewTingkatUsecase(repo tingkat.TingkatRepository) TingkatUsecase {
	return &tingkatUsecaseImpl{repo: repo}
}

// CreateTingkat creates a new tingkat
func (u *tingkatUsecaseImpl) CreateTingkat(nama string) (*entity.Tingkat, error) {
	if nama == "" {
		return nil, errors.New("nama cannot be empty")
	}

	tingkat := &entity.Tingkat{
		Nama: nama,
	}

	if err := u.repo.Create(tingkat); err != nil {
		return nil, err
	}

	return tingkat, nil
}

// GetTingkat gets a tingkat by ID
func (u *tingkatUsecaseImpl) GetTingkat(id int) (*entity.Tingkat, error) {
	if id <= 0 {
		return nil, errors.New("invalid ID")
	}

	return u.repo.GetByID(id)
}

// UpdateTingkat updates a tingkat
func (u *tingkatUsecaseImpl) UpdateTingkat(id int, nama string) (*entity.Tingkat, error) {
	if id <= 0 || nama == "" {
		return nil, errors.New("invalid ID or nama cannot be empty")
	}

	tingkat, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	tingkat.Nama = nama

	if err := u.repo.Update(tingkat); err != nil {
		return nil, err
	}

	return tingkat, nil
}

// DeleteTingkat deletes a tingkat
func (u *tingkatUsecaseImpl) DeleteTingkat(id int) error {
	if id <= 0 {
		return errors.New("invalid ID")
	}

	return u.repo.Delete(id)
}

// ListTingkat lists tingkat with pagination
func (u *tingkatUsecaseImpl) ListTingkat(page, pageSize int) ([]entity.Tingkat, *entity.PaginationResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	tingkats, total, err := u.repo.List(pageSize, offset)
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

	return tingkats, pagination, nil
}