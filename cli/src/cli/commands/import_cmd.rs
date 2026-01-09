//! Import command implementation
//!
//! Imports translations from the local messages directory to the YFlow backend.
//! Supports batching, retry logic with exponential backoff, and progress display.
//!
//! # Features
//!
//! - Batch processing with configurable batch size (default: 50)
//! - Automatic retry with exponential backoff for rate limiting
//! - Progress bar display for long-running imports
//! - Dry-run mode for previewing changes
//! - Language code mapping support

use crate::api::client::APIClient;
use crate::core::config::load_config;
use crate::core::language_mapping::LanguageMapper;
use crate::core::scanner::scan_messages_dir;
use crate::core::{ImportResult, Translations};
use crate::ui::progress::MultiProgressManager;
use anyhow::{Context, Result};
use clap::Parser;
use std::collections::HashMap;
use std::path::PathBuf;
use std::time::Duration;
use tokio::time::sleep;
use tracing::info;

/// 导入命令参数
///
/// 将本地 messages 目录的翻译导入到后端数据库。
#[derive(Parser, Debug)]
#[command(name = "import")]
#[command(about = "Import translations from local messages directory to backend", long_about = None)]
pub struct ImportCmd {
    /// 配置文件路径
    #[arg(short, long, value_name = "PATH")]
    pub config: Option<PathBuf>,

    /// 模拟运行 - 显示将要导入的内容但不实际修改
    #[arg(long)]
    pub dry_run: bool,
}

/// 导入翻译的批次大小
const BATCH_SIZE: usize = 50;

/// 批次间延迟（毫秒）- 避免速率限制
const BATCH_DELAY: Duration = Duration::from_millis(200);

/// 最大重试次数
const MAX_RETRIES: usize = 3;

impl ImportCmd {
    /// 执行导入命令
    ///
    /// # 处理流程
    ///
    /// 1. 加载配置
    /// 2. 创建 API 客户端
    /// 3. 验证认证
    /// 4. 扫描本地 messages 目录
    /// 5. 执行导入（或显示预览）
    ///
    /// # 参数
    ///
    /// * `global_config` - 可选的父级配置文件路径
    pub async fn run(&self, global_config: Option<PathBuf>) -> Result<ImportResult> {
        info!("Starting import to backend...");

        // 合并配置选项
        let config_path = self.config.clone().or(global_config);

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

        // 4. 扫描 messages 目录
        info!("Scanning messages directory: {}...", config.messages_dir.display());
        let scan_result = scan_messages_dir(&config.messages_dir)
            .await
            .context("Failed to scan messages directory")?;

        let languages: Vec<&str> = scan_result.translations.keys().map(|s| s.as_str()).collect();
        info!(
            "  - Scanned files: {}, keys: {}, languages: {}",
            scan_result.files.len(),
            scan_result.key_count,
            languages.join(", ")
        );

        if scan_result.key_count == 0 {
            info!("No translations found, skipping import.");
            return Ok(ImportResult::default());
        }

        // 5. 应用语言映射
        let mapped_translations = language_mapper.apply_to_translations(scan_result.translations);

        // 6. 执行导入或预览
        if self.dry_run {
            self.dry_run_import(&mapped_translations)?;
            Ok(ImportResult {
                added: scan_result.key_count,
                ..Default::default()
            })
        } else {
            self.execute_import(&client, mapped_translations).await
        }
    }

    /// 显示导入预览（dry-run 模式）
    ///
    /// 显示将要导入的翻译，但不实际调用 API。
    ///
    /// # 参数
    ///
    /// * `translations` - 要导入的翻译
    fn dry_run_import(&self, translations: &Translations) -> Result<()> {
        info!("=== DRY RUN ===");
        let mut total_keys = 0;

        info!("Translations to be imported:");
        for (lang, lang_translations) in translations {
            let count = lang_translations.len();
            info!("  {}: {} keys", lang, count);
            total_keys += count;

            // 显示前5个键作为预览
            let keys: Vec<&String> = lang_translations.keys().take(5).collect();
            if !keys.is_empty() {
                for key in &keys {
                    if let Some(value) = lang_translations.get(*key) {
                        let display_value = if value.len() > 50 {
                            &value[..50]
                        } else {
                            value.as_str()
                        };
                        info!("    - {}: \"{}\"", key, display_value);
                    }
                }
                if count > 5 {
                    info!("    ... and {} more keys", count - 5);
                }
            }
        }

        info!("Would import {} keys total", total_keys);

        Ok(())
    }

