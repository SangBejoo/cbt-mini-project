-- Migration: Create user limits tables
-- Date: 2025-01-17
-- Description: Add user limits and usage tracking tables

-- Create user_limits table
CREATE TABLE IF NOT EXISTS `user_limits` (
    `id` int NOT NULL AUTO_INCREMENT,
    `user_id` int NOT NULL,
    `limit_type` varchar(100) NOT NULL,
    `limit_value` int(11) NOT NULL DEFAULT 0,
    `current_used` int(11) NOT NULL DEFAULT 0,
    `reset_at` datetime NOT NULL,
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_user_limit_type` (`user_id`, `limit_type`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_reset_at` (`reset_at`),
    CONSTRAINT `fk_user_limits_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create user_limit_usage table
CREATE TABLE IF NOT EXISTS `user_limit_usage` (
    `id` int NOT NULL AUTO_INCREMENT,
    `user_id` int NOT NULL,
    `limit_type` varchar(100) NOT NULL,
    `action` varchar(100) NOT NULL,
    `resource_id` int NULL,
    `created_at` datetime NOT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_user_limit_usage_user_id` (`user_id`),
    KEY `idx_user_limit_usage_type` (`limit_type`),
    KEY `idx_user_limit_usage_created_at` (`created_at`),
    KEY `idx_user_limit_usage_resource_id` (`resource_id`),
    CONSTRAINT `fk_user_limit_usage_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default limits for existing users (optional - can be done via application)
-- This ensures all users have default limits when the system starts
INSERT IGNORE INTO `user_limits` (`user_id`, `limit_type`, `limit_value`, `current_used`, `reset_at`, `created_at`, `updated_at`)
SELECT
    u.id as user_id,
    'api_requests_per_hour' as limit_type,
    1000 as limit_value,
    0 as current_used,
    DATE_ADD(NOW(), INTERVAL 1 HOUR) as reset_at,
    NOW() as created_at,
    NOW() as updated_at
FROM `users` u;

INSERT IGNORE INTO `user_limits` (`user_id`, `limit_type`, `limit_value`, `current_used`, `reset_at`, `created_at`, `updated_at`)
SELECT
    u.id as user_id,
    'api_requests_per_day' as limit_type,
    10000 as limit_value,
    0 as current_used,
    DATE_ADD(DATE(NOW()), INTERVAL 1 DAY) as reset_at,
    NOW() as created_at,
    NOW() as updated_at
FROM `users` u;

INSERT IGNORE INTO `user_limits` (`user_id`, `limit_type`, `limit_value`, `current_used`, `reset_at`, `created_at`, `updated_at`)
SELECT
    u.id as user_id,
    'test_sessions_per_day' as limit_type,
    10 as limit_value,
    0 as current_used,
    DATE_ADD(DATE(NOW()), INTERVAL 1 DAY) as reset_at,
    NOW() as created_at,
    NOW() as updated_at
FROM `users` u;

INSERT IGNORE INTO `user_limits` (`user_id`, `limit_type`, `limit_value`, `current_used`, `reset_at`, `created_at`, `updated_at`)
SELECT
    u.id as user_id,
    'test_sessions_per_week' as limit_type,
    50 as limit_value,
    0 as current_used,
    DATE_ADD(DATE(NOW()), INTERVAL (8 - WEEKDAY(NOW())) DAY) as reset_at,
    NOW() as created_at,
    NOW() as updated_at
FROM `users` u;

INSERT IGNORE INTO `user_limits` (`user_id`, `limit_type`, `limit_value`, `current_used`, `reset_at`, `created_at`, `updated_at`)
SELECT
    u.id as user_id,
    'questions_per_day' as limit_type,
    100 as limit_value,
    0 as current_used,
    DATE_ADD(DATE(NOW()), INTERVAL 1 DAY) as reset_at,
    NOW() as created_at,
    NOW() as updated_at
FROM `users` u;