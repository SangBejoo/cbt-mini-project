package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
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

const (
	lmsEventsStream    = "lms_events"
	lmsEventsDLQStream = "lms_events_dlq"
	maxRetryCount      = 3
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
				Streams: []string{lmsEventsStream, lastID},
				Block:   5 * time.Second, // Block for 5 seconds max, then check context
				Count:   10,              // Process 10 messages at a time
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
					if err := w.processMessage(ctx, msg.Values); err != nil {
						slog.Error("failed to process LMS event", "message_id", msg.ID, "error", err)
						w.handleFailedMessage(ctx, msg, err)
					}
					lastID = msg.ID // Track last processed message ID
				}
			}
		}
	}
}

func (w *SyncWorker) processMessage(_ context.Context, data map[string]interface{}) error {
	eventType, _ := data["event"].(string)
	if eventType == "" {
		eventType, _ = data["type"].(string)
	}
	payload, _ := data["payload"].(string)
	if eventType == "" {
		return fmt.Errorf("missing event type")
	}
	if payload == "" {
		return fmt.Errorf("missing payload for event %s", eventType)
	}

	slog.Info("processing LMS event", "type", eventType)

	switch eventType {
	case "level_upsert":
		return w.handleLevelUpsert(payload)
	case "subject_upsert":
		return w.handleSubjectUpsert(payload)
	case "module_upsert":
		return w.handleModuleUpsert(payload)
	case "user_upsert":
		return w.handleUserUpsert(payload)
	case "exam_assignment_created":
		return w.handleExamAssignmentCreated(payload)
	case "exam_assignment_updated":
		return w.handleExamAssignmentUpdated(payload)
	case "exam_assignment_deleted":
		return w.handleExamAssignmentDeleted(payload)
	case "class_upsert":
		return w.handleClassUpsert(payload)
	case "class_deleted":
		return w.handleClassDeleted(payload)
	case "level_deleted":
		return w.handleLevelDeleted(payload)
	case "subject_deleted":
		return w.handleSubjectDeleted(payload)
	case "module_deleted":
		return w.handleModuleDeleted(payload)
	case "user_deleted":
		return w.handleUserDeleted(payload)
	case "class_student_joined":
		return w.handleClassStudentJoined(payload)
	case "class_student_left":
		return w.handleClassStudentLeft(payload)
	}

	slog.Warn("unknown LMS event type, skipping", "type", eventType)
	return nil
}

func (w *SyncWorker) handleFailedMessage(ctx context.Context, msg redis.XMessage, processErr error) {
	retryCount := parseRetryCount(msg.Values["retry_count"])
	eventType, _ := msg.Values["event"].(string)
	if eventType == "" {
		eventType, _ = msg.Values["type"].(string)
	}
	payload, _ := msg.Values["payload"].(string)

	if retryCount < maxRetryCount {
		values := map[string]interface{}{
			"event":           eventType,
			"type":            eventType,
			"payload":         payload,
			"retry_count":     retryCount + 1,
			"error":           processErr.Error(),
			"original_msg_id": msg.ID,
		}
		if err := infraRedis.RedisClient.XAdd(ctx, &redis.XAddArgs{
			Stream: lmsEventsStream,
			Values: values,
		}).Err(); err != nil {
			slog.Error("failed to requeue LMS event", "message_id", msg.ID, "retry_count", retryCount+1, "error", err)
		}
		return
	}

	dlqValues := map[string]interface{}{
		"event":           eventType,
		"type":            eventType,
		"payload":         payload,
		"error":           processErr.Error(),
		"retry_count":     retryCount,
		"original_msg_id": msg.ID,
		"failed_at":       time.Now().UTC().Format(time.RFC3339),
	}

	if err := infraRedis.RedisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: lmsEventsDLQStream,
		Values: dlqValues,
	}).Err(); err != nil {
		slog.Error("failed to write event to DLQ", "message_id", msg.ID, "error", err)
		return
	}

	slog.Error("moved LMS event to DLQ", "message_id", msg.ID, "event", eventType, "retry_count", retryCount)
}

func parseRetryCount(value interface{}) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		parsed, err := strconv.Atoi(typed)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

type LevelPayload struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	SchoolID int64  `json:"school_id"`
}

