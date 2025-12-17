package sesi_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/repository/jawaban_siswa"
	"cbt-test-mini-project/internal/repository/mata_pelajaran"
	"cbt-test-mini-project/internal/repository/sesi_test"
	"cbt-test-mini-project/internal/repository/soal"
)

// TestSessionUsecase defines business logic for test sessions
type TestSessionUsecase interface {
	// CreateTestSession starts a new test session for student
	CreateTestSession(ctx context.Context, namaPeserta string, tingkatan, idMataPelajaran, durasiMenit int, soalIDs []int) (*entity.TestSession, error)

	// GetTestSession retrieves session details
	GetTestSession(ctx context.Context, sessionToken string) (*entity.TestSession, error)

	// GetTestQuestion retrieves single question for student with navigation info
	GetTestQuestion(ctx context.Context, sessionToken string, nomorUrut int) (*TestQuestionDTO, error)

	// SubmitAnswer saves/updates student answer and validates it
	SubmitAnswer(ctx context.Context, sessionToken string, nomorUrut int, jawabanDipilih entity.JawabanOption) (*AnswerSubmitDTO, error)

	// CompleteTestSession marks test as completed and calculates score
	CompleteTestSession(ctx context.Context, sessionToken string) (*TestResultDTO, error)

	// GetTestResult retrieves completed test results with all answers
	GetTestResult(ctx context.Context, sessionToken string) (*TestResultDTO, error)

	// ListStudentHistory retrieves all test sessions for a student
	ListStudentHistory(ctx context.Context, namaPeserta string) ([]*entity.HistorySummary, error)

	// CheckSessionTimeout checks if session exceeded time limit
	CheckSessionTimeout(ctx context.Context, sessionToken string) (bool, error)

	// CancelTestSession cancels an ongoing test
	CancelTestSession(ctx context.Context, sessionToken string) error
}

type testSessionService struct {
	sessionRepo       sesi_test.TestSessionRepository
	sessionSoalRepo   sesi_test.TestSessionSoalRepository
	jawabanRepo       jawaban_siswa.JawabanSiswaRepository
	mataPelajaranRepo mata_pelajaran.MataPelajaranRepository
	soalRepo          soal.SoalRepository
}

// NewTestSessionUsecase creates instance of test session usecase
func NewTestSessionUsecase(
	sessionRepo sesi_test.TestSessionRepository,
	sessionSoalRepo sesi_test.TestSessionSoalRepository,
	jawabanRepo jawaban_siswa.JawabanSiswaRepository,
	mataPelajaranRepo mata_pelajaran.MataPelajaranRepository,
	soalRepo soal.SoalRepository,
) TestSessionUsecase {
	return &testSessionService{
		sessionRepo:       sessionRepo,
		sessionSoalRepo:   sessionSoalRepo,
		jawabanRepo:       jawabanRepo,
		mataPelajaranRepo: mataPelajaranRepo,
		soalRepo:          soalRepo,
	}
}

// DTOs for API responses
type TestQuestionDTO struct {
	SessionToken       string
	Soal               *entity.SoalForStudent
	TotalSoal          int
	CurrentNomorUrut   int
	DijawabCount       int
	IsAnsweredStatus   []bool
	BatasWaktu         time.Time
	TimeRemainingDetik int
}

type AnswerSubmitDTO struct {
	SessionToken   string
	NomorUrut      int
	JawabanDipilih entity.JawabanOption
	IsCorrect      bool
	DijawabPada    time.Time
}

type TestResultDTO struct {
	SessionToken          string
	NamaPeserta           string
	MataPelajaran         *entity.MataPelajaran
	Tingkatan             int
	WaktuMulai            time.Time
	WaktuSelesai          *time.Time
	DurasiPengerjaanDetik int
	NilaiAkhir            float64
	JumlahBenar           int
	TotalSoal             int
	PersentaseBenar       float64
	Status                entity.TestStatus
	JawabanDetails        []*entity.JawabanDetail
}

