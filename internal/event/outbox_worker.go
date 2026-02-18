package event

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	cbtOutboxStream        = "cbt_events"
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
	db     *sql.DB
	client *goredis.Client
}

func NewOutboxWorker(db *sql.DB, client *goredis.Client) *OutboxWorker {
	return &OutboxWorker{db: db, client: client}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	if w == nil || w.db == nil || w.client == nil {
		slog.Warn("cbt outbox worker disabled", "reason", "missing db or redis client")
		return
	}

	ticker := time.NewTicker(cbtOutboxPollInterval)
	defer ticker.Stop()

	slog.Info("CBT outbox worker started", "stream", cbtOutboxStream)

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
		err := w.publishRecord(ctx, rec)
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

func (w *OutboxWorker) publishRecord(ctx context.Context, rec outboxRecord) error {
	_, err := w.client.XAdd(ctx, &goredis.XAddArgs{
		Stream: cbtOutboxStream,
		Values: map[string]interface{}{
			"event":   rec.eventType,
			"type":    rec.eventType,
			"payload": rec.payload,
		},
	}).Result()
	if err != nil {
		return err
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
