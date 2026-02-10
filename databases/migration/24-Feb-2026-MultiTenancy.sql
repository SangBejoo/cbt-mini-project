-- Migration: Add multi-tenancy columns for class-scoped CBT data
-- Date: February 6, 2026

-- =============================================
-- NEW TABLE: classes (synced from LMS)
-- =============================================

CREATE TABLE IF NOT EXISTS classes (
    id SERIAL PRIMARY KEY,
    lms_class_id BIGINT UNIQUE NOT NULL,
    lms_school_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_classes_school ON classes (lms_school_id);
CREATE INDEX IF NOT EXISTS idx_classes_active ON classes (is_active);

-- =============================================
-- ADD COLUMNS TO EXISTING TABLES
-- =============================================

-- mata_pelajaran: add school/class scope
ALTER TABLE mata_pelajaran ADD COLUMN IF NOT EXISTS lms_school_id BIGINT;
ALTER TABLE mata_pelajaran ADD COLUMN IF NOT EXISTS lms_class_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_mp_school ON mata_pelajaran (lms_school_id);
CREATE INDEX IF NOT EXISTS idx_mp_class ON mata_pelajaran (lms_class_id);

-- tingkat: add school scope
ALTER TABLE tingkat ADD COLUMN IF NOT EXISTS lms_school_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_tingkat_school ON tingkat (lms_school_id);

-- materi: add class scope
ALTER TABLE materi ADD COLUMN IF NOT EXISTS lms_class_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_materi_class ON materi (lms_class_id);

-- soal: add class scope
ALTER TABLE soal ADD COLUMN IF NOT EXISTS lms_class_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_soal_class ON soal (lms_class_id);

-- soal_drag_drop: add class scope
ALTER TABLE soal_drag_drop ADD COLUMN IF NOT EXISTS lms_class_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_soal_dd_class ON soal_drag_drop (lms_class_id);

-- test_session: add class scope
ALTER TABLE test_session ADD COLUMN IF NOT EXISTS lms_class_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_session_class ON test_session (lms_class_id);

-- =============================================
-- TRIGGER FOR classes.updated_at
-- =============================================

CREATE TRIGGER update_classes_updated_at 
    BEFORE UPDATE ON classes 
    FOR EACH ROW 
    EXECUTE PROCEDURE update_updated_at_column();
