-- =============================================
-- COMPLETE OPTIMIZED DATABASE SCHEMA
-- CBT Mini Project - Computer Based Test System
-- Date: January 6, 2026
-- Optimized for Performance with Strategic Indexing
-- =============================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';
SET time_zone = '+00:00';

-- =============================================
-- DROP TABLES (for clean reinstall)
-- =============================================
DROP TABLE IF EXISTS `user_limit_usage`;
DROP TABLE IF EXISTS `user_limits`;
DROP TABLE IF EXISTS `jawaban_siswa`;
DROP TABLE IF EXISTS `test_session_soal`;
DROP TABLE IF EXISTS `test_session`;
DROP TABLE IF EXISTS `soal_gambar`;
DROP TABLE IF EXISTS `soal`;
DROP TABLE IF EXISTS `materi`;
DROP TABLE IF EXISTS `mata_pelajaran`;
DROP TABLE IF EXISTS `tingkat`;
DROP TABLE IF EXISTS `users`;

-- =============================================
-- MASTER TABLES
-- =============================================

-- Table: Users (Authentication & Authorization)
CREATE TABLE `users` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `email` VARCHAR(100) NOT NULL,
    `password_hash` VARCHAR(255) NOT NULL COMMENT 'Bcrypt hashed password',
    `nama` VARCHAR(100) NOT NULL,
    `role` ENUM('siswa','admin') NOT NULL DEFAULT 'siswa',
    `is_active` BOOLEAN NOT NULL DEFAULT TRUE,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_users_email` (`email`),
    KEY `idx_users_role` (`role`),
    KEY `idx_users_is_active` (`is_active`),
    KEY `idx_users_active_role` (`is_active`, `role`),
    KEY `idx_users_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='User authentication and authorization';

-- Table: Mata Pelajaran (Subject)
CREATE TABLE `mata_pelajaran` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `nama` VARCHAR(50) NOT NULL,
    `is_active` BOOLEAN NOT NULL DEFAULT TRUE,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_mata_pelajaran_nama` (`nama`),
    KEY `idx_mata_pelajaran_is_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='Subject/Course master data';

-- Table: Tingkat (Grade/Level)
CREATE TABLE `tingkat` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `nama` VARCHAR(50) NOT NULL,
    `is_active` BOOLEAN NOT NULL DEFAULT TRUE,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_tingkat_nama` (`nama`),
    KEY `idx_tingkat_is_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='Grade/level master data';

-- =============================================
-- CONTENT TABLES
-- =============================================

-- Table: Materi (Learning Material/Topic)
CREATE TABLE `materi` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `id_mata_pelajaran` INT UNSIGNED NOT NULL,
    `id_tingkat` INT UNSIGNED NOT NULL,
    `nama` VARCHAR(100) NOT NULL,
    `is_active` BOOLEAN NOT NULL DEFAULT TRUE,
    `default_durasi_menit` INT UNSIGNED NOT NULL DEFAULT 60,
    `default_jumlah_soal` INT UNSIGNED NOT NULL DEFAULT 20,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_materi_unique` (`id_mata_pelajaran`, `id_tingkat`, `nama`),
    KEY `idx_materi_mata_pelajaran` (`id_mata_pelajaran`),
    KEY `idx_materi_tingkat` (`id_tingkat`),
    KEY `idx_materi_is_active` (`is_active`),
    KEY `idx_materi_active_tingkat` (`is_active`, `id_tingkat`),
    KEY `idx_materi_active_mata_pelajaran` (`is_active`, `id_mata_pelajaran`),
    KEY `idx_materi_composite` (`id_mata_pelajaran`, `id_tingkat`, `is_active`),
    CONSTRAINT `fk_materi_mata_pelajaran` FOREIGN KEY (`id_mata_pelajaran`) 
        REFERENCES `mata_pelajaran` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_materi_tingkat` FOREIGN KEY (`id_tingkat`) 
        REFERENCES `tingkat` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='Learning materials/topics with test configuration';

-- Table: Soal (Questions)
CREATE TABLE `soal` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `id_materi` INT UNSIGNED NOT NULL,
    `id_tingkat` INT UNSIGNED NOT NULL,
    `pertanyaan` TEXT NOT NULL,
    `opsi_a` VARCHAR(500) NOT NULL,
    `opsi_b` VARCHAR(500) NOT NULL,
    `opsi_c` VARCHAR(500) NOT NULL,
    `opsi_d` VARCHAR(500) NOT NULL,
    `jawaban_benar` CHAR(1) NOT NULL,
    `pembahasan` TEXT NULL COMMENT 'Answer explanation',
    `is_active` BOOLEAN NOT NULL DEFAULT TRUE,
    `image_path` VARCHAR(255) NULL COMMENT 'Legacy filesystem path',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_soal_materi` (`id_materi`),
    KEY `idx_soal_tingkat` (`id_tingkat`),
    KEY `idx_soal_is_active` (`is_active`),
    KEY `idx_soal_active_materi` (`is_active`, `id_materi`),
    KEY `idx_soal_active_tingkat` (`is_active`, `id_tingkat`),
    KEY `idx_soal_materi_active` (`id_materi`, `is_active`),
    KEY `idx_soal_created_at` (`created_at`),
    CONSTRAINT `fk_soal_materi` FOREIGN KEY (`id_materi`) 
        REFERENCES `materi` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_soal_tingkat` FOREIGN KEY (`id_tingkat`) 
        REFERENCES `tingkat` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='Test questions with multiple choice answers';

