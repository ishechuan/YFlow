//! Configuration management module
//!
//! Handles loading and validating the YFlow configuration file (.i18nrc.json)

use anyhow::{Context, Result};
use serde::Serialize;
use std::env;
use std::fs;
use std::path::PathBuf;

use super::I18nConfig;

const CONFIG_FILENAME: &str = ".i18nrc.json";

/// 加载配置文件
///
/// 搜索路径（按优先级）：
/// 1. 命令行显式指定的路径
/// 2. 当前目录的 .i18nrc.json
/// 3. 用户主目录的 .i18nrc.json
///
/// # Arguments
///
/// * `config_path` - 可选的配置文件路径
///
/// # Errors
///
/// 如果配置文件不存在、无法读取或格式错误，返回错误
///
/// # Example
///
/// ```ignore
/// let config = load_config(None)?;  // 使用默认路径
/// let config = load_config(Some(PathBuf::from("/path/to/config")))?;
/// ```
pub fn load_config(config_path: Option<PathBuf>) -> Result<I18nConfig> {
    let path = resolve_config_path(config_path)?;

    let content = fs::read_to_string(&path)
        .with_context(|| format!("Failed to read config file: {}", path.display()))?;

    let config: I18nConfig = serde_json::from_str(&content)
        .with_context(|| format!("Invalid config file format: {}", path.display()))?;

    // 验证必需字段
    validate_config(&config)?;

    // 应用环境变量覆盖
    apply_env_overrides(config)
}

/// 解析配置文件路径
fn resolve_config_path(config_path: Option<PathBuf>) -> Result<PathBuf> {
    if let Some(path) = config_path {
        return Ok(path);
    }

    // 检查当前目录
    let current_dir = env::current_dir()?;
    let current_config = current_dir.join(CONFIG_FILENAME);
    if current_config.exists() {
        return Ok(current_config);
    }

    // 检查用户主目录
    if let Some(home_dir) = home::home_dir() {
        let home_config = home_dir.join(CONFIG_FILENAME);
        if home_config.exists() {
            return Ok(home_config);
        }
    }

    Err(anyhow::anyhow!(
        "Config file not found. Expected at: {} (current dir) or ~/.i18nrc.json",
        current_config.display()
    ))
}

/// 验证配置文件必需字段
fn validate_config(config: &I18nConfig) -> Result<()> {
    let mut errors = Vec::new();

    if config.messages_dir.as_os_str().is_empty() {
        errors.push("messagesDir (messages directory path) is required");
    }

    if config.project_id == 0 {
        errors.push("projectId must be a positive integer");
    }

    if config.api_url.is_empty() {
        errors.push("apiUrl (API URL) is required");
    }

    if config.api_key.is_empty() {
        errors.push("apiKey (API key) is required");
    }

    if errors.is_empty() {
        Ok(())
    } else {
        Err(anyhow::anyhow!(
            "Config validation failed:\n{}",
            errors.join("\n")
        ))
    }
}

/// 应用环境变量覆盖
///
/// 环境变量优先级高于配置文件：
/// - I18N_MESSAGES_DIR
/// - I18N_PROJECT_ID
/// - I18N_API_URL
/// - I18N_API_KEY
fn apply_env_overrides(config: I18nConfig) -> Result<I18nConfig> {
    Ok(I18nConfig {
        messages_dir: env::var("I18N_MESSAGES_DIR")
            .map(PathBuf::from)
            .unwrap_or_else(|_| config.messages_dir.clone()),
        project_id: env::var("I18N_PROJECT_ID")
            .ok()
            .and_then(|v| v.parse().ok())
            .unwrap_or(config.project_id),
        api_url: env::var("I18N_API_URL")
            .ok()
            .unwrap_or_else(|| config.api_url.clone()),
        api_key: env::var("I18N_API_KEY")
            .ok()
            .unwrap_or_else(|| config.api_key.clone()),
        language_mapping: config.language_mapping,
    })
}

/// 获取默认配置文件搜索路径
pub fn get_default_config_path() -> PathBuf {
    env::current_dir()
        .unwrap_or_else(|_| PathBuf::from("."))
        .join(CONFIG_FILENAME)
}

/// 检查配置文件是否存在
pub fn config_exists(config_path: Option<PathBuf>) -> bool {
    match resolve_config_path(config_path) {
        Ok(path) => path.exists(),
        Err(_) => false,
    }
}

/// 创建示例配置文件内容
///
/// # Example
///
/// ```ignore
/// let sample = create_sample_config();
/// std::fs::write(".i18nrc.json", sample).unwrap();
/// ```
pub fn create_sample_config() -> String {
    let sample = SampleConfig {
        messages_dir: "./src/locales".to_string(),
        project_id: 1,
        api_url: "http://localhost:8080/api".to_string(),
        api_key: "your-api-key-here".to_string(),
    };
    serde_json::to_string_pretty(&sample).unwrap()
}

/// 示例配置结构（用于生成 JSON）
#[derive(Serialize)]
struct SampleConfig {
    pub messages_dir: String,
    pub project_id: u64,
    pub api_url: String,
    pub api_key: String,
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs::File;
    use std::io::Write;
    use tempfile::TempDir;

    #[test]
    fn test_load_config_missing_file() {
        let result = load_config(Some(PathBuf::from("/nonexistent/path")));
        assert!(result.is_err());
    }

    #[test]
    fn test_load_config_invalid_json() {
        let temp_dir = TempDir::new().unwrap();
        let config_path = temp_dir.path().join(CONFIG_FILENAME);
        let mut file = File::create(&config_path).unwrap();
        writeln!(file, "invalid json {{").unwrap();

        let result = load_config(Some(config_path));
        assert!(result.is_err());
    }

    #[test]
    fn test_load_config_valid() {
        let temp_dir = TempDir::new().unwrap();
        let config_path = temp_dir.path().join(CONFIG_FILENAME);
        let config_content = r#"{
            "messagesDir": "./locales",
            "projectId": 1,
            "apiUrl": "http://localhost:8080/api",
            "apiKey": "test-key"
        }"#;
        let mut file = File::create(&config_path).unwrap();
        writeln!(file, "{}", config_content).unwrap();

        let result = load_config(Some(config_path)).unwrap();
        assert_eq!(result.project_id, 1);
        assert_eq!(result.api_key, "test-key");
    }

    #[test]
    fn test_env_override() {
        let temp_dir = TempDir::new().unwrap();
        let config_path = temp_dir.path().join(CONFIG_FILENAME);
        let config_content = r#"{
            "messagesDir": "./locales",
            "projectId": 1,
            "apiUrl": "http://localhost:8080/api",
            "apiKey": "original-key"
        }"#;
        let mut file = File::create(&config_path).unwrap();
        writeln!(file, "{}", config_content).unwrap();

        // 设置环境变量
        std::env::set_var("I18N_API_KEY", "env-override-key");

        let result = load_config(Some(config_path)).unwrap();
        assert_eq!(result.api_key, "env-override-key");

        // 清理环境变量
        std::env::remove_var("I18N_API_KEY");
    }
}
