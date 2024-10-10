
CREATE TABLE `kcserver`.`document_permission` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_at` DATETIME(6),

  `resource_type` TINYINT NOT NULL DEFAULT 0,
  `resource_id` BIGINT NOT NULL,
  `grantee_type` TINYINT NOT NULL DEFAULT 0,
  `grantee_id` BIGINT NOT NULL,
  `perm_type` TINYINT NOT NULL DEFAULT 1,
  `perm_source_type` TINYINT NOT NULL DEFAULT 0,

  CONSTRAINT `unique_index` UNIQUE (`resource_type`, `resource_id`, `grantee_type`, `grantee_id`)
)
CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;