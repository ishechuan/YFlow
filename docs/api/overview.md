# API 概述

yflow 提供 RESTful API 用于集成第三方应用和 CLI 工具。

## 基础信息

| 项目 | 值 |
|------|-----|
| 基础 URL | `/api` |
| 认证方式 | JWT Token / API Key |
| 响应格式 | JSON |
| 版本 | v1 |

## 响应格式

### 成功响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    // 业务数据
  }
}
```

### 错误响应

```json
{
  "code": 400,
  "message": "invalid parameters",
  "errors": [
    {
      "field": "email",
      "message": "invalid email format"
    }
  ]
}
```

## HTTP 状态码

| 状态码 | 描述 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未认证或 Token 无效 |
| 403 | 无权限访问 |
| 404 | 资源不存在 |
| 500 | 服务器错误 |

## 认证方式

### 用户认证 (JWT)

用于前端管理和用户操作：

```bash
# 登录获取 Token
POST /api/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}

# 响应
{
  "accessToken": "eyJhbG...",
  "refreshToken": "eyJhbG...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "role": "member"
  }
}
```

### CLI 认证 (API Key)

用于 CLI 工具和自动化脚本：

```bash
# 在请求头中添加 API Key
X-API-Key: your-api-key-here
```

## API 端点

### 认证

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/login` | 用户登录 |
| POST | `/api/register` | 用户注册（需邀请码） |
| POST | `/api/refresh` | 刷新 Token |

### 项目

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/projects` | 获取项目列表 |
| POST | `/api/projects` | 创建项目 |
| GET | `/api/projects/:id` | 获取项目详情 |
| PUT | `/api/projects/:id` | 更新项目 |
| DELETE | `/api/projects/:id` | 删除项目 |

### 翻译

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/translations` | 获取翻译矩阵 |
| POST | `/api/translations` | 创建翻译 |
| PUT | `/api/translations` | 更新翻译 |
| DELETE | `/api/translations` | 删除翻译 |
| POST | `/api/translations/batch` | 批量操作 |
| GET | `/api/exports/:project_id` | 导出翻译 |
| POST | `/api/imports/:project_id` | 导入翻译 |

### CLI 专用

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/cli/auth` | CLI 认证 |
| GET | `/api/cli/translations` | 获取翻译 |
| POST | `/api/cli/keys` | 推送翻译键 |

## 下一步

- [了解认证方式 →](/api/authentication)
- [查看完整端点 →](/api/endpoints)
