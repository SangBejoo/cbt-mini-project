-- Transactional outbox for reliable CBT -> LMS event publishing
CREATE TABLE IF NOT EXISTS cbt_outbox (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    aggregate_type VARCHAR(100),
    aggregate_id BIGINT,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cbt_outbox_status_next_attempt
    ON cbt_outbox (status, next_attempt_at, id);

CREATE INDEX IF NOT EXISTS idx_cbt_outbox_event_type
    ON cbt_outbox (event_type);
