package test_session_test

import (
	"context"
	"testing"
	"time"

	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/usecase/test_session"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) GetUserByID(ctx context.Context, id int32) (*base.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*base.User), args.Error(1)
}

func (m *MockUserRepo) FindOrCreateByLMSID(ctx context.Context, lmsID int64, email, name string, role int32) (*base.User, error) {
	args := m.Called(ctx, lmsID, email, name, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*base.User), args.Error(1)
}

func (m *MockUserRepo) CheckUserHasTestSessions(ctx context.Context, id int32) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepo) GetLMSUserIDByLocalID(ctx context.Context, id int32) (int64, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepo) DeleteUser(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepo) Login(ctx context.Context, email, password string) (*base.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*base.User), args.Error(1)
}

func (m *MockUserRepo) GetUserByEmail(ctx context.Context, email string) (*base.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*base.User), args.Error(1)
}

func (m *MockUserRepo) CreateUser(ctx context.Context, user *base.User) (*base.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*base.User), args.Error(1)
}

func (m *MockUserRepo) UpdateUser(ctx context.Context, id int32, updates map[string]interface{}) (*base.User, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*base.User), args.Error(1)
}

func (m *MockUserRepo) ListUsers(ctx context.Context, role int32, statusFilter int32, limit, offset int) ([]*base.User, int, error) {
	args := m.Called(ctx, role, statusFilter, limit, offset)
	return args.Get(0).([]*base.User), args.Get(1).(int), args.Error(2)
}

type MockTestSessionRepo struct {
	mock.Mock
}

func (m *MockTestSessionRepo) Create(session *entity.TestSession) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockTestSessionRepo) Update(session *entity.TestSession) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockTestSessionRepo) GetByToken(token string) (*entity.TestSession, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TestSession), args.Error(1)
}

func (m *MockTestSessionRepo) AssignRandomQuestions(sessionID, idMataPelajaran, idTingkat, jumlahSoal int) error {
	args := m.Called(sessionID, idMataPelajaran, idTingkat, jumlahSoal)
	return args.Error(0)
}

func (m *MockTestSessionRepo) GetTestSessionSoalByOrder(token string, nomorUrut int) (*entity.TestSessionSoal, error) {
	args := m.Called(token, nomorUrut)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TestSessionSoal), args.Error(1)
}

func (m *MockTestSessionRepo) GetQuestionByOrder(token string, nomorUrut int) (*entity.Soal, error) {
	args := m.Called(token, nomorUrut)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Soal), args.Error(1)
}

func (m *MockTestSessionRepo) SubmitAnswer(token string, nomorUrut int, jawaban entity.JawabanOption) error {
	args := m.Called(token, nomorUrut, jawaban)
	return args.Error(0)
}

func (m *MockTestSessionRepo) GetAllQuestionsForSession(token string) ([]entity.TestSessionSoal, error) {
	args := m.Called(token)
	return args.Get(0).([]entity.TestSessionSoal), args.Error(1)
}

func (m *MockTestSessionRepo) GetSessionQuestions(token string) ([]entity.TestSessionSoal, error) {
	args := m.Called(token)
	return args.Get(0).([]entity.TestSessionSoal), args.Error(1)
}

func (m *MockTestSessionRepo) GetSessionAnswers(token string) ([]entity.JawabanSiswa, error) {
	args := m.Called(token)
	return args.Get(0).([]entity.JawabanSiswa), args.Error(1)
}

func (m *MockTestSessionRepo) CreateUnansweredRecord(soalID, sessionID int) error {
	args := m.Called(soalID, sessionID)
	return args.Error(0)
}

func (m *MockTestSessionRepo) CompleteSession(token string, endTime time.Time, score *float64, correct *int, total *int) error {
	args := m.Called(token, endTime, score, correct, total)
	return args.Error(0)
}

func (m *MockTestSessionRepo) UpdateSessionStatus(token string, status entity.TestStatus) error {
	args := m.Called(token, status)
	return args.Error(0)
}

func (m *MockTestSessionRepo) List(tingkatan, idMataPelajaran *int, status *entity.TestStatus, limit, offset int) ([]entity.TestSession, int, error) {
	args := m.Called(tingkatan, idMataPelajaran, status, limit, offset)
	return args.Get(0).([]entity.TestSession), args.Get(1).(int), args.Error(2)
}

