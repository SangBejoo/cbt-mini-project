-- =============================================
-- Fix test_session_soal to allow NULL id_soal for drag-drop questions
-- Date: January 9, 2026
-- =============================================

SET FOREIGN_KEY_CHECKS = 0;

-- Drop the existing foreign key constraint
ALTER TABLE `test_session_soal` DROP FOREIGN KEY `test_session_soal_ibfk_2`;

-- Drop the index as well
ALTER TABLE `test_session_soal` DROP INDEX `test_session_soal_ibfk_2`;

-- Modify id_soal to allow NULL for drag-drop questions
ALTER TABLE `test_session_soal`
    MODIFY COLUMN `id_soal` INT UNSIGNED NULL 
    COMMENT 'FK to soal for multiple choice questions (NULL for drag-drop)';

-- Re-create the foreign key constraint with proper nullable support
ALTER TABLE `test_session_soal`
    ADD CONSTRAINT `fk_test_session_soal_soal` FOREIGN KEY (`id_soal`) 
        REFERENCES `soal` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE;

SET FOREIGN_KEY_CHECKS = 1;
