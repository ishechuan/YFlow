# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

YFlow is a full-stack internationalization management platform with:
- **admin-backend**: Go/Gin REST API for managing translations, languages, projects, and users
- **admin-frontend**: Vue 3 admin dashboard for platform management
- **cli**: Bun CLI tool for scanning and syncing translations

## Commands

### Backend (Go)

```bash
# Development with hot reload
cd admin-backend && air

# Run all tests
go test ./...

# Run tests in specific package
go test ./internal/service/...

# Run single test file
go test -v ./tests/service/cache_test.go

# Build binary
go build -o ./tmp/main.exe ./cmd/server

# Generate Swagger docs
swag init -g cmd/server/main.go
```

### Frontend (Vue 3)

```bash
cd admin-frontend

# Install dependencies (use pnpm, not npm)
pnpm install

# Development server
pnpm dev

# Type check
pnpm type-check

# Build for production
pnpm build

# Run tests
pnpm test:unit

# Lint (oxlint + eslint)
pnpm lint

# Format code
pnpm format
```

### CLI (Bun)

```bash
cd cli

# Development
bun run ./src/index.ts

# Build binary
bun build ./src/index.ts --outfile ./bin/yflow

# Run built binary
./bin/yflow
```

### Docker Compose

```bash
# Start all services
docker compose up -d

# View logs
docker compose logs -f

# Stop all services
docker compose down
```

## Architecture

### Backend (Go)

The backend follows a layered architecture with dependency injection via Uber FX:

```
admin-backend/internal/
├── api/           # HTTP handlers, middleware, routes, response utils
├── config/        # Configuration loading
├── container/     # FX dependency injection setup
├── di/            # FX module providers
├── domain/        # Entities, repository interfaces, errors
├── dto/           # Data transfer objects
├── repository/    # Data access layer (GORM with MySQL + Redis caching)
├── service/       # Business logic with cache decorators
└── utils/         # Utilities (JWT, security, logging)
```

Key patterns:
- Repository layer handles all database operations with Redis caching
- Services contain business logic and call repositories; cached variants wrap base services
- Handlers receive HTTP requests, call services, and format responses
- JWT authentication with dual tokens (access + refresh)
- FX module in `internal/di/module.go` registers all dependencies

Middleware (applied in `cmd/server/main.go`):
- Request ID, logging, security headers, rate limiting, SQL injection prevention
- XSS protection, input validation, CORS, error handling

### Frontend (Vue 3)

```
admin-frontend/src/
├── layouts/       # Page layouts (MainLayout)
├── router/        # Vue Router configuration
├── services/      # Axios API client wrappers
├── stores/        # Pinia state management
├── types/         # TypeScript interfaces
└── views/         # Page components
```

Key integrations:
- TanStack Vue Query for data fetching/caching
- Pinia for global state
- Element Plus for UI components

## API Documentation

Swagger UI available at: `http://localhost:8080/swagger/index.html`

### Authentication

Two authentication systems:
1. **User Auth**: JWT access + refresh tokens via `/api/login`, `/api/refresh`
2. **CLI Auth**: API Key authentication for `/api/cli/scan`

## Database

- MySQL 8.0 with GORM
- Redis for caching (configured in `internal/repository/redis_client.go`)
