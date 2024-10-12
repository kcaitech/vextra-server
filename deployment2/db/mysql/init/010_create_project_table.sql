CREATE TABLE `kcserver`.`project` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_at` DATETIME(6),

  `team_id` BIGINT NOT NULL,
  `name` VARCHAR(64) NOT NULL,
  `description` VARCHAR(128),
  `public_switch` BOOLEAN NOT NULL DEFAULT FALSE,
  `perm_type` TINYINT NOT NULL DEFAULT 1,
  `invited_switch` BOOLEAN NOT NULL DEFAULT FALSE,
  `need_approval` BOOLEAN NOT NULL DEFAULT TRUE
)
CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE `kcserver`.`project_favorite` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_at` DATETIME(6),

  `user_id` BIGINT NOT NULL,
  `project_id` BIGINT NOT NULL,
  `is_favor` BOOLEAN NOT NULL DEFAULT TRUE,

  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_project_id` (`project_id`),
  UNIQUE INDEX `idx_user_project` (`user_id`, `project_id`)
)
CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE `kcserver`.`project_join_request` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_at` DATETIME(6),

  `user_id` BIGINT NOT NULL,
  `project_id` BIGINT NOT NULL,
  `perm_type` TINYINT NOT NULL,
  `status` TINYINT NOT NULL DEFAULT 0,
  `first_displayed_at` DATETIME(6),
  `processed_at` DATETIME(6),
  `processed_by` BIGINT,
  `applicant_notes` VARCHAR(256),
  `processor_notes` VARCHAR(256),

  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_project_id` (`project_id`),
  UNIQUE INDEX `idx_unique_request` (`user_id`, `project_id`)
)
CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE `kcserver`.`project_join_request_message_show` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_at` DATETIME(6),

  `project_join_request_id` BIGINT NOT NULL,
  `user_id` BIGINT NOT NULL,
  `project_id` BIGINT NOT NULL,
  `first_displayed_at` DATETIME(6),

  INDEX `idx_project_join_request_id` (`project_join_request_id`),
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_project_id` (`project_id`),
  UNIQUE INDEX `idx_unique_show` (`project_join_request_id`, `user_id`)
)
CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE `kcserver`.`project_member` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_at` DATETIME(6),

  `project_id` BIGINT NOT NULL,
  `user_id` BIGINT NOT NULL,
  `perm_type` TINYINT NOT NULL DEFAULT 1,
  `perm_source_type` TINYINT NOT NULL DEFAULT 0,

  INDEX `idx_project_id` (`project_id`),
  INDEX `idx_user_id` (`user_id`),
  UNIQUE INDEX `idx_unique_member` (`project_id`, `user_id`)
)
CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;