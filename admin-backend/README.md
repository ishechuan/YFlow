# yflow 后端服务

yflow 是一个完整的国际化管理平台的后端服务，提供 RESTful API 接口，支持多语言翻译管理、用户权限控制、项目管理等核心功能。

## 技术栈

| 类别 | 技术 |
|------|------|
| **语言** | Go 1.23 |
| **Web 框架** | Gin 1.9.1 |
| **ORM** | GORM 1.30.0 |
| **数据库** | MySQL 8.0 |
| **缓存** | Redis 7.2 |
| **依赖注入** | Uber FX 1.20.0 |
| **日志** | Zap 1.27.0 |
| **API 文档** | Swaggo (Swagger) |
| **认证** | JWT (双令牌机制) |
| **限流** | Tollbooth |
| **验证** | govalidator, go-playground/validator |

## 项目结构

```
admin-backend/
├── cmd/
│   └── server/
│       └── main.go           # 应用入口点
├── internal/
│   ├── api/
│   │   ├── handlers/         # HTTP 请求处理器
│   │   ├── middleware/       # 中间件组件
│   │   ├── response/         # 统一响应格式
│   │   └── routes/           # 路由定义
│   ├── config/               # 配置管理
│   ├── container/            # FX 依赖注入容器
│   ├── di/                   # 依赖注入模块
│   ├── domain/               # 领域模型与接口
│   ├── dto/                  # 数据传输对象
│   ├── repository/           # 数据访问层
│   ├── service/              # 业务逻辑层
│   └── utils/                # 工具类
├── tests/                    # 测试目录
├── .air.toml                 # 热重载配置
├── .env.example              # 环境变量示例
├── docker-compose.yml        # Docker Compose 配置
├── Dockerfile                # Docker 构建文件
├── go.mod                    # Go 模块定义
└── go.sum                    # Go 依赖校验
```

### 架构说明

采用 **Clean Architecture** 架构设计，分层如下：

- **handlers/** - HTTP 请求处理层，负责接收和响应 HTTP 请求
- **middleware/** - 中间件层，包含认证、限流、安全等横切关注点
- **service/** - 业务逻辑层，实现核心业务规则
- **repository/** - 数据访问层，负责与数据库和缓存交互
- **domain/** - 领域层，定义实体和接口
- **dto/** - 数据传输对象，用于 API 请求和响应

## 快速开始

### 环境要求

- Go 1.23+
- MySQL 8.0
- Redis 7.2

### 安装依赖

```bash
go mod download
```

### 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，填入配置
```

### 开发模式运行

使用 air 实现热重载：

```bash
air
```

或直接运行：

```bash
go run cmd/server/main.go
```

### 构建生产版本

```bash
go build -o yflow ./cmd/server
./yflow
```

### 运行测试

```bash
go test ./...
go test ./... -coverprofile=coverage.out  # 带覆盖率
```

## 配置说明

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DB_USERNAME` | 数据库用户名 | root |
| `DB_PASSWORD` | 数据库密码 | - |
| `DB_HOST` | 数据库地址 | localhost |
| `DB_PORT` | 数据库端口 | 3306 |
| `DB_NAME` | 数据库名称 | i18n_flow |
| `JWT_SECRET` | JWT 访问令牌密钥 | - |
| `JWT_EXPIRATION_HOURS` | JWT 过期时间（小时） | 24 |
| `JWT_REFRESH_SECRET` | JWT 刷新令牌密钥 | - |
| `JWT_REFRESH_EXPIRATION_HOURS` | 刷新令牌过期时间（小时） | 168 |
| `CLI_API_KEY` | CLI 工具 API 密钥 | - |
| `ADMIN_USERNAME` | 初始管理员用户名 | admin |
| `ADMIN_PASSWORD` | 初始管理员密码 | admin123 |
| `REDIS_HOST` | Redis 地址 | localhost |
| `REDIS_PORT` | Redis 端口 | 6379 |
| `REDIS_PREFIX` | Redis 键前缀 | i18n_flow: |
| `LOG_LEVEL` | 日志级别 | info |
| `LOG_FORMAT` | 日志格式 | console |
| `LOG_OUTPUT` | 日志输出 | both |

### 密码复杂度要求

- **JWT Secret**: 至少 32 位，包含大小写字母、数字和特殊字符
- **API Key**: 至少 16 位

## API 文档

### 认证模块

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/login` | POST | 用户登录 |
| `/api/refresh` | POST | 刷新访问令牌 |
| `/api/user/info` | GET | 获取当前用户信息 |

### 用户管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/users` | GET | 获取用户列表 |
| `/api/users` | POST | 创建用户 |
| `/api/users/:id` | GET | 获取用户详情 |
| `/api/users/:id` | PUT | 更新用户 |
| `/api/users/:id` | DELETE | 删除用户 |

### 项目管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/projects` | GET | 获取项目列表 |
| `/api/projects` | POST | 创建项目 |
| `/api/projects/accessible` | GET | 获取可访问项目 |
| `/api/projects/:id` | GET | 获取项目详情 |
| `/api/projects/:id` | PUT | 更新项目 |
| `/api/projects/:id` | DELETE | 删除项目 |
| `/api/projects/:id/members` | GET | 获取项目成员 |
| `/api/projects/:id/members` | POST | 添加项目成员 |

### 语言管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/languages` | GET | 获取语言列表 |
| `/api/languages` | POST | 创建语言 |

### 翻译管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/translations/by-project/:id` | GET | 获取项目翻译 |
| `/api/translations/matrix/by-project/:id` | GET | 获取翻译矩阵视图 |
| `/api/translations/batch` | POST | 批量创建翻译 |
| `/api/translations/:id` | PUT | 更新翻译 |
| `/api/translations/:id` | DELETE | 删除翻译 |
| `/api/exports/project/:id` | GET | 导出翻译 |
| `/api/imports/project/:id` | POST | 导入翻译 |

### 邀请管理

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/invitations` | GET | 获取邀请列表 |
| `/api/invitations` | POST | 创建邀请 |
| `/api/invitations/:code` | GET | 使用邀请码注册 |

### CLI 接口

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/cli/scan` | POST | CLI 扫描接口（API Key 认证） |

### 监控接口

| 端点 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/stats` | GET | 统计信息 |
| `/stats/detailed` | GET | 详细统计 |
| `/swagger/*any` | GET | Swagger API 文档 |

### 响应格式

成功响应：

```json
{
  "success": true,
  "data": { ... }
}
```

错误响应：

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "错误描述",
    "details": "详细错误信息"
  }
}
```

## 安全特性

### 认证机制

- **JWT 双令牌机制**: 访问令牌（短期）+ 刷新令牌（长期）
- **API Key 认证**: 供 CLI 工具使用
- **密码加密**: 使用 BCrypt

### 权限控制

- **系统角色**: admin, member, viewer
- **项目角色**: owner, editor, viewer

### 限流策略

| 场景 | 限制 |
|------|------|
| 全局 | 100 请求/秒 |
| 登录 | 5 请求/秒 |
| API | 50 请求/秒 |
| 批量操作 | 2 请求/秒 |

### 安全中间件

- JWT 认证
- API Key 认证
- CORS 配置
- 安全 HTTP 头
- XSS 防护
- SQL 注入防护
- 请求验证

## Docker 部署

### 使用 Docker Compose

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 独立构建

```bash
# 构建镜像
docker build -t yflow-backend .

# 运行容器
docker run -p 8080:8080 yflow-backend
```

## 默认凭证

- **管理员用户名**: `admin`
- **管理员密码**: `admin123`

## 许可证

MIT License
