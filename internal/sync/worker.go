package sync

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	infraRedis "cbt-test-mini-project/init/infra/redis"
	"cbt-test-mini-project/internal/entity"
	authRepo "cbt-test-mini-project/internal/repository/auth"
	classRepo "cbt-test-mini-project/internal/repository/class"
	classStudentRepo "cbt-test-mini-project/internal/repository/class_student"
	mataPelajaranRepo "cbt-test-mini-project/internal/repository/mata_pelajaran"
	"cbt-test-mini-project/internal/repository/materi"
	testSessionRepo "cbt-test-mini-project/internal/repository/test_session"
	"cbt-test-mini-project/internal/repository/tingkat"

	"github.com/redis/go-redis/v9"
)

type SyncWorker struct {
	materiRepo       materi.MateriRepository
	tingkatRepo      tingkat.TingkatRepository
	subjectRepo      mataPelajaranRepo.MataPelajaranRepository
	authRepo         authRepo.AuthRepository
	testSessionRepo  testSessionRepo.TestSessionRepository
	classRepo        classRepo.ClassRepository
	classStudentRepo classStudentRepo.ClassStudentRepository
}

func NewSyncWorker(
	materiRepo materi.MateriRepository, 
	tingkatRepo tingkat.TingkatRepository, 
	subjectRepo mataPelajaranRepo.MataPelajaranRepository, 
	authRepo authRepo.AuthRepository, 
	testSessionRepo testSessionRepo.TestSessionRepository,
	classRepo classRepo.ClassRepository,
	classStudentRepo classStudentRepo.ClassStudentRepository,
) *SyncWorker {
	return &SyncWorker{
		materiRepo:       materiRepo,
		tingkatRepo:      tingkatRepo,
		subjectRepo:      subjectRepo,
		authRepo:         authRepo,
		testSessionRepo:  testSessionRepo,
		classRepo:        classRepo,
		classStudentRepo: classStudentRepo,
	}
}

func (w *SyncWorker) Start(ctx context.Context) {
	slog.Info("Sync worker started, listening for LMS events...")
	
	// Start from beginning to process any missed events, then track position
	lastID := "0"
	
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Read from Redis Stream
			// Using XREAD for simplicity here, but XREADGROUP is better for production
			streams, err := infraRedis.RedisClient.XRead(ctx, &redis.XReadArgs{
				Streams: []string{"lms_events", lastID},
				Block:   5 * time.Second, // Block for 5 seconds max, then check context
				Count:   10, // Process 10 messages at a time
			}).Result()

			if err != nil {
				// redis.Nil is returned when Block times out with no new messages
				if err.Error() == "redis: nil" {
					continue
				}
				slog.Error("failed to read from redis stream", "error", err)
				time.Sleep(2 * time.Second)
				continue
			}

			for _, stream := range streams {
				for _, msg := range stream.Messages {
					w.processMessage(ctx, msg.Values)
					lastID = msg.ID // Track last processed message ID
				}
			}
		}
	}
}

func (w *SyncWorker) processMessage(_ context.Context, data map[string]interface{}) {
	eventType, _ := data["event"].(string)
	payload, _ := data["payload"].(string)

	slog.Info("processing LMS event", "type", eventType)

	switch eventType {
	case "level_upsert":
		w.handleLevelUpsert(payload)
	case "subject_upsert":
		w.handleSubjectUpsert(payload)
	case "module_upsert":
		w.handleModuleUpsert(payload)
	case "user_upsert":
		w.handleUserUpsert(payload)
	case "exam_assignment_created":
		w.handleExamAssignmentCreated(payload)
	case "class_upsert":
		w.handleClassUpsert(payload)
	case "class_deleted":
		w.handleClassDeleted(payload)
	case "level_deleted":
		w.handleLevelDeleted(payload)
	case "subject_deleted":
		w.handleSubjectDeleted(payload)
	case "module_deleted":
		w.handleModuleDeleted(payload)
	case "user_deleted":
		w.handleUserDeleted(payload)
	case "class_student_joined":
		w.handleClassStudentJoined(payload)
	case "class_student_left":
		w.handleClassStudentLeft(payload)
	}
}

type LevelPayload struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	SchoolID int64  `json:"school_id"`
}

func (w *SyncWorker) handleLevelUpsert(payload string) {
	slog.Info("Processing level upsert event", "payload", payload)
	var p LevelPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		slog.Error("failed to unmarshal level payload", "error", err)
		return
	}
	
	slog.Info("Parsed level payload", "id", p.ID, "name", p.Name, "school_id", p.SchoolID)
	
	if err := w.tingkatRepo.UpsertByLMSID(p.ID, p.Name); err != nil {
		slog.Error("failed to sync level", "id", p.ID, "error", err)
		return
	}
	slog.Info("Synced Level from LMS", "id", p.ID, "name", p.Name, "school_id", p.SchoolID)
}