func (w *SyncWorker) handleLevelUpsert(payload string) error {
	slog.Info("Processing level upsert event", "payload", payload)
	var p LevelPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal level payload: %w", err)
	}

	slog.Info("Parsed level payload", "id", p.ID, "name", p.Name, "school_id", p.SchoolID)

	if err := w.tingkatRepo.UpsertByLMSID(p.ID, p.Name); err != nil {
		return fmt.Errorf("failed to sync level id=%d: %w", p.ID, err)
	}
	slog.Info("Synced Level from LMS", "id", p.ID, "name", p.Name, "school_id", p.SchoolID)
	return nil
}

type SubjectPayload struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	SchoolID int64  `json:"school_id"`
}

func (w *SyncWorker) handleSubjectUpsert(payload string) error {
	var p SubjectPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal subject payload: %w", err)
	}

	if err := w.subjectRepo.UpsertByLMSID(p.ID, p.Name, p.SchoolID); err != nil {
		return fmt.Errorf("failed to sync subject id=%d: %w", p.ID, err)
	}
	slog.Info("Synced Subject from LMS", "id", p.ID, "name", p.Name, "school_id", p.SchoolID)
	return nil
}

type ModulePayload struct {
	ID        int64  `json:"id"`
	SubjectID int64  `json:"subject_id"`
	LevelID   int64  `json:"level_id"`
	Name      string `json:"name"`
}

func (w *SyncWorker) handleModuleUpsert(payload string) error {
	var p ModulePayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal module payload: %w", err)
	}

	if err := w.materiRepo.UpsertByLMSID(p.ID, p.SubjectID, p.LevelID, p.Name); err != nil {
		return fmt.Errorf("failed to sync module id=%d: %w", p.ID, err)
	}
	slog.Info("Synced Module from LMS", "id", p.ID, "name", p.Name)
	return nil
}

type UserPayload struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	PasswordHash string `json:"password_hash"`
}

func (w *SyncWorker) handleUserUpsert(payload string) error {
	var p UserPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal user payload: %w", err)
	}

	roleInt := int32(1) // default SISWA
	if p.Role == "ADMIN" {
		roleInt = 2
	}

	_, err := w.authRepo.FindOrCreateByLMSID(context.Background(), p.ID, p.Email, p.Name, roleInt)
	if err != nil {
		return fmt.Errorf("failed to sync user id=%d: %w", p.ID, err)
	}
	slog.Info("Synced User from LMS", "id", p.ID, "email", p.Email)
	return nil
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

func (w *SyncWorker) handleExamAssignmentCreated(payload string) error {
	var p ExamAssignmentPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal exam assignment payload: %w", err)
	}

	if p.ModuleID == 0 {
		return fmt.Errorf("invalid module_id for assignment_id=%d", p.LMSAssignmentID)
	}

	// 1. Get module (materi) details to get duration and question count
	materi, err := w.materiRepo.GetByLMSID(p.ModuleID)
	if err != nil {
		return fmt.Errorf("failed to get materi details for module_id=%d: %w", p.ModuleID, err)
	}

	// 2. Get all students in the class
	studentIDs, err := w.classStudentRepo.GetStudentIDsByClassID(p.LMSClassID)
	if err != nil {
		return fmt.Errorf("failed to get students for class_id=%d: %w", p.LMSClassID, err)
	}

	// 3. Create test session for each student
	scheduledTime := parseScheduledTime(p.ScheduledTime)

	successCount := 0
	failureCount := 0
	var lastErr error
	for _, studentID := range studentIDs {
		created, err := w.testSessionRepo.CreateSessionForLMSUserIfMissing(
			p.LMSAssignmentID,
			p.LMSClassID,
			studentID,
			int(materi.IDMataPelajaran),
			int(materi.IDTingkat),
			materi.DefaultDurasiMenit,
			&materi.DefaultJumlahSoal,
			scheduledTime,
			entity.TestStatusScheduled,
		)
		if err != nil {
			slog.Error("failed to create test session", "assignment_id", p.LMSAssignmentID, "student_id", studentID, "error", err)
			failureCount++
			lastErr = err
			continue
		}
		if created {
			successCount++
		}
	}

	slog.Info("Synced Exam Assignment", "assignment_id", p.LMSAssignmentID, "class_id", p.LMSClassID, "sessions_created", successCount)
	if failureCount > 0 {
		return fmt.Errorf("failed creating %d sessions for assignment_id=%d: %w", failureCount, p.LMSAssignmentID, lastErr)
	}

	return nil
}

