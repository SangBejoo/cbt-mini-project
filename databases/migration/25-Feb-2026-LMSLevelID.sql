-- Migration: Add lms_level_id to tingkat table for LMS sync
-- Date: February 7, 2026
-- Purpose: Enable sync of level data from LMS to CBT

-- Add lms_level_id column to tingkat
ALTER TABLE tingkat ADD COLUMN IF NOT EXISTS lms_level_id BIGINT UNIQUE;

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_tingkat_lms_id ON tingkat (lms_level_id);

-- Add timestamps if not exists
ALTER TABLE tingkat ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE tingkat ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;

-- Create trigger for updated_at if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT FROM pg_trigger 
        WHERE tgname = 'update_tingkat_updated_at'
    ) THEN
        CREATE TRIGGER update_tingkat_updated_at 
            BEFORE UPDATE ON tingkat 
            FOR EACH ROW 
            EXECUTE PROCEDURE update_updated_at_column();
    END IF;
END
$$;
