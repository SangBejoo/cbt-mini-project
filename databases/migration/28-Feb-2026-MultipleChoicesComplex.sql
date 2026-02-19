-- Migration: Add schema support for multiple choices complex (checkbox style)
-- Date: 28-Feb-2026

-- 1) Extend question type enum
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_type t
        WHERE t.typname = 'question_type_enum'
    ) AND NOT EXISTS (
        SELECT 1
        FROM pg_type t
        JOIN pg_enum e ON t.oid = e.enumtypid
        WHERE t.typname = 'question_type_enum' AND e.enumlabel = 'multiple_choices_complex'
    ) THEN
        ALTER TYPE question_type_enum ADD VALUE 'multiple_choices_complex';
    END IF;
END
$$;

-- 2) English schema tables
ALTER TABLE IF EXISTS questions
    ADD COLUMN IF NOT EXISTS correct_options_complex JSONB;

ALTER TABLE IF EXISTS student_answers
    ADD COLUMN IF NOT EXISTS selected_options_complex JSONB;

CREATE INDEX IF NOT EXISTS idx_questions_correct_options_complex ON questions USING GIN (correct_options_complex);
CREATE INDEX IF NOT EXISTS idx_student_answers_selected_options_complex ON student_answers USING GIN (selected_options_complex);

-- 3) Legacy runtime tables (only when they are actual tables, not compatibility views)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relname = 'soal' AND c.relkind IN ('r', 'p')
    ) THEN
        ALTER TABLE soal ADD COLUMN IF NOT EXISTS jawaban_benar_complex JSONB;
        CREATE INDEX IF NOT EXISTS idx_soal_jawaban_benar_complex ON soal USING GIN (jawaban_benar_complex);
    END IF;
END
$$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relname = 'jawaban_siswa' AND c.relkind IN ('r', 'p')
    ) THEN
        ALTER TABLE jawaban_siswa ADD COLUMN IF NOT EXISTS jawaban_dipilih_complex JSONB;
        CREATE INDEX IF NOT EXISTS idx_jawaban_siswa_dipilih_complex ON jawaban_siswa USING GIN (jawaban_dipilih_complex);
    END IF;
END
$$;
