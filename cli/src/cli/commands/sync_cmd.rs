//! Sync command implementation
//!
//! Synchronizes translations from the YFlow backend to the local messages directory.
//! Preserves the original file structure and only updates translation keys.
//!
//! # Features
//!
//! - Downloads translations from backend
//! - Preserves original file structure
//! - Supports force overwrite mode
//! - Dry-run mode for previewing changes
//! - Progress bar display for file writing
//! - Language code mapping support

use crate::api::client::APIClient;
use crate::core::config::load_config;
use crate::core::language_mapping::LanguageMapper;
use crate::core::scanner::{scan_messages_dir, write_translations_with_structure};
use crate::core::{ScanResult, SyncResult, Translations};
use crate::ui::progress::MultiProgressManager;
use anyhow::{Context, Result};
use clap::Parser;
use std::path::PathBuf;
use tracing::info;

/// 同步命令参数
///
/// 将后端翻译同步到本地 messages 目录。
#[derive(Parser, Debug)]
#[command(name = "sync")]
#[command(about = "Sync translations from backend to local messages directory", long_about = None)]
pub struct SyncCmd {
    /// 配置文件路径
    #[arg(short, long, value_name = "PATH")]
    pub config: Option<PathBuf>,

    /// 模拟运行 - 显示将要同步的内容但不实际修改
    #[arg(long)]
    pub dry_run: bool,

    /// 强制覆盖所有现有翻译
    #[arg(long)]
    pub force: bool,
}

impl SyncCmd {
    /// 执行同步命令
    ///
    /// # 处理流程
    ///
    /// 1. 加载配置
    /// 2. 创建 API 客户端
    /// 3. 验证认证
    /// 4. 从后端获取翻译
    /// 5. 扫描本地 messages 目录
    /// 6. 执行同步（或显示差异）
    ///
    /// # 参数
    ///
    /// * `global_config` - 可选的父级配置文件路径
    pub async fn run(&self, global_config: Option<PathBuf>) -> Result<SyncResult> {
        // 合并配置选项
        let config_path = self.config.clone().or(global_config);

        info!("Starting sync from backend...");

        // 1. 加载配置
        info!("Loading configuration...");
        let config = load_config(config_path)?;
        info!("  - Messages directory: {}", config.messages_dir.display());
        info!("  - Project ID: {}", config.project_id);
        info!("  - API URL: {}", config.api_url);

        // 1.1 初始化语言映射器
        let language_mapper = LanguageMapper::new(Some(config.language_mapping));
        if language_mapper.needs_mapping() {
            info!("  - {}", language_mapper.get_description());
        }

        // 2. 创建 API 客户端
        let client = APIClient::new(
            config.api_url.clone(),
            config.api_key.clone(),
            config.project_id,
        )
        .context("Failed to create API client")?;

        // 3. 验证认证
        info!("Verifying API authentication...");
        if !client.check_auth()? {
            return Err(anyhow::anyhow!(
                "API authentication failed. Please check your API key."
            ));
        }
        info!("  - Authentication successful");

        // 4. 从后端获取翻译
        info!("Fetching translations from backend...");
        let backend_translations = client.get_translations()?;

        let total_keys: usize = backend_translations.values().map(|v| v.len()).sum();
        let languages: Vec<&str> = backend_translations.keys().map(|s| s.as_str()).collect();
        info!(
            "  - Fetched {} keys from {} languages: {}",
            total_keys,
            languages.len(),
            languages.join(", ")
        );

        if total_keys == 0 {
            info!("No translations found on backend, skipping sync.");
            return Ok(SyncResult::default());
        }

        // 4.1 应用反向语言映射（后端代码 -> 本地代码）
        let local_translations = language_mapper.reverse_translations(backend_translations);

        let local_key_count: usize = local_translations.values().map(|v| v.len()).sum();
        let lang_list: Vec<String> = local_translations.keys().cloned().collect();
        info!(
            "  - After mapping: {} keys, languages: {}",
            local_key_count,
            lang_list.join(", ")
        );

        // 5. 扫描本地 messages 目录
        let local_scan_result = match scan_messages_dir(&config.messages_dir).await {
            Ok(result) => result,
            Err(_) => {
                // 如果目录不存在，创建空结构
                info!("Local messages directory not found, creating empty structure.");
                crate::core::ScanResult {
                    translations: Translations::new(),
                    files: Vec::new(),
                    key_count: 0,
                }
            }
        };
        info!(
            "  - Local files: {}, local keys: {}",
            local_scan_result.files.len(),
            local_scan_result.key_count
        );

        // 6. 执行同步或显示差异
        if self.dry_run {
            self.show_sync_diff(&local_translations, &local_scan_result.translations)?;
            return Ok(SyncResult::default());
        }

        self.execute_sync(
            &config.messages_dir,
            &local_scan_result.files,
            &local_translations,
            &local_scan_result,
        )
        .await
    }

