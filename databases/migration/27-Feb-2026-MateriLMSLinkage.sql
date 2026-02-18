-- Migration: Add explicit LMS linkage columns for materi ownership scopes
-- Date: 27-Feb-2026

ALTER TABLE materi
    ADD COLUMN IF NOT EXISTS lms_book_id BIGINT,
    ADD COLUMN IF NOT EXISTS lms_teacher_material_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_materi_lms_book_id ON materi (lms_book_id);
CREATE INDEX IF NOT EXISTS idx_materi_lms_teacher_material_id ON materi (lms_teacher_material_id);

-- Keep module uniqueness while allowing book/teacher material overlays
CREATE UNIQUE INDEX IF NOT EXISTS uq_materi_lms_module_id
ON materi (lms_module_id)
WHERE lms_module_id IS NOT NULL;
