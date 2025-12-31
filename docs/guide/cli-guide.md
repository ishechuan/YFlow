# CLI 使用指南

学习如何使用 yflow CLI 工具同步和管理翻译。

## 安装 CLI

```bash
cd cli
bun install
bun link
```

验证安装：

```bash
yflow --help
```

## 初始化配置

在项目根目录创建配置文件：

```bash
yflow init
```

或者手动创建 `.i18nrc.json`：

```json
{
  "messagesDir": "./src/locales",
  "projectId": 1,
  "apiUrl": "http://localhost:8080/api",
  "apiKey": "your-api-key-here",
  "languageMapping": {
    "zh_CN": "zh",
    "zh_TW": "tw"
  }
}
```

## 配置说明

| 配置项 | 描述 | 示例 |
|--------|------|------|
| `messagesDir` | 本地翻译文件目录 | `"./src/locales"` |
| `projectId` | 项目 ID | `1` |
| `apiUrl` | API 地址 | `"http://localhost:8080/api"` |
| `apiKey` | CLI 认证密钥 | `"your-api-key"` |
| `languageMapping` | 语言代码映射（可选） | `{"zh_CN": "zh"}` |

### 语言代码映射

如果你的本地语言代码与后端不一致，可以使用 `languageMapping` 进行映射：

- 键：本地语言代码（JSON 文件名）
- 值：后端语言代码

例如：`"zh_CN": "zh"` 表示将本地的 `zh_CN.json` 映射到后端的 `zh` 语言。

### 环境变量覆盖

可以通过环境变量覆盖配置文件中的设置：

| 环境变量 | 对应配置项 |
|----------|------------|
| `I18N_MESSAGES_DIR` | `messagesDir` |
| `I18N_PROJECT_ID` | `projectId` |
| `I18N_API_URL` | `apiUrl` |
| `I18N_API_KEY` | `apiKey` |

## 命令

### sync - 同步翻译

从后端同步翻译到本地：

```bash
yflow sync
```

选项：
- `--dry-run` - 模拟运行，查看同步差异但不写入文件
- `--force` - 强制覆盖所有本地翻译
- `--config <path>` - 指定配置文件路径

示例：

```bash
# 同步翻译
yflow sync

# 查看同步差异（不实际写入）
yflow sync --dry-run

# 强制覆盖本地文件
yflow sync --force

# 指定配置文件
yflow sync --config /path/to/.i18nrc.json
```

同步后的文件结构：

```
src/locales/
├── en.json
├── zh-CN.json
└── ja-JP.json
```

### import - 导入翻译

将本地翻译导入到后端：

```bash
yflow import
```

工作流程：
1. 扫描 `messagesDir` 目录下的 JSON 文件
2. 读取所有翻译键值对
3. 分批推送到后端（每批 50 个键）

选项：
- `--dry-run` - 模拟运行，查看导入预览但不实际导入
- `--config <path>` - 指定配置文件路径

示例：

```bash
# 导入翻译
yflow import

# 查看导入预览（不实际导入）
yflow import --dry-run
```

### init - 初始化配置

创建示例配置文件：

```bash
yflow init
```

### help - 查看帮助

查看所有可用命令和选项：

```bash
yflow --help
```

### version - 查看版本

查看 CLI 版本：

```bash
yflow --version
```

## 获取 API 密钥

在管理界面的设置页面生成 CLI API 密钥：

1. 登录管理后台
2. 进入「设置」或「API 密钥」页面
3. 点击「生成新密钥」
4. 复制密钥并在本地配置

## 常见问题

### 同步失败

检查：
1. 后端服务是否运行
2. API 密钥是否正确
3. 网络连接是否正常

### 密钥冲突

当本地和远程有相同 key 但不同值时，默认跳过本地已有的键。如需强制覆盖：

```bash
yflow sync --force
```

### 配置文件不存在

运行 `yflow init` 创建示例配置文件，或手动创建 `.i18nrc.json`。

### 模拟运行

使用 `--dry-run` 选项可以在不实际修改数据的情况下预览操作结果：
- `sync --dry-run` - 预览将下载/跳过的键
- `import --dry-run` - 预览将导入的翻译

## 下一步

- [了解翻译管理 →](/guide/translation-guide)
- [团队协作 →](/guide/team-collaboration)