    /// 显示同步差异（dry-run 模式）
    ///
    /// 显示将要下载和将要跳过的键。
    ///
    /// # 参数
    ///
    /// * `backend` - 后端翻译（经过本地映射后）
    /// * `local` - 本地翻译
    fn show_sync_diff(
        &self,
        backend: &Translations,
        local: &Translations,
    ) -> Result<()> {
        info!("=== DRY RUN ===");

        let mut total_downloaded = 0;
        let mut total_skipped = 0;

        info!("Sync diff preview:");
        for (lang, translations) in backend {
            let local_lang = local.get(lang).cloned().unwrap_or_default();
            let new_keys: Vec<&String> = translations.keys()
                .filter(|k| !local_lang.contains_key(*k))
                .collect();
            let existing_keys: Vec<&String> = translations.keys()
                .filter(|k| local_lang.contains_key(*k))
                .collect();

            let new_count = new_keys.len();
            let existing_count = existing_keys.len();

            info!("  {}:", lang);
            if new_count > 0 {
                let preview: Vec<String> = new_keys.iter().take(5).map(|s| s.to_string()).collect();
                info!("    New ({}): {}", new_count, preview.join(", "));
            }
            if existing_count > 0 {
                let preview: Vec<String> = existing_keys.iter().take(3).map(|s| s.to_string()).collect();
                info!("    Existing ({}): {}", existing_count, preview.join(", "));
            }

            total_downloaded += new_count;
            total_skipped += existing_count;
        }

        info!("Summary:");
        info!("  - Would download: {}", total_downloaded);
        info!("  - Would skip: {}", total_skipped);

        Ok(())
    }

