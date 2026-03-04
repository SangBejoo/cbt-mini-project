package event

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"cbt-test-mini-project/internal/event/contracts"
)

const (
	cbtOutboxBatchSize     = 20
	cbtOutboxPollInterval  = 2 * time.Second
	cbtOutboxMaxRetryCount = 8
)

type outboxRecord struct {
	id         int64
	eventType  string
	payload    string
	retryCount int
}

type OutboxWorker struct {
	db *sql.DB
}

func NewOutboxWorker(db *sql.DB) *OutboxWorker {
	return &OutboxWorker{db: db}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	if w == nil || w.db == nil {
		slog.Warn("cbt outbox worker disabled", "reason", "missing db")
		return
	}

	ticker := time.NewTicker(cbtOutboxPollInterval)
	defer ticker.Stop()

	slog.Info("CBT outbox worker started", "mode", "db-projector")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.drainPending(ctx); err != nil {
				slog.Error("cbt outbox drain failed", "error", err)
			}
		}
	}
}

func (w *OutboxWorker) drainPending(ctx context.Context) error {
	records, err := w.claimPending(ctx, cbtOutboxBatchSize)
	if err != nil {
		return err
	}

	for _, rec := range records {
		err := w.projectRecord(ctx, rec)
		if err != nil {
			if failErr := w.markFailed(ctx, rec.id, rec.retryCount, err); failErr != nil {
				slog.Error("failed to mark outbox record as failed", "id", rec.id, "error", failErr)
			}
			continue
		}

		if err := w.markSent(ctx, rec.id); err != nil {
			slog.Error("failed to mark outbox record as sent", "id", rec.id, "error", err)
		}
	}

	return nil
}

func (w *OutboxWorker) claimPending(ctx context.Context, limit int) ([]outboxRecord, error) {
	if limit <= 0 {
		limit = cbtOutboxBatchSize
	}

	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	query := `
		WITH pick AS (
			SELECT id
			FROM cbt_outbox
			WHERE status IN ('pending', 'failed')
			  AND (next_attempt_at IS NULL OR next_attempt_at <= NOW())
			ORDER BY id ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE cbt_outbox o
		SET status = 'processing',
			updated_at = NOW()
		FROM pick
		WHERE o.id = pick.id
		RETURNING o.id, o.event_type, o.payload::text, o.retry_count
	`

	rows, err := tx.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]outboxRecord, 0)
	for rows.Next() {
		var rec outboxRecord
		if err := rows.Scan(&rec.id, &rec.eventType, &rec.payload, &rec.retryCount); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true

	return records, nil
}

func (w *OutboxWorker) projectRecord(ctx context.Context, rec outboxRecord) error {
	if rec.eventType != string(contracts.ExamResultCompleted) {
		return nil
	}

	var payload contracts.ExamResultPayload
	if err := json.Unmarshal([]byte(rec.payload), &payload); err != nil {
		return fmt.Errorf("failed to parse outbox payload: %w", err)
	}

	if payload.AssignmentID == 0 || payload.UserID == 0 {
		return nil
	}

	membershipID, err := w.getActiveStudentMembershipID(ctx, payload.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("failed to resolve student membership: %w", err)
	}

	completedAt := time.Now().UTC()
	if payload.CompletedAt != "" {
		parsed, parseErr := time.Parse(time.RFC3339, payload.CompletedAt)
		if parseErr == nil {
			completedAt = parsed.UTC()
		}
	}

	if err := w.upsertAssignmentAttemptScore(ctx, payload.AssignmentID, membershipID, payload.Score, completedAt); err != nil {
		return fmt.Errorf("failed to upsert assignment attempt score: %w", err)
	}

	if err := w.upsertGradebookEntry(ctx, payload.AssignmentID, membershipID, payload.Score, completedAt); err != nil {
		return fmt.Errorf("failed to upsert gradebook entry: %w", err)
	}

	return nil
}

