-- Add is_active to mata_pelajaran and tingkat for soft delete
ALTER TABLE `mata_pelajaran` ADD COLUMN `is_active` BOOLEAN DEFAULT TRUE AFTER `nama`;
ALTER TABLE `tingkat` ADD COLUMN `is_active` BOOLEAN DEFAULT TRUE AFTER `nama`;
ALTER TABLE `soal` ADD COLUMN `is_active` BOOLEAN DEFAULT TRUE AFTER `pembahasan`;
