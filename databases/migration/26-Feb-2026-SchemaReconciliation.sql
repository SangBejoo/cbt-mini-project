-- Migration: Reconcile schema with current repository implementation
-- Date: 26-Feb-2026
-- Purpose: Align columns/indexes used by runtime SQL (LMS sync + class-student sync)

-- Ensure enum supports scheduled sessions
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_type t
        JOIN pg_enum e ON t.oid = e.enumtypid
        WHERE t.typname = 'test_session_status_enum' AND e.enumlabel = 'scheduled'
    ) THEN
        ALTER TYPE test_session_status_enum ADD VALUE 'scheduled';
    END IF;
END
$$;

-- users
ALTER TABLE users ADD COLUMN IF NOT EXISTS lms_user_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_users_lms_id ON users (lms_user_id);

-- mata_pelajaran
ALTER TABLE mata_pelajaran ADD COLUMN IF NOT EXISTS lms_subject_id BIGINT;
ALTER TABLE mata_pelajaran ADD COLUMN IF NOT EXISTS lms_school_id BIGINT;
ALTER TABLE mata_pelajaran ADD COLUMN IF NOT EXISTS lms_class_id BIGINT;
CREATE UNIQUE INDEX IF NOT EXISTS uq_mata_pelajaran_lms_subject_id ON mata_pelajaran (lms_subject_id) WHERE lms_subject_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_mp_school ON mata_pelajaran (lms_school_id);
CREATE INDEX IF NOT EXISTS idx_mp_class ON mata_pelajaran (lms_class_id);

-- tingkat
ALTER TABLE tingkat ADD COLUMN IF NOT EXISTS lms_level_id BIGINT;
ALTER TABLE tingkat ADD COLUMN IF NOT EXISTS lms_school_id BIGINT;
CREATE UNIQUE INDEX IF NOT EXISTS uq_tingkat_lms_level_id ON tingkat (lms_level_id) WHERE lms_level_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tingkat_school ON tingkat (lms_school_id);

-- materi
ALTER TABLE materi ADD COLUMN IF NOT EXISTS lms_module_id BIGINT;
ALTER TABLE materi ADD COLUMN IF NOT EXISTS lms_class_id BIGINT;
ALTER TABLE materi ADD COLUMN IF NOT EXISTS owner_user_id INTEGER;
ALTER TABLE materi ADD COLUMN IF NOT EXISTS school_id BIGINT;
ALTER TABLE materi ADD COLUMN IF NOT EXISTS labels JSONB NOT NULL DEFAULT '[]'::jsonb;
CREATE UNIQUE INDEX IF NOT EXISTS uq_materi_lms_module_id ON materi (lms_module_id) WHERE lms_module_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_materi_class ON materi (lms_class_id);
CREATE INDEX IF NOT EXISTS idx_materi_school ON materi (school_id);

-- classes table used by LMS sync worker
CREATE TABLE IF NOT EXISTS classes (
    id SERIAL PRIMARY KEY,
    lms_class_id BIGINT UNIQUE NOT NULL,
    lms_school_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_classes_school ON classes (lms_school_id);
CREATE INDEX IF NOT EXISTS idx_classes_active ON classes (is_active);

-- class_students used by class_student repository
CREATE TABLE IF NOT EXISTS class_students (
    id SERIAL PRIMARY KEY,
    lms_class_id BIGINT,
    lms_user_id BIGINT,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE class_students ADD COLUMN IF NOT EXISTS lms_class_id BIGINT;
ALTER TABLE class_students ADD COLUMN IF NOT EXISTS lms_user_id BIGINT;
ALTER TABLE class_students ADD COLUMN IF NOT EXISTS joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP;

CREATE UNIQUE INDEX IF NOT EXISTS uq_class_students_lms_pair ON class_students (lms_class_id, lms_user_id)
WHERE lms_class_id IS NOT NULL AND lms_user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_class_students_lms_class ON class_students (lms_class_id);
CREATE INDEX IF NOT EXISTS idx_class_students_lms_user ON class_students (lms_user_id);

-- test_session LMS linkage
ALTER TABLE test_session ADD COLUMN IF NOT EXISTS lms_assignment_id BIGINT;
ALTER TABLE test_session ADD COLUMN IF NOT EXISTS lms_class_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_test_session_lms_assignment_id ON test_session (lms_assignment_id);
CREATE INDEX IF NOT EXISTS idx_session_class ON test_session (lms_class_id);