func (w *OutboxWorker) markSent(ctx context.Context, id int64) error {
	_, err := w.db.ExecContext(ctx, `
		UPDATE cbt_outbox
		SET status = 'sent',
			sent_at = NOW(),
			last_error = NULL,
			updated_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

func (w *OutboxWorker) markFailed(ctx context.Context, id int64, retryCount int, publishErr error) error {
	nextRetry := retryDelay(retryCount + 1)
	status := "failed"
	if retryCount+1 >= cbtOutboxMaxRetryCount {
		status = "dead"
	}

	_, err := w.db.ExecContext(ctx, `
		UPDATE cbt_outbox
		SET status = $2,
			retry_count = retry_count + 1,
			last_error = $3,
			next_attempt_at = $4,
			updated_at = NOW()
		WHERE id = $1
	`, id, status, truncateError(publishErr), time.Now().Add(nextRetry))
	return err
}

func retryDelay(retryCount int) time.Duration {
	if retryCount <= 1 {
		return 5 * time.Second
	}
	if retryCount <= 3 {
		return 15 * time.Second
	}
	if retryCount <= 5 {
		return 1 * time.Minute
	}
	return 5 * time.Minute
}

func truncateError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if len(msg) <= 1000 {
		return msg
	}
	return fmt.Sprintf("%s...", msg[:997])
}

func (w *OutboxWorker) getActiveStudentMembershipID(ctx context.Context, userID int64) (int64, error) {
	query := `
		SELECT sm.id
		FROM public.school_memberships sm
		WHERE sm.user_id = $1
		  AND sm.role = 'student'
		  AND sm.status = 'active'
		  AND sm.deleted_at IS NULL
		ORDER BY sm.created_at DESC
		LIMIT 1
	`

	var membershipID int64
	if err := w.db.QueryRowContext(ctx, query, userID).Scan(&membershipID); err != nil {
		return 0, err
	}

	return membershipID, nil
}

func (w *OutboxWorker) upsertAssignmentAttemptScore(ctx context.Context, assignmentID, studentMembershipID int64, score float64, submittedAt time.Time) error {
	var maxScore float64
	var assetID sql.NullInt64
	err := w.db.QueryRowContext(ctx,
		`SELECT a.max_score, ac.asset_id
		 FROM public.assignments a
		 LEFT JOIN public.assignment_components ac ON ac.assignment_id = a.id
		 WHERE a.id = $1
		 ORDER BY CASE WHEN ac.type = 'cbt_exam' THEN 0 ELSE 1 END, ac.order_no ASC
		 LIMIT 1`,
		assignmentID,
	).Scan(&maxScore, &assetID)
	if err != nil {
		return err
	}

	var latestAttemptID int64
	err = w.db.QueryRowContext(ctx,
		`SELECT id
		 FROM public.assignment_attempts
		 WHERE assignment_id = $1 AND student_membership_id = $2
		 ORDER BY attempt_no DESC
		 LIMIT 1`,
		assignmentID,
		studentMembershipID,
	).Scan(&latestAttemptID)

	if errors.Is(err, sql.ErrNoRows) {
		var nextAttempt int
		if err := w.db.QueryRowContext(ctx,
			`SELECT COALESCE(MAX(attempt_no), 0) + 1
			 FROM public.assignment_attempts
			 WHERE assignment_id = $1 AND student_membership_id = $2`,
			assignmentID,
			studentMembershipID,
		).Scan(&nextAttempt); err != nil {
			return err
		}

		var assetArg interface{}
		if assetID.Valid {
			assetArg = assetID.Int64
		}

		_, err = w.db.ExecContext(ctx,
			`INSERT INTO public.assignment_attempts (
				assignment_id, asset_id, student_membership_id, attempt_no,
				started_at, submitted_at, raw_score, max_score, status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			assignmentID,
			assetArg,
			studentMembershipID,
			nextAttempt,
			submittedAt,
			submittedAt,
			score,
			maxScore,
			"graded",
		)
		return err
	}

	if err != nil {
		return err
	}

	_, err = w.db.ExecContext(ctx,
		`UPDATE public.assignment_attempts
		 SET raw_score = $2,
		     max_score = $3,
		     status = 'graded',
		     submitted_at = COALESCE(submitted_at, $4),
		     asset_id = COALESCE(asset_id, $5)
		 WHERE id = $1`,
		latestAttemptID,
		score,
		maxScore,
		submittedAt,
		assetID,
	)
	return err
}

func (w *OutboxWorker) upsertGradebookEntry(ctx context.Context, assignmentID, studentMembershipID int64, score float64, gradedAt time.Time) error {
	updateQuery := `
		UPDATE public.gradebook_entries
		SET score = $3,
		    status = $4,
		    computed_from = $5,
		    graded_at = $6,
		    updated_at = $7
		WHERE assignment_id = $1
		  AND student_membership_id = $2
		  AND deleted_at IS NULL
	`

	result, err := w.db.ExecContext(ctx, updateQuery,
		assignmentID,
		studentMembershipID,
		score,
		"graded",
		"cbt_auto",
		gradedAt,
		time.Now().UTC(),
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		_, err = w.db.ExecContext(ctx,
			`INSERT INTO public.gradebook_entries (
				assignment_id, student_membership_id, score, status,
				computed_from, graded_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			assignmentID,
			studentMembershipID,
			score,
			"graded",
			"cbt_auto",
			gradedAt,
			time.Now().UTC(),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
