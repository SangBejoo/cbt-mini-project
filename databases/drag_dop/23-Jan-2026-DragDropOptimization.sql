-- Migration: Optimize drag-drop queries with proper indexes
-- Date: January 9, 2026
-- Purpose: Add indexes to improve performance of slow drag-drop queries

-- Note: This migration will fail if indexes already exist.
-- Run only once or modify to handle existing indexes.

-- Create indexes for soal_drag_drop queries filtering by is_active
CREATE INDEX idx_soal_drag_drop_is_active ON soal_drag_drop(is_active);

-- Composite index for common filter combinations
CREATE INDEX idx_soal_drag_drop_active_materi ON soal_drag_drop(is_active, id_materi);
CREATE INDEX idx_soal_drag_drop_active_tingkat ON soal_drag_drop(is_active, id_tingkat);

-- Index for drag_item queries (very important - this joins heavily)
CREATE INDEX idx_drag_item_soal_id ON drag_item(id_soal_drag_drop);
CREATE INDEX idx_drag_item_soal_urutan ON drag_item(id_soal_drag_drop, urutan);

-- Index for drag_slot queries
CREATE INDEX idx_drag_slot_soal_id ON drag_slot(id_soal_drag_drop);
CREATE INDEX idx_drag_slot_soal_urutan ON drag_slot(id_soal_drag_drop, urutan);

-- Index for drag_correct_answer queries (this has JOIN which is slow)
CREATE INDEX idx_drag_correct_answer_item ON drag_correct_answer(id_drag_item);
CREATE INDEX idx_drag_correct_answer_slot ON drag_correct_answer(id_drag_slot);
CREATE INDEX idx_drag_correct_answer_item_slot ON drag_correct_answer(id_drag_item, id_drag_slot);

-- Composite index for the complex JOIN query in GetCorrectAnswersBySoalID
CREATE INDEX idx_drag_correct_answer_soal_query ON drag_correct_answer(id_drag_item);
