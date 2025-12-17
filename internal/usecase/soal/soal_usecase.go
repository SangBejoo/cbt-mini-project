package soal

import (
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/repository/test_soal"
	"errors"
)

// soalUsecaseImpl implements SoalUsecase
type soalUsecaseImpl struct {
	repo test_soal.SoalRepository
}

// NewSoalUsecase creates a new SoalUsecase instance
func NewSoalUsecase(repo test_soal.SoalRepository) SoalUsecase {
	return &soalUsecaseImpl{repo: repo}
}

// CreateSoal creates a new soal
func (u *soalUsecaseImpl) CreateSoal(idMateri, idTingkat int, pertanyaan, opsiA, opsiB, opsiC, opsiD string, jawabanBenar entity.JawabanOption) (*entity.Soal, error) {
	if pertanyaan == "" || opsiA == "" || opsiB == "" || opsiC == "" || opsiD == "" {
		return nil, errors.New("all fields must be filled")
	}
	if jawabanBenar < entity.JawabanA || jawabanBenar > entity.JawabanD {
		return nil, errors.New("invalid jawaban benar")
	}

	s := &entity.Soal{
		IDMateri:   idMateri,
		IDTingkat:  idTingkat,
		Pertanyaan: pertanyaan,
		OpsiA:      opsiA,
		OpsiB:      opsiB,
		OpsiC:      opsiC,
		OpsiD:      opsiD,
		JawabanBenar: jawabanBenar,
	}
	err := u.repo.Create(s)
	if err != nil {
		return nil, err
	}
	return u.repo.GetByID(s.ID)
}

// GetSoal gets by ID
func (u *soalUsecaseImpl) GetSoal(id int) (*entity.Soal, error) {
	return u.repo.GetByID(id)
}

// UpdateSoal updates existing
func (u *soalUsecaseImpl) UpdateSoal(id, idMateri, idTingkat int, pertanyaan, opsiA, opsiB, opsiC, opsiD string, jawabanBenar entity.JawabanOption) (*entity.Soal, error) {
	if pertanyaan == "" || opsiA == "" || opsiB == "" || opsiC == "" || opsiD == "" {
		return nil, errors.New("all fields must be filled")
	}
	if jawabanBenar < entity.JawabanA || jawabanBenar > entity.JawabanD {
		return nil, errors.New("invalid jawaban benar")
	}

	s, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	s.IDMateri = idMateri
	s.IDTingkat = idTingkat
	s.Pertanyaan = pertanyaan
	s.OpsiA = opsiA
	s.OpsiB = opsiB
	s.OpsiC = opsiC
	s.OpsiD = opsiD
	s.JawabanBenar = jawabanBenar
	err = u.repo.Update(s)
	if err != nil {
		return nil, err
	}
	return u.repo.GetByID(s.ID)
}

// DeleteSoal deletes by ID
func (u *soalUsecaseImpl) DeleteSoal(id int) error {
	_, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	return u.repo.Delete(id)
}

// ListSoal lists with filters and pagination
func (u *soalUsecaseImpl) ListSoal(idMateri, tingkatan, idMataPelajaran int, page, pageSize int) ([]entity.Soal, *entity.PaginationResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	var idMateriPtr, tingPtr, idMataPtr *int
	if idMateri > 0 {
		idMateriPtr = &idMateri
	}
	if tingkatan > 0 {
		tingPtr = &tingkatan
	}
	if idMataPelajaran > 0 {
		idMataPtr = &idMataPelajaran
	}
	soals, total, err := u.repo.List(idMateriPtr, tingPtr, idMataPtr, pageSize, offset)
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

	return soals, pagination, nil
}