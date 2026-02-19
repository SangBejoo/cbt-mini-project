-- Migration: Backfill inferable LMS linkage columns from existing relational data
-- Date: 01-Mar-2026

-- 1) Backfill subject class scope from materi when missing
UPDATE subjects s
SET lms_class_id = src.lms_class_id,
    updated_at = CURRENT_TIMESTAMP
FROM (
  SELECT m.subject_id AS subject_id, MAX(m.lms_class_id) AS lms_class_id
  FROM materials m
    WHERE m.lms_class_id IS NOT NULL
  GROUP BY m.subject_id
) src
WHERE s.id = src.subject_id
  AND s.lms_class_id IS NULL;

-- 2) Backfill level school scope from classes through materi when missing
UPDATE grade_levels gl
SET lms_school_id = src.lms_school_id,
    updated_at = CURRENT_TIMESTAMP
FROM (
    SELECT m.grade_level_id AS level_id, MAX(c.lms_school_id) AS lms_school_id
    FROM materials m
    JOIN classes c ON c.lms_class_id = m.lms_class_id
    WHERE m.lms_class_id IS NOT NULL
      AND c.lms_school_id IS NOT NULL
    GROUP BY m.grade_level_id
) src
WHERE gl.id = src.level_id
  AND gl.lms_school_id IS NULL;

-- 3) Backfill materi class scope from subjects when missing
UPDATE materials m
SET lms_class_id = s.lms_class_id,
    updated_at = CURRENT_TIMESTAMP
FROM subjects s
WHERE m.subject_id = s.id
  AND m.lms_class_id IS NULL
  AND s.lms_class_id IS NOT NULL;

-- 4) Backfill question class scope from materi when missing
UPDATE questions q
SET lms_class_id = m.lms_class_id,
    updated_at = CURRENT_TIMESTAMP
FROM materials m
WHERE q.material_id = m.id
  AND q.lms_class_id IS NULL
  AND m.lms_class_id IS NOT NULL;

-- 5) Keep note: lms_module_id / lms_book_id / lms_teacher_material_id / lms_asset_id
-- are source-derived fields and cannot be safely inferred when source payload does not exist.
