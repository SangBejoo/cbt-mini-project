package test_session

import (
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/repository/auth"
	"cbt-test-mini-project/internal/repository/test_session"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	PublishExamResult(ctx context.Context, sessionID int, lmsAssignmentID, lmsUserID, lmsClassID int64, score float64, correctCount, totalCount int) error
}

// testSessionUsecaseImpl implements TestSessionUsecase
type testSessionUsecaseImpl struct {
	repo      test_session.TestSessionRepository
	userRepo  auth.AuthRepository
	publisher EventPublisher
}

// NewTestSessionUsecase creates a new TestSessionUsecase instance
func NewTestSessionUsecase(repo test_session.TestSessionRepository, userRepo auth.AuthRepository, publisher EventPublisher) TestSessionUsecase {
	return &testSessionUsecaseImpl{
		repo:      repo,
		userRepo:  userRepo,
		publisher: publisher,
	}
}

// ... existing code ...

// CompleteSession completes the session and calculates score
func (u *testSessionUsecaseImpl) CompleteSession(sessionToken string) (*entity.TestSession, error) {
	session, err := u.repo.GetByToken(sessionToken)
	if err != nil {
		return nil, err
	}

	if session.Status != entity.TestStatusOngoing && session.Status != entity.TestStatusTimeout {
		return nil, errors.New("session is already completed")
	}

	// Get all assigned questions
	allQuestions, err := u.repo.GetAllQuestionsForSession(sessionToken)
	if err != nil {
		return nil, err
	}

	// Get all answers
	answers, err := u.repo.GetSessionAnswers(sessionToken)
	if err != nil {
		return nil, err
	}

	// Create a map of answered question numbers
	answeredMap := make(map[int]bool)
	for _, ans := range answers {
		answeredMap[ans.TestSessionSoal.NomorUrut] = true
	}

	// For each unanswered question, create an entry directly in repository with nil answer
	for _, question := range allQuestions {
		if !answeredMap[question.NomorUrut] {
			// Create unanswered record with NULL jawaban_dipilih
			err := u.repo.CreateUnansweredRecord(question.ID, question.IDTestSession)
			if err != nil {
				// Log error but continue
				continue
			}
		}
	}

	// Get updated answers after filling in unanswered questions
	answers, err = u.repo.GetSessionAnswers(sessionToken)
	if err != nil {
		return nil, err
	}

	// Calculate score
	jumlahBenar := 0
	for _, ans := range answers {
		if ans.IsCorrect {
			jumlahBenar++
		}
	}

	totalSoal := len(allQuestions)
	var nilaiAkhir float64
	if totalSoal > 0 {
		nilaiAkhir = float64(jumlahBenar) / float64(totalSoal) * 100
	}

	err = u.repo.CompleteSession(sessionToken, time.Now(), &nilaiAkhir, &jumlahBenar, &totalSoal)
	if err != nil {
		return nil, err
	}

	// Publish logic
	updatedSession, err := u.repo.GetByToken(sessionToken)
	if err != nil {
		return nil, err
	}

	// Publish event if publisher is available and session has LMS linkage
	if u.publisher != nil && session.LMSAssignmentID != nil && session.LMSClassID != nil && session.UserID != nil {
		// Get User to retrieve LMS User ID
		user, err := u.userRepo.GetUserByID(context.Background(), int32(*session.UserID))
		if err != nil {
			// Log error but assume we can't publish
			fmt.Printf("Error getting user for publishing result: %v\n", err)
		} else if user != nil && user.LmsUserId != 0 {
			// Use correct LMS User ID (Note: Proto generated field is LmsUserId)
			err = u.publisher.PublishExamResult(
				context.Background(),
				session.ID,
				*session.LMSAssignmentID,
				user.LmsUserId,
				*session.LMSClassID,
				nilaiAkhir,
				jumlahBenar,
				totalSoal,
			)
			if err != nil {
				fmt.Printf("Error publishing exam result: %v\n", err)
			}
		}
	}
    
	return updatedSession, nil
}

