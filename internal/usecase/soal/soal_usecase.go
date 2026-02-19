package soal

import (
	"bytes"
	"cbt-test-mini-project/init/config"
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/repository/test_soal"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// soalUsecaseImpl implements SoalUsecase
type soalUsecaseImpl struct {
	repo   test_soal.SoalRepository
	config *config.Main
}

// NewSoalUsecase creates a new SoalUsecase instance
func NewSoalUsecase(repo test_soal.SoalRepository, config *config.Main) SoalUsecase {
	return &soalUsecaseImpl{repo: repo, config: config}
}

func normalizeQuestionType(questionType entity.QuestionType, pembahasan string) entity.QuestionType {
	if questionType == entity.QuestionTypeMultipleChoice || questionType == entity.QuestionTypeEssay || questionType == entity.QuestionTypeMultipleChoicesComplex {
		return questionType
	}
	if strings.HasPrefix(strings.TrimSpace(pembahasan), "[ESSAY]") {
		return entity.QuestionTypeEssay
	}
	return entity.QuestionTypeMultipleChoice
}

func validateComplexOptions(options []entity.JawabanOption) error {
	if len(options) < 2 {
		return errors.New("complex multiple-choice requires at least 2 correct answers")
	}
	seen := map[entity.JawabanOption]struct{}{}
	for _, option := range options {
		if option < entity.JawabanA || option > entity.JawabanD {
			return errors.New("invalid complex answer option")
		}
		if _, exists := seen[option]; exists {
			return errors.New("duplicate complex answer option")
		}
		seen[option] = struct{}{}
	}
	return nil
}

// saveImages saves multiple image files and returns list of SoalGambar entities
func (u *soalUsecaseImpl) saveImages(imageFilesBytes [][]byte) ([]entity.SoalGambar, error) {
	var gambar []entity.SoalGambar
	
	if len(imageFilesBytes) == 0 {
		return gambar, nil
	}

	// Initialize Cloudinary
	cld, err := cloudinary.NewFromParams(u.config.Cloudinary.Name, u.config.Cloudinary.Key, u.config.Cloudinary.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %v", err)
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

		// Upload to Cloudinary
		resp, err := cld.Upload.Upload(context.Background(), bytes.NewReader(imageBytes), uploader.UploadParams{
			Folder: "cbt/soal_images",
			PublicID: fmt.Sprintf("%d_%d_%d", time.Now().Unix(), time.Now().Nanosecond(), i),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to upload image to Cloudinary: %v", err)
		}

		urutan := i + 1
		gambar = append(gambar, entity.SoalGambar{
			NamaFile: resp.PublicID,
			FilePath: resp.SecureURL,
			FileSize: len(imageBytes),
			MimeType: mimeType,
			Urutan:   urutan,
			CloudId:  &u.config.Cloudinary.Name,
			PublicId: &resp.PublicID,
		})
	}

	return gambar, nil
}

// CreateSoal creates a new soal with multiple images
func (u *soalUsecaseImpl) CreateSoal(idMateri, idTingkat int, pertanyaan, opsiA, opsiB, opsiC, opsiD, pembahasan string, questionType entity.QuestionType, jawabanBenar entity.JawabanOption, jawabanBenarComplex []entity.JawabanOption, imageFilesBytes [][]byte) (*entity.Soal, error) {
	questionType = normalizeQuestionType(questionType, pembahasan)
	if pertanyaan == "" {
		return nil, errors.New("pertanyaan must be filled")
	}
	if questionType != entity.QuestionTypeEssay {
		if opsiA == "" || opsiB == "" || opsiC == "" || opsiD == "" {
			return nil, errors.New("all fields must be filled")
		}
	}
	if questionType == entity.QuestionTypeMultipleChoice {
		if jawabanBenar < entity.JawabanA || jawabanBenar > entity.JawabanD {
			return nil, errors.New("invalid jawaban benar")
		}
	}
	if questionType == entity.QuestionTypeMultipleChoicesComplex {
		if err := validateComplexOptions(jawabanBenarComplex); err != nil {
			return nil, err
		}
	}

	gambar, err := u.saveImages(imageFilesBytes)
	if err != nil {
		return nil, err
	}

	s := &entity.Soal{
		IDMateri:     idMateri,
		IDTingkat:    idTingkat,
		Pertanyaan:   pertanyaan,
		QuestionType: questionType,
		OpsiA:        opsiA,
		OpsiB:        opsiB,
		OpsiC:        opsiC,
		OpsiD:        opsiD,
		JawabanBenar: jawabanBenar,
		Pembahasan:   &pembahasan,
		Gambar:       gambar,
	}
	if questionType == entity.QuestionTypeEssay {
		essayKey := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(pembahasan), "[ESSAY]"))
		if essayKey != "" {
			s.JawabanEssayKey = &essayKey
		}
		s.OpsiA = "-"
		s.OpsiB = "-"
		s.OpsiC = "-"
		s.OpsiD = "-"
		s.JawabanBenar = entity.JawabanA
	}
	if questionType == entity.QuestionTypeMultipleChoicesComplex {
		if err := s.SetJawabanBenarComplex(jawabanBenarComplex); err != nil {
			return nil, err
		}
		s.JawabanBenar = entity.JawabanA
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
func (u *soalUsecaseImpl) UpdateSoal(id, idMateri, idTingkat int, pertanyaan, opsiA, opsiB, opsiC, opsiD, pembahasan string, questionType entity.QuestionType, jawabanBenar entity.JawabanOption, jawabanBenarComplex []entity.JawabanOption, imageFilesBytes [][]byte) (*entity.Soal, error) {
	questionType = normalizeQuestionType(questionType, pembahasan)
	if pertanyaan == "" {
		return nil, errors.New("pertanyaan must be filled")
	}
	if questionType != entity.QuestionTypeEssay {
		if opsiA == "" || opsiB == "" || opsiC == "" || opsiD == "" {
			return nil, errors.New("all fields must be filled")
		}
	}
	if questionType == entity.QuestionTypeMultipleChoice {
		if jawabanBenar < entity.JawabanA || jawabanBenar > entity.JawabanD {
			return nil, errors.New("invalid jawaban benar")
		}
	}
	if questionType == entity.QuestionTypeMultipleChoicesComplex {
		if err := validateComplexOptions(jawabanBenarComplex); err != nil {
			return nil, err
		}
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
	s.Pembahasan = &pembahasan
	s.QuestionType = questionType
	switch questionType {
	case entity.QuestionTypeEssay:
		essayKey := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(pembahasan), "[ESSAY]"))
		if essayKey != "" {
			s.JawabanEssayKey = &essayKey
		} else {
			s.JawabanEssayKey = nil
		}
		s.OpsiA = "-"
		s.OpsiB = "-"
		s.OpsiC = "-"
		s.OpsiD = "-"
		s.JawabanBenar = entity.JawabanA
		s.JawabanBenarComplex = nil
	case entity.QuestionTypeMultipleChoicesComplex:
		if err := s.SetJawabanBenarComplex(jawabanBenarComplex); err != nil {
			return nil, err
		}
		s.JawabanEssayKey = nil
		s.JawabanBenar = entity.JawabanA
	default:
		s.JawabanEssayKey = nil
		s.JawabanBenarComplex = nil
	}
	err = u.repo.Update(s)
	if err != nil {
		return nil, err
	}
	return u.repo.GetByID(s.ID)
}

// DeleteSoal soft deletes by setting is_active = false
func (u *soalUsecaseImpl) DeleteSoal(id int) error {
	s, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	s.IsActive = false
	return u.repo.Update(s)
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

	// Initialize Cloudinary
	cld, err := cloudinary.NewFromParams(u.config.Cloudinary.Name, u.config.Cloudinary.Key, u.config.Cloudinary.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}

	// Upload to Cloudinary
	resp, err := cld.Upload.Upload(context.Background(), bytes.NewReader(imageBytes), uploader.UploadParams{
		Folder:   "cbt/soal_images",
		PublicID: namaFile,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload image to Cloudinary: %v", err)
	}

	gambar := &entity.SoalGambar{
		IDSoal:     idSoal,
		NamaFile:   resp.PublicID,
		FilePath:   resp.SecureURL,
		FileSize:   len(imageBytes),
		MimeType:   mimeType,
		Urutan:     urutan,
		Keterangan: keterangan,
		CloudId:    &u.config.Cloudinary.Name,
		PublicId:   &resp.PublicID,
	}

	err = u.repo.CreateGambar(gambar)
	if err != nil {
		// Delete from Cloudinary if DB insert fails
		_, delErr := cld.Upload.Destroy(context.Background(), uploader.DestroyParams{PublicID: resp.PublicID})
		if delErr != nil {
			fmt.Printf("Warning: failed to delete image from Cloudinary %s: %v\n", resp.PublicID, delErr)
		}
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

	// Initialize Cloudinary
	cld, err := cloudinary.NewFromParams(u.config.Cloudinary.Name, u.config.Cloudinary.Key, u.config.Cloudinary.Secret)
	if err != nil {
		return fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}

	// Delete from Cloudinary
	publicID := gambar.NamaFile // fallback to NamaFile for backward compatibility
	if gambar.PublicId != nil {
		publicID = *gambar.PublicId
	}
	_, err = cld.Upload.Destroy(context.Background(), uploader.DestroyParams{PublicID: publicID})
	if err != nil {
		// Log error but continue to delete DB record
		fmt.Printf("Warning: failed to delete image from Cloudinary %s: %v\n", publicID, err)
	}

	return u.repo.DeleteGambar(idGambar)
}

// UpdateImageInSoal updates an image in a soal
func (u *soalUsecaseImpl) UpdateImageInSoal(idGambar int, urutan int, keterangan *string) error {
	return u.repo.UpdateGambar(idGambar, urutan, keterangan)
}

// GetQuestionCountsByTopic returns the count of questions per topic
func (u *soalUsecaseImpl) GetQuestionCountsByTopic() (map[int]int, error) {
	return u.repo.GetQuestionCountsByTopic()
}