package materi

import (
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/repository/materi"
	"errors"
)

// materiUsecaseImpl implements MateriUsecase
type materiUsecaseImpl struct {
	repo materi.MateriRepository
}

// NewMateriUsecase creates a new MateriUsecase instance
func NewMateriUsecase(repo materi.MateriRepository) MateriUsecase {
	return &materiUsecaseImpl{repo: repo}
}

// CreateMateri creates a new materi
func (u *materiUsecaseImpl) CreateMateri(idMataPelajaran int, nama string, idTingkat int, isActive bool, defaultDurasiMenit, defaultJumlahSoal int) (*entity.Materi, error) {
	if nama == "" {
		return nil, errors.New("nama cannot be empty")
	}
	if idTingkat < 1 {
		return nil, errors.New("idTingkat must be positive")
	}
	if defaultDurasiMenit < 1 {
		defaultDurasiMenit = 60
	}
	if defaultJumlahSoal < 1 {
		defaultJumlahSoal = 20
	}

	m := &entity.Materi{
		IDMataPelajaran:    idMataPelajaran,
		IDTingkat:          idTingkat,
		Nama:               nama,
		IsActive:           isActive,
		DefaultDurasiMenit: defaultDurasiMenit,
		DefaultJumlahSoal:  defaultJumlahSoal,
	}
	err := u.repo.Create(m)
	if err != nil {
		return nil, err
	}
	return u.repo.GetByID(m.ID) // To preload
}

// GetMateri gets by ID
func (u *materiUsecaseImpl) GetMateri(id int) (*entity.Materi, error) {
	return u.repo.GetByID(id)
}

// UpdateMateri updates existing
func (u *materiUsecaseImpl) UpdateMateri(id, idMataPelajaran int, nama string, idTingkat int, isActive bool, defaultDurasiMenit, defaultJumlahSoal int) (*entity.Materi, error) {
	if nama == "" {
		return nil, errors.New("nama cannot be empty")
	}
	if idTingkat < 1 {
		return nil, errors.New("idTingkat must be positive")
	}
	if defaultDurasiMenit < 1 {
		defaultDurasiMenit = 60
	}
	if defaultJumlahSoal < 1 {
		defaultJumlahSoal = 20
	}

	m, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	m.IDMataPelajaran = idMataPelajaran
	m.IDTingkat = idTingkat
	m.Nama = nama
	m.IsActive = isActive
	m.DefaultDurasiMenit = defaultDurasiMenit
	m.DefaultJumlahSoal = defaultJumlahSoal
	err = u.repo.Update(m)
	if err != nil {
		return nil, err
	}
	return u.repo.GetByID(m.ID)
}

// DeleteMateri soft deletes by setting is_active = false
func (u *materiUsecaseImpl) DeleteMateri(id int) error {
	m, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	m.IsActive = false
	return u.repo.Update(m)
}

// ListMateri lists with filters and pagination
func (u *materiUsecaseImpl) ListMateri(idMataPelajaran, idTingkat int, page, pageSize int) ([]entity.Materi, *entity.PaginationResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 100 // Default to larger page size for dynamic loading
	}
	// Allow large page sizes for virtual scrolling (cap at 1000 to prevent abuse)
	if pageSize > 1000 {
		pageSize = 1000
	}

	offset := (page - 1) * pageSize
	var idPtr, tingPtr *int
	if idMataPelajaran > 0 {
		idPtr = &idMataPelajaran
	}
	if idTingkat > 0 {
		tingPtr = &idTingkat
	}
	materis, total, err := u.repo.List(idPtr, tingPtr, pageSize, offset)
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

	return materis, pagination, nil
}