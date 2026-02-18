package sync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	infraRedis "cbt-test-mini-project/init/infra/redis"
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/event/contracts"
	authRepo "cbt-test-mini-project/internal/repository/auth"
	classRepo "cbt-test-mini-project/internal/repository/class"
	classStudentRepo "cbt-test-mini-project/internal/repository/class_student"
	mataPelajaranRepo "cbt-test-mini-project/internal/repository/mata_pelajaran"
	"cbt-test-mini-project/internal/repository/materi"
	testSessionRepo "cbt-test-mini-project/internal/repository/test_session"
	"cbt-test-mini-project/internal/repository/tingkat"

	goredis "github.com/redis/go-redis/v9"
)

const (
	lmsEventsCriticalStream = "lms_events_critical"
	lmsEventsGeneralStream  = "lms_events_general"
	lmsEventsLegacyStream   = "lms_events"
	lmsEventsDLQStream   = "lms_events_dlq"
	lmsConsumerGroup     = "cbt_lms_sync_group"
	consumerReadBatch    = int64(25)
	consumerReadBlock    = 2 * time.Second
	consumerMinClaimIdle = 30 * time.Second
	maxRetryCount        = 3
	processedMessageTTL  = 7 * 24 * time.Hour
)

var lmsInputStreams = []string{
	lmsEventsCriticalStream,
	lmsEventsGeneralStream,
	lmsEventsLegacyStream,
}

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
	if err := w.ensureConsumerGroup(ctx); err != nil {
		slog.Error("failed to initialize consumer group", "error", err)
		return
	}

	consumerName := fmt.Sprintf("cbt-sync-consumer-%d", time.Now().UnixNano())

	for {
		select {
		case <-ctx.Done():
			return
		default:
			w.claimStalePending(ctx, consumerName)
			w.readOwnPending(ctx, consumerName)

			for _, streamName := range lmsInputStreams {
				streams, err := infraRedis.RedisClient.XReadGroup(ctx, &goredis.XReadGroupArgs{
					Group:    lmsConsumerGroup,
					Consumer: consumerName,
					Streams:  []string{streamName, ">"},
					Block:    consumerReadBlock,
					Count:    consumerReadBatch,
				}).Result()

				if err != nil {
					if errors.Is(err, goredis.Nil) {
						continue
					}
					if strings.Contains(err.Error(), "NOGROUP") {
						if groupErr := w.ensureConsumerGroup(ctx); groupErr != nil {
							slog.Error("failed to recreate consumer group", "error", groupErr)
						}
						continue
					}
					slog.Error("failed to read from redis stream group", "stream", streamName, "error", err)
					time.Sleep(2 * time.Second)
					continue
				}

				w.processStreamMessages(ctx, streams)
			}
		}
	}
}

func (w *SyncWorker) processStreamMessages(ctx context.Context, streams []goredis.XStream) {
	for _, stream := range streams {
		for _, msg := range stream.Messages {
			if err := w.processMessage(ctx, stream.Stream, msg); err != nil {
				slog.Error("failed to process LMS event", "message_id", msg.ID, "error", err)
				w.handleFailedMessage(ctx, stream.Stream, msg, err)
			}
		}
	}
}

