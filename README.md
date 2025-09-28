# RoleChat - AI角色聊天应用

RoleChat是一个基于Vue 3和Go语言的AI角色聊天应用，支持用户创建和管理聊天对话，并与AI角色进行互动。

## 🏗️ 架构设计

### 整体架构
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端 (Vue 3)  │    │  后端 (Go+Gin)  │    │ 数据库 (PostgreSQL)│
│                 │<-->│                 │<-->│                 │
│ - Vue Router    │    │ - REST API      │    │ - 用户数据       │
│ - TypeScript    │    │ - JWT认证       │    │ - 聊天记录       │
│ - Vite          │    │ - GORM ORM      │    │ - 角色信息       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                               │
                               v
                    ┌─────────────────┐
                    │   AI服务        │
                    │ (智谱AI GLM-4)  │
                    │ - 文本生成       │
                    │ - 流式回复       │
                    └─────────────────┘
```

### 技术栈

#### 前端 (rolechat/)
- **框架**: Vue 3 + TypeScript
- **构建工具**: Vite
- **路由**: Vue Router 4
- **UI**: 自定义CSS + TailwindCSS
- **3D效果**: Three.js + OGL

#### 后端 (rolechat_back/)
- **语言**: Go 1.24
- **Web框架**: Gin
- **数据库ORM**: GORM
- **认证**: JWT (golang-jwt/jwt)
- **AI服务**: 智谱AI API
- **配置管理**: Viper
- **日志**: Zap

#### 数据库
- **主数据库**: PostgreSQL 16
- **连接池**: GORM + pgx驱动

### 目录结构

```
rolechat/
├── rolechat/              # 前端应用
│   ├── src/
│   │   ├── components/    # Vue组件
│   │   ├── views/         # 页面视图
│   │   ├── router/        # 路由配置
│   │   ├── layouts/       # 布局组件
│   │   └── assets/        # 静态资源
│   └── public/            # 公共资源
│
├── rolechat_back/         # 后端应用
│   ├── cmd/server/        # 应用入口
│   ├── internal/          # 内部包
│   │   ├── handler/       # HTTP处理器
│   │   ├── service/       # 业务逻辑层
│   │   ├── repository/    # 数据访问层
│   │   ├── models/        # 数据模型
│   │   └── app/           # 应用配置
│   ├── pkg/               # 公共包
│   │   ├── ai/            # AI客户端
│   │   ├── config/        # 配置管理
│   │   ├── logger/        # 日志工具
│   │   └── utils/         # 工具函数
│   └── configs/           # 配置文件
│
└── db/                    # 数据库相关
    └── docker-compose.yml # 数据库容器配置
```

## 🚀 快速开始

### 环境要求

- **Node.js**: >= 18.0.0
- **Go**: >= 1.24.0  
- **PostgreSQL**: >= 16.0
- **Docker** (可选): 用于快速启动数据库

### 1. 克隆项目

```bash
git clone https://github.com/karma0aliu/RoleChat.git
cd RoleChat
```

### 2. 数据库设置

#### 方式一：使用Docker (推荐)

```bash
cd db
docker-compose up -d
```

这将启动一个PostgreSQL数据库实例：
- 端口: 5432
- 数据库名: db_rolechat
- 用户名: rolechat_user01
- 密码: rolechatuser01

#### 方式二：手动安装PostgreSQL

1. 安装PostgreSQL 16
2. 创建数据库和用户：
```sql
CREATE DATABASE db_rolechat;
CREATE USER rolechat_user01 WITH PASSWORD 'rolechatuser01';
GRANT ALL PRIVILEGES ON DATABASE db_rolechat TO rolechat_user01;
```

### 3. 后端配置与启动

#### 配置文件
编辑 `rolechat_back/configs/config.yaml`：

```yaml
server:
  port: "8080"
  mode: "debug"

database:
  host: "localhost"
  port: 5432
  user: "rolechat_user01"
  password: "rolechatuser01"
  dbname: "db_rolechat"
  sslmode: "disable"