// generateSessionToken creates unique session token
func (s *testSessionService) generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *testSessionService) CreateTestSession(ctx context.Context, namaPeserta string, tingkatan, idMataPelajaran, durasiMenit int, soalIDs []int) (*entity.TestSession, error) {
	// Validation
	if namaPeserta == "" {
		return nil, errors.New("INVALID_INPUT: nama_peserta cannot be empty")
	}
	if tingkatan <= 0 || tingkatan > 6 {
		return nil, errors.New("INVALID_INPUT: tingkatan must be between 1 and 6")
	}
	if idMataPelajaran <= 0 {
		return nil, errors.New("INVALID_INPUT: id_mata_pelajaran must be greater than 0")
	}
	if durasiMenit <= 0 || durasiMenit > 480 {
		return nil, errors.New("INVALID_INPUT: durasi_menit must be between 1 and 480")
	}
	if len(soalIDs) == 0 {
		return nil, errors.New("INVALID_INPUT: at least one soal must be selected")
	}

	// Check if subject exists
	mp, err := s.mataPelajaranRepo.GetByID(ctx, idMataPelajaran)
	if err != nil || mp == nil {
		return nil, errors.New("NOT_FOUND: mata_pelajaran not found")
	}

	// Generate unique session token
	sessionToken, err := s.generateSessionToken()
	if err != nil {
		return nil, errors.New("INTERNAL_ERROR: failed to generate session token")
	}

	// Create session
	session := &entity.TestSession{
		SessionToken:    sessionToken,
		NamaPeserta:     namaPeserta,
		Tingkatan:       tingkatan,
		IDMataPelajaran: idMataPelajaran,
		DurasiMenit:     durasiMenit,
		Status:          entity.TestStatusOngoing,
		TotalSoal:       intPtr(len(soalIDs)),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, errors.New("DATABASE_ERROR: failed to create test session")
	}

	// Add soal to session
	for i, idSoal := range soalIDs {
		// Validate soal exists
		soalEntity, err := s.soalRepo.GetByID(ctx, idSoal)
		if err != nil || soalEntity == nil {
			return nil, fmt.Errorf("NOT_FOUND: soal with ID %d not found", idSoal)
		}

		sessionSoal := &entity.TestSessionSoal{
			IDTestSession: session.ID,
			IDSoal:        idSoal,
			NomorUrut:     i + 1,
		}

		if err := s.sessionSoalRepo.Create(ctx, sessionSoal); err != nil {
			return nil, errors.New("DATABASE_ERROR: failed to add soal to session")
		}

		// Initialize empty answer record
		jawaban := &entity.JawabanSiswa{
			IDTestSessionSoal: sessionSoal.ID,
			IsCorrect:         false,
		}
		if err := s.jawabanRepo.CreateOrUpdate(ctx, jawaban); err != nil {
			return nil, errors.New("DATABASE_ERROR: failed to initialize answer record")
		}
	}

	// Reload session with relations
	return s.GetTestSession(ctx, sessionToken)
}

func (s *testSessionService) GetTestSession(ctx context.Context, sessionToken string) (*entity.TestSession, error) {
	if sessionToken == "" {
		return nil, errors.New("INVALID_INPUT: session_token cannot be empty")
	}

	session, err := s.sessionRepo.GetBySessionToken(ctx, sessionToken)
	if err != nil {
		return nil, errors.New("NOT_FOUND: test session not found")
	}

	return session, nil
}