-- Table: Soal Gambar (Question Images - Cloudinary)
CREATE TABLE `soal_gambar` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `id_soal` INT UNSIGNED NOT NULL,
    `nama_file` VARCHAR(255) NOT NULL COMMENT 'Original filename',
    `file_path` VARCHAR(500) NOT NULL COMMENT 'Relative path from storage root',
    `file_size` INT UNSIGNED NOT NULL COMMENT 'File size in bytes',
    `mime_type` VARCHAR(50) NOT NULL COMMENT 'image/jpeg, image/png, etc',
    `urutan` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT 'Order for multiple images',
    `keterangan` VARCHAR(255) NULL COMMENT 'Image description/caption',
    `cloud_id` VARCHAR(255) NULL COMMENT 'Cloudinary resource ID',
    `public_id` VARCHAR(500) NULL COMMENT 'Cloudinary public ID',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_soal_gambar_soal` (`id_soal`),
    KEY `idx_soal_gambar_soal_urutan` (`id_soal`, `urutan`),
    KEY `idx_soal_gambar_cloud_id` (`cloud_id`),
    KEY `idx_soal_gambar_public_id` (`public_id`(191)),
    CONSTRAINT `fk_soal_gambar_soal` FOREIGN KEY (`id_soal`) 
        REFERENCES `soal` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='Question images metadata (Cloudinary storage)';

-- =============================================
-- TEST SESSION TABLES
-- =============================================

-- Table: Test Session
CREATE TABLE `test_session` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `session_token` VARCHAR(64) NOT NULL,
    `nama_peserta` VARCHAR(100) NOT NULL,
    `id_tingkat` INT UNSIGNED NOT NULL,
    `id_mata_pelajaran` INT UNSIGNED NOT NULL,
    `user_id` INT UNSIGNED NULL COMMENT 'Linked user (NULL for anonymous)',
    `waktu_mulai` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `waktu_selesai` TIMESTAMP NULL,
    `durasi_menit` INT UNSIGNED NOT NULL,
    `nilai_akhir` DECIMAL(5,2) NULL,
    `jumlah_benar` INT UNSIGNED NULL,
    `total_soal` INT UNSIGNED NULL,
    `status` ENUM('ongoing','completed','timeout') NOT NULL DEFAULT 'ongoing',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_test_session_token` (`session_token`),
    KEY `idx_test_session_tingkat` (`id_tingkat`),
    KEY `idx_test_session_mata_pelajaran` (`id_mata_pelajaran`),
    KEY `idx_test_session_user` (`user_id`),
    KEY `idx_test_session_waktu_mulai` (`waktu_mulai`),
    KEY `idx_test_session_status` (`status`),
    KEY `idx_test_session_user_status` (`user_id`, `status`),
    KEY `idx_test_session_status_waktu` (`status`, `waktu_mulai`),
    CONSTRAINT `fk_test_session_tingkat` FOREIGN KEY (`id_tingkat`) 
        REFERENCES `tingkat` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_test_session_mata_pelajaran` FOREIGN KEY (`id_mata_pelajaran`) 
        REFERENCES `mata_pelajaran` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_test_session_user` FOREIGN KEY (`user_id`) 
        REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='Test/exam sessions';

-- Table: Test Session Soal (Questions in Session)
CREATE TABLE `test_session_soal` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `id_test_session` INT UNSIGNED NOT NULL,
    `id_soal` INT UNSIGNED NOT NULL,
    `nomor_urut` SMALLINT UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_test_session_soal_unique` (`id_test_session`, `nomor_urut`),
    KEY `idx_test_session_soal_session` (`id_test_session`),
    KEY `idx_test_session_soal_soal` (`id_soal`),
    CONSTRAINT `fk_test_session_soal_session` FOREIGN KEY (`id_test_session`) 
        REFERENCES `test_session` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_test_session_soal_soal` FOREIGN KEY (`id_soal`) 
        REFERENCES `soal` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='Questions assigned to test sessions';

