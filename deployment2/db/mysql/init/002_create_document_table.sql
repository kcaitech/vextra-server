CREATE TABLE `kcserver`.`documents` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_at` DATETIME(6),

  `user_id` BIGINT NOT NULL,
  `path` VARCHAR(64) NOT NULL,
  `doc_type` TINYINT NOT NULL DEFAULT 0,
  `name` VARCHAR(64) NOT NULL,
  `size` BIGINT UNSIGNED NOT NULL,
  `purged_at` DATETIME(6),
  `delete_by` BIGINT,
  `version_id` VARCHAR(64) NOT NULL,
  `team_id` BIGINT NOT NULL,
  `project_id` BIGINT NOT NULL,
  `locked_at` DATETIME(6),
  `locked_reason` VARCHAR(255),
  `locked_words` VARCHAR(255),

  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_team_id` (`team_id`),
  INDEX `idx_project_id` (`project_id`)
)
CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;