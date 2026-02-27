-- Migration: Role alignment + ownership linkage support + weighted question points
-- Date: 27-Feb-2026

-- ============================================================
-- 1) ROLE NORMALIZATION
-- ============================================================

DO $$
BEGIN
    -- If modern enum exists, ensure values are present
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role_enum') THEN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_type t
            JOIN pg_enum e ON t.oid = e.enumtypid
            WHERE t.typname = 'user_role_enum' AND e.enumlabel = 'teacher'
        ) THEN
            ALTER TYPE user_role_enum ADD VALUE 'teacher';
        END IF;

        IF NOT EXISTS (
            SELECT 1
            FROM pg_type t
            JOIN pg_enum e ON t.oid = e.enumtypid
            WHERE t.typname = 'user_role_enum' AND e.enumlabel = 'superadmin'
        ) THEN
            ALTER TYPE user_role_enum ADD VALUE 'superadmin';
        END IF;
    END IF;
END
$$;

-- Best-effort normalization for legacy role labels
UPDATE users
SET role = CASE
    WHEN role::text IN ('admin', 'ADMIN') THEN 'superadmin'::user_role_enum
    WHEN role::text IN ('teacher', 'TEACHER', 'guru', 'GURU') THEN 'teacher'::user_role_enum
    WHEN role::text IN ('siswa', 'SISWA', 'student', 'STUDENT') THEN 'student'::user_role_enum
    ELSE role
END
WHERE role::text IN ('admin','ADMIN','teacher','TEACHER','guru','GURU','siswa','SISWA','student','STUDENT');

-- ============================================================
-- 2) QUESTION POINTS (weighted scoring)
-- ============================================================

-- Materi question selection mode
ALTER TABLE IF EXISTS materials
    ADD COLUMN IF NOT EXISTS randomize_questions BOOLEAN NOT NULL DEFAULT TRUE;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relname = 'materi' AND c.relkind IN ('r', 'p')
    ) THEN
        ALTER TABLE materi ADD COLUMN IF NOT EXISTS randomize_questions BOOLEAN NOT NULL DEFAULT TRUE;
    END IF;
END
$$;

-- Always apply on modern base tables
ALTER TABLE IF EXISTS questions
    ADD COLUMN IF NOT EXISTS point NUMERIC(10,2) NOT NULL DEFAULT 1;

ALTER TABLE IF EXISTS drag_drop_questions
    ADD COLUMN IF NOT EXISTS point NUMERIC(10,2) NOT NULL DEFAULT 1;

ALTER TABLE IF EXISTS exam_session_questions
    ADD COLUMN IF NOT EXISTS point NUMERIC(10,2) NOT NULL DEFAULT 1;

ALTER TABLE IF EXISTS questions
    ADD COLUMN IF NOT EXISTS urutan INTEGER NOT NULL DEFAULT 0;

ALTER TABLE IF EXISTS drag_drop_questions
    ADD COLUMN IF NOT EXISTS urutan INTEGER NOT NULL DEFAULT 0;

-- Apply on legacy tables only when they are real tables (not views)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relname = 'soal' AND c.relkind IN ('r', 'p')
    ) THEN
        ALTER TABLE soal ADD COLUMN IF NOT EXISTS point NUMERIC(10,2) NOT NULL DEFAULT 1;
        ALTER TABLE soal ADD COLUMN IF NOT EXISTS urutan INTEGER NOT NULL DEFAULT 0;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relname = 'soal_drag_drop' AND c.relkind IN ('r', 'p')
    ) THEN
        ALTER TABLE soal_drag_drop ADD COLUMN IF NOT EXISTS point NUMERIC(10,2) NOT NULL DEFAULT 1;
        ALTER TABLE soal_drag_drop ADD COLUMN IF NOT EXISTS urutan INTEGER NOT NULL DEFAULT 0;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relname = 'test_session_soal' AND c.relkind IN ('r', 'p')
    ) THEN
        ALTER TABLE test_session_soal ADD COLUMN IF NOT EXISTS point NUMERIC(10,2) NOT NULL DEFAULT 1;
    END IF;
END
$$;

-- Backfill on modern tables
UPDATE questions SET point = 1 WHERE point <= 0 OR point IS NULL;
UPDATE drag_drop_questions SET point = 1 WHERE point <= 0 OR point IS NULL;
UPDATE exam_session_questions SET point = 1 WHERE point <= 0 OR point IS NULL;
UPDATE questions SET urutan = id WHERE urutan <= 0 OR urutan IS NULL;
UPDATE drag_drop_questions SET urutan = id WHERE urutan <= 0 OR urutan IS NULL;

UPDATE exam_session_questions esq
SET point = COALESCE(src.point, 1)
FROM (
    SELECT id, point FROM questions
) src
WHERE esq.question_id = src.id
  AND (esq.point IS NULL OR esq.point <= 0);

UPDATE exam_session_questions esq
SET point = COALESCE(src.point, 1)
FROM (
    SELECT id, point FROM drag_drop_questions
) src
WHERE esq.drag_drop_question_id = src.id
  AND (esq.point IS NULL OR esq.point <= 0);

-- Backfill on legacy table only when it is a real table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relname = 'test_session_soal' AND c.relkind IN ('r', 'p')
    ) THEN
        UPDATE test_session_soal tss
        SET point = COALESCE(src.point, 1)
        FROM (
            SELECT id, point FROM soal
        ) src
        WHERE tss.id_soal = src.id
          AND (tss.point IS NULL OR tss.point <= 0);

        UPDATE test_session_soal tss
        SET point = COALESCE(src.point, 1)
        FROM (
            SELECT id, point FROM soal_drag_drop
        ) src
        WHERE tss.id_soal_drag_drop = src.id
          AND (tss.point IS NULL OR tss.point <= 0);

        UPDATE test_session_soal SET point = 1 WHERE point <= 0 OR point IS NULL;
    END IF;
END
$$;

-- Optional performance helpers on base tables
CREATE INDEX IF NOT EXISTS idx_questions_point ON questions (point);
CREATE INDEX IF NOT EXISTS idx_drag_drop_questions_point ON drag_drop_questions (point);
CREATE INDEX IF NOT EXISTS idx_exam_session_questions_point ON exam_session_questions (point);
CREATE INDEX IF NOT EXISTS idx_questions_urutan ON questions (material_id, urutan);
CREATE INDEX IF NOT EXISTS idx_drag_drop_questions_urutan ON drag_drop_questions (material_id, urutan);

-- Optional performance helpers on legacy tables only when table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relname = 'soal' AND c.relkind IN ('r', 'p')
    ) THEN
        CREATE INDEX IF NOT EXISTS idx_soal_point ON soal (point);
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relname = 'soal_drag_drop' AND c.relkind IN ('r', 'p')
    ) THEN
        CREATE INDEX IF NOT EXISTS idx_soal_drag_drop_point ON soal_drag_drop (point);
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public' AND c.relname = 'test_session_soal' AND c.relkind IN ('r', 'p')
    ) THEN
        CREATE INDEX IF NOT EXISTS idx_test_session_soal_point ON test_session_soal (point);
    END IF;
END
$$;
