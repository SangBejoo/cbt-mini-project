-- Migration: Add LMS Assignment ID to test_session table
-- Date: 06-Feb-2026
-- Description: Add lms_assignment_id column to support LMS integration

ALTER TABLE test_session ADD COLUMN IF NOT EXISTS lms_assignment_id BIGINT;

-- Add index for performance
CREATE INDEX IF NOT EXISTS idx_test_session_lms_assignment_id ON test_session(lms_assignment_id);

-- Update enum to include 'scheduled' status
DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM pg_type t
		JOIN pg_enum e ON t.oid = e.enumtypid
		WHERE t.typname = 'test_session_status_enum' AND e.enumlabel = 'scheduled'
	) THEN
		ALTER TYPE test_session_status_enum ADD VALUE 'scheduled';
	END IF;
END
$$;