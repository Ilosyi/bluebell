INSERT INTO `community` (`community_id`, `community_name`, `introduction`)
VALUES
  (1, 'Go', 'Go 语言后端开发、并发编程与工程实践'),
  (2, 'Gin', 'Gin 框架、中间件、路由与接口设计'),
  (3, 'MySQL', '关系型数据库设计、查询优化与事务实践'),
  (4, 'Redis', '缓存、排行榜、投票与高性能数据结构'),
  (5, 'Vue', '前端页面、组件化与用户交互体验')
ON DUPLICATE KEY UPDATE
  `community_name` = VALUES(`community_name`),
  `introduction` = VALUES(`introduction`);
