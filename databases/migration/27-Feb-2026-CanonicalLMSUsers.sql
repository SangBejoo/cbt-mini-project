-- Migration: Canonical LMS users for CBT
-- Date: 27-Feb-2026
-- Purpose:
-- 1) Move CBT user ownership to LMS public.users
-- 2) Derive CBT role from LMS school_memberships
-- 3) Keep compatibility via cbt.users view for existing CBT queries

CREATE SCHEMA IF NOT EXISTS cbt;

-- Keep a copy of legacy CBT users table for audit/backfill if present.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'cbt' AND table_name = 'users'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'cbt' AND table_name = 'users_legacy'
    ) THEN
        EXECUTE 'ALTER TABLE cbt.users RENAME TO users_legacy';
    END IF;
END
$$;

-- Ensure LMS users has rows for every legacy CBT user reference.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'cbt' AND table_name = 'users_legacy'
    ) THEN
        INSERT INTO public.users (id, email, password_hash, full_name, is_active, created_at, updated_at)
        SELECT
            COALESCE(u.lms_user_id, u.id::bigint) AS id,
            COALESCE(NULLIF(trim(u.email), ''), format('cbt_legacy_%s@intelecto.local', u.id)) AS email,
            COALESCE(NULLIF(u.password_hash, ''), crypt('legacy_password', gen_salt('bf'))) AS password_hash,
            COALESCE(NULLIF(trim(u.full_name), ''), format('Legacy User %s', u.id)) AS full_name,
            COALESCE(u.is_active, true) AS is_active,
            COALESCE(u.created_at, NOW()) AS created_at,
            COALESCE(u.updated_at, NOW()) AS updated_at
        FROM cbt.users_legacy u
        ON CONFLICT (id) DO NOTHING;
    END IF;
END
$$;

-- Re-map user_id columns to LMS user IDs before adding new FK constraints.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'cbt' AND table_name = 'users_legacy'
    ) THEN
        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema='cbt' AND table_name='exam_sessions' AND column_name='user_id'
        ) THEN
            EXECUTE '
                UPDATE cbt.exam_sessions es
                SET user_id = COALESCE(u.lms_user_id, u.id::bigint)
                FROM cbt.users_legacy u
                WHERE es.user_id::bigint = u.id::bigint
            ';
        END IF;

        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema='cbt' AND table_name='user_limits' AND column_name='user_id'
        ) THEN
            EXECUTE '
                UPDATE cbt.user_limits ul
                SET user_id = COALESCE(u.lms_user_id, u.id::bigint)
                FROM cbt.users_legacy u
                WHERE ul.user_id::bigint = u.id::bigint
            ';
        END IF;

        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema='cbt' AND table_name='user_limit_usage' AND column_name='user_id'
        ) THEN
            EXECUTE '
                UPDATE cbt.user_limit_usage ulu
                SET user_id = COALESCE(u.lms_user_id, u.id::bigint)
                FROM cbt.users_legacy u
                WHERE ulu.user_id::bigint = u.id::bigint
            ';
        END IF;

        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema='cbt' AND table_name='exam_sessions' AND column_name='user_id'
        ) THEN
            EXECUTE '
                UPDATE cbt.exam_sessions es
                SET user_id = COALESCE(u.lms_user_id, u.id::bigint)
                FROM cbt.users_legacy u
                WHERE es.user_id::bigint = u.id::bigint
            ';
        END IF;
    END IF;
END
$$;

-- cbt.test_session is a compatibility view over cbt.exam_sessions and depends on exam_sessions.user_id.
DROP VIEW IF EXISTS cbt.test_session;

-- user_id must be BIGINT to reference public.users(id).
ALTER TABLE IF EXISTS cbt.exam_sessions ALTER COLUMN user_id TYPE BIGINT USING user_id::bigint;
ALTER TABLE IF EXISTS cbt.user_limits ALTER COLUMN user_id TYPE BIGINT USING user_id::bigint;
ALTER TABLE IF EXISTS cbt.user_limit_usage ALTER COLUMN user_id TYPE BIGINT USING user_id::bigint;

