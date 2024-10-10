CREATE TABLE `kcserver`.`document_permission_requests` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_at` DATETIME(6),

  `user_id` BIGINT NOT NULL,
  `document_id` BIGINT NOT NULL,
  `perm_type` TINYINT NOT NULL,
  `status` TINYINT NOT NULL,
  `first_displayed_at` DATETIME(6),
  `processed_at` DATETIME(6),
  `processed_by` BIGINT,
  `applicant_notes` VARCHAR(256),
  `processor_notes` VARCHAR(256),

  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_document_id` (`document_id`)
)
CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;