func (w *SyncWorker) processMessage(ctx context.Context, streamName string, msg goredis.XMessage) error {
	eventType := extractEventType(msg.Values)
	payload := extractStringValue(msg.Values["payload"])
	if eventType == "" {
		return fmt.Errorf("missing event type")
	}
	if payload == "" {
		return fmt.Errorf("missing payload for event %s", eventType)
	}

	originalMessageID := getOriginalMessageID(msg.Values, msg.ID)
	processed, err := w.isMessageProcessed(ctx, originalMessageID)
	if err != nil {
		return fmt.Errorf("failed to check dedupe key: %w", err)
	}
	if processed {
		return w.ackMessage(ctx, streamName, msg.ID)
	}

	slog.Debug("processing LMS event", "type", eventType)

	var processErr error
	switch eventType {
	case "level_upsert":
		processErr = w.handleLevelUpsert(payload)
	case "subject_upsert":
		processErr = w.handleSubjectUpsert(payload)
	case string(contracts.ModuleUpsert):
		processErr = w.handleModuleUpsert(payload)
	case "user_upsert":
		processErr = w.handleUserUpsert(payload)
	case string(contracts.ExamAssignmentCreated):
		processErr = w.handleExamAssignmentCreated(payload)
	case string(contracts.ExamAssignmentUpdated):
		processErr = w.handleExamAssignmentUpdated(payload)
	case string(contracts.ExamAssignmentDeleted):
		processErr = w.handleExamAssignmentDeleted(payload)
	case string(contracts.ClassUpsert):
		processErr = w.handleClassUpsert(payload)
	case string(contracts.ClassDeleted):
		processErr = w.handleClassDeleted(payload)
	case "level_deleted":
		processErr = w.handleLevelDeleted(payload)
	case "subject_deleted":
		processErr = w.handleSubjectDeleted(payload)
	case string(contracts.ModuleDeleted):
		processErr = w.handleModuleDeleted(payload)
	case "user_deleted":
		processErr = w.handleUserDeleted(payload)
	case string(contracts.ClassStudentJoined):
		processErr = w.handleClassStudentJoined(payload)
	case string(contracts.ClassStudentLeft):
		processErr = w.handleClassStudentLeft(payload)
	default:
		slog.Warn("unknown LMS event type, skipping", "type", eventType)
	}

	if processErr != nil {
		return processErr
	}

	if err := w.markMessageProcessed(ctx, originalMessageID); err != nil {
		return fmt.Errorf("failed to mark message as processed: %w", err)
	}

	return w.ackMessage(ctx, streamName, msg.ID)
}

func (w *SyncWorker) handleFailedMessage(ctx context.Context, streamName string, msg goredis.XMessage, processErr error) {
	retryCount := parseRetryCount(msg.Values["retry_count"]) + 1
	eventType := extractEventType(msg.Values)
	payload := extractStringValue(msg.Values["payload"])
	originalMessageID := getOriginalMessageID(msg.Values, msg.ID)

	if retryCount <= maxRetryCount {
		values := map[string]interface{}{
			"event":           eventType,
			"type":            eventType,
			"payload":         payload,
			"retry_count":     retryCount,
			"error":           processErr.Error(),
			"original_msg_id": originalMessageID,
			"failed_at":       time.Now().UTC().Format(time.RFC3339),
		}
		if err := infraRedis.RedisClient.XAdd(ctx, &goredis.XAddArgs{
			Stream: streamName,
			Values: values,
		}).Err(); err != nil {
			slog.Error("failed to requeue LMS event", "message_id", msg.ID, "retry_count", retryCount, "error", err)
			return
		}
		if err := w.ackMessage(ctx, streamName, msg.ID); err != nil {
			slog.Error("failed to ack requeued LMS event", "message_id", msg.ID, "error", err)
		}
		return
	}

	dlqValues := map[string]interface{}{
		"event":           eventType,
		"type":            eventType,
		"payload":         payload,
		"error":           processErr.Error(),
		"retry_count":     retryCount,
		"original_msg_id": originalMessageID,
		"failed_at":       time.Now().UTC().Format(time.RFC3339),
	}

	if err := infraRedis.RedisClient.XAdd(ctx, &goredis.XAddArgs{
		Stream: lmsEventsDLQStream,
		Values: dlqValues,
	}).Err(); err != nil {
		slog.Error("failed to write event to DLQ", "message_id", msg.ID, "error", err)
		return
	}

	if err := w.ackMessage(ctx, streamName, msg.ID); err != nil {
		slog.Error("failed to ack DLQ event", "message_id", msg.ID, "error", err)
	}

	slog.Error("moved LMS event to DLQ", "message_id", msg.ID, "event", eventType, "retry_count", retryCount)
}

