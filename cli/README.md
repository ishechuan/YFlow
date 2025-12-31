# @yflow/cli

yflow CLI 工具，用于管理前端国际化翻译与后端数据库之间的同步。

## 安装

```bash
# 使用 Bun 运行
bun run ./src/index.ts <命令>

# 或全局安装后使用
bun install -g .
yflow <命令>
```

## 快速开始

1. **初始化配置文件**

   在前端项目根目录运行：

   ```bash
   yflow init
   ```

2. **编辑配置文件**

   编辑生成的 `.i18nrc.json` 文件：

   ```json
   {
     "messagesDir": "./src/locales",
     "projectId": 1,
     "apiUrl": "http://localhost:8080/api",
     "apiKey": "your-api-key-here"
   }
   ```

3. **导入翻译到后端**

   ```bash
   yflow import
   ```

4. **从后端同步翻译**

   ```bash
   yflow sync
   ```

## 命令

### init

创建示例配置文件。

```bash
yflow init
```

### import

将前端 `messages` 目录的翻译导入到后端数据库。

```bash
yflow import          # 执行导入
yflow import --dry-run  # 模拟运行，不实际修改
```

**工作流程：**
1. 扫描 `messages` 目录下的所有 JSON 文件
2. 展平嵌套的 JSON 结构（如 `{ "error": { "notFound": "xxx" } }` 变为 `"error.notFound": "xxx"`）
3. 按语言分组收集翻译
4. 调用后端 API 批量导入/更新

### sync

从后端同步翻译到前端 `messages` 目录。

```bash
yflow sync           # 同步翻译（跳过已存在的）
yflow sync --force   # 强制覆盖所有翻译
yflow sync --dry-run # 模拟运行，预览差异
```

**工作流程：**
1. 从后端获取所有翻译数据
2. 扫描本地 `messages` 目录获取原始文件结构
3. 按语言拆分翻译数据
4. 保持原始目录结构写入文件

### help

显示帮助信息。

```bash
yflow --help
```

## 配置文件

`.i18nrc.json` 配置文件包含以下字段：

| 字段 | 必填 | 说明 |
|------|------|------|
| `messagesDir` | 是 | messages 目录路径，相对于配置文件位置 |
| `projectId` | 是 | 后端项目 ID |
| `apiUrl` | 是 | 后端 API 地址 |
| `apiKey` | 是 | API 密钥 |
| `languageMapping` | 否 | 语言代码映射表（见下方说明） |

### 语言代码映射

如果你的前端项目使用的语言代码与后端不同，可以使用 `languageMapping` 进行映射：

```json
{
  "messagesDir": "./messages",
  "projectId": 1,
  "apiUrl": "http://localhost:8080/api",
  "apiKey": "your-api-key",
  "languageMapping": {
    "zh_CN": "zh",
    "zh_TW": "tw",
    "en_US": "en"
  }
}
```

**映射规则：**
- 导入 (`import`) 时：将本地目录名（如 `zh_CN`）映射为后端语言代码（如 `zh`）
- 同步 (`sync`) 时：将后端语言代码反向映射为本地目录名

**示例：**
- 本地目录 `zh_CN/` → 后端语言 `zh`
- 本地目录 `zh_TW/` → 后端语言 `tw`
- 本地目录 `en_US/` → 后端语言 `en`

### 环境变量覆盖

可以通过环境变量覆盖配置：

| 环境变量 | 对应配置 |
|----------|----------|
| `I18N_MESSAGES_DIR` | messagesDir |
| `I18N_PROJECT_ID` | projectId |
| `I18N_API_URL` | apiUrl |
| `I18N_API_KEY` | apiKey |

## messages 目录结构

CLI 期望的目录结构：

```
messages/
├── en/
│   ├── common.json
│   ├── auth.json
│   └── ...
├── zh/
│   ├── common.json
│   ├── auth.json
│   └── ...
└── ...
```

JSON 文件格式支持嵌套：

```json
{
  "error": {
    "notFound": "Page not found",
    "fetchFailed": "Failed to load data"
  },
  "common": {
    "save": "Save",
    "cancel": "Cancel"
  }
}
```

## 常见问题

### 配置文件找不到

确保在包含 `.i18nrc.json` 文件的目录运行 CLI，或使用 `--config` 参数指定路径：

```bash
yflow import --config /path/to/.i18nrc.json
```

### API 认证失败

检查配置文件中的 `apiKey` 是否正确，并确保后端已配置相同的 API Key（环境变量 `CLI_API_KEY`）。

### 导入时语言不匹配

确保后端已创建对应的语言，且语言代码与 messages 子目录名称一致（如 `en`、`zh`、`ko`）。

## 开发

```bash
# 安装依赖
bun install

# 运行 CLI
bun run ./src/index.ts import

# 构建可执行文件
bun build ./src/index.ts --outfile ./bin/yflow
```
