# Bluebell 后端学习导读

这篇文档面向 Go 初学者，目标不是一次讲完所有细节，而是帮你建立“这套后端是怎么跑起来的、代码应该从哪里开始读”的整体认识。

## 1. 项目是什么

`bluebell` 是一个典型的社区论坛后端，当前已经具备这些能力：

- 用户注册、登录、JWT 鉴权
- 社区列表、社区详情
- 发帖、帖子详情、帖子列表
- 热度排序、投票
- 我的帖子、草稿箱、编辑/删除/发布草稿
- 用户资料页
- 搜索公开帖子

技术栈：

- Web 框架：`Gin`
- 数据库：`MySQL`
- 缓存/排序/投票：`Redis`
- 配置：`Viper`
- 日志：`Zap`
- 鉴权：`JWT`
- 分布式 ID：`Snowflake`

## 2. 先建立一个总图

这个项目采用分层结构：

1. `routes/`
   负责“有哪些路由、哪些接口需要登录”。

2. `middlewares/`
   负责“请求进入 controller 前的公共处理”，例如 JWT 鉴权。

3. `controller/`
   负责“收请求、校验参数、返回响应”。

4. `logic/`
   负责“业务编排”，也就是决定先查 MySQL 还是 Redis、要不要做权限判断、错误怎么转换。

5. `dao/mysql/` 和 `dao/redis/`
   负责“真正访问存储层”。

6. `models/`
   负责定义结构体，包括请求参数、数据库模型、接口响应模型。

你可以把它理解成：

- `controller` 像柜台窗口
- `logic` 像后台业务员
- `dao` 像仓库管理员
- `models` 像各种单据模板

## 3. 建议的阅读顺序

如果你时间有限，按下面顺序读最有效：

1. [main.go](/Users/losyi/CodeHub/bluebell/main.go)
   先看程序怎么启动。

2. [routes/routes.go](/Users/losyi/CodeHub/bluebell/routes/routes.go)
   看路由表，知道有哪些接口。

3. [middlewares/auth.go](/Users/losyi/CodeHub/bluebell/middlewares/auth.go)
   看登录后的接口是怎么拿到当前用户的。

4. [controller/post.go](/Users/losyi/CodeHub/bluebell/controller/post.go)
   这是业务最多的入口文件。

5. [logic/post.go](/Users/losyi/CodeHub/bluebell/logic/post.go)
   看帖子相关真正业务流程。

6. [dao/mysql/post.go](/Users/losyi/CodeHub/bluebell/dao/mysql/post.go)
   学 SQL 是怎么写的。

7. [dao/redis/vote.go](/Users/losyi/CodeHub/bluebell/dao/redis/vote.go)
   学 Redis 在这个项目里怎么用。

8. [models/](/Users/losyi/CodeHub/bluebell/models)
   回头再对照结构体，理解每层在传什么。

## 4. 程序启动流程

入口在 [main.go](/Users/losyi/CodeHub/bluebell/main.go)。

启动顺序是：

1. `settings.Init()`
   读取配置文件。

2. `controller.InitTrans("zh")`
   初始化参数校验错误的中文翻译。

3. `logger.Init(...)`
   初始化全局日志。

4. `mysql.Init()`
   初始化 MySQL 连接池。

5. `redis.Init()`
   初始化 Redis 客户端。

6. `snowflake.Init(...)`
   初始化雪花算法 ID 生成器。

7. `routes.Setup()`
   注册路由和中间件。

8. `srv.ListenAndServe()`
   启动 HTTP 服务。

这里最值得记住的点：

- `settings.GlobalConfig` 是全局配置入口
- `zap.L()` 是全局日志入口
- `mysql.GetDB()` / `redis.GetRDB()` 分别拿到全局数据库客户端

## 5. 一次请求是怎么走的

以“创建帖子”为例：

1. 前端请求 `POST /api/v1/post`
2. 路由命中 [routes/routes.go](/Users/losyi/CodeHub/bluebell/routes/routes.go)
3. 因为这个接口在 JWT 中间件之后，所以先进入 [middlewares/auth.go](/Users/losyi/CodeHub/bluebell/middlewares/auth.go)
4. 中间件解析 token，把 `userID` 写进 `Gin Context`
5. 然后进入 [controller/post.go](/Users/losyi/CodeHub/bluebell/controller/post.go) 的 `CreatePostHandler`
6. controller 绑定 JSON、校验参数、读取当前登录用户 ID
7. controller 调用 `logic.CreatePost`
8. logic 生成雪花 ID，先写 MySQL，再写 Redis 索引
9. controller 返回统一 JSON 响应

你以后看任何接口，都可以用这个套路追：

`routes -> middleware -> controller -> logic -> dao`

## 6. 每层到底该写什么

### controller 层

适合写：

- `ShouldBindJSON`
- `ShouldBindQuery`
- 参数错误返回
- 调 `logic`
- 把业务错误映射成响应码

不适合写：

- 大量 SQL
- 复杂业务判断
- 跨 MySQL/Redis 的编排逻辑

### logic 层

适合写：

- 权限判断
- 状态流转
- 先查 MySQL 再查 Redis 的组合流程
- 领域错误转换

不适合写：

