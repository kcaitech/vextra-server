CREATE TABLE `kcserver`.`document_version` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_at` DATETIME(6),

  `document_id` BIGINT NOT NULL,
  `version_id` VARCHAR(64) NOT NULL,
  `last_cmd_id` BIGINT,

  INDEX `idx_document_id` (`document_id`),
  INDEX `idx_version_id` (`version_id`)
)
CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;