jwt:
  access_secret: "your-access-secret-key"
  refresh_secret: "your-refresh-secret-key"
  access_expires_mins: 5
  refresh_expires_hours: 24

api_key:
  zhipuai_api_key: "your-zhipu-ai-api-key"  # 智谱AI API密钥
```

#### 启动后端服务

```bash
cd rolechat_back

# 下载依赖
go mod tidy

# 运行服务
go run cmd/server/main.go
```

后端服务将在 `http://localhost:8080` 启动。

### 4. 前端配置与启动

```bash
cd rolechat

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端应用将在 `http://localhost:5173` 启动。

## 📱 功能特性

### 核心功能

1. **用户系统**
   - 用户注册/登录
   - JWT令牌认证
   - 用户配置文件管理

2. **聊天系统**
   - 创建聊天话题
   - 发送和接收消息
   - 消息历史记录
   - 实时消息更新

3. **AI角色对话**
   - 基于智谱AI GLM-4模型
   - 自定义角色人格
   - 流式文本生成
   - 角色一致性维护

4. **用户界面**
   - 响应式设计
   - 3D卡片效果
   - 现代化UI组件
   - 暗色/亮色主题

### API接口

#### 认证相关
- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/refresh` - 刷新令牌

#### 用户相关
- `GET /api/me` - 获取用户信息

#### 聊天相关  
- `POST /api/chat/message` - 发送消息
- `GET /api/chat/topics` - 获取话题列表（默认返回100个）
- `GET /api/chat/topics/limit?n=数量` - 根据参数获取前n个话题
- `GET /api/chat/topics/:id/messages` - 获取消息列表

#### AI相关
- `POST /api/chat/role-reply` - AI角色回复
- `POST /api/chat/role-reply/stream` - 流式AI回复

## 🛠️ 开发指南

### 开发环境设置

1. **后端开发**
```bash
cd rolechat_back
go mod tidy
go run cmd/server/main.go
```

2. **前端开发**
```bash  
cd rolechat
npm install
npm run dev
```

3. **构建生产版本**
```bash
# 前端构建
cd rolechat
npm run build

# 后端构建
cd rolechat_back
go build -o bin/server cmd/server/main.go
```

### 数据库迁移

应用启动时会自动创建所需的数据表：
- users (用户表)
- topics (话题表) 
- messages (消息表)
- role_personas (角色人格表)

### 环境变量

可以通过环境变量覆盖配置文件设置：
- `SERVER_PORT`: 服务端口
- `DB_HOST`: 数据库主机
- `DB_PORT`: 数据库端口
- `ZHIPU_API_KEY`: 智谱AI API密钥

## 📦 部署

### Docker部署 (推荐)

创建 `Dockerfile` 用于容器化部署：

```dockerfile
# 后端Dockerfile示例
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY rolechat_back/ .
RUN go mod tidy && go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs
CMD ["./server"]
```

### 传统部署

1. **构建应用**
```bash
# 构建前端
cd rolechat && npm run build

# 构建后端  
cd rolechat_back && go build -o bin/server cmd/server/main.go
```

2. **配置反向代理**
使用Nginx等反向代理服务器：
```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        root /path/to/rolechat/dist;
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 🔗 相关链接

- [Vue 3 文档](https://vuejs.org/)
- [Gin Web Framework](https://gin-gonic.com/)
- [GORM 文档](https://gorm.io/)
- [智谱AI API](https://open.bigmodel.cn/)

## ❓ 常见问题

### Q: 如何获取智谱AI API密钥？
A: 访问 [智谱AI开放平台](https://open.bigmodel.cn/) 注册账号并申请API密钥。

### Q: 数据库连接失败怎么办？
A: 检查PostgreSQL服务是否正在运行，以及配置文件中的数据库连接信息是否正确。

### Q: 前端页面无法加载？
A: 确保后端API服务正在运行，并检查网络连接和CORS配置。

### Q: AI回复功能不工作？
A: 验证智谱AI API密钥是否有效，并检查API调用频率是否超限。
