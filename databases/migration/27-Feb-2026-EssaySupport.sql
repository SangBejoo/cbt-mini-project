-- Migration: Essay support and grading lifecycle
-- Date: 27-Feb-2026

-- 1) Extend question/session enums (legacy schema)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_type t
        JOIN pg_enum e ON t.oid = e.enumtypid
        WHERE t.typname = 'question_type_enum' AND e.enumlabel = 'essay'
    ) THEN
        ALTER TYPE question_type_enum ADD VALUE 'essay';
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_type t
        JOIN pg_enum e ON t.oid = e.enumtypid
        WHERE t.typname = 'test_session_status_enum' AND e.enumlabel = 'grading_in_progress'
    ) THEN
        ALTER TYPE test_session_status_enum ADD VALUE 'grading_in_progress';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_type t
        JOIN pg_enum e ON t.oid = e.enumtypid
        WHERE t.typname = 'test_session_status_enum' AND e.enumlabel = 'graded'
    ) THEN
        ALTER TYPE test_session_status_enum ADD VALUE 'graded';
    END IF;
END
$$;

-- 2) Add essay fields on soal and jawaban_siswa (legacy runtime tables)
ALTER TABLE soal
    ADD COLUMN IF NOT EXISTS question_type question_type_enum NOT NULL DEFAULT 'multiple_choice',
    ADD COLUMN IF NOT EXISTS jawaban_essay_key TEXT;

ALTER TABLE jawaban_siswa
    ADD COLUMN IF NOT EXISTS jawaban_essay TEXT,
    ADD COLUMN IF NOT EXISTS nilai_essay DECIMAL(5,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS feedback_teacher TEXT;

-- 3) Align question type to avoid null/mixed values
UPDATE soal
SET question_type = 'multiple_choice'
WHERE question_type IS NULL;

CREATE INDEX IF NOT EXISTS idx_soal_question_type ON soal (question_type);
CREATE INDEX IF NOT EXISTS idx_jawaban_siswa_essay_score ON jawaban_siswa (nilai_essay);