func (w *SyncWorker) handleExamAssignmentUpdated(payload string) error {
	var p ExamAssignmentPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal exam assignment updated payload: %w", err)
	}

	if p.LMSAssignmentID == 0 {
		return fmt.Errorf("missing lms_assignment_id on exam_assignment_updated")
	}

	if p.ModuleID == 0 {
		deleted, err := w.testSessionRepo.DeleteSessionsByAssignment(p.LMSAssignmentID)
		if err != nil {
			return fmt.Errorf("failed to delete sessions for assignment without CBT module id=%d: %w", p.LMSAssignmentID, err)
		}
		slog.Info("Deleted assignment sessions due to removed CBT component", "assignment_id", p.LMSAssignmentID, "deleted_sessions", deleted)
		return nil
	}

	materiData, err := w.materiRepo.GetByLMSID(p.ModuleID)
	if err != nil {
		return fmt.Errorf("failed to resolve module for assignment update module_id=%d: %w", p.ModuleID, err)
	}

	scheduledTime := parseScheduledTime(p.ScheduledTime)
	updatedRows, err := w.testSessionRepo.UpdateScheduledSessionsByAssignment(
		p.LMSAssignmentID,
		p.LMSClassID,
		int(materiData.IDMataPelajaran),
		int(materiData.IDTingkat),
		materiData.DefaultDurasiMenit,
		&materiData.DefaultJumlahSoal,
		scheduledTime,
	)
	if err != nil {
		return fmt.Errorf("failed to update scheduled sessions for assignment_id=%d: %w", p.LMSAssignmentID, err)
	}

	studentIDs, err := w.classStudentRepo.GetStudentIDsByClassID(p.LMSClassID)
	if err != nil {
		return fmt.Errorf("failed to get class students for class_id=%d: %w", p.LMSClassID, err)
	}

	createdCount := 0
	failureCount := 0
	var lastErr error
	for _, studentID := range studentIDs {
		created, createErr := w.testSessionRepo.CreateSessionForLMSUserIfMissing(
			p.LMSAssignmentID,
			p.LMSClassID,
			studentID,
			int(materiData.IDMataPelajaran),
			int(materiData.IDTingkat),
			materiData.DefaultDurasiMenit,
			&materiData.DefaultJumlahSoal,
			scheduledTime,
			entity.TestStatusScheduled,
		)
		if createErr != nil {
			slog.Error("failed to ensure session for assignment update", "assignment_id", p.LMSAssignmentID, "student_id", studentID, "error", createErr)
			failureCount++
			lastErr = createErr
			continue
		}
		if created {
			createdCount++
		}
	}

	slog.Info("Synced exam assignment update", "assignment_id", p.LMSAssignmentID, "updated_sessions", updatedRows, "created_sessions", createdCount)
	if failureCount > 0 {
		return fmt.Errorf("failed ensuring %d sessions on assignment update id=%d: %w", failureCount, p.LMSAssignmentID, lastErr)
	}

	return nil
}

func (w *SyncWorker) handleExamAssignmentDeleted(payload string) error {
	var p ExamAssignmentPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal exam assignment deleted payload: %w", err)
	}
	if p.LMSAssignmentID == 0 {
		return fmt.Errorf("missing lms_assignment_id on exam_assignment_deleted")
	}

	deletedRows, err := w.testSessionRepo.DeleteSessionsByAssignment(p.LMSAssignmentID)
	if err != nil {
		return fmt.Errorf("failed deleting sessions for assignment_id=%d: %w", p.LMSAssignmentID, err)
	}

	slog.Info("Deleted sessions for assignment", "assignment_id", p.LMSAssignmentID, "deleted_sessions", deletedRows)
	return nil
}

func parseScheduledTime(raw string) time.Time {
	if raw == "" {
		return time.Now()
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Now()
	}
	return parsed
}

type DeletePayload struct {
	ID int64 `json:"id"`
}

