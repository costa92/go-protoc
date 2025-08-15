-- 创建数据库
CREATE DATABASE protoc;

-- Migration: 001_create_users_table.sql
-- Description: 创建用户表，支持完整的用户管理功能
-- Created: 2025-08-15

use protoc;

-- 创建用户表
CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID，主键自增',
  `username` VARCHAR(64) NOT NULL COMMENT '用户名，唯一标识',
  `email` VARCHAR(255) NOT NULL COMMENT '邮箱地址',
  `password_hash` VARCHAR(255) NOT NULL COMMENT '密码哈希值',
  `phone` VARCHAR(20) DEFAULT NULL COMMENT '手机号码',
  `full_name` VARCHAR(100) DEFAULT NULL COMMENT '用户全名',
  `avatar_url` VARCHAR(500) DEFAULT NULL COMMENT '头像URL',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '用户状态: 0=禁用, 1=启用, 2=锁定',
  `email_verified` BOOLEAN NOT NULL DEFAULT FALSE COMMENT '邮箱是否已验证',
  `phone_verified` BOOLEAN NOT NULL DEFAULT FALSE COMMENT '手机是否已验证',
  `login_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '登录次数',
  `last_login_at` TIMESTAMP NULL DEFAULT NULL COMMENT '最后登录时间',
  `last_login_ip` VARCHAR(45) DEFAULT NULL COMMENT '最后登录IP',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` TIMESTAMP NULL DEFAULT NULL COMMENT '软删除时间',

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_username` (`username`),
  UNIQUE KEY `uk_users_email` (`email`),
  KEY `idx_users_phone` (`phone`),
  KEY `idx_users_status` (`status`),
  KEY `idx_users_created_at` (`created_at`),
  KEY `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 创建用户配置表
CREATE TABLE IF NOT EXISTS `user_profiles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '配置ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `bio` TEXT DEFAULT NULL COMMENT '个人简介',
  `website` VARCHAR(255) DEFAULT NULL COMMENT '个人网站',
  `location` VARCHAR(100) DEFAULT NULL COMMENT '所在地',
  `timezone` VARCHAR(50) DEFAULT 'UTC' COMMENT '时区',
  `language` VARCHAR(10) DEFAULT 'en' COMMENT '语言偏好',
  `theme` VARCHAR(20) DEFAULT 'light' COMMENT '主题偏好: light, dark',
  `notification_email` BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否接收邮件通知',
  `notification_sms` BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否接收短信通知',
  `privacy_profile` TINYINT NOT NULL DEFAULT 1 COMMENT '资料隐私设置: 0=私密, 1=公开, 2=好友可见',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_profiles_user_id` (`user_id`),
  KEY `idx_user_profiles_language` (`language`),
  KEY `idx_user_profiles_timezone` (`timezone`),

  CONSTRAINT `fk_user_profiles_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户配置表';

-- 创建用户角色表
CREATE TABLE IF NOT EXISTS `user_roles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '角色ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `role` VARCHAR(50) NOT NULL COMMENT '角色名称: admin, user, moderator, guest',
  `scope` VARCHAR(100) DEFAULT 'global' COMMENT '角色作用域',
  `granted_by` BIGINT UNSIGNED DEFAULT NULL COMMENT '授权人ID',
  `granted_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',
  `expires_at` TIMESTAMP NULL DEFAULT NULL COMMENT '角色过期时间',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_roles_user_role_scope` (`user_id`, `role`, `scope`),
  KEY `idx_user_roles_role` (`role`),
  KEY `idx_user_roles_scope` (`scope`),
  KEY `idx_user_roles_expires_at` (`expires_at`),

  CONSTRAINT `fk_user_roles_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_roles_granted_by` FOREIGN KEY (`granted_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色表';

-- 插入示例数据
INSERT IGNORE INTO `users` (`id`, `username`, `email`, `password_hash`, `full_name`, `status`) VALUES
(1, 'admin', 'admin@example.com', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iYqiSjZdW8nviYjBX5Zl.4Z9E6D.', 'System Administrator', 1),
(2, 'demo', 'demo@example.com', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iYqiSjZdW8nviYjBX5Zl.4Z9E6D.', 'Demo User', 1);

-- 插入用户配置
INSERT IGNORE INTO `user_profiles` (`user_id`, `bio`, `language`, `timezone`) VALUES
(1, 'System administrator with full access privileges', 'en', 'UTC'),
(2, 'Demo user for testing purposes', 'en', 'UTC');

-- 插入用户角色
INSERT IGNORE INTO `user_roles` (`user_id`, `role`, `scope`) VALUES
(1, 'admin', 'global'),
(2, 'user', 'global');