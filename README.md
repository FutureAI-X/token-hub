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

### 运行

```bash
./token-hub.exe
```

或者直接运行：

```bash
go run main.go
```

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `PORT` | 服务端口 | `3000` |
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
├── go.mod               # Go模块文件
└── go.sum               # 依赖校验文件
```

## 开发计划

- [ ] 用户认证系统
- [ ] 模型管理
- [ ] API Key管理
- [ ] 使用量统计
- [ ] 计费系统