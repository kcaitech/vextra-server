# Vextra Server

Server-side implementation of Vextra, providing backend services for core functionalities such as document collaboration, team management, real-time communication, and more.

## Project Overview

Vextra Server is a high-performance collaboration server developed in Go, providing complete backend service support for the Vextra platform. The project adopts a modern microservice architecture, supporting core functionalities such as real-time document collaboration, team management, permission control, file storage, and more.

## Features

- 📄 **Document Collaboration**: Supports real-time multi-user collaborative editing with WebSocket-based real-time communication
- 👥 **Team Management**: Complete team creation, member management, and permission allocation system
- 🔐 **Permission Control**: Fine-grained document access permission control with support for multiple permission levels
- 💬 **Comment System**: Document comments and feedback functionality
- 📁 **File Management**: Supports multiple storage backends (MinIO, Alibaba Cloud OSS, AWS S3)
- 🔍 **Content Review**: Integrated content security review from Alibaba Cloud and Baidu Cloud
- 📊 **Project Management**: Project creation, categorization, favorites, and other management features
- 🚀 **High Performance**: High-performance HTTP service based on Gin framework
- 🔄 **Real-time Sync**: WebSocket real-time data synchronization and status updates

## Tech Stack

### Backend
- **Language**: Go 1.22+
- **Web Framework**: Gin
- **Database**: MySQL + MongoDB
- **Cache**: Redis
- **Storage**: MinIO / Alibaba Cloud OSS / AWS S3
- **Real-time Communication**: WebSocket (Gorilla)
- **Authentication**: JWT
- **Configuration Management**: YAML

### Client
- **Language**: TypeScript
- **Build Tool**: Vite
- **HTTP Client**: Axios
- **Type Validation**: Zod

## Quick Start

### Requirements

- Go 1.22+
- MySQL 8.0+
- MongoDB 6.0+
- Redis 6.0+
- MinIO (optional, for local storage)

### Install Dependencies

```bash
# Set Go proxy (recommended for users in China)
go env -w GOPROXY=https://goproxy.cn,direct
# Or use Alibaba Cloud proxy
go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct

# Install dependencies
go mod tidy
```

### Configuration

Copy the configuration file template and modify it:

```bash
cp config/config.yaml.example config/config.yaml
```

Main configuration items:
- Database connection information
- Redis connection configuration
- Storage service configuration
- Authentication server configuration
- Middleware settings

### Run Service

```bash
# Run with default configuration
go run main.go

# Specify configuration file
go run main.go -config config/config.yaml

# Specify port
go run main.go -port 8080

# Specify frontend file path
go run main.go -web /path/to/web/files
```

### Build

```bash
# Build executable file
go build -o kcserver main.go

# Run the built file
./kcserver -config config/config.yaml -port 8080
```

## Project Structure

```
kcserver/
├── api/                    # API routes and handlers
│   ├── v1/                # API v1 version
│   └── index.go           # Route registration
├── client/                 # TypeScript client library
│   ├── request/           # HTTP request encapsulation
│   ├── ws/                # WebSocket client
│   └── tests/             # Client tests
├── common/                 # Common constants and utilities
├── config/                 # Configuration management
├── handlers/               # Business logic handlers
│   ├── document/          # Document-related handlers
│   ├── user/              # User-related handlers
│   └── ws/                # WebSocket handlers
├── middlewares/            # Middlewares
├── models/                 # Data models
├── providers/              # External service providers
│   ├── auth/              # Authentication service
│   ├── mongo/             # MongoDB connection
│   ├── redis/             # Redis connection
│   ├── safereview/        # Content security review
│   └── storage/           # Storage service
├── services/               # Business service layer
├── utils/                  # Utility functions
└── main.go                # Program entry point
```

## API Documentation

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh token
- `POST /api/v1/auth/logout` - User logout

### Document Management
- `GET /api/v1/document` - Get document list
- `POST /api/v1/document` - Create document
- `GET /api/v1/document/:id` - Get document details
- `PUT /api/v1/document/:id` - Update document
- `DELETE /api/v1/document/:id` - Delete document

### Team Management
- `GET /api/v1/team` - Get team list
- `POST /api/v1/team` - Create team
- `GET /api/v1/team/:id` - Get team details
- `PUT /api/v1/team/:id` - Update team information

### WebSocket Interface
- `/ws` - WebSocket connection endpoint
- Supports real-time document collaboration, comments, selection synchronization, and other features

## Development Guide

### Adding New API Endpoints

1. Create new handler files in the `api/v1/` directory
2. Register routes in `api/index.go`
3. Implement business logic in the `handlers/` directory
4. Define data models in the `models/` directory

### Adding New Storage Backends

1. Implement the interface in `providers/storage/base.go`
2. Register new providers in `providers/storage/storage.go`
3. Update configuration files to support new storage options

## License

This project is licensed under the AGPL-3.0 License - see the [LICENSE](LICENSE.txt) file for details.

## Contact

- Website: [https://kcaitech.com](https://kcaitech.com)