func (w *SyncWorker) handleLevelDeleted(payload string) error {
	var p DeletePayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal level delete payload: %w", err)
	}

	if err := w.tingkatRepo.DeleteByLMSID(p.ID); err != nil {
		return fmt.Errorf("failed to delete level id=%d: %w", p.ID, err)
	}
	slog.Info("Deleted Level from LMS", "id", p.ID)
	return nil
}

func (w *SyncWorker) handleSubjectDeleted(payload string) error {
	var p DeletePayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal subject delete payload: %w", err)
	}

	if err := w.subjectRepo.DeleteByLMSID(p.ID); err != nil {
		return fmt.Errorf("failed to delete subject id=%d: %w", p.ID, err)
	}
	slog.Info("Deleted Subject from LMS", "id", p.ID)
	return nil
}

func (w *SyncWorker) handleModuleDeleted(payload string) error {
	var p DeletePayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal module delete payload: %w", err)
	}

	if err := w.materiRepo.DeleteByLMSID(p.ID); err != nil {
		return fmt.Errorf("failed to delete module id=%d: %w", p.ID, err)
	}
	slog.Info("Deleted Module from LMS", "id", p.ID)
	return nil
}

func (w *SyncWorker) handleUserDeleted(payload string) error {
	var p DeletePayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal user delete payload: %w", err)
	}

	if err := w.authRepo.DeleteUser(context.Background(), int32(p.ID)); err != nil {
		return fmt.Errorf("failed to delete user id=%d: %w", p.ID, err)
	}
	slog.Info("Deleted User from LMS", "id", p.ID)
	return nil
}

type ClassPayload struct {
	ID       int64  `json:"id"`
	SchoolID int64  `json:"school_id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

func (w *SyncWorker) handleClassUpsert(payload string) error {
	var p ClassPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal class payload: %w", err)
	}

	if err := w.classRepo.UpsertByLMSID(p.ID, p.SchoolID, p.Name, p.IsActive); err != nil {
		return fmt.Errorf("failed to sync class id=%d: %w", p.ID, err)
	}
	slog.Info("Synced Class from LMS", "id", p.ID, "name", p.Name)
	return nil
}

func (w *SyncWorker) handleClassDeleted(payload string) error {
	var p DeletePayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal class delete payload: %w", err)
	}

	if err := w.classRepo.DeleteByLMSID(p.ID); err != nil {
		return fmt.Errorf("failed to delete class id=%d: %w", p.ID, err)
	}
	slog.Info("Deleted Class from LMS", "id", p.ID)
	return nil
}

// ClassStudentPayload represents the payload for class student events
type ClassStudentPayload struct {
	LMSClassID int64 `json:"lms_class_id"`
	LMSUserID  int64 `json:"lms_user_id"`
}

func (w *SyncWorker) handleClassStudentJoined(payload string) error {
	var p ClassStudentPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal class_student_joined payload: %w", err)
	}

	if err := w.classStudentRepo.AddStudent(p.LMSClassID, p.LMSUserID); err != nil {
		return fmt.Errorf("failed to add student to class class_id=%d user_id=%d: %w", p.LMSClassID, p.LMSUserID, err)
	}

	created, err := w.testSessionRepo.BackfillSessionsForJoinedStudent(p.LMSClassID, p.LMSUserID)
	if err != nil {
		return fmt.Errorf("failed to backfill sessions for joined student class_id=%d user_id=%d: %w", p.LMSClassID, p.LMSUserID, err)
	}
	slog.Info("Synced student join to class from LMS", "class_id", p.LMSClassID, "user_id", p.LMSUserID, "sessions_backfilled", created)
	return nil
}

func (w *SyncWorker) handleClassStudentLeft(payload string) error {
	var p ClassStudentPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal class_student_left payload: %w", err)
	}

	if err := w.classStudentRepo.RemoveStudent(p.LMSClassID, p.LMSUserID); err != nil {
		return fmt.Errorf("failed to remove student from class class_id=%d user_id=%d: %w", p.LMSClassID, p.LMSUserID, err)
	}
	slog.Info("Synced student leave from class from LMS", "class_id", p.LMSClassID, "user_id", p.LMSUserID)
	return nil
}
