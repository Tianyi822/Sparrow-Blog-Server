# 麻雀博客后端系统 (Sparrow-Server)

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-GPL--3.0-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](#)

一个功能完整、高性能的 Go 语言博客后端系统，专为个人博客和小型网站设计。

## ✨ 项目特性

- 🚀 **高性能架构**：基于 Gin 框架，支持高并发访问
- 💾 **内置缓存系统**：可持久化缓存，无需外部 Redis 等缓存组件
- 🔍 **智能搜索引擎**：基于 Bleve 的全文搜索，支持中文分词
- 🔐 **安全可靠**：JWT 认证，HTTPS 支持，完善的权限控制
- 📁 **云存储支持**：集成阿里云 OSS 对象存储
- 📊 **完善日志系统**：结构化日志，支持日志轮转和压缩
- ⚙️ **灵活配置**：YAML 配置文件，支持多环境部署
- 🎯 **RESTful API**：标准化 API 设计，易于前端集成

## 🛠 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.23+ | 后端开发语言 |
| Gin | 1.10+ | Web 框架 |
| GORM | 1.26+ | ORM 框架 |
| MySQL | 8.0+ | 主数据库 |
| Bleve | 2.5+ | 全文搜索引擎 |
| JWT | 5.2+ | 身份认证 |
| Zap | 1.27+ | 日志框架 |
| Lumberjack | 2.2+ | 日志轮转 |
| Gomail | 2.0+ | 邮件发送 |

## 📁 项目结构

```
H2Blog-Server/
├── cache/                  # 缓存模块
│   ├── aof/               # AOF 持久化
│   └── common/            # 缓存通用类型
├── env/                   # 环境配置
├── internal/              # 内部模块
│   ├── model/             # 数据模型
│   │   ├── dto/           # 数据传输对象
│   │   ├── po/            # 持久化对象
│   │   └── vo/            # 视图对象
│   ├── repositories/      # 数据访问层
│   └── services/          # 业务逻辑层
├── pkg/                   # 公共包
│   ├── config/            # 配置管理
│   ├── email/             # 邮件服务
│   ├── logger/            # 日志服务
│   ├── utils/             # 工具函数
│   └── webjwt/            # JWT 认证
├── routers/               # 路由层
│   ├── adminrouter/       # 管理员路由
│   ├── middleware/        # 中间件
│   ├── resp/              # 响应处理
│   └── webrouter/         # Web 路由
├── searchengine/          # 搜索引擎
├── storage/               # 存储层
│   ├── db/                # 数据库
│   └── ossstore/          # 对象存储
└── main.go               # 程序入口
```

## 🚀 快速开始

### 环境要求

- Go 1.23 或更高版本
- MySQL 8.0 或更高版本
- Git

### 安装步骤

1. **克隆项目**
   ```bash
   git clone https://github.com/your-username/H2Blog-Server.git
   cd H2Blog-Server
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **配置数据库**
   - 创建 MySQL 数据库
   - 配置数据库连接信息（见配置说明）

4. **启动服务**
   ```bash
   # 开发环境
   go run main.go --env debug
   
   # 生产环境
   go run main.go
   ```

5. **验证安装**
   - 开发环境：访问 `http://localhost:8080`
   - 生产环境：访问 `https://your-domain.com`

## ⚙️ 配置说明

系统使用 YAML 格式的配置文件，位于用户主目录的 `.h2blog/config/sparrow_blog_config.yaml`。

### 主要配置项

```yaml
# 服务器配置
server:
  port: 8080                    # 服务端口
  token_key: "your-secret-key"  # JWT 密钥
  token_expire_duration: 7      # Token 过期时间（天）
  ssl:
    cert_file: "/path/to/cert.pem"  # SSL 证书文件
    key_file: "/path/to/key.pem"    # SSL 私钥文件

# 数据库配置
mysql:
  host: "localhost"
  port: 3306
  user: "root"
  password: "your-password"
  database: "h2blog"
  max_open: 100
  max_idle: 10

# 缓存配置
cache:
  aof:
    enable: true
    path: "./data/cache.aof"
    max_size: 100
    compress: true

# 搜索引擎配置
search_engine:
  index_path: "./data/search_index"

# OSS 配置（可选）
oss:
  endpoint: "oss-cn-hangzhou.aliyuncs.com"
  access_key_id: "your-access-key"
  access_key_secret: "your-secret-key"
  bucket: "your-bucket"
```

## 📚 API 文档

### 认证接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/admin/login` | 管理员登录 |
| POST | `/api/admin/logout` | 管理员登出 |

### 博客管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/web/blogs` | 获取博客列表 |
| GET | `/api/web/blog/:id` | 获取博客详情 |
| POST | `/api/admin/blog` | 创建博客 |
| PUT | `/api/admin/blog/:id` | 更新博客 |
| DELETE | `/api/admin/blog/:id` | 删除博客 |

### 分类标签

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/web/categories` | 获取分类列表 |
| GET | `/api/web/tags` | 获取标签列表 |
| POST | `/api/admin/category` | 创建分类 |
| POST | `/api/admin/tag` | 创建标签 |

### 搜索功能

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/web/search` | 搜索博客 |

## 🚀 部署指南

### Docker 部署

1. **构建镜像**
   ```bash
   docker build -t h2blog-server .
   ```

2. **运行容器**
   ```bash
   docker run -d \
     --name h2blog \
     -p 8080:8080 \
     -v /path/to/config:/app/config \
     -v /path/to/data:/app/data \
     h2blog-server
   ```

### 传统部署

1. **编译程序**
   ```bash
   go build -o h2blog-server main.go
   ```

2. **配置系统服务**
   ```bash
   # 创建 systemd 服务文件
   sudo vim /etc/systemd/system/h2blog.service
   ```

3. **启动服务**
   ```bash
   sudo systemctl enable h2blog
   sudo systemctl start h2blog
   ```

### Nginx 反向代理

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 🔧 开发指南

### 开发环境搭建

1. **安装开发工具**
   ```bash
   # 安装 Air（热重载）
   go install github.com/cosmtrek/air@latest
   
   # 安装 golangci-lint（代码检查）
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

2. **启动开发服务**
   ```bash
   air
   ```

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 使用 `golangci-lint` 进行代码检查
- 编写单元测试，保持测试覆盖率 > 80%

### 测试

```bash
# 运行所有测试
go test ./...

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🤝 贡献指南

我们欢迎所有形式的贡献！

### 贡献方式

1. **报告 Bug**：在 Issues 中详细描述问题
2. **功能建议**：提出新功能的想法和建议
3. **代码贡献**：提交 Pull Request
4. **文档改进**：完善项目文档

### 提交流程

1. Fork 本仓库
2. 创建功能分支：`git checkout -b feature/amazing-feature`
3. 提交更改：`git commit -m 'Add some amazing feature'`
4. 推送分支：`git push origin feature/amazing-feature`
5. 提交 Pull Request

### 代码提交规范

```
type(scope): description

[optional body]

[optional footer]
```

类型说明：
- `feat`: 新功能
- `fix`: 修复 Bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

## 📄 许可证

本项目采用 [GPL-3.0](LICENSE) 许可证。

## 🙏 致谢

感谢以下开源项目：

- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [GORM](https://github.com/go-gorm/gorm) - ORM 库
- [Bleve](https://github.com/blevesearch/bleve) - 全文搜索引擎
- [Zap](https://github.com/uber-go/zap) - 日志库

---

⭐ 如果这个项目对你有帮助，请给我们一个 Star！