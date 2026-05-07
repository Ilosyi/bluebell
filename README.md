# Bluebell

`bluebell` 是一个由 [Q1mi/bluebell](https://github.com/Q1mi/bluebell) 衍生并持续扩展的社区论坛项目。它保留了原项目“Gin + MySQL + Redis”的教学骨架，同时更新了前端页面、用户资料、帖子管理、搜索等更接近真实产品的能力。

当前仓库是一个前后端分离项目：

- 后端：Go + Gin 提供 REST API
- 前端：Vue 3 + Element Plus 提供论坛界面
- 
## 快速开始

可前往cnb.cool平台fork[仓库](https://cnb.cool/HUST_losyi/bluebell)
点击云原生开发，在终端运行下面命令
```bash
docker compose up -d
```

启动后，在 **PORTS** 中添加 `80` 端口（如果使用 Docker Desktop），然后打开 **Forwarded Address**（如 `https://zsh7nt5ioi-80.cnb.run/`）即可体验论坛。

## 项目目标

这个项目一方面适合作为 Go Web 后端学习样例，另一方面也在逐步向一个可运行的轻论坛产品靠近。

它重点覆盖的是这几类场景：

- 用户注册、登录、登录态保持
- 社区浏览与帖子阅读
- 发帖、编辑帖子、删除帖子、草稿箱
- 用户资料维护
- 帖子搜索、排序与投票

## 当前功能

### 用户系统

- 用户注册
- 用户登录
- 支持“用户名 / 昵称”登录
- JWT 鉴权
- 账号中心
- 修改昵称、头像、个人简介

### 社区与帖子

- 社区列表
- 社区详情
- 帖子详情
- 最新 / 热门排序
- 按社区筛选
- 首页关键字搜索

### 帖子管理

- 发布帖子
- 保存草稿
- 编辑已发布帖子
- 编辑草稿
- 发布草稿
- 删除帖子
- 我的帖子
- 草稿箱

### 互动能力

- 帖子投票
- Redis 热度排序

## 架构概览

### 整体结构

```text
frontend (Vue 3 SPA)
        |
        v
Gin Router -> Controller -> Logic -> DAO(MySQL / Redis)
                                |
                                +-> MySQL: 用户、社区、帖子、草稿、搜索
                                |
                                +-> Redis: 排序、投票、社区帖子集合
```

### 后端分层说明

- `routes/`
  - 注册接口与中间件。
- `middlewares/`
  - 处理 JWT 鉴权等公共逻辑。
- `controller/`
  - 接收请求、绑定参数、校验参数、返回统一响应。
- `logic/`
  - 负责业务编排，是后端主要的业务层。
- `dao/mysql/`
  - 负责 MySQL 数据读写。
- `dao/redis/`
  - 负责 Redis 排序、投票、集合操作。
- `models/`
  - 定义请求参数、领域模型、Swagger 结构。

### 前端结构说明

- `frontend/src/views/`
  - 页面级视图。
- `frontend/src/components/`
  - 可复用论坛组件。
- `frontend/src/composables/`
  - 可复用数据流逻辑。
- `frontend/src/api/`
  - API 请求封装。
- `frontend/src/router/`
  - 前端路由和登录拦截逻辑。

## 主要技术点

### 1. 雪花算法

项目通过 `pkg/snowflake` 封装 `bwmarrin/snowflake`，为用户 ID、帖子 ID 生成全局唯一业务主键。这样做的好处是业务层在写入数据库前就能拿到唯一 ID，不依赖 MySQL 自增列，也更适合后续扩展到分布式场景。

### 2. Gin 框架

后端 Web 框架使用 Gin。项目通过 `routes/routes.go` 统一注册路由，通过 `controller/` 处理请求，通过 `middlewares/` 挂载 JWT 鉴权等公共逻辑。整体结构清晰，适合学习 Go Web 项目的基础分层。

### 3. Zap 日志库

项目使用 `go.uber.org/zap` 作为日志方案，并在 `logger/` 中做了统一初始化与 Gin 日志接入。它主要用于：

- 启动阶段日志
- 请求日志
- panic 恢复日志
- 业务错误日志

相比标准库日志，Zap 更适合结构化日志输出，也更方便后续接入生产环境日志系统。

### 4. Viper 配置管理

项目通过 `settings/settings.go` 使用 Viper 读取 `settings/config.yaml`，并把配置反序列化到全局 `settings.GlobalConfig`。当前配置内容包括：

- 应用端口与运行模式
- MySQL 连接信息
- Redis 连接信息
- JWT 密钥与过期时间
- 日志配置

这部分很适合初学者理解 Go 项目的“配置集中管理”方式。

### 5. Swagger 生成文档

项目已经接入 Swagger。接口注释写在 `controller/` 层，通过 `swag` 生成文档数据，并通过 Gin 暴露 `/swagger/index.html` 页面，便于直接查看和调试 API。

### 6. JWT 认证

项目通过 `pkg/jwt` 负责 token 生成与解析，通过 `middlewares/auth.go` 负责请求鉴权。用户登录成功后会拿到 JWT，前端在后续请求中通过 `Authorization: Bearer <token>` 传递身份信息。账号中心、发帖、编辑帖子、投票等接口都依赖这套机制。

### 7. 令牌桶限流

项目已经通过 `github.com/juju/ratelimit` 接入全局令牌桶限流中间件。它在 `routes.Setup()` 中作为全局中间件挂载，所有进入 Gin 的请求都会先尝试从同一个令牌桶中取令牌。

当前限流参数来自 `settings/config.yaml`：

- `rate_limit.enabled`：是否启用限流
- `rate_limit.rate`：每秒补充多少个令牌
- `rate_limit.capacity`：令牌桶最大容量

当桶内没有令牌时，接口会返回 HTTP `429 Too Many Requests`，响应体仍然保持项目统一格式，业务码为 `CodeTooManyRequests`。

### 8. Go 语言操作 MySQL（sqlx）

项目使用 `github.com/jmoiron/sqlx` 操作 MySQL。`sqlx` 相比标准库 `database/sql` 更方便的地方在于：

- 支持结构体映射
- 支持更简洁的查询封装
- 写批量查询和结果绑定时更顺手

当前用户、社区、帖子等核心数据都保存在 MySQL 中，对应代码主要在 `dao/mysql/`。

### 9. Go 语言操作 Redis（go-redis）

项目使用 `github.com/redis/go-redis/v9` 操作 Redis。Redis 在本项目里不负责正文存储，而主要负责：

- 最新帖子排序
- 热门帖子排序
- 用户投票记录
- 社区帖子集合

这部分代码主要在 `dao/redis/`，尤其适合学习 `zset`、`set`、pipeline 等用法。

### 10. MySQL 与 Redis 的职责分工

项目没有把所有数据都塞进一个存储里，而是做了明确分工：

- MySQL 负责：
  - 用户资料
  - 社区信息
  - 帖子正文
  - 草稿数据
  - 搜索查询
- Redis 负责：
  - 最新帖子排序
  - 热门帖子排序
  - 用户投票记录
  - 社区帖子集合

这是本项目很值得学习的一个架构点：正文数据与排行榜数据分别落在最适合它们的存储里。

### 11. 帖子搜索

主页搜索目前接入增强帖子列表接口 `/api/v1/posts2`。搜索关键字会匹配公开帖子相关信息，主要用于首页内容流检索。


## 目录结构

```text
bluebell/
├── controller/          # HTTP 处理层
├── dao/
│   ├── mysql/           # MySQL 数据访问
│   └── redis/           # Redis 数据访问
├── docs/                # 项目文档
├── frontend/            # Vue 3 前端
├── logger/              # Zap 日志封装
├── logic/               # 业务逻辑层
├── middlewares/         # 中间件
├── models/              # 模型与建表 SQL
├── pkg/
│   ├── jwt/             # JWT 工具
│   └── snowflake/       # 雪花算法 ID
├── routes/              # 路由注册
├── settings/            # 配置读取
├── AGENT.md             # 给 agent 的项目速览
└── README.md            # 仓库说明
```

## 核心接口

统一前缀：`/api/v1`

### 公开接口

- `POST /signup`
- `POST /login`
- `GET /community`
- `GET /community/:id`
- `GET /post/:id`
- `GET /posts`
- `GET /posts2`

### 需要登录的接口

- `GET /me`
- `PUT /me`
- `POST /post`
- `POST /post/draft`
- `GET /my/posts`
- `GET /my/posts/:id`
- `PUT /post/:id`
- `PUT /post/:id/draft`
- `POST /post/:id/publish`
- `DELETE /post/:id`
- `POST /vote`

## 快速开始

### 1. 环境要求

- Go 1.26+
- Node.js 18+
- MySQL 8.x
- Redis 7.x

### 2. 初始化数据库

使用下面的建表文件初始化数据库：

- `models/create_tables.sql`

### 3. 配置后端

项目会优先读取：

- `settings/config.yaml`

如果本地配置不存在，则回退到：

- `settings/config.example.yaml`

建议先复制示例配置，再按本地环境修改 MySQL、Redis、JWT 等参数。

### 4. 启动后端

```bash
go run ./
```

或：

```bash
make run
```

启动后可访问：

- Swagger：`http://localhost:8080/swagger/index.html`
- 健康检查：`http://localhost:8080/api/v1/ping`

### 5. 启动前端

```bash
cd frontend
npm install
npm run dev
```

默认通过 Vite 启动本地开发服务。

## Docker 部署

项目提供了一套面向部署的 Docker Compose 编排，包含：

- `nginx`：托管前端静态资源，并反向代理 `/api/` 到后端。
- `backend`：运行 Go API 服务。
- `mysql`：保存用户、社区、帖子等持久化数据。
- `redis`：保存帖子排序、投票记录、社区帖子集合等数据。

### 1. 生产配置

Docker 环境使用单独配置文件：

- `settings/config.docker.yaml`

该配置默认：

- `app.mode` 设置为 `release`
- `app.enable_swagger` 设置为 `false`
- 日志级别为 `info`
- MySQL 主机名为 Compose 服务名 `mysql`
- Redis 主机名为 Compose 服务名 `redis`

部署前必须修改：

- `settings/config.docker.yaml` 中的 `jwt.secret`
- `docker-compose.yml` 中的 MySQL 密码
- `settings/config.docker.yaml` 中对应的 MySQL 密码
- `docker-compose.yml` 中的 Redis `--requirepass`
- `settings/config.docker.yaml` 中对应的 Redis 密码

### 2. 启动服务

```bash
docker compose up -d --build
```

启动后访问：

- 前端页面：`http://localhost/`
- 后端健康检查：`http://localhost/api/v1/ping`

### 3. 查看状态和日志

```bash
docker compose ps
docker compose logs -f backend
docker compose logs -f nginx
```

### 4. 停止服务

```bash
docker compose down
```

如果需要连同 MySQL 和 Redis 数据一起删除：

```bash
docker compose down -v
```

注意：`docker-compose.yml` 会在 MySQL 首次初始化时执行 `models/create_tables.sql` 和 `deploy/mysql/02_seed.sql`。如果 `mysql_data` 数据卷已经存在，MySQL 官方镜像不会重复执行初始化 SQL。

## 开发建议

如果你想快速读懂这个项目，推荐阅读顺序如下：

1. `main.go`
2. `routes/routes.go`
3. `controller/post.go`
4. `logic/post.go`
5. `dao/mysql/post.go`
6. `dao/redis/vote.go`
7. `frontend/src/router/index.js`
8. `frontend/src/views/Home.vue`
9. `frontend/src/views/Profile.vue`

## 测试与校验

后端常用校验命令：

```bash
go test ./...
```

前端常用校验命令：

```bash
cd frontend
npm run build
```

## 相关文档

- [后端学习导读](./docs/backend-learning-guide.md)
- [用户资料与帖子管理数据库变更](./docs/profile-post-management-migration.md)
- [项目 Review 与后续规划](./docs/review-and-roadmap.md)

## 致谢

- 原始项目：[Q1mi/bluebell](https://github.com/Q1mi/bluebell)
- 当前仓库在其基础上扩展了前端页面、资料页、帖子管理、搜索、样式统一等能力。
