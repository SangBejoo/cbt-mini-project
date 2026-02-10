-- Migration: Add LMS Assignment ID to test_session table
-- Date: 06-Feb-2026
-- Description: Add lms_assignment_id column to support LMS integration

ALTER TABLE test_session ADD COLUMN lms_assignment_id BIGINT;

-- Add index for performance
CREATE INDEX idx_test_session_lms_assignment_id ON test_session(lms_assignment_id);

-- Update enum to include 'scheduled' status
ALTER TABLE test_session MODIFY COLUMN status ENUM('scheduled','ongoing','completed','timeout') DEFAULT 'scheduled';