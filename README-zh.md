# Vextra Server

Vextra 的服务器端实现，提供文档协作、团队管理、实时通信等核心功能的后端服务。

## 项目简介

Vextra Server 是一个基于 Go 语言开发的高性能协作服务器，为 Vextra 平台提供完整的后端服务支持。项目采用现代化的微服务架构，支持文档实时协作、团队管理、权限控制、文件存储等核心功能。

## 功能特性

- 📄 **文档协作**: 支持实时多人协作编辑，基于 WebSocket 的实时通信
- 👥 **团队管理**: 完整的团队创建、成员管理、权限分配系统
- 🔐 **权限控制**: 细粒度的文档访问权限控制，支持多种权限级别
- 💬 **评论系统**: 文档内评论和反馈功能
- 📁 **文件管理**: 支持多种存储后端（MinIO、阿里云OSS、AWS S3）
- 🔍 **内容审核**: 集成阿里云和百度云的内容安全审核
- 📊 **项目管理**: 项目创建、分类、收藏等管理功能
- 🚀 **高性能**: 基于 Gin 框架的高性能 HTTP 服务
- 🔄 **实时同步**: WebSocket 实时数据同步和状态更新

## 技术栈

### 后端
- **语言**: Go 1.22+
- **Web框架**: Gin
- **数据库**: MySQL + MongoDB
- **缓存**: Redis
- **存储**: MinIO / 阿里云OSS / AWS S3
- **实时通信**: WebSocket (Gorilla)
- **认证**: JWT
- **配置管理**: YAML

### 客户端
- **语言**: TypeScript
- **构建工具**: Vite
- **HTTP客户端**: Axios
- **类型验证**: Zod

## 快速开始

### 环境要求

- Go 1.22+
- MySQL 8.0+
- MongoDB 6.0+
- Redis 6.0+
- MinIO (可选，用于本地存储)

### 安装依赖

```bash
# 设置 Go 代理（国内用户推荐）
go env -w GOPROXY=https://goproxy.cn,direct
# 或者使用阿里云代理
go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct

# 安装依赖
go mod tidy
```

### 配置

复制配置文件模板并修改：

```bash
cp config/config.yaml.example config/config.yaml
```

主要配置项：
- 数据库连接信息
- Redis 连接配置
- 存储服务配置
- 认证服务器配置
- 中间件设置

### 运行服务

```bash
# 使用默认配置运行
go run main.go

# 指定配置文件
go run main.go -config config/config.yaml

# 指定端口
go run main.go -port 8080

# 指定前端文件路径
go run main.go -web /path/to/web/files
```

### 构建

```bash
# 构建可执行文件
go build -o kcserver main.go

# 运行构建后的文件
./kcserver -config config/config.yaml -port 8080
```

## 项目结构

```
kcserver/
├── api/                    # API 路由和处理器
│   ├── v1/                # API v1 版本
│   └── index.go           # 路由注册
├── client/                 # TypeScript 客户端库
│   ├── request/           # HTTP 请求封装
│   ├── ws/                # WebSocket 客户端
│   └── tests/             # 客户端测试
├── common/                 # 通用常量和工具
├── config/                 # 配置管理
├── handlers/               # 业务逻辑处理器
│   ├── document/          # 文档相关处理
│   ├── user/              # 用户相关处理
│   └── ws/                # WebSocket 处理
├── middlewares/            # 中间件
├── models/                 # 数据模型
├── providers/              # 外部服务提供者
│   ├── auth/              # 认证服务
│   ├── mongo/             # MongoDB 连接
│   ├── redis/             # Redis 连接
│   ├── safereview/        # 内容安全审核
│   └── storage/           # 存储服务
├── services/               # 业务服务层
├── utils/                  # 工具函数
└── main.go                # 程序入口
```

## API 文档

### 认证相关
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/refresh` - 刷新令牌
- `POST /api/v1/auth/logout` - 用户登出

### 文档管理
- `GET /api/v1/document` - 获取文档列表
- `POST /api/v1/document` - 创建文档
- `GET /api/v1/document/:id` - 获取文档详情
- `PUT /api/v1/document/:id` - 更新文档
- `DELETE /api/v1/document/:id` - 删除文档

### 团队管理
- `GET /api/v1/team` - 获取团队列表
- `POST /api/v1/team` - 创建团队
- `GET /api/v1/team/:id` - 获取团队详情
- `PUT /api/v1/team/:id` - 更新团队信息

### WebSocket 接口
- `/ws` - WebSocket 连接端点
- 支持实时文档协作、评论、选择同步等功能

## 开发指南

### 添加新的 API 端点

1. 在 `api/v1/` 目录下创建新的处理器文件
2. 在 `api/index.go` 中注册路由
3. 在 `handlers/` 目录下实现业务逻辑
4. 在 `models/` 目录下定义数据模型

### 添加新的存储后端

1. 实现 `providers/storage/base.go` 中的接口
2. 在 `providers/storage/storage.go` 中注册新的提供者
3. 更新配置文件以支持新的存储选项


## 许可证

本项目采用 AGPL-3.0 许可证 - 查看 [LICENSE](LICENSE.txt) 文件了解详情。

## 联系方式

- Website: [https://kcaitech.com](https://kcaitech.com)