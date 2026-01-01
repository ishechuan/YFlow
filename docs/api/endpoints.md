# API 端点

完整的 API 端点参考。

## 认证端点

### 登录

```http
POST /api/login
```

**请求体**：

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**响应**：

```json
{
  "accessToken": "string",
  "refreshToken": "string",
  "user": {
    "id": "number",
    "email": "string",
    "role": "string"
  }
}
```

### 刷新 Token

```http
POST /api/refresh
```

**请求体**：

```json
{
  "refreshToken": "string"
}
```

### 注册

```http
POST /api/register
```

**请求体**：

```json
{
  "email": "user@example.com",
  "password": "password123",
  "invitationCode": "string"
}
```

## 项目端点

### 获取项目列表

```http
GET /api/projects
```

**响应**：

```json
{
  "data": [
    {
      "id": 1,
      "name": "My Project",
      "description": "Description",
      "defaultLanguage": "en",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### 创建项目

```http
POST /api/projects
```

**请求体**：

```json
{
  "name": "New Project",
  "description": "Project description",
  "defaultLanguage": "en"
}
```

### 获取项目详情

```http
GET /api/projects/:id
```

### 更新项目

```http
PUT /api/projects/:id
```

**请求体**：

```json
{
  "name": "Updated Name",
  "description": "Updated description"
}
```

### 删除项目

```http
DELETE /api/projects/:id
```

## 翻译端点

### 获取翻译矩阵

```http
GET /api/translations?project_id=1
```

**响应**：

```json
{
  "data": {
    "keys": ["greeting", "user.name"],
    "translations": {
      "greeting": {
        "en": "Hello",
        "zh-CN": "你好"
      },
      "user.name": {
        "en": "Name",
        "zh-CN": "姓名"
      }
    }
  }
}
```

### 创建翻译

```http
POST /api/translations
```

**请求体**：

```json
{
  "project_id": 1,
  "key": "new.key",
  "values": {
    "en": "Hello",
    "zh-CN": "你好"
  }
}
```

### 更新翻译

```http
PUT /api/translations
```

**请求体**：

```json
{
  "id": 1,
  "key": "existing.key",
  "values": {
    "en": "Updated",
    "zh-CN": "已更新"
  }
}
```

### 批量操作

```http
POST /api/translations/batch
```

**请求体**：

```json
{
  "project_id": 1,
  "actions": [
    {
      "action": "create",
      "key": "new.key",
      "values": { "en": "New", "zh-CN": "新建" }
    },
    {
      "action": "delete",
      "key": "old.key"
    }
  ]
}
```

### 导出翻译

```http
GET /api/exports/:project_id?format=json&languages=en,zh-CN
```

### 导入翻译

```http
POST /api/imports/:project_id
Content-Type: multipart/form-data

file: translations.json
```

## CLI 专用端点

### CLI 认证

```http
GET /api/cli/auth
X-API-Key: your-api-key
```

### 获取翻译 (CLI)

```http
GET /api/cli/translations?project_id=1
X-API-Key: your-api-key
```

### 推送翻译键 (CLI)

```http
POST /api/cli/keys
X-API-Key: your-api-key

{
  "project_id": 1,
  "translations": {
    "key": {
      "en": "Value",
      "zh-CN": "值"
    }
  }
}
```

## 语言端点

### 获取语言列表

```http
GET /api/languages
```

**响应**：

```json
{
  "data": [
    {
      "code": "en",
      "name": "English",
      "is_default": true
    },
    {
      "code": "zh-CN",
      "name": "简体中文",
      "is_default": false
    }
  ]
}
```

## 机器翻译端点

### 自动填充语言翻译

自动填充项目中缺失的目标语言翻译。

```http
POST /api/projects/:project_id/auto-fill-language
```

**请求体**：

```json
{
  "source_lang": "en",
  "target_lang": "zh-CN"
}
```

| 参数 | 类型 | 必填 | 描述 |
|-----|------|-----|------|
| source_lang | string | 否 | 源语言代码，如不提供则使用项目默认语言 |
| target_lang | string | 是 | 目标语言代码 |

**响应**：

```json
{
  "total": 50,
  "success_count": 48,
  "failed_count": 2,
  "message": "Successfully translated 48 missing translations"
}
```

### 获取支持的语言列表

获取机器翻译服务支持的语言列表。

```http
GET /api/translations/machine-translate/languages
```

**响应**：

```json
[
  {
    "code": "en",
    "name": "English"
  },
  {
    "code": "zh",
    "name": "Chinese"
  },
  {
    "code": "ja",
    "name": "Japanese"
  }
]
```

### 检查服务健康状态

检查机器翻译服务是否可用。

```http
GET /api/translations/machine-translate/health
```

**响应**：

```json
{
  "available": true
}
```

## 错误码参考

| 错误码 | 描述 |
|--------|------|
| 10001 | 无效的 Token |
| 10002 | Token 已过期 |
| 10003 | 无效的 API Key |
| 20001 | 项目不存在 |
| 20002 | 无权限访问项目 |
| 30001 | 翻译键不存在 |
| 30002 | 翻译键已存在 |
| 40001 | 语言不存在 |
