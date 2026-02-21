-- ══════════════════════════════════════════════════════════════════════════════
-- SOFT DELETE FOUNDATION MIGRATION — CBT
-- Date: 2026-02-21
-- Purpose: Add deleted_at columns to soft-deletable CBT tables.
--          Note: CBT keeps is_active for business state (question availability),
--          deleted_at is for administrative deletion only.
-- ══════════════════════════════════════════════════════════════════════════════

BEGIN;

-- ── CONTENT TABLES ────────────────────────────────
ALTER TABLE materials           ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE questions           ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE drag_drop_questions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;

-- ── EXAM SESSION TABLES ───────────────────────────
ALTER TABLE exam_sessions       ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE student_answers     ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;


-- ══════════════════════════════════════════════════
-- Partial indexes for query performance
-- ══════════════════════════════════════════════════

CREATE INDEX IF NOT EXISTS idx_materials_active
  ON materials(id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_materials_subject_active
  ON materials(subject_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_materials_school_active
  ON materials(school_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_questions_material_active
  ON questions(material_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_dd_questions_material_active_sd
  ON drag_drop_questions(material_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_exam_sessions_active
  ON exam_sessions(id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_exam_sessions_user_active
  ON exam_sessions(user_id) WHERE deleted_at IS NULL;

COMMIT;