func (w *SyncWorker) ensureConsumerGroup(ctx context.Context) error {
	for _, streamName := range lmsInputStreams {
		err := infraRedis.RedisClient.XGroupCreateMkStream(ctx, streamName, lmsConsumerGroup, "0").Err()
		if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
			return err
		}
	}
	return nil
}

func (w *SyncWorker) readOwnPending(ctx context.Context, consumerName string) {
	for _, streamName := range lmsInputStreams {
		streams, err := infraRedis.RedisClient.XReadGroup(ctx, &goredis.XReadGroupArgs{
			Group:    lmsConsumerGroup,
			Consumer: consumerName,
			Streams:  []string{streamName, "0"},
			Count:    consumerReadBatch,
		}).Result()
		if err != nil && !errors.Is(err, goredis.Nil) {
			slog.Warn("failed reading own pending entries", "stream", streamName, "error", err)
			continue
		}
		if len(streams) > 0 {
			w.processStreamMessages(ctx, streams)
		}
	}
}

func (w *SyncWorker) claimStalePending(ctx context.Context, consumerName string) {
	for _, streamName := range lmsInputStreams {
		w.claimStalePendingFromStream(ctx, consumerName, streamName)
	}
}

func (w *SyncWorker) claimStalePendingFromStream(ctx context.Context, consumerName, streamName string) {
	start := "0-0"
	for {
		messages, next, err := infraRedis.RedisClient.XAutoClaim(ctx, &goredis.XAutoClaimArgs{
			Stream:   streamName,
			Group:    lmsConsumerGroup,
			Consumer: consumerName,
			MinIdle:  consumerMinClaimIdle,
			Start:    start,
			Count:    consumerReadBatch,
		}).Result()
		if err != nil && !errors.Is(err, goredis.Nil) {
			slog.Warn("failed auto-claiming pending entries", "stream", streamName, "error", err)
			return
		}
		if len(messages) == 0 {
			return
		}

		w.processStreamMessages(ctx, []goredis.XStream{{Stream: streamName, Messages: messages}})

		if next == "0-0" || next == start {
			return
		}
		start = next
	}
}

func (w *SyncWorker) ackMessage(ctx context.Context, streamName, messageID string) error {
	if err := infraRedis.RedisClient.XAck(ctx, streamName, lmsConsumerGroup, messageID).Err(); err != nil {
		return err
	}
	if err := infraRedis.RedisClient.XDel(ctx, streamName, messageID).Err(); err != nil {
		slog.Warn("failed to delete acknowledged stream message", "message_id", messageID, "error", err)
	}
	return nil
}

func (w *SyncWorker) processedKey(originalMessageID string) string {
	return "lms_events:processed:" + originalMessageID
}

func (w *SyncWorker) isMessageProcessed(ctx context.Context, originalMessageID string) (bool, error) {
	result, err := infraRedis.RedisClient.Exists(ctx, w.processedKey(originalMessageID)).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (w *SyncWorker) markMessageProcessed(ctx context.Context, originalMessageID string) error {
	_, err := infraRedis.RedisClient.SetNX(ctx, w.processedKey(originalMessageID), "1", processedMessageTTL).Result()
	return err
}

func extractEventType(values map[string]interface{}) string {
	eventType := extractStringValue(values["event"])
	if eventType == "" {
		eventType = extractStringValue(values["type"])
	}
	return eventType
}

func extractStringValue(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		if value == nil {
			return ""
		}
		return fmt.Sprint(value)
	}
}

func getOriginalMessageID(values map[string]interface{}, fallback string) string {
	original := extractStringValue(values["original_msg_id"])
	if original != "" {
		return original
	}
	return fallback
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
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err != nil {
			return 0
		}
		return parsed
	default:
		parsed, err := strconv.Atoi(strings.TrimSpace(fmt.Sprint(value)))
		if err != nil {
			return 0
		}
		return parsed
	}
}

type LevelPayload struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	SchoolID int64  `json:"school_id"`
}

