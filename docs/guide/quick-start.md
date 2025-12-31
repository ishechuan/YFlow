# 快速开始

本指南将帮助你快速搭建和运行 yflow 项目。

## 环境要求

- **Go** 1.20+
- **Node.js** 18+
- **pnpm** 9+
- **Bun** 1.0+
- **MySQL** 8.0+
- **Redis** 7.0+

## 1. 克隆项目

```bash
git clone https://github.com/cerebralatlas/yflow.git
cd yflow
```

## 2. 配置数据库

确保 MySQL 和 Redis 已启动，并创建数据库：

```sql
CREATE DATABASE i18n_flow DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## 3. 配置环境变量

### 后端配置

```bash
cd admin-backend
cp .env.example .env
```

编辑 `.env` 文件：

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=i18n_flow

REDIS_HOST=localhost
REDIS_PORT=6379

JWT_SECRET=your_jwt_secret
JWT_ACCESS_EXPIRY=24h
JWT_REFRESH_EXPIRY=7d

API_KEY=your_api_key_for_cli
```

### 前端配置

前端无需额外配置，API 地址可在开发时修改。

## 4. 启动后端

```bash
cd admin-backend

# 安装依赖
go mod tidy

# 启动开发服务器（带热重载）
air
```

后端将在 `http://localhost:8080` 启动。

## 5. 启动前端

```bash
cd admin-frontend

# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev
```

前端将在 `http://localhost:5170` 启动。

## 6. 首次访问

1. 打开浏览器访问 `http://localhost:5170`
2. 使用管理员账户登录（或联系系统管理员创建）
3. 开始创建你的第一个项目！

## 7. 安装 CLI 工具

```bash
cd cli

# 安装依赖
bun install

# 全局链接
bun link
```

现在可以使用 `yflow` 命令了。

## 下一步

- [了解基本概念 →](/guide/getting-started)
- [学习项目管理 →](/guide/project-guide)
- [使用 CLI 工具 →](/guide/cli-guide)
