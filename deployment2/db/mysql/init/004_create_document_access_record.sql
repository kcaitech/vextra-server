CREATE TABLE `kcserver`.`document_access_records` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_at` DATETIME(6),

  `user_id` BIGINT NOT NULL,
  `document_id` BIGINT NOT NULL,
  `last_access_time` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),

  UNIQUE INDEX `idx_user_document` (`user_id`, `document_id`)
)
CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;