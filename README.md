# Token Hub

下一代LLM网关和AI资产管理系统

## 快速开始

### 安装依赖

```bash
go mod tidy
```

### 编译

```bash
go build -o token-hub.exe .
```

### 运行后端

```bash
./token-hub.exe
```

或者直接运行：

```bash
go run main.go
```

### 运行前端

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

## API 接口

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
- **数据库**: 待定
- **缓存**: 待定

## 项目结构

```
token-hub/
├── main.go              # 主入口文件
├── router/              # 路由配置
│   └── router.go
├── controller/          # 控制器
│   └── model.go
├── model/               # 数据模型（待实现）
├── web/                 # 前端项目
│   ├── src/
│   │   ├── App.tsx      # 主页组件
│   │   ├── App.css      # 主页样式（现代化设计）
│   │   └── main.tsx     # 入口文件
│   ├── index.html       # HTML模板
│   ├── package.json     # 前端依赖
│   └── vite.config.ts   # Vite配置
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

- [ ] 用户认证系统
- [ ] 模型管理
- [ ] API Key管理
- [ ] 使用量统计
- [ ] 计费系统