type SubjectPayload struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	SchoolID int64  `json:"school_id"`
}

func (w *SyncWorker) handleSubjectUpsert(payload string) {
	var p SubjectPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		slog.Error("failed to unmarshal subject payload", "error", err)
		return
	}
	
	if err := w.subjectRepo.UpsertByLMSID(p.ID, p.Name, p.SchoolID); err != nil {
		slog.Error("failed to sync subject", "id", p.ID, "error", err)
		return
	}
	slog.Info("Synced Subject from LMS", "id", p.ID, "name", p.Name, "school_id", p.SchoolID)
}

type ModulePayload struct {
	ID        int64  `json:"id"`
	SubjectID int64  `json:"subject_id"`
	LevelID   int64  `json:"level_id"`
	Name      string `json:"name"`
}

func (w *SyncWorker) handleModuleUpsert(payload string) {
	var p ModulePayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		slog.Error("failed to unmarshal module payload", "error", err)
		return
	}
	
	if err := w.materiRepo.UpsertByLMSID(p.ID, p.SubjectID, p.LevelID, p.Name); err != nil {
		slog.Error("failed to sync module", "id", p.ID, "error", err)
		return
	}
	slog.Info("Synced Module from LMS", "id", p.ID, "name", p.Name)
}

type UserPayload struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	PasswordHash string `json:"password_hash"`
}

func (w *SyncWorker) handleUserUpsert(payload string) {
	var p UserPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		slog.Error("failed to unmarshal user payload", "error", err)
		return
	}
	
	roleInt := int32(1) // default SISWA
	if p.Role == "ADMIN" {
		roleInt = 2
	}
	
	_, err := w.authRepo.FindOrCreateByLMSID(context.Background(), p.ID, p.Email, p.Name, roleInt)
	if err != nil {
		slog.Error("failed to sync user", "id", p.ID, "error", err)
		return
	}
	slog.Info("Synced User from LMS", "id", p.ID, "email", p.Email)
}

// ExamAssignmentPayload represents the payload for exam assignment created events
type ExamAssignmentPayload struct {
	LMSAssignmentID int64  `json:"lms_assignment_id"`
	LMSClassID      int64  `json:"lms_class_id"`
	Title           string `json:"title"`
	MaxScore        int    `json:"max_score"`
	ModuleID        int64  `json:"module_id"`
	ScheduledTime   string `json:"scheduled_time"`
}

func (w *SyncWorker) handleExamAssignmentCreated(payload string) {
	var p ExamAssignmentPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		slog.Error("failed to unmarshal exam assignment payload", "error", err)
		return
	}

	if p.ModuleID == 0 {
		slog.Warn("skipping exam assignment sync: missing module_id", "assignment_id", p.LMSAssignmentID)
		return
	}

	// 1. Get module (materi) details to get duration and question count
	materi, err := w.materiRepo.GetByLMSID(p.ModuleID)
	if err != nil {
		slog.Error("failed to get materi details", "module_id", p.ModuleID, "error", err)
		// Fallback defaults if not found? Or return?
		// Better to return as we need valid config
		return
	}

	// 2. Get all students in the class
	studentIDs, err := w.classStudentRepo.GetStudentIDsByClassID(p.LMSClassID)
	if err != nil {
		slog.Error("failed to get students for class", "class_id", p.LMSClassID, "error", err)
		return
	}

	// 3. Create test session for each student
	scheduledTime := time.Now()
	if p.ScheduledTime != "" {
		if t, err := time.Parse(time.RFC3339, p.ScheduledTime); err == nil {
			scheduledTime = t
		}
	}

	successCount := 0
	for _, studentID := range studentIDs {
		// user_id in test_session is the local user ID. 
		// classStudentRepo returns local user IDs (mapped from LMS user IDs via sync).
		
		stdID := int(studentID)
		session := &entity.TestSession{
			UserID:          &stdID,
			IDMataPelajaran: int(materi.IDMataPelajaran), // From materi
			IDTingkat:       int(materi.IDTingkat),       // From materi
			LMSAssignmentID: &p.LMSAssignmentID,
			WaktuMulai:      scheduledTime,
			DurasiMenit:     materi.DefaultDurasiMenit,
			TotalSoal:       &materi.DefaultJumlahSoal,
			Status:          entity.TestStatusScheduled,
		}

		if err := w.testSessionRepo.Create(session); err != nil {
			slog.Error("failed to create test session", "assignment_id", p.LMSAssignmentID, "student_id", studentID, "error", err)
			continue
		}
		successCount++
	}

	slog.Info("Synced Exam Assignment", "assignment_id", p.LMSAssignmentID, "class_id", p.LMSClassID, "sessions_created", successCount)
}

