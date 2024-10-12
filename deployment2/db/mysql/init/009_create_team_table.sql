CREATE TABLE `kcserver`.`team` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_at` DATETIME(6),

  `name` VARCHAR(64) NOT NULL,
  `description` VARCHAR(128),
  `avatar` VARCHAR(256),
  `uid` VARCHAR(64) NOT NULL UNIQUE,
  `invited_perm_type` TINYINT NOT NULL DEFAULT 0,
  `invited_switch` BOOLEAN NOT NULL DEFAULT FALSE
)
CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;


CREATE TABLE `kcserver`.`team_member` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_at` DATETIME(6),

  `team_id` BIGINT NOT NULL,
  `user_id` BIGINT NOT NULL,
  `perm_type` TINYINT NOT NULL,
  `nickname` VARCHAR(64) NOT NULL,

  INDEX `idx_team_id` (`team_id`),
  INDEX `idx_user_id` (`user_id`),
  UNIQUE INDEX `idx_team_user` (`team_id`, `user_id`)
)
CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;


CREATE TABLE `kcserver`.`team_join_request` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_at` DATETIME(6),

  `user_id` BIGINT NOT NULL,
  `team_id` BIGINT NOT NULL,
  `perm_type` TINYINT NOT NULL,
  `status` TINYINT NOT NULL DEFAULT 0,
  `first_displayed_at` DATETIME(6),
  `processed_at` DATETIME(6),
  `processed_by` BIGINT,
  `applicant_notes` VARCHAR(256),
  `processor_notes` VARCHAR(256),

  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_team_id` (`team_id`),
  UNIQUE INDEX `idx_unique_request` (`user_id`, `team_id`)
)
CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;


CREATE TABLE `kcserver`.`team_join_request_message_show` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_at` DATETIME(6),

  `team_join_request_id` BIGINT NOT NULL,
  `user_id` BIGINT NOT NULL,
  `team_id` BIGINT NOT NULL,
  `first_displayed_at` DATETIME(6),

  INDEX `idx_team_join_request_id` (`team_join_request_id`),
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_team_id` (`team_id`),
  UNIQUE INDEX `idx_unique_show` (`team_join_request_id`, `user_id`)
)
CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;