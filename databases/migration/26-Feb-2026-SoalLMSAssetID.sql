-- Migration: Add LMS asset linkage for CBT multiple-choice questions
-- Date: 26-Feb-2026
-- Purpose: Keep traceability between CBT `soal` and LMS content assets

ALTER TABLE soal
ADD COLUMN IF NOT EXISTS lms_asset_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_soal_lms_asset_id ON soal (lms_asset_id);