    /// 执行实际导入操作
    ///
    /// 分批导入翻译，支持重试逻辑和速率限制处理。
    /// 为每种语言显示进度条。
    ///
    /// # 参数
    ///
    /// * `client` - 用于发送请求的 API 客户端
    /// * `translations` - 要导入的翻译
    async fn execute_import(
        &self,
        client: &APIClient,
        translations: Translations,
    ) -> Result<ImportResult> {
        info!("Importing translations to backend...");

        // 初始化进度管理器
        let progress_manager = MultiProgressManager::new();
        let show_progress = progress_manager.is_enabled();

        let mut result = ImportResult::default();
        let total_languages = translations.len();
        let mut current_lang_index = 0;

        for (lang_code, lang_translations) in translations {
            current_lang_index += 1;
            let total_keys = lang_translations.len();
            if total_keys == 0 {
                continue;
            }

            if show_progress {
                info!("Importing {} ({}/{})...", lang_code, current_lang_index, total_languages);
            } else {
                info!("Importing {} ({} keys)...", lang_code, total_keys);
            }

            // 为该语言创建进度条
            let mut lang_progress = progress_manager.create_bar(&lang_code, total_keys as u64);

            // 将翻译拆分为多个批次
            let chunks: Vec<HashMap<String, String>> = lang_translations
                .into_iter()
                .collect::<Vec<_>>()
                .chunks(BATCH_SIZE)
                .map(|chunk| chunk.iter().cloned().collect())
                .collect();

            let total_batches = chunks.len();
            for (batch_idx, chunk) in chunks.iter().enumerate() {
                let batch_num = batch_idx + 1;
                let is_last_batch = batch_num == total_batches;

                // 将批次包装为 Translations 格式以供 API 使用
                let batch_translations: Translations = [(lang_code.clone(), chunk.clone())]
                    .iter()
                    .cloned()
                    .collect();

                // 带指数退避的重试循环
                let mut retry_count = 0;
                let mut success = false;

                while !success && retry_count < MAX_RETRIES {
                    match client.push_translations(batch_translations.clone()) {
                        Ok(response) => {
                            // 记录结果
                            result.added += response.added.len();
                            result.updated += response.existed.len();
                            result.failed += response.failed.len();

                            // 更新进度条
                            let processed_in_batch = response.added.len() + response.existed.len() + response.failed.len();
                            lang_progress.inc_by(processed_in_batch as u64);

                            // 记录失败的键
                            if !response.failed.is_empty() {
                                let failed_keys = response
                                    .failed
                                    .iter()
                                    .take(10)
                                    .map(|s| s.as_str())
                                    .collect::<Vec<_>>()
                                    .join(", ");
                                result.errors.push(format!(
                                    "{}[{}]: failed keys - {}",
                                    lang_code,
                                    batch_num,
                                    failed_keys
                                ));
                                if response.failed.len() > 10 {
                                    result.errors.push(format!(
                                        "  ... and {} more",
                                        response.failed.len() - 10
                                    ));
                                }
                            }

                            if show_progress {
                                info!(
                                    "  Batch {}: +{}, ~{}, ✗{}",
                                    batch_num,
                                    response.added.len(),
                                    response.existed.len(),
                                    response.failed.len()
                                );
                            }

                            success = true;
                        }
                        Err(e) => {
                            // 检查是否为速率限制错误（429）
                            if is_rate_limit_error(&e) && retry_count < MAX_RETRIES - 1 {
                                retry_count += 1;
                                let wait_time = BATCH_DELAY.as_millis() as u64 * (retry_count as u64 * 2);
                                info!(
                                    "  Rate limited, waiting {}ms before retry ({}/{})",
                                    wait_time, retry_count, MAX_RETRIES
                                );
                                sleep(Duration::from_millis(wait_time)).await;
                            } else {
                                // 记录错误并继续下一个批次
                                result.failed += chunk.len();
                                result.errors.push(format!("{}[{}]: {}", lang_code, batch_num, e));
                                info!("  Batch {}: FAILED - {}", batch_num, e);
                                lang_progress.inc_by(chunk.len() as u64);
                                success = true; // 即使失败也继续下一个批次
                            }
                        }
                    }
                }

                // 批次间延迟（除了最后一个）
                if !is_last_batch {
                    sleep(BATCH_DELAY).await;
                }
            }

            // 完成该语言的进度条
            lang_progress.finish();
        }

        // 停止所有进度条
        progress_manager.stop();

        info!("Import complete:");
        info!("  - Added: {}", result.added);
        info!("  - Updated: {}", result.updated);
        info!("  - Failed: {}", result.failed);

        if !result.errors.is_empty() {
            info!("  - Errors: {} detail(s)", result.errors.len());
        }

        Ok(result)
    }
}

/// 检查错误是否为速率限制错误（HTTP 429）
///
/// # 参数
///
/// * `error` - 要检查的错误
///
/// # 返回
///
/// 如果错误表示速率限制则返回 true
fn is_rate_limit_error(error: &anyhow::Error) -> bool {
    let error_msg = error.to_string().to_lowercase();
    error_msg.contains("429")
        || error_msg.contains("rate limit")
        || error_msg.contains("too many requests")
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_import_cmd_default() {
        let cmd = ImportCmd {
            config: None,
            dry_run: false,
        };
        assert!(!cmd.dry_run);
    }

    #[test]
    fn test_import_cmd_with_dry_run() {
        let cmd = ImportCmd {
            config: None,
            dry_run: true,
        };
        assert!(cmd.dry_run);
    }

    #[test]
    fn test_is_rate_limit_error_429() {
        let error = anyhow::anyhow!("HTTP 429: Too Many Requests");
        assert!(is_rate_limit_error(&error));
    }

    #[test]
    fn test_is_rate_limit_error_case_insensitive() {
        let error1 = anyhow::anyhow!("rate limit exceeded");
        let error2 = anyhow::anyhow!("RATE LIMIT");
        let error3 = anyhow::anyhow!("Too Many Requests");

        assert!(is_rate_limit_error(&error1));
        assert!(is_rate_limit_error(&error2));
        assert!(is_rate_limit_error(&error3));
    }

    #[test]
    fn test_is_rate_limit_error_negative() {
        let error = anyhow::anyhow!("Not Found: 404");
        assert!(!is_rate_limit_error(&error));
    }

    #[test]
    fn test_constants() {
        assert_eq!(BATCH_SIZE, 50);
        assert_eq!(BATCH_DELAY.as_millis(), 200);
        assert_eq!(MAX_RETRIES, 3);
    }
}
