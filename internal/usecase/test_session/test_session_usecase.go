package test_session

import (
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/repository/test_session"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

// testSessionUsecaseImpl implements TestSessionUsecase
type testSessionUsecaseImpl struct {
	repo test_session.TestSessionRepository
}

// NewTestSessionUsecase creates a new TestSessionUsecase instance
func NewTestSessionUsecase(repo test_session.TestSessionRepository) TestSessionUsecase {
	return &testSessionUsecaseImpl{repo: repo}
}

// CreateTestSession creates a new test session with random questions
func (u *testSessionUsecaseImpl) CreateTestSession(namaPeserta string, tingkatan, idMataPelajaran, durasiMenit, jumlahSoal int) (*entity.TestSession, error) {
	if namaPeserta == "" || tingkatan < 1 || idMataPelajaran < 1 || durasiMenit < 1 || jumlahSoal < 1 {
		return nil, errors.New("invalid input parameters")
	}

	// Generate unique session token
	token, err := u.generateToken()
	if err != nil {
		return nil, err
	}

	session := &entity.TestSession{
		SessionToken:    token,
		NamaPeserta:     namaPeserta,
		IDTingkat:       tingkatan,
		IDMataPelajaran: idMataPelajaran,
		DurasiMenit:     durasiMenit,
		Status:          entity.TestStatusOngoing,
	}

	err = u.repo.Create(session)
	if err != nil {
		return nil, err
	}

	// Select and assign random questions
	err = u.repo.AssignRandomQuestions(session.ID, idMataPelajaran, tingkatan, jumlahSoal)
	if err != nil {
		// Cleanup session if question assignment fails
		u.repo.Delete(session.ID)
		return nil, err
	}

	return u.repo.GetByToken(token)
}

// GetTestSession gets by token
func (u *testSessionUsecaseImpl) GetTestSession(sessionToken string) (*entity.TestSession, error) {
	session, err := u.repo.GetByToken(sessionToken)
	if err != nil {
		return nil, err
	}

	// Check if session is still active
	if session.Status != entity.TestStatusOngoing {
		return nil, errors.New("session is not active")
	}

	// Check timeout
	if time.Now().After(session.BatasWaktu()) {
		u.repo.CompleteSession(sessionToken, time.Now(), nil, nil, nil)
		session.Status = entity.TestStatusTimeout
	}

	return session, nil
}

// GetTestQuestions gets a single question for the student
func (u *testSessionUsecaseImpl) GetTestQuestions(sessionToken string, nomorUrut int) (*entity.SoalForStudent, error) {
	_, err := u.GetTestSession(sessionToken)
	if err != nil {
		return nil, err
	}

	soal, err := u.repo.GetQuestionByOrder(sessionToken, nomorUrut)
	if err != nil {
		return nil, err
	}

	// Get existing answer if any
	answers, _ := u.repo.GetSessionAnswers(sessionToken)
	var jawabanDipilih *entity.JawabanOption
	isAnswered := false
	for _, ans := range answers {
		if ans.TestSessionSoal.NomorUrut == nomorUrut {
			jawabanDipilih = ans.JawabanDipilih
			isAnswered = true
			break
		}
	}

	soalForStudent := &entity.SoalForStudent{
		ID:             soal.ID,
		NomorUrut:      nomorUrut,
		Pertanyaan:     soal.Pertanyaan,
		OpsiA:          soal.OpsiA,
		OpsiB:          soal.OpsiB,
		OpsiC:          soal.OpsiC,
		OpsiD:          soal.OpsiD,
		JawabanDipilih: jawabanDipilih,
		IsAnswered:     isAnswered,
		Materi:         soal.Materi,
		Gambar:         soal.Gambar,
	}

	return soalForStudent, nil
}

// GetAllTestQuestions gets all questions for the session
func (u *testSessionUsecaseImpl) GetAllTestQuestions(sessionToken string) ([]entity.SoalForStudent, error) {
	_, err := u.GetTestSession(sessionToken)
	if err != nil {
		return nil, err
	}

	soals, err := u.repo.GetAllQuestionsForSession(sessionToken)
	if err != nil {
		return nil, err
	}

	// Get existing answers
	answers, _ := u.repo.GetSessionAnswers(sessionToken)

	var soalForStudents []entity.SoalForStudent
	for _, soal := range soals {
		var jawabanDipilih *entity.JawabanOption
		isAnswered := false
		for _, ans := range answers {
			if ans.TestSessionSoal.NomorUrut == soal.NomorUrut {
				jawabanDipilih = ans.JawabanDipilih
				isAnswered = true
				break
			}
		}

		soalForStudents = append(soalForStudents, entity.SoalForStudent{
			ID:             soal.Soal.ID,
			NomorUrut:      soal.NomorUrut,
			Pertanyaan:     soal.Soal.Pertanyaan,
			OpsiA:          soal.Soal.OpsiA,
			OpsiB:          soal.Soal.OpsiB,
			OpsiC:          soal.Soal.OpsiC,
			OpsiD:          soal.Soal.OpsiD,
			JawabanDipilih: jawabanDipilih,
			IsAnswered:     isAnswered,
			Materi:         soal.Soal.Materi,
			Gambar:         soal.Soal.Gambar,
		})
	}

	return soalForStudents, nil
}

// GetSessionAnswers gets all answers for the session
func (u *testSessionUsecaseImpl) GetSessionAnswers(sessionToken string) ([]entity.JawabanSiswa, error) {
	_, err := u.GetTestSession(sessionToken)
	if err != nil {
		return nil, err
	}

	return u.repo.GetSessionAnswers(sessionToken)
}

// SubmitAnswer submits or updates an answer
func (u *testSessionUsecaseImpl) SubmitAnswer(sessionToken string, nomorUrut int, jawaban entity.JawabanOption) error {
	_, err := u.GetTestSession(sessionToken)
	if err != nil {
		return err
	}

	return u.repo.SubmitAnswer(sessionToken, nomorUrut, jawaban)
}

// ClearAnswer clears an answer
func (u *testSessionUsecaseImpl) ClearAnswer(sessionToken string, nomorUrut int) error {
	_, err := u.GetTestSession(sessionToken)
	if err != nil {
		return err
	}

	return u.repo.ClearAnswer(sessionToken, nomorUrut)
}

// CompleteSession completes the session and calculates score
func (u *testSessionUsecaseImpl) CompleteSession(sessionToken string) (*entity.TestSession, error) {
	session, err := u.repo.GetByToken(sessionToken)
	if err != nil {
		return nil, err
	}

	if session.Status != entity.TestStatusOngoing {
		return nil, errors.New("session is already completed")
	}

	answers, err := u.repo.GetSessionAnswers(sessionToken)
	if err != nil {
		return nil, err
	}

	jumlahBenar := 0
	for _, ans := range answers {
		if ans.IsCorrect {
			jumlahBenar++
		}
	}

	nilaiAkhir := float64(jumlahBenar) / float64(len(answers)) * 100
	totalSoal := len(answers)

	err = u.repo.CompleteSession(sessionToken, time.Now(), &nilaiAkhir, &jumlahBenar, &totalSoal)
	if err != nil {
		return nil, err
	}

	return u.repo.GetByToken(sessionToken)
}

// GetTestResult gets the test result
func (u *testSessionUsecaseImpl) GetTestResult(sessionToken string) (*entity.TestSession, []entity.JawabanDetail, error) {
	session, err := u.repo.GetByToken(sessionToken)
	if err != nil {
		return nil, nil, err
	}

	if session.Status == entity.TestStatusOngoing {
		return nil, nil, errors.New("session is not completed")
	}

	// Get detailed answers
	var details []entity.JawabanDetail
	answers, err := u.repo.GetSessionAnswers(sessionToken)
	if err != nil {
		return nil, nil, err
	}

	for _, ans := range answers {
		detail := entity.JawabanDetail{
			NomorUrut:      ans.TestSessionSoal.NomorUrut,
			Pertanyaan:     ans.TestSessionSoal.Soal.Pertanyaan,
			OpsiA:          ans.TestSessionSoal.Soal.OpsiA,
			OpsiB:          ans.TestSessionSoal.Soal.OpsiB,
			OpsiC:          ans.TestSessionSoal.Soal.OpsiC,
			OpsiD:          ans.TestSessionSoal.Soal.OpsiD,
			JawabanDipilih: ans.JawabanDipilih,
			JawabanBenar:   ans.TestSessionSoal.Soal.JawabanBenar,
			IsCorrect:      ans.IsCorrect,
		}
		details = append(details, detail)
	}

	return session, details, nil
}

// ListTestSessions lists sessions for admin
func (u *testSessionUsecaseImpl) ListTestSessions(tingkatan, idMataPelajaran *int, status *entity.TestStatus, page, pageSize int) ([]entity.TestSession, *entity.PaginationResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	sessions, total, err := u.repo.List(tingkatan, idMataPelajaran, status, pageSize, offset)
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

	return sessions, pagination, nil
}

// Helper functions

func (u *testSessionUsecaseImpl) generateToken() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}