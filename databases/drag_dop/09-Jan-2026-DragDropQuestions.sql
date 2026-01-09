-- =============================================
-- DRAG AND DROP QUESTIONS MIGRATION
-- CBT Mini Project - Computer Based Test System
-- Date: January 9, 2026
-- Adds support for drag-and-drop question types
-- =============================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- =============================================
-- NEW TABLES FOR DRAG-DROP QUESTIONS
-- =============================================

-- Table: Soal Drag Drop (Drag-Drop Questions)
CREATE TABLE `soal_drag_drop` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `id_materi` INT UNSIGNED NOT NULL,
    `id_tingkat` INT UNSIGNED NOT NULL,
    `pertanyaan` TEXT NOT NULL,
    `drag_type` ENUM('ordering','matching') NOT NULL COMMENT 'ordering=sequence, matching=categorize',
    `pembahasan` TEXT NULL COMMENT 'Answer explanation',
    `is_active` BOOLEAN NOT NULL DEFAULT TRUE,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_soal_drag_drop_materi` (`id_materi`),
    KEY `idx_soal_drag_drop_tingkat` (`id_tingkat`),
    KEY `idx_soal_drag_drop_is_active` (`is_active`),
    KEY `idx_soal_drag_drop_type` (`drag_type`),
    KEY `idx_soal_drag_drop_active_materi` (`is_active`, `id_materi`),
    CONSTRAINT `fk_soal_drag_drop_materi` FOREIGN KEY (`id_materi`) 
        REFERENCES `materi` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_soal_drag_drop_tingkat` FOREIGN KEY (`id_tingkat`) 
        REFERENCES `tingkat` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='Drag and drop questions (ordering/matching types)';

-- Table: Drag Item (Draggable elements)
CREATE TABLE `drag_item` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `id_soal_drag_drop` INT UNSIGNED NOT NULL,
    `label` VARCHAR(255) NOT NULL COMMENT 'Text label for the item',
    `image_url` VARCHAR(500) NULL COMMENT 'Optional image URL (Cloudinary)',
    `urutan` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT 'Display order',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_drag_item_soal` (`id_soal_drag_drop`),
    KEY `idx_drag_item_urutan` (`id_soal_drag_drop`, `urutan`),
    CONSTRAINT `fk_drag_item_soal` FOREIGN KEY (`id_soal_drag_drop`) 
        REFERENCES `soal_drag_drop` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='Draggable items for drag-drop questions';

-- Table: Drag Slot (Drop zones)
CREATE TABLE `drag_slot` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `id_soal_drag_drop` INT UNSIGNED NOT NULL,
    `label` VARCHAR(255) NOT NULL COMMENT 'Slot label (1,2,3 or category name)',
    `urutan` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT 'Display order',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_drag_slot_soal` (`id_soal_drag_drop`),
    KEY `idx_drag_slot_urutan` (`id_soal_drag_drop`, `urutan`),
    CONSTRAINT `fk_drag_slot_soal` FOREIGN KEY (`id_soal_drag_drop`) 
        REFERENCES `soal_drag_drop` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='Drop zones/slots for drag-drop questions';

-- Table: Drag Correct Answer (Item-to-Slot mapping)
CREATE TABLE `drag_correct_answer` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `id_drag_item` INT UNSIGNED NOT NULL,
    `id_drag_slot` INT UNSIGNED NOT NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_drag_correct_item_slot` (`id_drag_item`, `id_drag_slot`),
    KEY `idx_drag_correct_item` (`id_drag_item`),
    KEY `idx_drag_correct_slot` (`id_drag_slot`),
    CONSTRAINT `fk_drag_correct_item` FOREIGN KEY (`id_drag_item`) 
        REFERENCES `drag_item` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_drag_correct_slot` FOREIGN KEY (`id_drag_slot`) 
        REFERENCES `drag_slot` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='Correct answer mapping: which item goes to which slot';

-- =============================================
-- MODIFY EXISTING TABLES
-- =============================================

-- Add question_type to test_session_soal to support mixed question types
ALTER TABLE `test_session_soal`
ADD COLUMN `question_type` ENUM('multiple_choice','drag_drop') NOT NULL DEFAULT 'multiple_choice' 
    COMMENT 'Type of question' AFTER `id_soal`,
ADD COLUMN `id_soal_drag_drop` INT UNSIGNED NULL 
    COMMENT 'FK to soal_drag_drop for drag-drop questions' AFTER `question_type`,
ADD KEY `idx_test_session_soal_type` (`question_type`),
ADD KEY `idx_test_session_soal_drag_drop` (`id_soal_drag_drop`),
ADD CONSTRAINT `fk_test_session_soal_drag_drop` FOREIGN KEY (`id_soal_drag_drop`) 
    REFERENCES `soal_drag_drop` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE;

-- Add drag_drop answer support to jawaban_siswa
ALTER TABLE `jawaban_siswa`
ADD COLUMN `question_type` ENUM('multiple_choice','drag_drop') NOT NULL DEFAULT 'multiple_choice' 
    COMMENT 'Type of question answered' AFTER `jawaban_dipilih`,
ADD COLUMN `jawaban_drag_drop` JSON NULL 
    COMMENT 'Drag-drop answer: {"item_id": slot_id, ...}' AFTER `question_type`,
ADD KEY `idx_jawaban_siswa_type` (`question_type`);

SET FOREIGN_KEY_CHECKS = 1;

-- =============================================
-- END OF MIGRATION
-- =============================================
