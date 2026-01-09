# YFlow CLI (Rust)

YFlow CLI 的 Rust 重写版本，使用 Rust 语言开发的高性能翻译管理命令行工具。

## 功能特性

- **导入翻译**: 将本地 `messages` 目录的翻译导入到 YFlow 后端数据库
- **同步翻译**: 从 YFlow 后端同步翻译到本地 `messages` 目录
- **批量处理**: 支持大批量翻译导入，默认每批 50 个键
- **重试机制**: 遇到速率限制时自动重试，支持指数退避
- **进度显示**: 实时显示导入/同步进度
- **语言映射**: 支持本地语言代码与后端语言代码之间的映射
- **Dry-run 模式**: 预览操作结果而不实际执行
- **强制覆盖**: 支持强制覆盖现有翻译

## 安装

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/yflow/yflow.git
cd yflow/rs-cli

# 编译release版本
cargo build --release

# 编译产物位于: target/release/yflow
```

### 使用 pre-built 二进制

从 [Releases](https://github.com/yflow/yflow/releases) 页面下载预编译的二进制文件。

## 快速开始

### 1. 初始化配置

```bash
# 在当前目录创建默认配置文件
yflow init

# 指定输出路径
yflow init --output /path/to/.i18nrc.json
```

### 2. 编辑配置文件

编辑生成的 `.i18nrc.json` 文件：

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

### 3. 导入翻译

将本地翻译导入到后端：

```bash
# 正常导入
yflow import

# 模拟运行（预览将要导入的内容）
yflow import --dry-run

# 指定配置文件
yflow import --config /path/to/.i18nrc.json
```

### 4. 同步翻译

从后端同步翻译到本地：

```bash
# 正常同步
yflow sync

# 模拟运行（预览将要同步的内容）
yflow sync --dry-run

# 强制覆盖所有现有翻译
yflow sync --force

# 指定配置文件
yflow sync --config /path/to/.i18nrc.json
```

## 命令参考

### 全局选项

| 选项 | 描述 |
|------|------|
| `-c, --config <PATH>` | 指定配置文件路径 |
| `-v, --verbose` | 启用详细日志输出 |
| `-h, --help` | 显示帮助信息 |
| `-V, --version` | 显示版本信息 |

### import 命令

导入本地翻译到后端数据库。

```bash
yflow import [OPTIONS]
```

| 选项 | 描述 |
|------|------|
| `--dry-run` | 模拟运行，仅显示预览而不实际导入 |

### sync 命令

从后端同步翻译到本地目录。

```bash
yflow sync [OPTIONS]
```

| 选项 | 描述 |
|------|------|
| `--dry-run` | 模拟运行，仅显示预览而不实际写入 |
| `--force` | 强制覆盖所有现有翻译 |

### init 命令

创建示例配置文件。

```bash
yflow init [OPTIONS]
```

| 选项 | 描述 |
|------|------|
| `-o, --output <PATH>` | 输出文件路径（默认: `./i18nrc.json`） |

### version 命令

显示版本信息。

```bash
yflow version
```

### help-cmd 命令

显示帮助信息。

```bash
yflow help-cmd        # 显示所有命令
yflow help-cmd import # 显示 import 命令帮助
```

## 配置文件

### 文件格式

配置文件使用 JSON 格式，文件名必须为 `.i18nrc.json`。

### 配置项

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `messagesDir` | string | 是 | 本地 messages 目录路径 |
| `projectId` | number | 是 | YFlow 项目 ID |
| `apiUrl` | string | 是 | YFlow API 地址 |
| `apiKey` | string | 是 | API 密钥 |
| `languageMapping` | object | 否 | 语言代码映射表 |

### 语言映射示例

```json
{
  "languageMapping": {
    "zh_CN": "zh",
    "zh_TW": "tw",
    "en_US": "en"
  }
}
```

### 搜索顺序

配置文件按以下顺序查找：

1. 命令行 `--config` 参数指定的路径
2. 当前目录的 `.i18nrc.json`
3. 用户主目录的 `.i18nrc.json`

## 环境变量

可以通过环境变量覆盖配置文件中的值（优先级更高）：

| 环境变量 | 对应配置项 |
|----------|------------|
| `I18N_MESSAGES_DIR` | `messagesDir` |
| `I18N_PROJECT_ID` | `projectId` |
| `I18N_API_URL` | `apiUrl` |
| `I18N_API_KEY` | `apiKey` |

示例：

```bash
export I18N_API_KEY="your-api-key"
export I18N_PROJECT_ID=2
yflow import
```

## 目录结构

期望的目录结构：

```
messages/
├── en/
│   ├── common.json
│   └── errors.json
├── zh_CN/
│   ├── common.json
│   └── errors.json
└── ...
```

JSON 文件示例：

```json
{
  "greeting": "Hello",
  "user": {
    "name": "User Name",
    "email": "user@example.com"
  }
}
```

## 开发

### 运行测试

```bash
cargo test
```

### 构建

```bash
# debug 版本
cargo build

# release 版本（推荐用于生产）
cargo build --release
```

### 代码检查

```bash
cargo clippy
cargo fmt
```

## 技术栈

- **语言**: Rust 2021 Edition
- **CLI 框架**: clap 4.4
- **HTTP 客户端**: ureq (基于 rustls)
- **异步运行时**: tokio
- **进度条**: indicatif
- **日志**: tracing
- **JSON 处理**: serde + serde_json

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
