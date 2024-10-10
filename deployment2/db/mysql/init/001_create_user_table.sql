CREATE TABLE `kcserver`.`user` (
  `id` BIGINT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `created_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  `updated_at` DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  `deleted_at` DATETIME(6),

  `nickname` VARCHAR(64),
  `avatar` VARCHAR(256),
  `uid` VARCHAR(64) NOT NULL UNIQUE,
  `is_activated` BOOLEAN NOT NULL DEFAULT FALSE,
--   `web_app_channel` VARCHAR(64),

  `wx_open_id` VARCHAR(64) NOT NULL,
  `wx_access_token` VARCHAR(255),
  `wx_access_token_create_time` DATETIME(6),
  `wx_refresh_token` VARCHAR(255),
  `wx_refresh_token_create_time` DATETIME(6),
  `wx_login_code` VARCHAR(64),

  `wx_mp_open_id` VARCHAR(64) NOT NULL,
  `wx_mp_session_key` VARCHAR(255),
  `wx_mp_session_key_create_time` DATETIME(6),
  `wx_mp_login_code` VARCHAR(64),

  `wx_union_id` VARCHAR(64) NOT NULL UNIQUE,

  UNIQUE INDEX `wx_openid_unique` (`wx_open_id`, `wx_mp_open_id`)
)
CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;