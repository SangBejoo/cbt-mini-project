-- Migration: Add soal_drag_drop_gambar table for drag-drop question images
-- Date: 10-Jan-2026

CREATE TABLE IF NOT EXISTS `soal_drag_drop_gambar` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `id_soal_drag_drop` INT UNSIGNED NOT NULL,
    `nama_file` VARCHAR(255) NOT NULL,
    `file_path` VARCHAR(500) NOT NULL,
    `file_size` INT NOT NULL,
    `mime_type` VARCHAR(50) NOT NULL,
    `urutan` TINYINT UNSIGNED NOT NULL DEFAULT 1,
    `keterangan` VARCHAR(255) NULL,
    `cloud_id` VARCHAR(255) NULL,
    `public_id` VARCHAR(500) NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX `idx_soal_drag_drop_gambar_soal` (`id_soal_drag_drop`),
    CONSTRAINT `fk_soal_drag_drop_gambar_soal` 
        FOREIGN KEY (`id_soal_drag_drop`) 
        REFERENCES `soal_drag_drop`(`id`) 
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
