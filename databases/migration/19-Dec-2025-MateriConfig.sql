-- Migration: Add configuration fields to materi table
-- Date: 2025-12-21
-- Description: Allow admin to set default duration, number of questions, and active status per materi

-- Add columns to materi table
ALTER TABLE `materi` ADD COLUMN `is_active` BOOLEAN DEFAULT TRUE AFTER `nama`;
ALTER TABLE `materi` ADD COLUMN `default_durasi_menit` INT DEFAULT 60 AFTER `is_active`;
ALTER TABLE `materi` ADD COLUMN `default_jumlah_soal` INT DEFAULT 20 AFTER `default_durasi_menit`;

-- Add indexes for filtering
CREATE INDEX `materi_index_active` ON `materi` (`is_active`);
CREATE INDEX `materi_index_active_tingkat` ON `materi` (`is_active`, `id_tingkat`);
CREATE INDEX `materi_index_active_mata_pelajaran` ON `materi` (`is_active`, `id_mata_pelajaran`);