func (w *SyncWorker) handleLevelUpsert(payload string) error {
	slog.Debug("processing level upsert event", "payload", payload)
	var p LevelPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal level payload: %w", err)
	}

	slog.Debug("parsed level payload", "id", p.ID, "name", p.Name, "school_id", p.SchoolID)

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

func (w *SyncWorker) handleModuleUpsert(payload string) error {
	var p contracts.ModuleUpsertPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal module payload: %w", err)
	}

	if err := w.materiRepo.UpsertByLMSID(p.ID, p.SubjectID, p.LevelID, p.ClassID, p.Name); err != nil {
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

func (w *SyncWorker) handleExamAssignmentCreated(payload string) error {
	var p contracts.ExamAssignmentPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal exam assignment payload: %w", err)
	}

	if p.ModuleID == 0 {
		return fmt.Errorf("invalid module_id for assignment_id=%d", p.LMSAssignmentID)
	}

	// 1. Resolve module reference to materi details
	materi, err := w.resolveMateriByModuleReference(p.ModuleID, p.ModuleRefType)
	if err != nil {
		return fmt.Errorf("failed to resolve materi details for module_id=%d: %w", p.ModuleID, err)
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
	var p contracts.ExamAssignmentPayload
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

	materiData, err := w.resolveMateriByModuleReference(p.ModuleID, p.ModuleRefType)
	if err != nil {
		return fmt.Errorf("failed to resolve materi for assignment update module_id=%d: %w", p.ModuleID, err)
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
	var p contracts.ExamAssignmentPayload
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

func (w *SyncWorker) resolveMateriByModuleReference(moduleID int64, moduleRefType string) (*entity.Materi, error) {
	if moduleID == 0 {
		return nil, fmt.Errorf("module_id is required")
	}

	normalizedRefType := strings.ToLower(strings.TrimSpace(moduleRefType))

	switch normalizedRefType {
	case "lms_module_id", "teacher_material_id":
		return w.materiRepo.GetByLMSID(moduleID)
	case "cbt_materi_id":
		return w.materiRepo.GetByID(int(moduleID))
	}

	if materiData, err := w.materiRepo.GetByLMSID(moduleID); err == nil {
		return materiData, nil
	}

	materiData, err := w.materiRepo.GetByID(int(moduleID))
	if err != nil {
		return nil, err
	}

	return materiData, nil
}

func (w *SyncWorker) handleLevelDeleted(payload string) error {
	var p contracts.DeletePayload
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
	var p contracts.DeletePayload
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
	var p contracts.DeletePayload
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
	var p contracts.DeletePayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal user delete payload: %w", err)
	}

	if err := w.authRepo.DeleteUser(context.Background(), int32(p.ID)); err != nil {
		return fmt.Errorf("failed to delete user id=%d: %w", p.ID, err)
	}
	slog.Info("Deleted User from LMS", "id", p.ID)
	return nil
}

func (w *SyncWorker) handleClassUpsert(payload string) error {
	var p contracts.ClassPayload
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
	var p contracts.DeletePayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal class delete payload: %w", err)
	}

	if err := w.classRepo.DeleteByLMSID(p.ID); err != nil {
		return fmt.Errorf("failed to delete class id=%d: %w", p.ID, err)
	}
	slog.Info("Deleted Class from LMS", "id", p.ID)
	return nil
}

func (w *SyncWorker) handleClassStudentJoined(payload string) error {
	var p contracts.ClassStudentPayload
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
	var p contracts.ClassStudentPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return fmt.Errorf("failed to unmarshal class_student_left payload: %w", err)
	}

	if err := w.classStudentRepo.RemoveStudent(p.LMSClassID, p.LMSUserID); err != nil {
		return fmt.Errorf("failed to remove student from class class_id=%d user_id=%d: %w", p.LMSClassID, p.LMSUserID, err)
	}
	slog.Info("Synced student leave from class from LMS", "class_id", p.LMSClassID, "user_id", p.LMSUserID)
	return nil
}