-- Table: Jawaban Siswa (Student Answers)
CREATE TABLE `jawaban_siswa` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `id_test_session_soal` INT UNSIGNED NOT NULL,
    `jawaban_dipilih` CHAR(1) NULL COMMENT 'Selected answer (A/B/C/D)',
    `is_correct` BOOLEAN NOT NULL DEFAULT FALSE,
    `dijawab_pada` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_jawaban_siswa_unique` (`id_test_session_soal`),
    KEY `idx_jawaban_siswa_is_correct` (`is_correct`),
    KEY `idx_jawaban_siswa_dijawab_pada` (`dijawab_pada`),
    CONSTRAINT `fk_jawaban_siswa_test_session_soal` FOREIGN KEY (`id_test_session_soal`) 
        REFERENCES `test_session_soal` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='Student answers for test questions';

-- =============================================
-- USER LIMITS & USAGE TRACKING
-- =============================================

-- Table: User Limits
CREATE TABLE `user_limits` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` INT UNSIGNED NOT NULL,
    `limit_type` VARCHAR(100) NOT NULL COMMENT 'api_requests_per_hour, test_sessions_per_day, etc',
    `limit_value` INT UNSIGNED NOT NULL DEFAULT 0,
    `current_used` INT UNSIGNED NOT NULL DEFAULT 0,
    `reset_at` TIMESTAMP NOT NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_user_limits_unique` (`user_id`, `limit_type`),
    KEY `idx_user_limits_user` (`user_id`),
    KEY `idx_user_limits_reset_at` (`reset_at`),
    KEY `idx_user_limits_user_type` (`user_id`, `limit_type`),
    KEY `idx_user_limits_type_reset` (`limit_type`, `reset_at`),
    CONSTRAINT `fk_user_limits_user` FOREIGN KEY (`user_id`) 
        REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='User usage limits and quotas';

-- Table: User Limit Usage
CREATE TABLE `user_limit_usage` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` INT UNSIGNED NOT NULL,
    `limit_type` VARCHAR(100) NOT NULL,
    `action` VARCHAR(100) NOT NULL COMMENT 'Action performed',
    `resource_id` INT UNSIGNED NULL COMMENT 'Related resource ID',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_user_limit_usage_user` (`user_id`),
    KEY `idx_user_limit_usage_type` (`limit_type`),
    KEY `idx_user_limit_usage_created_at` (`created_at`),
    KEY `idx_user_limit_usage_resource_id` (`resource_id`),
    KEY `idx_user_limit_usage_user_created` (`user_id`, `created_at`),
    KEY `idx_user_limit_usage_user_type_created` (`user_id`, `limit_type`, `created_at`),
    CONSTRAINT `fk_user_limit_usage_user` FOREIGN KEY (`user_id`) 
        REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='Usage tracking for limits';

-- =============================================
-- PERFORMANCE OPTIMIZATION SETTINGS
-- =============================================

-- Optimize InnoDB settings (these should be in my.cnf/my.ini)
-- innodb_buffer_pool_size = 1G (or 70-80% of available RAM)
-- innodb_log_file_size = 256M
-- innodb_flush_log_at_trx_commit = 2
-- innodb_flush_method = O_DIRECT
-- query_cache_type = 0 (for MySQL 5.7)
-- max_connections = 200

SET FOREIGN_KEY_CHECKS = 1;

-- =============================================
-- INITIAL DATA (Optional - Comment out if not needed)
-- =============================================

-- Insert default admin user (password: admin123)
INSERT INTO `users` (`email`, `password_hash`, `nama`, `role`, `is_active`) VALUES
('admin@erlangga.com', '$2a$10$YourBcryptHashHere', 'Administrator', 'admin', TRUE);

-- Insert sample grades
INSERT INTO `tingkat` (`nama`, `is_active`) VALUES
('Kelas 7', TRUE),
('Kelas 8', TRUE),
('Kelas 9', TRUE),
('Kelas 10', TRUE),
('Kelas 11', TRUE),
('Kelas 12', TRUE);

-- Insert sample subjects
INSERT INTO `mata_pelajaran` (`nama`, `is_active`) VALUES
('Matematika', TRUE),
('Bahasa Indonesia', TRUE),
('Bahasa Inggris', TRUE),
('IPA', TRUE),
('IPS', TRUE),
('Fisika', TRUE),
('Kimia', TRUE),
('Biologi', TRUE);

-- =============================================
-- MAINTENANCE QUERIES (For DBA Reference)
-- =============================================

-- Check table sizes
-- SELECT 
--     table_name AS 'Table',
--     ROUND(((data_length + index_length) / 1024 / 1024), 2) AS 'Size (MB)'
-- FROM information_schema.TABLES
-- WHERE table_schema = DATABASE()
-- ORDER BY (data_length + index_length) DESC;

-- Check index usage
-- SELECT 
--     TABLE_NAME,
--     INDEX_NAME,
--     SEQ_IN_INDEX,
--     COLUMN_NAME,
--     CARDINALITY
-- FROM information_schema.STATISTICS
-- WHERE TABLE_SCHEMA = DATABASE()
-- ORDER BY TABLE_NAME, INDEX_NAME, SEQ_IN_INDEX;

-- Analyze and optimize tables (run periodically)
-- ANALYZE TABLE users, mata_pelajaran, tingkat, materi, soal, soal_gambar, 
--                test_session, test_session_soal, jawaban_siswa, 
--                user_limits, user_limit_usage;

-- OPTIMIZE TABLE users, mata_pelajaran, tingkat, materi, soal, soal_gambar,
--                 test_session, test_session_soal, jawaban_siswa,
--                 user_limits, user_limit_usage;

-- =============================================
-- END OF SCHEMA
-- =============================================