// CreateTestSession creates a new test session with random questions
func (u *testSessionUsecaseImpl) CreateTestSession(userID, tingkatan, idMataPelajaran, durasiMenit, jumlahSoal int) (*entity.TestSession, error) {
	fmt.Printf("=== USECASE CreateTestSession: userID=%d, tingkatan=%d, idMataPelajaran=%d ===\n", userID, tingkatan, idMataPelajaran)
	
	if userID < 1 || tingkatan < 1 || idMataPelajaran < 1 || durasiMenit < 1 || jumlahSoal < 1 {
		return nil, fmt.Errorf("invalid input parameters: userID=%d, tingkatan=%d, idMataPelajaran=%d, durasiMenit=%d, jumlahSoal=%d", userID, tingkatan, idMataPelajaran, durasiMenit, jumlahSoal)
	}

	// Check if userRepo is initialized
	if u.userRepo == nil {
		return nil, fmt.Errorf("userRepo is nil - dependency injection issue")
	}

	// Get user to set nama peserta
	fmt.Printf("=== USECASE: Getting user by ID %d ===\n", userID)
	user, err := u.userRepo.GetUserByID(context.Background(), int32(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID %d: %w", userID, err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found with ID %d", userID)
	}
	fmt.Printf("=== USECASE: Got user: %s ===\n", user.Nama)

	// Generate unique session token
	token, err := u.generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	fmt.Printf("=== USECASE: Generated token: %s ===\n", token)

	waktuMulai := time.Now()
	session := &entity.TestSession{
		SessionToken:    token,
		UserID:          &userID,
		NamaPeserta:     user.Nama,
		IDTingkat:       tingkatan,
		IDMataPelajaran: idMataPelajaran,
		WaktuMulai:      waktuMulai,
		DurasiMenit:     durasiMenit,
		Status:          entity.TestStatusOngoing,
	}

	fmt.Printf("=== USECASE: Creating session in DB ===\n")
	err = u.repo.Create(session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	fmt.Printf("=== USECASE: Session created with ID %d ===\n", session.ID)

	// Select and assign random questions
	fmt.Printf("=== USECASE: Assigning random questions ===\n")
	err = u.repo.AssignRandomQuestions(session.ID, idMataPelajaran, tingkatan, jumlahSoal)
	if err != nil {
		// Cleanup session if question assignment fails
		u.repo.Delete(session.ID)
		return nil, fmt.Errorf("failed to assign questions: %w", err)
	}
	fmt.Printf("=== USECASE: Questions assigned ===\n")

	result, err := u.repo.GetByToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to get created session: %w", err)
	}
	fmt.Printf("=== USECASE: Session retrieved successfully ===\n")
	return result, nil
}

// GetTestSession gets by token
func (u *testSessionUsecaseImpl) GetTestSession(sessionToken string) (*entity.TestSession, error) {
	session, err := u.repo.GetByToken(sessionToken)
	if err != nil {
		return nil, err
	}

	// Check timeout only if session is still ongoing
	if session.Status == entity.TestStatusOngoing && time.Now().After(session.BatasWaktu()) {
		// Just mark as timeout, don't complete yet (let auto-submit or manual completion handle scoring)
		u.repo.UpdateSessionStatus(sessionToken, entity.TestStatusTimeout)
		session.Status = entity.TestStatusTimeout
	}

	return session, nil
}

// GetTestQuestions gets a single question for the student
func (u *testSessionUsecaseImpl) GetTestQuestions(sessionToken string, nomorUrut int) (*entity.QuestionForStudent, error) {
	_, err := u.GetTestSession(sessionToken)
	if err != nil {
		return nil, err
	}

	// Get the TestSessionSoal to determine question type
	tss, err := u.repo.GetTestSessionSoalByOrder(sessionToken, nomorUrut)
	if err != nil {
		return nil, err
	}

	question := &entity.QuestionForStudent{
		NomorUrut:    nomorUrut,
		QuestionType: tss.QuestionType,
		Materi:       tss.Soal.Materi, // Will be nil for drag-drop, need to handle this
		IsAnswered:   false, // Will be set below
	}

	// Get existing answer if any
	answers, _ := u.repo.GetSessionAnswers(sessionToken)
	for _, ans := range answers {
		if ans.TestSessionSoal.NomorUrut == nomorUrut {
			question.IsAnswered = true
			break
		}
	}

	if tss.QuestionType == entity.QuestionTypeMultipleChoice && tss.IDSoal != nil {
		// Handle multiple choice question
		soal, err := u.repo.GetQuestionByOrder(sessionToken, nomorUrut)
		if err != nil {
			return nil, err
		}

		var jawabanDipilih *entity.JawabanOption
		for _, ans := range answers {
			if ans.TestSessionSoal.NomorUrut == nomorUrut && ans.JawabanDipilih != nil {
				jawabanDipilih = ans.JawabanDipilih
				break
			}
		}

		question.Materi = soal.Materi
		question.MCID = &soal.ID
		question.MCPertanyaan = &soal.Pertanyaan
		question.MCOpsiA = &soal.OpsiA
		question.MCOpsiB = &soal.OpsiB
		question.MCOpsiC = &soal.OpsiC
		question.MCOpsiD = &soal.OpsiD
		question.MCJawabanDipilih = jawabanDipilih
		question.MCGambar = soal.Gambar

	} else if tss.QuestionType == entity.QuestionTypeDragDrop && tss.IDSoalDragDrop != nil {
		// Handle drag-drop question
		soalDD, err := u.repo.GetSoalDragDropByID(*tss.IDSoalDragDrop)
		if err != nil {
			return nil, err
		}

		var userAnswer map[int]int
		for _, ans := range answers {
			if ans.TestSessionSoal.NomorUrut == nomorUrut && ans.QuestionType == entity.QuestionTypeDragDrop {
				userAnswer = ans.GetDragDropAnswer()
				break
			}
		}

		question.Materi = soalDD.Materi
		question.DDID = &soalDD.ID
		question.DDPertanyaan = &soalDD.Pertanyaan
		question.DDDDragType = &soalDD.DragType
		question.DDItems = soalDD.Items
		question.DDSlots = soalDD.Slots
		question.DDUserAnswer = userAnswer
	}

	return question, nil
}

