CREATE TABLE `user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `username` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
  `password` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
  `nickname` varchar(32) COLLATE utf8mb4_general_ci NOT NULL,
  `avatar_url` varchar(512) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `bio` varchar(160) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `email` varchar(64) COLLATE utf8mb4_general_ci,
  `gender` tinyint(4) NOT NULL DEFAULT '0',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`) USING BTREE,
  UNIQUE KEY `idx_user_id` (`user_id`) USING BTREE,
  UNIQUE KEY `idx_nickname` (`nickname`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE `community` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `community_id` int unsigned NOT NULL,
  `community_name` varchar(128) COLLATE utf8mb4_general_ci NOT NULL,
  `introduction` varchar(256) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_community_id` (`community_id`) USING BTREE,
  UNIQUE KEY `idx_community_name` (`community_name`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE `post` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `post_id` bigint(20) NOT NULL,
  `title` varchar(128) COLLATE utf8mb4_general_ci NOT NULL,
  `content` varchar(8192) COLLATE utf8mb4_general_ci NOT NULL,
  `author_id` bigint(20) NOT NULL,
  `community_id` bigint(20) NOT NULL,
  `status` tinyint(4) NOT NULL DEFAULT '1',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_post_id` (`post_id`) USING BTREE,
  KEY `idx_author_id` (`author_id`) USING BTREE,
  KEY `idx_community_id` (`community_id`) USING BTREE,
  KEY `idx_author_status` (`author_id`,`status`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
