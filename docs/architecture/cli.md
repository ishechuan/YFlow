# CLI 架构

了解 Rust CLI 工具的技术实现和架构设计。

## 技术栈

- **Rust 2021 Edition** - 开发语言
- **Clap 4.4** - CLI 框架
- **Ureq** - HTTP 客户端（基于 rustls）
- **Tokio** - 异步运行时
- **Indicatif** - 进度条
- **Serde** - JSON 序列化/反序列化

## 目录结构

```
cli/src/
├── cli/               # 命令实现
│   ├── commands/      # 子命令
│   │   ├── sync.rs    # 同步命令
│   │   └── import.rs  # 导入命令
│   └── mod.rs         # 命令模块
├── api/               # API 客户端
│   ├── client.rs      # HTTP 客户端实现
│   └── mod.rs         # API 模块
├── core/              # 核心功能
│   ├── config.rs      # 配置管理
│   ├── scanner.rs     # 文件扫描器
│   ├── flatten.rs     # 翻译展平工具
│   ├── language_mapping.rs  # 语言映射
│   └── mod.rs         # 核心模块
├── ui/                # UI 组件
│   ├── progress.rs    # 进度显示
│   ├── spinner.rs     # 加载动画
│   └── mod.rs         # UI 模块
├── main.rs            # CLI 入口
└── lib.rs             # 库入口
```

## 核心模块

### CLI 入口

```rust
// main.rs
use clap::Parser;
use cli::Commands;

#[derive(Parser, version)]
struct Args {
    #[command(subcommand)]
    command: Commands,
}

fn main() {
    let args = Args::parse();
    args.command.run();
}
```

### 命令定义

```rust
// cli/mod.rs
use clap::{Parser, Subcommand};

#[derive(Subcommand)]
pub enum Commands {
    Sync(sync::SyncArgs),
    Import(import::ImportArgs),
    Init(init::InitArgs),
}

impl Commands {
    pub fn run(&self) {
        match self {
            Commands::Sync(args) => sync::run(args),
            Commands::Import(args) => import::run(args),
            Commands::Init(args) => init::run(args),
        }
    }
}
```

### 配置管理

```rust
// core/config.rs
use serde::Deserialize;
use std::path::Path;

#[derive(Deserialize)]
pub struct Config {
    pub messages_dir: String,
    pub project_id: u64,
    pub api_url: String,
    pub api_key: String,
    pub language_mapping: Option<HashMap<String, String>>,
}

impl Config {
    pub fn load(config_path: &Path) -> Result<Self, Box<dyn Error>> {
        let content = std::fs::read_to_string(config_path)?;
        let config: Config = serde_json::from_str(&content)?;
        Ok(config)
    }

    pub fn find_config() -> Option<PathBuf> {
        // 1. 当前目录
        let current = PathBuf::from(".i18nrc.json");
        if current.exists() {
            return Some(current);
        }
        // 2. 用户主目录
        if let Ok(home) = home::home_dir() {
            let home_config = home.join(".i18nrc.json");
            if home_config.exists() {
                return Some(home_config);
            }
        }
        None
    }
}
```

### API 客户端

```rust
// api/client.rs
use ureq::Agent;

pub struct ApiClient {
    agent: Agent,
    base_url: String,
    api_key: String,
}

impl ApiClient {
    pub fn new(base_url: String, api_key: String) -> Self {
        let agent = Agent::new();
        Self { agent, base_url, api_key }
    }

    pub fn get_translations(&self, project_id: u64) -> Result<serde_json::Value, ApiError> {
        let url = format!("{}/cli/translations", self.base_url);
        let response = self.agent
            .get(&url)
            .query("project_id", &project_id.to_string())
            .set("X-API-Key", &self.api_key)
            .call()?;

        Ok(response.into_json()?)
    }

    pub fn push_translations(
        &self,
        project_id: u64,
        translations: &serde_json::Value,
    ) -> Result<(), ApiError> {
        let url = format!("{}/cli/keys", self.base_url);
        let payload = serde_json::json!({
            "project_id": project_id,
            "translations": translations
        });

        self.agent
            .post(&url)
            .set("X-API-Key", &self.api_key)
            .send_json(payload)?;

        Ok(())
    }
}
```

### 文件扫描器

```rust
// core/scanner.rs
use std::fs;
use std::path::Path;

#[derive(Debug)]
pub struct TranslationFile {
    pub language: String,
    pub content: serde_json::Value,
}

pub fn scan_translation_files(messages_dir: &Path) -> Result<Vec<TranslationFile>, std::io::Error> {
    let entries = fs::read_dir(messages_dir)?;

    entries
        .filter_map(|entry| {
            let path = entry.ok()?.path();
            if path.extension().map(|ext| ext == "json").unwrap_or(false) {
                Some(read_translation_file(&path))
            } else {
                None
            }
        })
        .collect()
}

fn read_translation_file(path: &Path) -> Option<TranslationFile> {
    let language = path.file_stem()?.to_string_lossy().to_string();
    let content = fs::read_to_string(path).ok()?;
    let json: serde_json::Value = serde_json::from_str(&content).ok()?;

    Some(TranslationFile { language, content: json })
}
```

