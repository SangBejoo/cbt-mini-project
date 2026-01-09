package test_session

import (
	"cbt-test-mini-project/internal/entity"
)

// TestSessionUsecase defines the interface for TestSession usecase operations
type TestSessionUsecase interface {
	CreateTestSession(userID, tingkatan, idMataPelajaran, durasiMenit, jumlahSoal int) (*entity.TestSession, error)
	GetTestSession(sessionToken string) (*entity.TestSession, error)
	GetTestQuestions(sessionToken string, nomorUrut int) (*entity.QuestionForStudent, error)
	GetAllTestQuestions(sessionToken string) ([]entity.QuestionForStudent, error)
	GetSessionAnswers(sessionToken string) ([]entity.JawabanSiswa, error)
	SubmitAnswer(sessionToken string, nomorUrut int, jawaban entity.JawabanOption) error
	SubmitDragDropAnswer(sessionToken string, nomorUrut int, answer map[int]int) error // NEW: for drag-drop
	ClearAnswer(sessionToken string, nomorUrut int) error
	CompleteSession(sessionToken string) (*entity.TestSession, error)
	GetTestResult(sessionToken string) (*entity.TestSession, []entity.JawabanDetail, error)
	ListTestSessions(tingkatan, idMataPelajaran *int, status *entity.TestStatus, page, pageSize int) ([]entity.TestSession, *entity.PaginationResponse, error)
}