func (m *MockTestSessionRepo) GetSoalDragDropByID(id int) (*entity.SoalDragDrop, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SoalDragDrop), args.Error(1)
}

func (m *MockTestSessionRepo) GetDragDropCorrectAnswers(soalID int) ([]entity.DragCorrectAnswer, error) {
	args := m.Called(soalID)
	return args.Get(0).([]entity.DragCorrectAnswer), args.Error(1)
}

func (m *MockTestSessionRepo) SubmitDragDropAnswer(token string, nomorUrut int, answer map[int]int, isCorrect bool) error {
	args := m.Called(token, nomorUrut, answer, isCorrect)
	return args.Error(0)
}

func (m *MockTestSessionRepo) ClearAnswer(token string, nomorUrut int) error {
	args := m.Called(token, nomorUrut)
	return args.Error(0)
}

func (m *MockTestSessionRepo) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) PublishExamResult(ctx context.Context, sessionID int, lmsAssignmentID, lmsUserID, lmsClassID int64, score float64, correctCount, totalCount int) error {
	args := m.Called(ctx, sessionID, lmsAssignmentID, lmsUserID, lmsClassID, score, correctCount, totalCount)
	return args.Error(0)
}

// --- Test ---

func TestCompleteSession_Integration(t *testing.T) {
	// Setup
	mockRepo := new(MockTestSessionRepo)
	mockUserRepo := new(MockUserRepo)
	mockPublisher := new(MockPublisher)

	usecase := test_session.NewTestSessionUsecase(mockRepo, mockUserRepo, mockPublisher)

	// Data
	token := "valid-token"
	userID := int(123)
	lmsAssignmentID := int64(1001)
	lmsClassID := int64(2002)
	lmsUserID := int64(9999)

	session := &entity.TestSession{
		ID:              1,
		SessionToken:    token,
		Status:          entity.TestStatusOngoing,
		UserID:          &userID,
		LMSAssignmentID: &lmsAssignmentID,
		LMSClassID:      &lmsClassID,
	}

	// Mock Expectations

	// 1. GetByToken (initial check)
	mockRepo.On("GetByToken", token).Return(session, nil).Once()

	// 2. GetAllQuestionsForSession
	questions := []entity.TestSessionSoal{
		{ID: 1, IDTestSession: 1, NomorUrut: 1, QuestionType: "multiple_choice"},
	}
	mockRepo.On("GetAllQuestionsForSession", token).Return(questions, nil)

	// 3. GetSessionAnswers (1st call - before filling unanswered)
	answers1 := []entity.JawabanSiswa{} // empty
	mockRepo.On("GetSessionAnswers", token).Return(answers1, nil).Once()

	// 4. CreateUnansweredRecord (since answer is empty)
	mockRepo.On("CreateUnansweredRecord", 1, 1).Return(nil)

	// 5. GetSessionAnswers (2nd call - after filling)
	answers2 := []entity.JawabanSiswa{
		{TestSessionSoal: entity.TestSessionSoal{ID: 1, NomorUrut: 1}, IsCorrect: true},
	}
	mockRepo.On("GetSessionAnswers", token).Return(answers2, nil).Once()

	// 6. CompleteSession (update DB)
	// We use Matcher because time.Now() is variable
	mockRepo.On("CompleteSession", token, mock.AnythingOfType("time.Time"), mock.MatchedBy(func(s *float64) bool { return *s == 100.0 }), mock.MatchedBy(func(c *int) bool { return *c == 1 }), mock.MatchedBy(func(tot *int) bool { return *tot == 1 })).Return(nil)

	// 7. GetByToken (retrieve for return value)
	completedSession := *session
	completedSession.Status = entity.TestStatusCompleted
	mockRepo.On("GetByToken", token).Return(&completedSession, nil).Once()

	// 8. GetUserByID (for publishing)
	user := &base.User{Id: int32(userID), LmsUserId: lmsUserID}
	mockUserRepo.On("GetUserByID", mock.Anything, int32(userID)).Return(user, nil)

	// 9. PublishExamResult
	mockPublisher.On("PublishExamResult", mock.Anything, 1, lmsAssignmentID, lmsUserID, lmsClassID, 100.0, 1, 1).Return(nil)

	// Execute
	res, err := usecase.CompleteSession(token)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, entity.TestStatusCompleted, res.Status)
	mockPublisher.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}
