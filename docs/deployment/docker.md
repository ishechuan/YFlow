# Docker 部署

使用 Docker 容器化部署 YFlow。

## 前置条件

- Docker Engine 20.10+
- Docker Compose V2
- MySQL 8.0 (外部或容器内)
- Redis 7.0 (外部或容器内)

### 使用 Docker Compose

#### 1. 准备环境变量

创建 `.env` 文件：

```env
# 数据库
MYSQL_ROOT_PASSWORD=your_root_password
MYSQL_DATABASE=i18n_flow
MYSQL_USER=i18n_flow
MYSQL_PASSWORD=your_db_password

# Redis
REDIS_PASSWORD=your_redis_password

# 后端
JWT_SECRET=your_jwt_secret
API_KEY=your_api_key

# 前端
VITE_API_URL=http://localhost:8080/api
```

#### 2. docker-compose.yml

```yaml
version: '3.8'

services:
  # MySQL 数据库
  mysql:
    image: mysql:8.0
    container_name: i18n-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: ${MYSQL_DATABASE}
      MYSQL_USER: ${MYSQL_USER}
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
    volumes:
      - mysql_data:/var/lib/mysql
    ports:
      - "3306:3306"
    command: --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis 缓存
  redis:
    image: redis:7-alpine
    container_name: i18n-redis
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # 后端 API
  backend:
    build:
      context: ./admin-backend
      dockerfile: Dockerfile
    container_name: i18n-backend
    restart: unless-stopped
    environment:
      DB_HOST: mysql
      DB_PORT: 3306
      DB_USER: ${MYSQL_USER}
      DB_PASSWORD: ${MYSQL_PASSWORD}
      DB_NAME: ${MYSQL_DATABASE}
      REDIS_HOST: redis
      REDIS_PORT: 6379
      REDIS_PASSWORD: ${REDIS_PASSWORD}
      JWT_SECRET: ${JWT_SECRET}
      API_KEY: ${API_KEY}
    ports:
      - "8080:8080"
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy

  # 前端
  frontend:
    build:
      context: ./admin-frontend
      dockerfile: Dockerfile
    container_name: i18n-frontend
    restart: unless-stopped
    environment:
      VITE_API_URL: /api
    ports:
      - "80:80"
    depends_on:
      - backend

volumes:
  mysql_data:
  redis_data:
```

#### 3. 后端 Dockerfile

```dockerfile
# admin-backend/Dockerfile
FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/.env.example .env.example

EXPOSE 8080

CMD ["./server"]
```

#### 4. 前端 Dockerfile

```dockerfile
# admin-frontend/Dockerfile
FROM node:18-alpine AS builder

WORKDIR /app

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm build

FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

#### 5. Nginx 配置

```nginx
# admin-frontend/nginx.conf
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://backend:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

#### 6. 启动服务

```bash
# 构建并启动所有服务
docker compose up -d --build

# 查看日志
docker compose logs -f

# 停止服务
docker compose down

# 停止并删除数据卷
docker compose down -v
```

### 生产环境优化

#### 1. 健康检查

```yaml
services:
  backend:
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

#### 2. 资源限制

```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
```

#### 3. 日志管理

```yaml
services:
  backend:
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "5"
```

### 验证部署

#### 检查服务状态

```bash
# 查看运行中的容器
docker compose ps

# 检查健康状态
docker inspect --format='{{.State.Health.Status}}' i18n-backend
```

#### 测试 API

```bash
# 登录 API
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password"}'
```

### 数据备份

#### 备份 MySQL

```bash
docker exec i18n-mysql mysqldump -u root -p i18n_flow > backup.sql
```

#### 备份 Redis

```bash
docker exec i18n-redis redis-cli BGSAVE
docker cp i18n-redis:/data/dump.rdb ./redis_backup.rdb
```

## 下一步

- [快速开始 →](/guide/quick-start)
- [架构概述 →](/architecture/overview)