### 翻译展平工具

```rust
// core/flatten.rs
use serde_json::Value;

pub fn flatten_translations(
    files: &[TranslationFile],
) -> Value {
    let mut result = Value::Object(serde_json::Map::new());

    for file in files {
        flatten_object(&file.content, &file.language, "", &mut result);
    }

    result
}

fn flatten_object(
    obj: &Value,
    language: &str,
    prefix: &str,
    result: &mut Value,
) {
    match obj {
        Value::Object(map) => {
            for (key, value) in map {
                let new_key = if prefix.is_empty() {
                    key.clone()
                } else {
                    format!("{}.{}", prefix, key)
                };
                flatten_object(value, language, &new_key, result);
            }
        }
        Value::String(s) => {
            if let Some(map) = result.as_object_mut() {
                map.insert(format!("{}.{}", language, prefix.to_string()), Value::String(s.clone()));
            }
        }
        _ => {}
    }
}
```

## 命令实现

### sync 命令

```rust
// cli/commands/sync.rs
use clap::Parser;
use core::config::Config;
use core::scanner::scan_translation_files;

#[derive(Parser)]
pub struct SyncArgs {
    /// 模拟运行，不实际写入文件
    #[arg(long)]
    pub dry_run: bool,
    /// 强制覆盖所有现有翻译
    #[arg(long)]
    pub force: bool,
    /// 配置文件路径
    #[arg(short, long)]
    pub config: Option<PathBuf>,
}

pub fn run(args: &SyncArgs) {
    let config_path = args.config.clone()
        .or_else(|| Config::find_config())
        .expect("Config file not found. Run 'yflow init' first.");

    let config = Config::load(&config_path)
        .expect("Failed to load config");

    println!("Syncing translations for project {}...", config.project_id);

    let client = ApiClient::new(config.api_url, config.api_key);
    let translations = client.get_translations(config.project_id)
        .expect("Failed to fetch translations");

    if args.dry_run {
        println!("[Dry run] Would write {} translations", translations.as_array()
            .map(|arr| arr.len())
            .unwrap_or(0));
        return;
    }

    // 写入本地文件
    for (language, values) in translations.as_object().expect("Invalid format") {
        let file_path = Path::new(&config.messages_dir).join(format!("{}.json", language));
        std::fs::write(&file_path, serde_json::to_string_pretty(values).unwrap())
            .expect("Failed to write file");
        println!("  Updated {}.json", language);
    }

    println!("Sync completed!");
}
```

### import 命令

```rust
// cli/commands/import.rs
use clap::Parser;
use core::{config::Config, scanner::scan_translation_files, flatten::flatten_translations};

#[derive(Parser)]
pub struct ImportArgs {
    /// 模拟运行，不实际导入
    #[arg(long)]
    pub dry_run: bool,
    /// 配置文件路径
    #[arg(short, long)]
    pub config: Option<PathBuf>,
}

pub fn run(args: &ImportArgs) {
    let config_path = args.config.clone()
        .or_else(|| Config::find_config())
        .expect("Config file not found. Run 'yflow init' first.");

    let config = Config::load(&config_path)
        .expect("Failed to load config");

    println!("Scanning local translation files...");

    let messages_dir = Path::new(&config.messages_dir);
    let files = scan_translation_files(messages_dir)
        .expect("Failed to scan translation files");

    let translations = flatten_translations(&files);

    let keys_count = translations.as_object()
        .map(|m| m.len())
        .unwrap_or(0);
    println!("Found {} translation keys", keys_count);

    if args.dry_run {
        println!("[Dry run] Would import {} keys", keys_count);
        return;
    }

    // 分批导入
    let batch_size = 50;
    let keys: Vec<_> = translations.as_object().unwrap().keys().cloned().collect();

    for chunk in keys.chunks(batch_size) {
        let batch: serde_json::Value = serde_json::json!({
            "keys": chunk
        });

        let client = ApiClient::new(config.api_url.clone(), config.api_key.clone());
        client.push_keys(config.project_id, &batch)
            .expect("Failed to import batch");
    }

    println!("Import completed!");
}
```

## 工作流程

### sync 命令流程

```
yflow sync
        │
        ▼
加载配置文件 (.i18nrc.json)
        │
        ▼
请求 /api/cli/translations
        │
        ▼
接收翻译数据
        │
        ▼
写入本地 JSON 文件
        │
        ▼
完成
```

### import 命令流程

```
yflow import
        │
        ▼
扫描 messagesDir 目录
        │
        ▼
读取所有 JSON 文件
        │
        ▼
扁平化翻译数据
        │
        ▼
分批 POST 到 /api/cli/keys
        │
        ▼
完成
```

## 下一步

- [CLI 使用指南 →](/guide/cli-guide)
- [部署指南 →](/deployment/docker)