-- Drop old FK constraints if present.
ALTER TABLE IF EXISTS cbt.exam_sessions DROP CONSTRAINT IF EXISTS exam_sessions_user_id_fkey;
ALTER TABLE IF EXISTS cbt.user_limits DROP CONSTRAINT IF EXISTS user_limits_user_id_fkey;
ALTER TABLE IF EXISTS cbt.user_limit_usage DROP CONSTRAINT IF EXISTS user_limit_usage_user_id_fkey;

-- Add canonical FK constraints to LMS users.
ALTER TABLE IF EXISTS cbt.exam_sessions
    ADD CONSTRAINT exam_sessions_user_id_public_users_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id)
    ON UPDATE CASCADE ON DELETE SET NULL;

ALTER TABLE IF EXISTS cbt.user_limits
    ADD CONSTRAINT user_limits_user_id_public_users_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id)
    ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE IF EXISTS cbt.user_limit_usage
    ADD CONSTRAINT user_limit_usage_user_id_public_users_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id)
    ON UPDATE CASCADE ON DELETE CASCADE;

-- Recreate compatibility view after exam_sessions.user_id alteration.
CREATE VIEW cbt.test_session AS
SELECT
    id,
    session_token,
    student_name AS nama_peserta,
    grade_level_id AS id_tingkat,
    subject_id AS id_mata_pelajaran,
    user_id,
    started_at AS waktu_mulai,
    finished_at AS waktu_selesai,
    duration_minutes AS durasi_menit,
    final_score AS nilai_akhir,
    total_correct AS jumlah_benar,
    total_questions AS total_soal,
    status,
    lms_assignment_id,
    lms_class_id,
    created_at,
    updated_at
FROM cbt.exam_sessions;

-- Role derivation helper.
CREATE OR REPLACE FUNCTION cbt.resolve_user_role(p_user_id BIGINT)
RETURNS TEXT
LANGUAGE sql
STABLE
AS $$
    WITH ranked_membership AS (
        SELECT sm.role::text,
               CASE sm.role::text
                   WHEN 'school_admin' THEN 3
                   WHEN 'teacher' THEN 2
                   WHEN 'student' THEN 1
                   ELSE 0
               END AS rank_order,
               COALESCE(sm.updated_at, sm.created_at, NOW()) AS rank_time
        FROM public.school_memberships sm
        WHERE sm.user_id = p_user_id
          AND sm.deleted_at IS NULL
          AND COALESCE(sm.status, 'active') = 'active'
    )
    SELECT COALESCE(
        (
            SELECT CASE role
                WHEN 'school_admin' THEN 'superadmin'
                WHEN 'teacher' THEN 'teacher'
                ELSE 'student'
            END
            FROM ranked_membership
            ORDER BY rank_order DESC, rank_time DESC
            LIMIT 1
        ),
        'student'
    );
$$;

-- Compatibility projection for existing CBT queries.
DROP VIEW IF EXISTS cbt.users;
CREATE VIEW cbt.users AS
SELECT
    u.id,
    u.email,
    u.password_hash,
    u.full_name,
    u.full_name AS nama,
    cbt.resolve_user_role(u.id) AS role,
    COALESCE(u.is_active, true) AS is_active,
    COALESCE(u.created_at, NOW()) AS created_at,
    COALESCE(u.updated_at, NOW()) AS updated_at,
    u.id AS lms_user_id
FROM public.users u
WHERE u.deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_exam_sessions_user_id_bigint ON cbt.exam_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_limits_user_id_bigint ON cbt.user_limits(user_id);
CREATE INDEX IF NOT EXISTS idx_user_limit_usage_user_id_bigint ON cbt.user_limit_usage(user_id);
