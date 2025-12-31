# 认证

了解 yflow 的两种认证方式。

## 用户认证 (JWT)

适用于前端管理界面和用户操作。

### 登录

```http
POST /api/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "your_password"
}
```

**响应**：

```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "role": "member"
  }
}
```

### 使用 Token

将 Access Token 放入请求头：

```http
GET /api/projects
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

### 刷新 Token

Access Token 过期时，使用 Refresh Token 获取新的 Token：

```http
POST /api/refresh
Content-Type: application/json

{
  "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
}
```

**响应**：

```json
{
  "accessToken": "new_access_token...",
  "refreshToken": "new_refresh_token..."
}
```

### 注册

使用邀请码注册新用户：

```http
POST /api/register
Content-Type: application/json

{
  "email": "newuser@example.com",
  "password": "password123",
  "invitationCode": "INVITE123"
}
```

## CLI 认证 (API Key)

适用于 CLI 工具和自动化脚本。

### 获取 API 密钥

在管理后台的「设置」→「API 密钥」页面生成。

### 使用 API Key

在请求头中添加 `X-API-Key`：

```http
GET /api/cli/translations?project_id=1
X-API-Key: your-api-key-here
```

### API Key 权限

API Key 权限由生成时分配：
- **只读**：只能获取翻译
- **读写**：可以获取和推送翻译

## 权限级别

### 系统级权限

| 角色 | 权限 |
|------|------|
| admin | 访问所有 API，包括用户管理、语言管理 |
| member | 访问基础功能，不能管理系统设置 |

### 项目级权限

需要额外在项目级别授权：

| 角色 | 权限 |
|------|------|
| owner | 完全控制项目 |
| editor | 读取和编辑翻译 |
| viewer | 只读访问 |

## 错误处理

### Token 过期

```json
{
  "code": 401,
  "message": "token expired"
}
```

处理：使用 Refresh Token 获取新 Token。

### 无权限

```json
{
  "code": 403,
  "message": "access denied"
}
```

处理：检查用户角色和项目权限。

### 无效 API Key

```json
{
  "code": 401,
  "message": "invalid api key"
}
```

处理：检查 API Key 是否正确且有效。
