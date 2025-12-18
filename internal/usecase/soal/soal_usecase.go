package soal

import (
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/repository/test_soal"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// soalUsecaseImpl implements SoalUsecase
type soalUsecaseImpl struct {
	repo test_soal.SoalRepository
}

// NewSoalUsecase creates a new SoalUsecase instance
func NewSoalUsecase(repo test_soal.SoalRepository) SoalUsecase {
	return &soalUsecaseImpl{repo: repo}
}

// saveImages saves multiple image files and returns list of SoalGambar entities
func (u *soalUsecaseImpl) saveImages(imageFilesBytes [][]byte) ([]entity.SoalGambar, error) {
	var gambar []entity.SoalGambar
	
	if len(imageFilesBytes) == 0 {
		return gambar, nil
	}

	dir := "uploads/images"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}

	for i, imageBytes := range imageFilesBytes {
		if len(imageBytes) == 0 {
			continue
		}

		// Validate image type
		mimeType := http.DetectContentType(imageBytes)
		if mimeType != "image/jpeg" && mimeType != "image/png" {
			return nil, fmt.Errorf("invalid image type: %s, only JPG and PNG are allowed", mimeType)
		}

		// Determine extension
		var ext string
		if mimeType == "image/jpeg" {
			ext = ".jpg"
		} else {
			ext = ".png"
		}

		filename := fmt.Sprintf("%d_%d_%d%s", time.Now().Unix(), time.Now().Nanosecond(), i, ext)
		filePath := filepath.Join(dir, filename)

		if err := os.WriteFile(filePath, imageBytes, 0644); err != nil {
			return nil, err
		}

		urutan := i + 1
		gambar = append(gambar, entity.SoalGambar{
			NamaFile: filename,
			FilePath: filePath,
			FileSize: len(imageBytes),
			MimeType: mimeType,
			Urutan:   urutan,
		})
	}

	return gambar, nil
}

// CreateSoal creates a new soal with multiple images
func (u *soalUsecaseImpl) CreateSoal(idMateri, idTingkat int, pertanyaan, opsiA, opsiB, opsiC, opsiD string, jawabanBenar entity.JawabanOption, imageFilesBytes [][]byte) (*entity.Soal, error) {
	if pertanyaan == "" || opsiA == "" || opsiB == "" || opsiC == "" || opsiD == "" {
		return nil, errors.New("all fields must be filled")
	}
	if jawabanBenar < entity.JawabanA || jawabanBenar > entity.JawabanD {
		return nil, errors.New("invalid jawaban benar")
	}

	gambar, err := u.saveImages(imageFilesBytes)
	if err != nil {
		return nil, err
	}

	s := &entity.Soal{
		IDMateri:     idMateri,
		IDTingkat:    idTingkat,
		Pertanyaan:   pertanyaan,
		OpsiA:        opsiA,
		OpsiB:        opsiB,
		OpsiC:        opsiC,
		OpsiD:        opsiD,
		JawabanBenar: jawabanBenar,
		Gambar:       gambar,
	}
	err = u.repo.Create(s)
	if err != nil {
		return nil, err
	}
	return u.repo.GetByID(s.ID)
}

// GetSoal gets by ID
func (u *soalUsecaseImpl) GetSoal(id int) (*entity.Soal, error) {
	return u.repo.GetByID(id)
}

// UpdateSoal updates existing with multiple images
func (u *soalUsecaseImpl) UpdateSoal(id, idMateri, idTingkat int, pertanyaan, opsiA, opsiB, opsiC, opsiD string, jawabanBenar entity.JawabanOption, imageFilesBytes [][]byte) (*entity.Soal, error) {
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

	gambar, err := u.saveImages(imageFilesBytes)
	if err != nil {
		return nil, err
	}
	if len(gambar) > 0 {
		s.Gambar = gambar
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

// UploadImageToSoal uploads an image to a soal
func (u *soalUsecaseImpl) UploadImageToSoal(idSoal int, imageBytes []byte, namaFile string, urutan int, keterangan *string) (*entity.SoalGambar, error) {
	if len(imageBytes) == 0 {
		return nil, errors.New("image bytes cannot be empty")
	}

	// Validate image type
	mimeType := http.DetectContentType(imageBytes)
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		return nil, fmt.Errorf("invalid image type: %s, only JPG and PNG are allowed", mimeType)
	}

	dir := "uploads/images"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}

	// Determine extension
	var ext string
	if mimeType == "image/jpeg" {
		ext = ".jpg"
	} else {
		ext = ".png"
	}

	// Generate filename if not provided or ensure it has correct extension
	if namaFile == "" {
		namaFile = fmt.Sprintf("%d_%d%s", time.Now().Unix(), time.Now().Nanosecond(), ext)
	} else {
		// Ensure extension matches
		if !strings.HasSuffix(strings.ToLower(namaFile), ext) {
			namaFile = strings.TrimSuffix(namaFile, filepath.Ext(namaFile)) + ext
		}
	}

	filePath := filepath.Join(dir, namaFile)

	if err := os.WriteFile(filePath, imageBytes, 0644); err != nil {
		return nil, err
	}

	gambar := &entity.SoalGambar{
		IDSoal:     idSoal,
		NamaFile:   namaFile,
		FilePath:   filePath,
		FileSize:   len(imageBytes),
		MimeType:   mimeType,
		Urutan:     urutan,
		Keterangan: keterangan,
	}

	err := u.repo.CreateGambar(gambar)
	if err != nil {
		// Clean up file if DB insert fails
		os.Remove(filePath)
		return nil, err
	}

	return gambar, nil
}

// DeleteImageFromSoal deletes an image from a soal
func (u *soalUsecaseImpl) DeleteImageFromSoal(idGambar int) error {
	gambar, err := u.repo.GetGambarByID(idGambar)
	if err != nil {
		return err
	}

	// Delete file
	if err := os.Remove(gambar.FilePath); err != nil {
		// Log error but continue to delete DB record
		fmt.Printf("Warning: failed to delete file %s: %v\n", gambar.FilePath, err)
	}

	return u.repo.DeleteGambar(idGambar)
}

// UpdateImageInSoal updates an image in a soal
func (u *soalUsecaseImpl) UpdateImageInSoal(idGambar int, urutan int, keterangan *string) error {
	return u.repo.UpdateGambar(idGambar, urutan, keterangan)
}