func (s *testSessionService) GetTestQuestion(ctx context.Context, sessionToken string, nomorUrut int) (*TestQuestionDTO, error) {
	if sessionToken == "" {
		return nil, errors.New("INVALID_INPUT: session_token cannot be empty")
	}
	if nomorUrut <= 0 {
		return nil, errors.New("INVALID_INPUT: nomor_urut must be greater than 0")
	}

	// Get session
	session, err := s.sessionRepo.GetBySessionToken(ctx, sessionToken)
	if err != nil {
		return nil, errors.New("NOT_FOUND: test session not found")
	}

	// Check if session is still ongoing
	if session.Status != entity.TestStatusOngoing {
		return nil, errors.New("VALIDATION_ERROR: test session is not ongoing")
	}

	// Check timeout
	batasWaktu := session.BatasWaktu()
	if time.Now().After(batasWaktu) {
		// Auto timeout
		session.Status = entity.TestStatusTimeout
		session.WaktuSelesai = timePtr(time.Now())
		_ = s.sessionRepo.Update(ctx, session)
		return nil, errors.New("TIMEOUT: test session time limit exceeded")
	}

	// Get all session soal for this session
	sessionSoalList, err := s.sessionSoalRepo.ListByTestSession(ctx, session.ID)
	if err != nil {
		return nil, errors.New("DATABASE_ERROR: failed to fetch questions")
	}

	if len(sessionSoalList) == 0 {
		return nil, errors.New("NOT_FOUND: no questions found in session")
	}

	// Find soal with matching nomor_urut
	var targetSessionSoal *entity.TestSessionSoal
	for _, ss := range sessionSoalList {
		if ss.NomorUrut == nomorUrut {
			targetSessionSoal = ss
			break
		}
	}

	if targetSessionSoal == nil {
		return nil, fmt.Errorf("NOT_FOUND: question number %d not found in session", nomorUrut)
	}

	// Get student's answer if exists
	jawaban, _ := s.jawabanRepo.GetByTestSessionSoal(ctx, targetSessionSoal.ID)

	// Convert to SoalForStudent (no answer exposed)
	soalDTO := &entity.SoalForStudent{
		ID:         targetSessionSoal.Soal.ID,
		NomorUrut:  nomorUrut,
		Pertanyaan: targetSessionSoal.Soal.Pertanyaan,
		OpsiA:      targetSessionSoal.Soal.OpsiA,
		OpsiB:      targetSessionSoal.Soal.OpsiB,
		OpsiC:      targetSessionSoal.Soal.OpsiC,
		OpsiD:      targetSessionSoal.Soal.OpsiD,
		IsAnswered: jawaban != nil && jawaban.JawabanDipilih != nil,
	}

	if jawaban != nil && jawaban.JawabanDipilih != nil {
		soalDTO.JawabanDipilih = jawaban.JawabanDipilih
	}

	// Get answered status array for sidebar
	isAnsweredStatus, _ := s.jawabanRepo.GetAnsweredStatusArray(ctx, session.ID, len(sessionSoalList))

	// Count answered
	dijawabCount, _ := s.jawabanRepo.CountAnsweredByTestSession(ctx, session.ID)

	// Calculate time remaining
	timeRemainingDetik := int(batasWaktu.Sub(time.Now()).Seconds())
	if timeRemainingDetik < 0 {
		timeRemainingDetik = 0
	}

	return &TestQuestionDTO{
		SessionToken:       sessionToken,
		Soal:               soalDTO,
		TotalSoal:          len(sessionSoalList),
		CurrentNomorUrut:   nomorUrut,
		DijawabCount:       int(dijawabCount),
		IsAnsweredStatus:   isAnsweredStatus,
		BatasWaktu:         batasWaktu,
		TimeRemainingDetik: timeRemainingDetik,
	}, nil
}

