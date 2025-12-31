# admin-frontend

yflow 管理后台前端，基于 Vue 3 + TypeScript + Element Plus 构建。

## 技术栈

- **框架**: Vue 3 (Composition API)
- **语言**: TypeScript
- **UI 组件库**: Element Plus
- **状态管理**: Pinia
- **路由**: Vue Router
- **数据获取**: TanStack Vue Query
- **构建工具**: Vite
- **代码质量**: ESLint + oxlint + Prettier
- **单元测试**: Vitest

## 项目结构

```
admin-frontend/
├── src/
│   ├── assets/           # 静态资源
│   ├── layouts/          # 布局组件
│   ├── router/           # 路由配置
│   ├── services/         # API 服务层
│   ├── stores/           # Pinia 状态管理
│   ├── types/            # TypeScript 类型定义
│   ├── views/            # 页面组件
│   ├── App.vue           # 根组件
│   └── main.ts           # 应用入口
├── public/               # 公共静态资源
├── index.html            # HTML 模板
├── vite.config.ts        # Vite 配置
├── tsconfig.json         # TypeScript 配置
└── package.json          # 项目依赖
```

## 功能特性

### 页面功能

| 页面 | 路径 | 说明 |
|------|------|------|
| 登录 | `/login` | 用户登录 |
| 注册 | `/register` | 用户注册 |
| 仪表板 | `/dashboard` | 系统概览和统计信息 |
| 项目管理 | `/projects` | 项目 CRUD 操作 |
| 语言管理 | `/languages` | 语言配置管理 |
| 翻译管理 | `/translations` | 翻译内容管理 |
| 用户管理 | `/users` | 用户管理 (仅管理员) |
| 邀请管理 | `/invitations` | 邀请码管理 (仅管理员) |

### 核心功能

- **认证系统**: JWT 认证，支持 Token 刷新
- **权限控制**: 基于角色的访问控制 (RBAC)
- **响应式布局**: 适配桌面和移动端
- **数据可视化**: 仪表板统计卡片
- **主题支持**: 深色侧边栏 + 浅色内容区

## 快速开始

### 环境要求

- Node.js: `^20.19.0 || >=22.12.0`
- pnpm: 推荐使用 pnpm 管理依赖

### 安装依赖

```bash
pnpm install
```

### 开发模式

```bash
pnpm dev
```

启动后访问 <http://localhost:5173>

### 类型检查

```bash
pnpm type-check
```

### 代码检查

```bash
# ESLint
pnpm lint:eslint

# oxlint
pnpm lint:oxlint

# 完整检查
pnpm lint
```

### 代码格式化

```bash
pnpm format
```

### 构建生产版本

```bash
pnpm build
```

### 预览生产版本

```bash
pnpm preview
```

### 运行单元测试

```bash
pnpm test:unit
```

## API 集成

前端通过 `src/services/api.ts` 中的 Axios 实例与后端 API 通信：

- 自动附加 JWT Token 到请求头
- 响应拦截器统一处理错误
- 支持 Token 自动刷新

## 认证流程

1. 用户登录后，Token 和 Refresh Token 存储于 localStorage
2. Pinia store (`src/stores/auth.ts`) 管理认证状态
3. 路由守卫检查页面访问权限
4. 管理员可访问用户管理和邀请管理页面

## 主要依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| vue | ^3.5.25 | 核心框架 |
| vue-router | ^4.6.3 | 路由管理 |
| pinia | ^3.0.4 | 状态管理 |
| element-plus | ^2.13.0 | UI 组件库 |
| axios | ^1.13.2 | HTTP 客户端 |
| @tanstack/vue-query | ^5.92.1 | 数据获取/缓存 |
| @vueuse/core | ^14.1.0 | 组合式工具函数 |

## License

MIT