// GetAllTestQuestions gets all questions for the session
func (u *testSessionUsecaseImpl) GetAllTestQuestions(sessionToken string) ([]entity.QuestionForStudent, error) {
	_, err := u.GetTestSession(sessionToken)
	if err != nil {
		return nil, err
	}

	sessionSoals, err := u.repo.GetAllQuestionsForSession(sessionToken)
	if err != nil {
		return nil, err
	}

	// Get existing answers
	answers, _ := u.repo.GetSessionAnswers(sessionToken)

	var questions []entity.QuestionForStudent
	for _, tss := range sessionSoals {
		question := &entity.QuestionForStudent{
			NomorUrut:    tss.NomorUrut,
			QuestionType: tss.QuestionType,
			IsAnswered:   false, // Will be set below
		}

		// Check if answered
		for _, ans := range answers {
			if ans.TestSessionSoal.NomorUrut == tss.NomorUrut {
				question.IsAnswered = true
				break
			}
		}

		if tss.QuestionType == entity.QuestionTypeMultipleChoice && tss.Soal.ID > 0 {
			// Handle multiple choice question
			var jawabanDipilih *entity.JawabanOption
			for _, ans := range answers {
				if ans.TestSessionSoal.NomorUrut == tss.NomorUrut && ans.JawabanDipilih != nil {
					jawabanDipilih = ans.JawabanDipilih
					break
				}
			}

			question.Materi = tss.Soal.Materi
			question.MCID = &tss.Soal.ID
			question.MCPertanyaan = &tss.Soal.Pertanyaan
			question.MCOpsiA = &tss.Soal.OpsiA
			question.MCOpsiB = &tss.Soal.OpsiB
			question.MCOpsiC = &tss.Soal.OpsiC
			question.MCOpsiD = &tss.Soal.OpsiD
			question.MCJawabanDipilih = jawabanDipilih
			question.MCGambar = tss.Soal.Gambar


		} else if tss.QuestionType == entity.QuestionTypeDragDrop && tss.SoalDragDrop != nil && tss.SoalDragDrop.ID > 0 {
			// Handle drag-drop question
			var userAnswer map[int]int
			for _, ans := range answers {
				if ans.TestSessionSoal.NomorUrut == tss.NomorUrut && ans.QuestionType == entity.QuestionTypeDragDrop {
					userAnswer = ans.GetDragDropAnswer()
					break
				}
			}

			question.Materi = tss.SoalDragDrop.Materi
			question.DDID = &tss.SoalDragDrop.ID
			question.DDPertanyaan = &tss.SoalDragDrop.Pertanyaan
			question.DDDDragType = &tss.SoalDragDrop.DragType
			question.DDItems = tss.SoalDragDrop.Items
			question.DDSlots = tss.SoalDragDrop.Slots
			question.DDUserAnswer = userAnswer
		}

		questions = append(questions, *question)
	}

	return questions, nil
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

// SubmitDragDropAnswer submits a drag-drop answer with all-or-nothing scoring
func (u *testSessionUsecaseImpl) SubmitDragDropAnswer(sessionToken string, nomorUrut int, answer map[int]int) error {
	_, err := u.GetTestSession(sessionToken)
	if err != nil {
		return err
	}

	// Get the TestSessionSoal to find the drag-drop question
	tss, err := u.repo.GetTestSessionSoalByOrder(sessionToken, nomorUrut)
	if err != nil {
		return err
	}

	// Verify this is a drag-drop question
	if tss.QuestionType != entity.QuestionTypeDragDrop || tss.IDSoalDragDrop == nil {
		return errors.New("this is not a drag-drop question")
	}

	// Get correct answers
	correctAnswers, err := u.repo.GetDragDropCorrectAnswers(*tss.IDSoalDragDrop)
	if err != nil {
		return err
	}

	// All-or-nothing scoring: check if all answers are correct
	isCorrect := u.checkDragDropAnswer(correctAnswers, answer)

	return u.repo.SubmitDragDropAnswer(sessionToken, nomorUrut, answer, isCorrect)
}

// checkDragDropAnswer implements all-or-nothing scoring
func (u *testSessionUsecaseImpl) checkDragDropAnswer(correctAnswers []entity.DragCorrectAnswer, userAnswer map[int]int) bool {
	if len(userAnswer) != len(correctAnswers) {
		return false // Not all items answered
	}
	for _, correct := range correctAnswers {
		userSlot, exists := userAnswer[correct.IDDragItem]
		if !exists || userSlot != correct.IDDragSlot {
			return false // Wrong or missing answer
		}
	}
	return true // All correct!
}

// ClearAnswer clears an answer
func (u *testSessionUsecaseImpl) ClearAnswer(sessionToken string, nomorUrut int) error {
	_, err := u.GetTestSession(sessionToken)
	if err != nil {
		return err
	}

	return u.repo.ClearAnswer(sessionToken, nomorUrut)
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

	// Get all assigned questions (for proper ordering)
	allQuestions, err := u.repo.GetAllQuestionsForSession(sessionToken)
	if err != nil {
		return nil, nil, err
	}

	// Get detailed answers
	var details []entity.JawabanDetail
	answers, err := u.repo.GetSessionAnswers(sessionToken)
	if err != nil {
		return nil, nil, err
	}

	// Create a map of answers for quick lookup
	answersMap := make(map[int]*entity.JawabanSiswa)
	for i, ans := range answers {
		answersMap[ans.TestSessionSoal.NomorUrut] = &answers[i]
	}

	// Process all questions in order
	// Process all questions in order
	for _, question := range allQuestions {
		var detail entity.JawabanDetail
		detail.NomorUrut = question.NomorUrut
		detail.QuestionType = question.QuestionType

		// Check type
		if question.QuestionType == entity.QuestionTypeDragDrop && question.SoalDragDrop != nil {
			// DRAG DROP Handling
			dd := question.SoalDragDrop
			detail.Pertanyaan = dd.Pertanyaan
			detail.Pembahasan = dd.Pembahasan
			detail.DragType = &dd.DragType
			detail.DragItems = dd.Items
			detail.DragSlots = dd.Slots

			// Map DragDrop Images to generic SoalGambar for ID
			if len(dd.Gambar) > 0 {
				convertedImages := make([]entity.SoalGambar, len(dd.Gambar))
				for i, img := range dd.Gambar {
					convertedImages[i] = entity.SoalGambar{
						ID:         img.ID,
						NamaFile:   img.NamaFile,
						FilePath:   img.FilePath,
						FileSize:   img.FileSize,
						MimeType:   img.MimeType,
						Urutan:     img.Urutan,
						Keterangan: img.Keterangan,
						CreatedAt:  img.CreatedAt,
					}
				}
				detail.Gambar = convertedImages
			}

			// Fetch correct answers (needed for result view)
			correctAnswersList, err := u.repo.GetDragDropCorrectAnswers(dd.ID)
			if err == nil {
				detail.CorrectDragAnswer = make(map[int]int)
				for _, ca := range correctAnswersList {
					detail.CorrectDragAnswer[ca.IDDragItem] = ca.IDDragSlot
				}
			}

			// Get User Answer
			if ans, exists := answersMap[question.NomorUrut]; exists {
				detail.UserDragAnswer = ans.GetDragDropAnswer()
				detail.IsCorrect = ans.IsCorrect
				detail.IsAnswered = (len(detail.UserDragAnswer) > 0)
			} else {
				detail.IsCorrect = false
				detail.IsAnswered = false
			}

		} else {
			// MULTIPLE CHOICE Handling (Default)
			// Use loaded Soal info directly
			
			detail.Pertanyaan = question.Soal.Pertanyaan
			detail.OpsiA = question.Soal.OpsiA
			detail.OpsiB = question.Soal.OpsiB
			detail.OpsiC = question.Soal.OpsiC
			detail.OpsiD = question.Soal.OpsiD
			detail.JawabanBenar = question.Soal.JawabanBenar
			detail.Pembahasan = question.Soal.Pembahasan
			detail.Gambar = question.Soal.Gambar

			// Add answer details if exists
			if ans, exists := answersMap[question.NomorUrut]; exists {
				detail.JawabanDipilih = ans.JawabanDipilih
				detail.IsCorrect = ans.IsCorrect
				detail.IsAnswered = ans.JawabanDipilih != nil
			} else {
				// No answer provided
				detail.IsCorrect = false
				detail.IsAnswered = false
			}
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