func (s *testSessionService) SubmitAnswer(ctx context.Context, sessionToken string, nomorUrut int, jawabanDipilih entity.JawabanOption) (*AnswerSubmitDTO, error) {
	if sessionToken == "" {
		return nil, errors.New("INVALID_INPUT: session_token cannot be empty")
	}
	if nomorUrut <= 0 {
		return nil, errors.New("INVALID_INPUT: nomor_urut must be greater than 0")
	}
	if jawabanDipilih != entity.JawabanA && jawabanDipilih != entity.JawabanB && jawabanDipilih != entity.JawabanC && jawabanDipilih != entity.JawabanD {
		return nil, errors.New("INVALID_INPUT: jawaban_dipilih must be A, B, C, or D")
	}

	// Get session
	session, err := s.sessionRepo.GetBySessionToken(ctx, sessionToken)
	if err != nil {
		return nil, errors.New("NOT_FOUND: test session not found")
	}

	// Check if session is ongoing
	if session.Status != entity.TestStatusOngoing {
		return nil, errors.New("VALIDATION_ERROR: test session is not ongoing")
	}

	// Check timeout
	batasWaktu := session.BatasWaktu()
	if time.Now().After(batasWaktu) {
		session.Status = entity.TestStatusTimeout
		session.WaktuSelesai = timePtr(time.Now())
		_ = s.sessionRepo.Update(ctx, session)
		return nil, errors.New("TIMEOUT: test session time limit exceeded")
	}

	// Get session soal
	sessionSoal, err := s.sessionSoalRepo.GetBySessionAndNomorUrut(ctx, session.ID, nomorUrut)
	if err != nil {
		return nil, fmt.Errorf("NOT_FOUND: question number %d not found in session", nomorUrut)
	}

	// Check if answer is correct
	isCorrect := sessionSoal.Soal.JawabanBenar == jawabanDipilih

	// Save/update answer
	jawaban := &entity.JawabanSiswa{
		IDTestSessionSoal: sessionSoal.ID,
		JawabanDipilih:    &jawabanDipilih,
		IsCorrect:         isCorrect,
	}

	if err := s.jawabanRepo.CreateOrUpdate(ctx, jawaban); err != nil {
		return nil, errors.New("DATABASE_ERROR: failed to save answer")
	}

	// Reload to get dijawab_pada
	savedJawaban, _ := s.jawabanRepo.GetByTestSessionSoal(ctx, sessionSoal.ID)

	return &AnswerSubmitDTO{
		SessionToken:   sessionToken,
		NomorUrut:      nomorUrut,
		JawabanDipilih: jawabanDipilih,
		IsCorrect:      isCorrect,
		DijawabPada:    savedJawaban.DijawabPada,
	}, nil
}

func (s *testSessionService) CompleteTestSession(ctx context.Context, sessionToken string) (*TestResultDTO, error) {
	if sessionToken == "" {
		return nil, errors.New("INVALID_INPUT: session_token cannot be empty")
	}

	// Get session
	session, err := s.sessionRepo.GetBySessionToken(ctx, sessionToken)
	if err != nil {
		return nil, errors.New("NOT_FOUND: test session not found")
	}

	// Check if already completed
	if session.Status == entity.TestStatusCompleted {
		return s.GetTestResult(ctx, sessionToken)
	}

	// Calculate scores
	totalSoal := *session.TotalSoal
	jumlahBenar, _ := s.jawabanRepo.CountCorrectByTestSession(ctx, session.ID)
	nilaiAkhir := (float64(jumlahBenar) / float64(totalSoal)) * 100

	// Update session
	now := time.Now()
	session.Status = entity.TestStatusCompleted
	session.WaktuSelesai = &now
	session.JumlahBenar = intPtr(int(jumlahBenar))
	session.NilaiAkhir = &nilaiAkhir

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, errors.New("DATABASE_ERROR: failed to complete session")
	}

	return s.GetTestResult(ctx, sessionToken)
}

func (s *testSessionService) GetTestResult(ctx context.Context, sessionToken string) (*TestResultDTO, error) {
	if sessionToken == "" {
		return nil, errors.New("INVALID_INPUT: session_token cannot be empty")
	}

	// Get session
	session, err := s.sessionRepo.GetBySessionToken(ctx, sessionToken)
	if err != nil {
		return nil, errors.New("NOT_FOUND: test session not found")
	}

	// Only return results if completed
	if session.Status != entity.TestStatusCompleted {
		return nil, errors.New("VALIDATION_ERROR: test session is not completed")
	}

	// Get all answers with details
	answerDetails, _ := s.jawabanRepo.ListByTestSessionWithAnsweredDetails(ctx, session.ID)

	var jawabanDetails []*entity.JawabanDetail
	for _, detail := range answerDetails {
		jawabanDetails = append(jawabanDetails, &entity.JawabanDetail{
			NomorUrut:    detail["nomor_urut"].(int),
			Pertanyaan:   detail["pertanyaan"].(string),
			OpsiA:        detail["opsi_a"].(string),
			OpsiB:        detail["opsi_b"].(string),
			OpsiC:        detail["opsi_c"].(string),
			OpsiD:        detail["opsi_d"].(string),
			JawabanBenar: entity.JawabanOption(detail["jawaban_benar"].(string)),
			IsCorrect:    detail["is_correct"].(bool),
		})
	}

	durasi := int(0)
	if session.WaktuSelesai != nil {
		durasi = int(session.WaktuSelesai.Sub(session.WaktuMulai).Seconds())
	}

	persentaseBenar := 0.0
	if session.NilaiAkhir != nil {
		persentaseBenar = *session.NilaiAkhir
	}

	return &TestResultDTO{
		SessionToken:          sessionToken,
		NamaPeserta:           session.NamaPeserta,
		MataPelajaran:         &session.MataPelajaran,
		Tingkatan:             session.Tingkatan,
		WaktuMulai:            session.WaktuMulai,
		WaktuSelesai:          session.WaktuSelesai,
		DurasiPengerjaanDetik: durasi,
		NilaiAkhir:            *session.NilaiAkhir,
		JumlahBenar:           *session.JumlahBenar,
		TotalSoal:             *session.TotalSoal,
		PersentaseBenar:       persentaseBenar,
		Status:                session.Status,
		JawabanDetails:        jawabanDetails,
	}, nil
}

