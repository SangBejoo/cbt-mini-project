package test_session

import (
	"cbt-test-mini-project/internal/entity"
	"time"
)

// TestSessionRepository defines the interface for TestSession repository operations
type TestSessionRepository interface {
	// Create a new test session
	Create(session *entity.TestSession) error

	// Get session by token
	GetByToken(token string) (*entity.TestSession, error)

	// Update existing session
	Update(session *entity.TestSession) error

	// Delete session by ID
	Delete(id int) error

	// Complete session
	CompleteSession(token string, waktuSelesai time.Time, nilaiAkhir *float64, jumlahBenar, totalSoal *int) error

	// Assign random questions to session
	AssignRandomQuestions(sessionID, idMataPelajaran, tingkatan, jumlahSoal int) error

	// List sessions with filters
	List(tingkatan, idMataPelajaran *int, status *entity.TestStatus, limit, offset int) ([]entity.TestSession, int, error)

	// Get questions for session
	GetSessionQuestions(token string) ([]entity.TestSessionSoal, error)

	// Get all questions for session with soal data
	GetAllQuestionsForSession(token string) ([]entity.TestSessionSoal, error)

	// Get single question by order
	GetQuestionByOrder(token string, nomorUrut int) (*entity.Soal, error)

	// Submit answer
	SubmitAnswer(token string, nomorUrut int, jawaban entity.JawabanOption) error

	// Clear answer
	ClearAnswer(token string, nomorUrut int) error

	// Get answers for session
	GetSessionAnswers(token string) ([]entity.JawabanSiswa, error)

	// Create unanswered record with NULL jawaban_dipilih
	CreateUnansweredRecord(sessionSoalID, testSessionID int) error
}