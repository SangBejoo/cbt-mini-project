-- =============================================
-- DRAG AND DROP SLOT IMAGES MIGRATION
-- CBT Mini Project - Computer Based Test System
-- Date: January 26, 2026
-- Adds image support for drag-drop slot categories/labels
-- =============================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- Add image_url column to drag_slot table for supporting images in matching type questions
ALTER TABLE `drag_slot`
ADD COLUMN `image_url` VARCHAR(500) NULL COMMENT 'Optional image URL for slot (Cloudinary)' AFTER `label`,
ADD KEY `idx_drag_slot_urutan_updated` (`id_soal_drag_drop`, `urutan`);

SET FOREIGN_KEY_CHECKS = 1;

-- =============================================
-- END OF MIGRATION
-- =============================================
