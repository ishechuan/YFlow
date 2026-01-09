# 语流 - 国际化管理平台

[![Go版本](https://img.shields.io/badge/Go-1.23-blue)](https://go.dev/)
[![Vue 3](https://img.shields.io/badge/Vue-3-4FC08D)](https://vuejs.org/)
[![Rust](https://img.shields.io/badge/Rust-yellow)](https://www.rust-lang.org/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

简体中文 | [English](./README.md)

语流（YFlow）是一个专为现代开发团队设计的全栈国际化管理平台，提供完整的翻译管理、多语言工作流和自动化同步解决方案。

## 什么是语流？

语流简化了整个i18n生命周期——从翻译管理到基于CLI的同步自动化。平台采用 Go、Vue 3 和 Rust 构建，提供强大的管理后台、稳健的API接口和开发者友好的CLI工具，助您自动化本地化流程。

## 核心功能

- **团队协作** - 多语言管理、基于角色的权限控制、邀请系统和实时协作功能
- **CLI自动化** - 扫描源代码、自动检测缺失翻译，并通过CI/CD管道同步
- **管理后台** - 可视化翻译编辑器、语言管理、项目组织和综合数据分析
- **开发者体验** - Redis缓存API、Swagger文档、JWT认证和Docker部署支持

## 快速开始

### Docker Compose（推荐）

```bash
# 克隆仓库
git clone https://github.com/your-org/yflow.git
cd yflow

# 启动所有服务
docker compose up -d

# 访问平台
# 管理后台: http://localhost:5173
# API文档: http://localhost:8080/swagger/index.html
```

### 手动安装

```bash
# 后端
cd admin-backend
cp .env.example .env
go mod tidy
air

# 前端
cd admin-frontend
pnpm install
pnpm dev

# CLI
cd cli
cargo build --release
./target/release/yflow --help
```

## 项目结构

| 组件 | 描述 | 技术栈 | 链接 |
|------|------|--------|------|
| admin-backend | REST API 后端 | Go / Gin / GORM | [README](./admin-backend/README.md) |
| admin-frontend | 管理后台前端 | Vue 3 / Element Plus | [README](./admin-frontend/README.md) |
| cli | CLI 同步工具 | Rust / Clap | [README](./cli/README.md) |
| docs | 文档站点 | VitePress | [查看文档](./docs/) |

## 文档导航

- [快速开始](./docs/guide/quick-start.md) - 5分钟快速上手
- [入门指南](./docs/guide/getting-started.md) - 完整安装配置指南
- [架构设计](./docs/architecture/overview.md) - 系统架构概览
- [API参考](./docs/api/overview.md) - REST API 文档
- [部署指南](./docs/deployment/docker.md) - Docker 部署说明
- [CLI使用指南](./docs/guide/cli-guide.md) - CLI 工具教程

## 相关文档

- [后端详细文档](./admin-backend/README.md) - Go 后端开发指南
- [前端详细文档](./admin-frontend/README.md) - Vue 3 前端开发指南
- [CLI详细文档](./cli/README.md) - CLI 工具使用说明

## 默认账号

部署完成后，您可以使用以下账号登录：

- **用户名：** admin
- **密码：** admin123

## 贡献指南

欢迎贡献代码！请阅读我们的[贡献指南](./docs/guide/getting-started.md)了解行为规范和提交Pull Request的流程。

## 社区

- [GitHub Issues](https://github.com/your-org/yflow/issues) - 报告Bug或提出功能建议
- [Discord](https://discord.gg/your-server) - 加入社区讨论

## 开源协议

本项目采用 MIT 许可证 - 请查看 [LICENSE](LICENSE) 文件了解详情。
