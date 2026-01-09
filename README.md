# YFlow - Internationalization Management Platform

[![Go Version](https://img.shields.io/badge/Go-1.23-blue)](https://go.dev/)
[![Vue 3](https://img.shields.io/badge/Vue-3-4FC08D)](https://vuejs.org/)
[![Rust](https://img.shields.io/badge/Rust-yellow)](https://www.rust-lang.org/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

[简体中文](./README.zh-CN.md) | English

YFlow is a full-stack internationalization management platform designed for modern development teams. It provides a complete solution for managing translations, languages, and localization workflows across your applications.

## What is YFlow?

YFlow streamlines the entire i18n lifecycle — from translation management to CLI-based sync automation. Built with Go, Vue 3, and Rust, it offers a powerful admin dashboard, robust API, and developer-friendly CLI tools to automate your localization pipeline.

## Features

- **Team Collaboration** - Multi-language management with role-based access control, invitation system, and real-time collaboration features
- **CLI Automation** - Scan source code, auto-detect missing translations, and sync with CI/CD pipelines using the Rust-based CLI
- **Admin Dashboard** - Visual translation editor, language management, project organization, and comprehensive analytics
- **Developer Experience** - Redis-cached APIs, Swagger documentation, JWT authentication, and Docker deployment support

## Quick Start

### Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/your-org/yflow.git
cd yflow

# Start all services
docker compose up -d

# Access the platform
# Admin Dashboard: http://localhost:5173
# API Docs: http://localhost:8080/swagger/index.html
```

### Manual Setup

```bash
# Backend
cd admin-backend
cp .env.example .env
go mod tidy
air

# Frontend
cd admin-frontend
pnpm install
pnpm dev

# CLI
cd cli
cargo build --release
./target/release/yflow --help
```

## Project Structure

| Component | Description | Tech Stack | Links |
|-----------|-------------|------------|-------|
| admin-backend | REST API backend | Go / Gin / GORM | [README](./admin-backend/README.md) |
| admin-frontend | Admin dashboard | Vue 3 / Element Plus | [README](./admin-frontend/README.md) |
| cli | CLI sync tool | Rust / Clap | [README](./cli/README.md) |
| docs | Documentation site | VitePress | [View Docs](./docs/) |

## Documentation

- [Getting Started](./docs/guide/getting-started.md) - Complete setup guide
- [Quick Start](./docs/guide/quick-start.md) - 5-minute quick start
- [Architecture](./docs/architecture/overview.md) - System design overview
- [API Reference](./docs/api/overview.md) - REST API documentation
- [Deployment](./docs/deployment/docker.md) - Docker deployment guide
- [CLI Guide](./docs/guide/cli-guide.md) - CLI usage tutorial

## Default Credentials

After deployment, you can log in with:

- **Username:** admin
- **Password:** admin123

## Contributing

Contributions are welcome! Please read our [contributing guide](CONTRIBUTING.md) for details.

## Community

- [GitHub Issues](https://github.com/ishechuan/yflow/issues) - Report bugs or request features
- [GitHub Discussions](https://github.com/ishechuan/yflow/discussions) - Ask questions and discuss
- [Code of Conduct](.github/CODE_OF_CONDUCT.md) - Our community guidelines

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
