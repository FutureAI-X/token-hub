# Token Hub

下一代LLM网关和AI资产管理系统

## PostgreSQL 管理

### 配置文件

`docker-compose.yml`：

```yaml
services:
  postgres:
    image: postgres:15
    container_name: token-hub-postgres
    environment:
      POSTGRES_USER: token_hub
      POSTGRES_PASSWORD: token_hub_123
      POSTGRES_DB: token_hub
    ports:
      - "5432:5432"
    volumes:
      - pg_data:/var/lib/postgresql/data

volumes:
  pg_data:
```

### 常用命令

```bash
# 启动（后台运行）
docker-compose up -d

# 停止
docker-compose down

# 停止并删除数据卷（慎用，会清除所有数据）
docker-compose down -v

# 查看运行状态
docker-compose ps

# 查看日志
docker-compose logs postgres

# 实时跟踪日志
docker-compose logs -f postgres

# 重启
docker-compose restart postgres
```

### 连接信息

| 项目 | 值 |
|------|-----|
| 主机 | `localhost` |
| 端口 | `5432` |
| 用户名 | `token_hub` |
| 密码 | `token_hub_123` |
| 数据库 | `token_hub` |
| 连接字符串 | `postgres://token_hub:token_hub_123@localhost:5432/token_hub?sslmode=disable` |

### 数据持久化

- 数据存储在 Docker 卷 `pg_data` 中
- 执行 `docker-compose down` 不会丢失数据
- 只有执行 `docker-compose down -v` 才会删除数据

### 使用 pgAdmin 管理（可选）

如需图形化管理工具，可在 `docker-compose.yml` 中添加：

```yaml
services:
  # ... postgres 配置 ...

  pgadmin:
    image: dpage/pgadmin4
    container_name: token-hub-pgadmin
    environment:
      PGADMIN_DEFAULT_EMAIL: admin@admin.com
      PGADMIN_DEFAULT_PASSWORD: admin
    ports:
      - "5050:80"
```

访问 http://localhost:5050 即可使用 pgAdmin。

## 快速开始

### 1. 启动 PostgreSQL

```bash
docker-compose up -d
```

### 2. 配置环境变量

创建 `.env` 文件：

```bash
SQL_DSN=postgres://token_hub:token_hub_123@localhost:5432/token_hub?sslmode=disable
PORT=3001
GIN_MODE=debug
```

### 3. 安装依赖

```bash
go mod tidy
```

### 4. 运行后端

```bash
go run main.go
```

### 5. 运行前端（可选）

```bash
cd web
npm install
npm run dev
```

前端开发服务器将在 http://localhost:5173 启动，并自动代理 API 请求到后端。

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `PORT` | 服务端口 | `3001` |
| `GIN_MODE` | Gin模式 (debug/release) | `release` |
| `SQL_DSN` | PostgreSQL 连接字符串 | 必须设置 |
| `SQL_MAX_IDLE_CONNS` | 最大空闲连接数 | `10` |
| `SQL_MAX_OPEN_CONNS` | 最大打开连接数 | `100` |
| `SQL_MAX_LIFETIME` | 连接最大生存时间(秒) | `60` |
| `JWT_SECRET` | JWT 密钥（用于生成和验证 Token） | `token-hub-jwt-secret-change-me` |
| `DEBUG` | 调试模式 | `false` |

## API 接口

### 用户登录

```bash
POST /api/auth/login
```

请求体：

```json
{
  "username": "root",
  "password": "123456"
}
```

响应示例：

```json
{
  "success": true,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "root",
      "display_name": "Root User",
      "role": 100,
      "status": 1,
      "email": "",
      "quota": 100000000,
      "used_quota": 0
    }
  }
}
```

### 获取用户信息（需要登录）

```bash
GET /api/user/info
Authorization: Bearer <token>
```

### 获取模型列表

```bash
GET /v1/models
```

响应示例：

```json
{
  "object": "list",
  "data": [
    {
      "id": "deepseek-v4-flash",
      "object": "model",
      "owned_by": "token-hub"
    }
  ]
}
```

### 健康检查

```bash
GET /health
```

响应示例：

```json
{
  "status": "ok"
}
```

## 技术栈

- **后端**: Go + Gin
- **前端**: React + TypeScript + Vite
- **数据库**: PostgreSQL
- **缓存**: 待定

## 项目结构

```
token-hub/
├── main.go              # 主入口文件
├── common/              # 公共工具
│   ├── database.go      # 数据库类型定义
│   ├── env.go           # 环境变量工具
│   ├── log.go           # 日志工具
│   ├── crypto.go        # 密码加密工具
│   └── jwt.go           # JWT Token 工具
├── router/              # 路由配置
│   └── router.go
├── controller/          # 控制器
│   ├── model.go         # 模型控制器
│   └── auth.go          # 认证控制器
├── middleware/           # 中间件
│   └── auth.go          # 认证中间件
├── model/               # 数据模型
│   ├── main.go          # 数据库初始化
│   ├── user.go          # 用户模型
│   └── model.go         # 模型管理
├── web/                 # 前端项目
│   ├── src/
│   │   ├── App.tsx      # 主页组件
│   │   ├── App.css      # 主页样式
│   │   ├── index.css    # Tailwind CSS
│   │   ├── lib/utils.ts # 工具函数
│   │   └── main.tsx     # 入口文件
│   ├── index.html       # HTML模板
│   ├── package.json     # 前端依赖
│   └── vite.config.ts   # Vite配置
├── docker-compose.yml   # Docker Compose 配置（PostgreSQL）
├── .env.example         # 环境变量示例
├── go.mod               # Go模块文件
└── go.sum               # 依赖校验文件
```

## 主页功能

现代化的主页 UI，参考 New API 项目设计：

- ✅ 顶部导航栏（毛玻璃效果、响应式）
- ✅ Hero 区域（渐变背景、网格图案、动画效果）
- ✅ 终端演示（实时显示 API 响应）
- ✅ 功能特性展示（Bento 网格布局）
- ✅ 模型列表（卡片式设计、悬停效果）
- ✅ 快速开始代码示例
- ✅ 暗色模式支持（跟随系统）
- ✅ 响应式设计（支持移动端）
- ✅ 流畅动画（淡入效果）

## 开发计划

- [x] PostgreSQL 数据库集成
- [x] 模型管理（从数据库读取）
- [x] 用户认证系统（JWT + bcrypt）
- [ ] API Key管理
- [ ] 使用量统计
- [ ] 计费系统