    /// 执行同步操作
    ///
    /// 将翻译写入本地文件，同时保留原始文件结构。
    /// 显示每个语言写入的进度条。
    ///
    /// # Arguments
    ///
    /// * `messages_dir` - Messages 目录路径
    /// * `local_files` - 本地文件列表（相对路径）
    /// * `translations` - 要写入的翻译数据
    /// * `local_scan_result` - 本地扫描结果（包含现有翻译，用于统计计算）
    ///
    /// # 统计计算
    ///
    /// 统计逻辑说明：
    /// - `downloaded`: 新下载的键数量（force=true 或本地不存在的键）
    /// - `skipped`: 跳过的键数量（force=false 且本地已存在的键）
    /// - `written`: 写入的文件数量
    ///
    /// 通过传入 `local_scan_result` 避免重复扫描目录，提高性能。
    async fn execute_sync(
        &self,
        messages_dir: &PathBuf,
        local_files: &[PathBuf],
        translations: &Translations,
        local_scan_result: &ScanResult,
    ) -> Result<SyncResult> {
        // 初始化进度管理器
        let progress_manager = MultiProgressManager::new();
        let show_progress = progress_manager.is_enabled();

        info!("Writing translations to local files...");

        // 跟踪进度的回调
        let total_languages = translations.len();
        let _total_languages = total_languages; // 抑制警告

        // 创建进度回调（使用线程安全的内部可变性）
        let progress_callback: crate::core::scanner::ProgressCallback = if show_progress {
            let bar = progress_manager.create_bar("", 1);
            use std::sync::{Arc, Mutex};
            let bar = Arc::new(Mutex::new(bar));
            Box::new(move |_lang: String, _index: usize, _total: usize| {
                let mut bar = bar.lock().unwrap();
                bar.inc();
                bar.finish();
            })
        } else {
            Box::new(|_lang: String, _index: usize, _total: usize| {})
        };

        // 写入翻译（保留文件结构）
        let written = write_translations_with_structure(
            messages_dir,
            local_files,
            translations,
            self.force,
            Some(progress_callback),
        )
        .await
        .context("Failed to write translations")?;

        // 停止进度显示
        progress_manager.stop();

        // 计算统计结果
        // 使用传入的 local_scan_result，避免重复扫描目录
        let mut result = SyncResult::default();
        result.written = written.len();

        for (lang, translations) in translations {
            // 从本地扫描结果获取该语言的现有翻译
            let local_translations = local_scan_result
                .translations
                .get(lang)
                .cloned()
                .unwrap_or_default();

            // 遍历所有翻译键，计算下载/跳过数量
            for key in translations.keys() {
                if self.force || !local_translations.contains_key(key) {
                    // force=true 或键不存在于本地 -> 下载
                    result.downloaded += 1;
                } else {
                    // force=false 且键已存在于本地 -> 跳过
                    result.skipped += 1;
                }
            }
        }

        info!("Sync complete:");
        info!("  - Downloaded: {}", result.downloaded);
        info!("  - Skipped: {}", result.skipped);
        info!("  - Files written: {}", result.written);

        Ok(result)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::TempDir;
    use std::path::PathBuf;

    #[test]
    fn test_sync_cmd_default() {
        let cmd = SyncCmd {
            config: None,
            dry_run: false,
            force: false,
        };
        assert!(!cmd.dry_run);
        assert!(!cmd.force);
    }

    #[test]
    fn test_sync_cmd_with_force() {
        let cmd = SyncCmd {
            config: None,
            dry_run: false,
            force: true,
        };
        assert!(cmd.force);
    }

    #[test]
    fn test_sync_cmd_with_dry_run() {
        let cmd = SyncCmd {
            config: None,
            dry_run: true,
            force: false,
        };
        assert!(cmd.dry_run);
    }

    #[test]
    fn test_sync_cmd_with_both_options() {
        let cmd = SyncCmd {
            config: Some(PathBuf::from("/custom/path")),
            dry_run: true,
            force: true,
        };
        assert!(cmd.dry_run);
        assert!(cmd.force);
        assert_eq!(cmd.config, Some(PathBuf::from("/custom/path")));
    }

    // ========== execute_sync 统计逻辑测试 ==========

    /// 测试 force=true 时，所有键都应该被下载
    #[tokio::test]
    async fn test_execute_sync_stats_with_force() {
        let temp_dir = TempDir::new().unwrap();
        let messages_dir = temp_dir.path().join("messages");
        std::fs::create_dir_all(&messages_dir).unwrap();

        // 创建本地文件（en 目录和文件）
        let en_dir = messages_dir.join("en");
        std::fs::create_dir_all(&en_dir).unwrap();
        let en_file = en_dir.join("common.json");
        std::fs::write(&en_file, r#"{"greeting": "Hello"}"#).unwrap();

        // 创建本地扫描结果（现有 greeting 键）
        let local_scan_result = ScanResult {
            translations: [(
                "en".to_string(),
                [("greeting".to_string(), "Hello".to_string())]
                    .iter()
                    .cloned()
                    .collect(),
            )]
            .iter()
            .cloned()
            .collect(),
            files: vec![PathBuf::from("en/common.json")],
            key_count: 1,
        };

        // 要写入的翻译（包含现有键和新键）
        let translations: Translations = [(
            "en".to_string(),
            [
                ("greeting".to_string(), "Hello Updated".to_string()),
                ("new_key".to_string(), "New Value".to_string()),
            ]
            .iter()
            .cloned()
            .collect(),
        )]
        .iter()
        .cloned()
        .collect();

        let cmd = SyncCmd {
            config: None,
            dry_run: false,
            force: true,
        };

        let result = cmd
            .execute_sync(&messages_dir, &local_scan_result.files, &translations, &local_scan_result)
            .await
            .unwrap();

        // force=true 时，所有键都应该被下载（包括已存在的）
        assert_eq!(result.downloaded, 2);
        assert_eq!(result.skipped, 0);
        assert!(result.written >= 1);
    }

    /// 测试 force=false 时，已存在的键被跳过
    #[tokio::test]
    async fn test_execute_sync_stats_without_force() {
        let temp_dir = TempDir::new().unwrap();
        let messages_dir = temp_dir.path().join("messages");
        std::fs::create_dir_all(&messages_dir).unwrap();

        // 创建本地文件
        let en_dir = messages_dir.join("en");
        std::fs::create_dir_all(&en_dir).unwrap();
        let en_file = en_dir.join("common.json");
        std::fs::write(&en_file, r#"{"greeting": "Hello"}"#).unwrap();

        // 创建本地扫描结果（现有 greeting 键）
        let local_scan_result = ScanResult {
            translations: [(
                "en".to_string(),
                [("greeting".to_string(), "Hello".to_string())]
                    .iter()
                    .cloned()
                    .collect(),
            )]
            .iter()
            .cloned()
            .collect(),
            files: vec![PathBuf::from("en/common.json")],
            key_count: 1,
        };

        // 要写入的翻译
        let translations: Translations = [(
            "en".to_string(),
            [
                ("greeting".to_string(), "Hello Updated".to_string()),
                ("new_key".to_string(), "New Value".to_string()),
            ]
            .iter()
            .cloned()
            .collect(),
        )]
        .iter()
        .cloned()
        .collect();

        let cmd = SyncCmd {
            config: None,
            dry_run: false,
            force: false,
        };

        let result = cmd
            .execute_sync(&messages_dir, &local_scan_result.files, &translations, &local_scan_result)
            .await
            .unwrap();

        // force=false 时，已存在的键被跳过，新键被下载
        assert_eq!(result.downloaded, 1); // new_key
        assert_eq!(result.skipped, 1); // greeting 已存在
        assert!(result.written >= 1);
    }

    /// 测试多语言统计
    #[tokio::test]
    async fn test_execute_sync_stats_multiple_languages() {
        let temp_dir = TempDir::new().unwrap();
        let messages_dir = temp_dir.path().join("messages");
        std::fs::create_dir_all(&messages_dir).unwrap();

        // 创建 en 和 zh_CN 目录
        let en_dir = messages_dir.join("en");
        let zh_dir = messages_dir.join("zh_CN");
        std::fs::create_dir_all(&en_dir).unwrap();
        std::fs::create_dir_all(&zh_dir).unwrap();

        // 创建本地文件
        std::fs::write(en_dir.join("common.json"), r#"{"greeting": "Hello"}"#).unwrap();
        std::fs::write(zh_dir.join("common.json"), r#"{"greeting": "你好"}"#).unwrap();

        // 创建本地扫描结果
        let local_scan_result = ScanResult {
            translations: [
                ("en".to_string(), [("greeting".to_string(), "Hello".to_string())].iter().cloned().collect()),
                ("zh_CN".to_string(), [("greeting".to_string(), "你好".to_string())].iter().cloned().collect()),
            ]
            .iter()
            .cloned()
            .collect(),
            files: vec![
                PathBuf::from("en/common.json"),
                PathBuf::from("zh_CN/common.json"),
            ],
            key_count: 2,
        };

        // 要写入的翻译
        let translations: Translations = [
            ("en".to_string(), [
                ("greeting".to_string(), "Hello Updated".to_string()),
                ("new_en".to_string(), "New EN".to_string()),
            ].iter().cloned().collect()),
            ("zh_CN".to_string(), [
                ("greeting".to_string(), "你好更新".to_string()),
                ("new_zh".to_string(), "新中文".to_string()),
            ].iter().cloned().collect()),
        ]
        .iter()
        .cloned()
        .collect();

        let cmd = SyncCmd {
            config: None,
            dry_run: false,
            force: false,
        };

        let result = cmd
            .execute_sync(&messages_dir, &local_scan_result.files, &translations, &local_scan_result)
            .await
            .unwrap();

        // force=false 时，每种语言各跳过 1 个（greeting），下载 1 个（new_*）
        assert_eq!(result.downloaded, 2); // new_en, new_zh
        assert_eq!(result.skipped, 2); // en.greeting, zh_CN.greeting
        assert!(result.written >= 2);
    }

    /// 测试新语言的统计（本地不存在该语言）
    #[tokio::test]
    async fn test_execute_sync_stats_new_language() {
        let temp_dir = TempDir::new().unwrap();
        let messages_dir = temp_dir.path().join("messages");
        std::fs::create_dir_all(&messages_dir).unwrap();

        // 只创建 en 目录
        let en_dir = messages_dir.join("en");
        std::fs::create_dir_all(&en_dir).unwrap();
        std::fs::write(en_dir.join("common.json"), r#"{"greeting": "Hello"}"#).unwrap();

        // 本地扫描结果只有 en
        let local_scan_result = ScanResult {
            translations: [(
                "en".to_string(),
                [("greeting".to_string(), "Hello".to_string())]
                    .iter()
                    .cloned()
                    .collect(),
            )]
            .iter()
            .cloned()
            .collect(),
            files: vec![PathBuf::from("en/common.json")],
            key_count: 1,
        };

        // 要写入 en 和新语言 ja_JP
        let translations: Translations = [
            ("en".to_string(), [
                ("greeting".to_string(), "Hello Updated".to_string()),
            ].iter().cloned().collect()),
            ("ja_JP".to_string(), [
                ("greeting".to_string(), "こんにちは".to_string()),
            ].iter().cloned().collect()),
        ]
        .iter()
        .cloned()
        .collect();

        let cmd = SyncCmd {
            config: None,
            dry_run: false,
            force: false,
        };

        let result = cmd
            .execute_sync(&messages_dir, &local_scan_result.files, &translations, &local_scan_result)
            .await
            .unwrap();

        // en: 跳过 1 个（已存在），ja_JP: 下载 1 个（新语言）
        assert_eq!(result.downloaded, 1);
        assert_eq!(result.skipped, 1);
    }

    /// 测试空翻译场景
    #[tokio::test]
    async fn test_execute_sync_empty_translations() {
        let temp_dir = TempDir::new().unwrap();
        let messages_dir = temp_dir.path().join("messages");
        std::fs::create_dir_all(&messages_dir).unwrap();

        let local_scan_result = ScanResult {
            translations: std::collections::HashMap::new(),
            files: vec![],
            key_count: 0,
        };

        let translations: Translations = std::collections::HashMap::new();

        let cmd = SyncCmd {
            config: None,
            dry_run: false,
            force: false,
        };

        let result = cmd
            .execute_sync(&messages_dir, &local_scan_result.files, &translations, &local_scan_result)
            .await
            .unwrap();

        assert_eq!(result.downloaded, 0);
        assert_eq!(result.skipped, 0);
        assert_eq!(result.written, 0);
    }
}