- 原始 SQL 语句
- 直接拼 HTTP 响应

### dao 层

适合写：

- SQL
- Redis key 读写
- pipeline / zset / set 操作

不适合写：

- 当前用户是不是作者
- 某个错误该返回什么业务码

## 7. 你最应该重点学的几个文件

### 7.1 用户登录

- [controller/user.go](/Users/losyi/CodeHub/bluebell/controller/user.go)
- [logic/user.go](/Users/losyi/CodeHub/bluebell/logic/user.go)
- [dao/mysql/user.go](/Users/losyi/CodeHub/bluebell/dao/mysql/user.go)
- [pkg/jwt/jwt.go](/Users/losyi/CodeHub/bluebell/pkg/jwt/jwt.go)

你可以重点看：

- 用户名/昵称登录怎么查
- 密码怎么加密比对
- JWT 是什么时候生成的

### 7.2 发帖和帖子列表

- [controller/post.go](/Users/losyi/CodeHub/bluebell/controller/post.go)
- [logic/post.go](/Users/losyi/CodeHub/bluebell/logic/post.go)
- [dao/mysql/post.go](/Users/losyi/CodeHub/bluebell/dao/mysql/post.go)
- [dao/redis/vote.go](/Users/losyi/CodeHub/bluebell/dao/redis/vote.go)

你可以重点看：

- 帖子为什么先写 MySQL 再写 Redis
- 为什么搜索走 MySQL，热榜走 Redis
- 为什么详情页票数还要单独查 Redis

### 7.3 中间件和上下文

- [middlewares/auth.go](/Users/losyi/CodeHub/bluebell/middlewares/auth.go)
- [controller/request.go](/Users/losyi/CodeHub/bluebell/controller/request.go)

你可以重点看：

- 中间件如何把 `userID` 放进 `Context`
- controller 如何取出 `userID`

## 8. 为什么这个项目要同时用 MySQL 和 Redis

这是初学者最容易疑惑的地方。

### MySQL 负责

- 用户资料
- 帖子正文
- 社区信息
- 草稿数据
- 搜索

特点：

- 适合持久化保存
- 适合复杂查询
- 数据可靠

### Redis 负责

- 最新帖子排序
- 热门帖子排序
- 用户投票记录
- 社区帖子集合

特点：

- 读写快
- 很适合排行榜、计数、集合运算

一句话理解：

- 正文和资料放 MySQL
- 排序和热度放 Redis

## 9. Redis 在本项目里的 4 类 key

见 [dao/redis/key.go](/Users/losyi/CodeHub/bluebell/dao/redis/key.go)。

### 1. `bluebell:post:time`

- 类型：`zset`
- member：`post_id`
- score：发布时间时间戳

作用：

- 做“最新帖子”排序

### 2. `bluebell:post:score`

- 类型：`zset`
- member：`post_id`
- score：帖子热度分

作用：

- 做“热门帖子”排序

### 3. `bluebell:post:voted:<post_id>`

- 类型：`zset`
- member：`user_id`
- score：投票方向（1 / -1）

作用：

- 判断某个用户对某个帖子投过什么票

### 4. `bluebell:community:<community_id>`

- 类型：`set`
- member：`post_id`

作用：

- 记录某个社区下有哪些帖子

## 10. 为什么很多地方用“函数变量”

例如：

```go
var createPost = logic.CreatePost
```

这不是多余写法，主要是为了测试。

测试里可以临时替换：

```go
createPost = func(p *models.Post) error {
    return nil
}
```

这样就能只测 controller，不依赖真实数据库。

## 11. 为什么很多响应都返回 HTTP 200

这个项目采用的是“HTTP 状态码固定 + 业务 code 区分成功失败”的风格。

例如：

```json
{
  "code": 1004,
  "msg": "账号或密码错误",
  "data": null
}
```

优点：

- 前后端统一解析风格
- 对教学项目比较直观

缺点：

- 不完全符合 REST 风格
- 需要前端额外判断 `code`

## 12. 你现在应该怎么学这份代码

建议按这个节奏：

1. 先只看启动链路：`main -> routes -> middleware`
2. 再选 1 个接口完整追踪，例如“登录”或“发帖”
3. 画出调用链
4. 自己尝试改一个小需求
5. 再看对应测试理解预期行为

不要一开始就逐行啃完整项目，否则会很容易失去主线。

## 13. 建议你做的几个练习

1. 自己追踪“登录接口”完整流程，写下 8-10 个步骤。
2. 自己解释为什么帖子列表要补 Redis 投票数。
3. 给帖子新增一个“浏览量”字段，先只存在 MySQL。
4. 给社区详情加“该社区帖子数”。
5. 自己写一个简单接口，例如“获取当前登录用户名”。

## 14. 后续你可以继续问我的方式

如果你想继续学，可以直接让我做下面这些事：

- “带我逐行讲解 `logic/post.go`”
- “画出登录流程图”
- “解释 Redis 投票算法”
- “把某个函数改成更适合初学者理解的写法”
- “给我出 5 个基于当前项目的 Go 练习题”

如果你愿意，我下一步可以继续做两件事中的一个：

1. 把剩余后端运行时代码继续补到更细粒度注释
2. 单独写一篇“从登录到发帖”的源码走读文档