func (s *testSessionService) ListStudentHistory(ctx context.Context, namaPeserta string) ([]*entity.HistorySummary, error) {
	if namaPeserta == "" {
		return nil, errors.New("INVALID_INPUT: nama_peserta cannot be empty")
	}

	sessions, err := s.sessionRepo.ListByStudentName(ctx, namaPeserta)
	if err != nil {
		return nil, errors.New("DATABASE_ERROR: failed to fetch history")
	}

	var history []*entity.HistorySummary
	for _, sess := range sessions {
		if sess.Status != entity.TestStatusCompleted {
			continue
		}

		durasi := int(0)
		if sess.WaktuSelesai != nil {
			durasi = int(sess.WaktuSelesai.Sub(sess.WaktuMulai).Seconds())
		}

		history = append(history, &entity.HistorySummary{
			ID:                    sess.ID,
			SessionToken:          sess.SessionToken,
			MataPelajaran:         sess.MataPelajaran,
			WaktuMulai:            sess.WaktuMulai,
			WaktuSelesai:          sess.WaktuSelesai,
			DurasiPengerjaanDetik: durasi,
			NilaiAkhir:            *sess.NilaiAkhir,
			JumlahBenar:           *sess.JumlahBenar,
			TotalSoal:             *sess.TotalSoal,
			Status:                sess.Status,
		})
	}

	return history, nil
}

func (s *testSessionService) CheckSessionTimeout(ctx context.Context, sessionToken string) (bool, error) {
	if sessionToken == "" {
		return false, errors.New("INVALID_INPUT: session_token cannot be empty")
	}

	session, err := s.sessionRepo.GetBySessionToken(ctx, sessionToken)
	if err != nil {
		return false, errors.New("NOT_FOUND: test session not found")
	}

	batasWaktu := session.BatasWaktu()
	isTimeout := time.Now().After(batasWaktu)

	if isTimeout && session.Status == entity.TestStatusOngoing {
		session.Status = entity.TestStatusTimeout
		session.WaktuSelesai = timePtr(time.Now())
		_ = s.sessionRepo.Update(ctx, session)
	}

	return isTimeout, nil
}

func (s *testSessionService) CancelTestSession(ctx context.Context, sessionToken string) error {
	if sessionToken == "" {
		return errors.New("INVALID_INPUT: session_token cannot be empty")
	}

	session, err := s.sessionRepo.GetBySessionToken(ctx, sessionToken)
	if err != nil {
		return errors.New("NOT_FOUND: test session not found")
	}

	if session.Status != entity.TestStatusOngoing {
		return errors.New("VALIDATION_ERROR: can only cancel ongoing sessions")
	}

	// Delete session and all related data
	if err := s.sessionRepo.Delete(ctx, session.ID); err != nil {
		return errors.New("DATABASE_ERROR: failed to cancel session")
	}

	return nil
}

// Helper functions
func intPtr(v int) *int {
	return &v
}

func timePtr(v time.Time) *time.Time {
	return &v
}