type DeletePayload struct {
	ID int64 `json:"id"`
}

func (w *SyncWorker) handleLevelDeleted(payload string) {
	var p DeletePayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		slog.Error("failed to unmarshal level delete payload", "error", err)
		return
	}
	
	if err := w.tingkatRepo.DeleteByLMSID(p.ID); err != nil {
		slog.Error("failed to delete level", "id", p.ID, "error", err)
		return
	}
	slog.Info("Deleted Level from LMS", "id", p.ID)
}

func (w *SyncWorker) handleSubjectDeleted(payload string) {
	var p DeletePayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		slog.Error("failed to unmarshal subject delete payload", "error", err)
		return
	}
	
	if err := w.subjectRepo.DeleteByLMSID(p.ID); err != nil {
		slog.Error("failed to delete subject", "id", p.ID, "error", err)
		return
	}
	slog.Info("Deleted Subject from LMS", "id", p.ID)
}

func (w *SyncWorker) handleModuleDeleted(payload string) {
	var p DeletePayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		slog.Error("failed to unmarshal module delete payload", "error", err)
		return
	}
	
	if err := w.materiRepo.DeleteByLMSID(p.ID); err != nil {
		slog.Error("failed to delete module", "id", p.ID, "error", err)
		return
	}
	slog.Info("Deleted Module from LMS", "id", p.ID)
}

func (w *SyncWorker) handleUserDeleted(payload string) {
	var p DeletePayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		slog.Error("failed to unmarshal user delete payload", "error", err)
		return
	}
	
	if err := w.authRepo.DeleteUser(context.Background(), int32(p.ID)); err != nil {
		slog.Error("failed to delete user", "id", p.ID, "error", err)
		return
	}
	slog.Info("Deleted User from LMS", "id", p.ID)
}

type ClassPayload struct {
	ID       int64  `json:"id"`
	SchoolID int64  `json:"school_id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

func (w *SyncWorker) handleClassUpsert(payload string) {
	var p ClassPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		slog.Error("failed to unmarshal class payload", "error", err)
		return
	}
	
	if err := w.classRepo.UpsertByLMSID(p.ID, p.SchoolID, p.Name, p.IsActive); err != nil {
		slog.Error("failed to sync class", "id", p.ID, "error", err)
		return
	}
	slog.Info("Synced Class from LMS", "id", p.ID, "name", p.Name)
}

func (w *SyncWorker) handleClassDeleted(payload string) {
	var p DeletePayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		slog.Error("failed to unmarshal class delete payload", "error", err)
		return
	}
	
	if err := w.classRepo.DeleteByLMSID(p.ID); err != nil {
		slog.Error("failed to delete class", "id", p.ID, "error", err)
		return
	}
	slog.Info("Deleted Class from LMS", "id", p.ID)
}

// ClassStudentPayload represents the payload for class student events
type ClassStudentPayload struct {
	LMSClassID int64 `json:"lms_class_id"`
	LMSUserID  int64 `json:"lms_user_id"`
}

func (w *SyncWorker) handleClassStudentJoined(payload string) {
	var p ClassStudentPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		slog.Error("failed to unmarshal class_student_joined payload", "error", err)
		return
	}
	
	if err := w.classStudentRepo.AddStudent(p.LMSClassID, p.LMSUserID); err != nil {
		slog.Error("failed to add student to class", "class_id", p.LMSClassID, "user_id", p.LMSUserID, "error", err)
		return
	}
	slog.Info("Synced student join to class from LMS", "class_id", p.LMSClassID, "user_id", p.LMSUserID)
}

func (w *SyncWorker) handleClassStudentLeft(payload string) {
	var p ClassStudentPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		slog.Error("failed to unmarshal class_student_left payload", "error", err)
		return
	}
	
	if err := w.classStudentRepo.RemoveStudent(p.LMSClassID, p.LMSUserID); err != nil {
		slog.Error("failed to remove student from class", "class_id", p.LMSClassID, "user_id", p.LMSUserID, "error", err)
		return
	}
	slog.Info("Synced student leave from class from LMS", "class_id", p.LMSClassID, "user_id", p.LMSUserID)
}
