# 用户资料与帖子管理数据库变更

如果是全新建库，直接使用 `models/create_tables.sql`。

如果是已有数据库，需要先执行下面的迁移 SQL：

```sql
ALTER TABLE `user`
  ADD COLUMN `nickname` varchar(32) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' AFTER `password`,
  ADD COLUMN `avatar_url` varchar(512) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' AFTER `nickname`,
  ADD COLUMN `bio` varchar(160) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' AFTER `avatar_url`;

UPDATE `user`
SET `nickname` = `username`
WHERE `nickname` = '';

ALTER TABLE `user`
  MODIFY COLUMN `nickname` varchar(32) COLLATE utf8mb4_general_ci NOT NULL,
  ADD UNIQUE KEY `idx_nickname` (`nickname`) USING BTREE;

ALTER TABLE `post`
  MODIFY COLUMN `status` tinyint(4) NOT NULL DEFAULT '1',
  ADD INDEX `idx_author_status` (`author_id`, `status`);
```

状态约定：

- `status = 0`: 草稿
- `status = 1`: 已发布
