//! API client implementation
//!
//! Handles all HTTP communication with the YFlow backend API.
//! Provides methods for authentication, fetching translations, and pushing translations.

use anyhow::{Context, Result};
use std::collections::HashMap;

use crate::core::Translations;

/// API 客户端
///
/// 负责与后端 API 通信，包括：
/// - 认证检查
/// - 获取翻译
/// - 推送翻译
///
/// # Example
///
/// ```ignore
/// let client = APIClient::new(
///     "http://localhost:8080/api".to_string(),
///     "your-api-key".to_string(),
///     1,
/// )?;
///
/// // 检查认证
/// if client.check_auth()? {
///     println!("Authenticated!");
/// }
///
/// // 获取翻译
/// let translations = client.get_translations()?;
/// ```
#[derive(Debug, Clone)]
pub struct APIClient {
    /// API 基础 URL
    base_url: String,
    /// API 密钥
    api_key: String,
    /// 项目 ID
    project_id: u64,
}

impl APIClient {
    /// 创建新的 API 客户端
    ///
    /// # Arguments
    ///
    /// * `base_url` - API 基础 URL（会自动移除末尾斜杠）
    /// * `api_key` - API 密钥
    /// * `project_id` - 项目 ID
    ///
    /// # Errors
    ///
    /// 如果 URL 为空或格式无效，返回错误
    ///
    /// # URL 规范化说明
    ///
    /// 该方法会自动处理以下情况：
    /// - 移除 URL 末尾的斜杠（如 `http://localhost/api/` → `http://localhost/api`）
    /// - 去除首尾空白字符
    /// - 验证 URL 必须以 `http://` 或 `https://` 开头
    pub fn new(base_url: String, api_key: String, project_id: u64) -> Result<Self> {
        // 验证 URL 不为空
        if base_url.trim().is_empty() {
            return Err(anyhow::anyhow!("API URL cannot be empty"));
        }

        // 规范化 URL：移除末尾斜杠和空白字符
        // 这样可以容忍用户配置的 URL 末尾有或没有斜杠，保持与 TypeScript 实现一致
        let normalized_url = base_url
            .trim()
            .trim_end_matches('/')
            .to_string();

        // 验证 URL 格式（简单验证，必须以 http:// 或 https:// 开头）
        if !normalized_url.starts_with("http://") && !normalized_url.starts_with("https://") {
            return Err(anyhow::anyhow!(
                "API URL must start with 'http://' or 'https://', got: {}",
                normalized_url
            ));
        }

        // 验证项目 ID 为正数
        if project_id == 0 {
            return Err(anyhow::anyhow!("Project ID must be a positive integer"));
        }

        Ok(Self {
            base_url: normalized_url,
            api_key,
            project_id,
        })
    }

    /// 获取 API 基础 URL
    pub fn base_url(&self) -> &str {
        &self.base_url
    }

    /// 获取 API 密钥
    pub fn api_key(&self) -> &str {
        &self.api_key
    }

    /// 获取项目 ID
    pub fn project_id(&self) -> u64 {
        self.project_id
    }

    /// 检查 API 认证状态
    ///
    /// 向后端发送认证检查请求。
    ///
    /// # Returns
    ///
    /// 认证成功返回 `true`，失败返回 `false`
    ///
    /// # Errors
    ///
    /// 如果网络请求失败，返回错误
    pub fn check_auth(&self) -> Result<bool> {
        let url = format!("{}/cli/auth", self.base_url);
        let agent = ureq::Agent::new();

        let response = agent
            .get(&url)
            .set("X-API-Key", &self.api_key)
            .call();

        match response {
            Ok(_) => Ok(true),
            Err(ureq::Error::Status(401, _)) => Ok(false),
            Err(e) => Err(anyhow::anyhow!("Auth check failed: {}", e)),
        }
    }

    /// 获取所有翻译
    ///
    /// 从后端获取指定项目的所有翻译数据。
    ///
    /// # Returns
    ///
    /// 翻译数据，格式为 `{语言代码: {键: 值}}`
    ///
    /// # Errors
    ///
    /// 如果请求失败或响应格式错误，返回错误
    pub fn get_translations(&self) -> Result<Translations> {
        let url = format!(
            "{}/cli/translations?project_id={}",
            self.base_url, self.project_id
        );
        let agent = ureq::Agent::new();

        let response = agent
            .get(&url)
            .set("X-API-Key", &self.api_key)
            .call()
            .context("Failed to fetch translations")?;

        let status = response.status();
        if status == 401 {
            return Err(anyhow::anyhow!("API authentication failed"));
        }

        if status < 200 || status >= 300 {
            let error_text = response.into_string()?;
            return Err(anyhow::anyhow!("API error ({}): {}", status, error_text));
        }

        let json: serde_json::Value = response
            .into_json()
            .context("Failed to parse response as JSON")?;

        // 解析响应
        let data = json.get("data")
            .ok_or_else(|| anyhow::anyhow!("Missing 'data' field in response"))?;

        // 处理空响应
        if data.is_null() {
            return Ok(HashMap::new());
        }

        // API 返回键中心化格式: {key: {lang: value}}
        // 需要转换为语言中心化格式: {lang: {key: value}}
        let translations = Self::transform_translations_format(data.clone())?;

        Ok(translations)
    }

    /// 转换翻译数据格式
    ///
    /// 从键中心化格式 `{key: {lang: value}}` 转换为语言中心化格式 `{lang: {key: value}}`
    fn transform_translations_format(data: serde_json::Value) -> Result<Translations> {
        let mut result: Translations = HashMap::new();

        if let Some(keys_map) = data.as_object() {
            for (key, langs_map) in keys_map {
                if let Some(langs) = langs_map.as_object() {
                    for (lang_code, value) in langs {
                        let lang_translations = result
                            .entry(lang_code.clone())
                            .or_insert_with(HashMap::new);
                        if let Some(value_str) = value.as_str() {
                            lang_translations.insert(key.clone(), value_str.to_string());
                        }
                    }
                }
            }
        }

        Ok(result)
    }

    /// 获取指定语言的翻译
    ///
    /// # Arguments
    ///
    /// * `locale` - 语言代码（如 "en"、"zh_CN"）
    ///
    /// # Returns
    ///
    /// 该语言的翻译数据
    ///
    /// # Errors
    ///
    /// 如果请求失败，返回错误
    pub fn get_translations_by_locale(&self, locale: &str) -> Result<HashMap<String, String>> {
        let url = format!(
            "{}/cli/translations?project_id={}&locale={}",
            self.base_url, self.project_id, locale
        );
        let agent = ureq::Agent::new();

        let response = agent
            .get(&url)
            .set("X-API-Key", &self.api_key)
            .call()
            .context("Failed to fetch translations by locale")?;

        let status = response.status();
        if status < 200 || status >= 300 {
            let error_text = response.into_string()?;
            return Err(anyhow::anyhow!("API error ({}): {}", status, error_text));
        }

        let json: serde_json::Value = response
            .into_json()
            .context("Failed to parse response as JSON")?;

        let data = json.get("data")
            .ok_or_else(|| anyhow::anyhow!("Missing 'data' field in response"))?;

        let translations: HashMap<String, String> = serde_json::from_value(data.clone())
            .context("Failed to parse translations data")?;

        Ok(translations)
    }

    /// 批量推送翻译
    ///
    /// 将翻译数据批量导入到后端数据库。
    /// 支持增量更新，只导入不存在的键或值发生变化的键。
    ///
    /// # Arguments
    ///
    /// * `translations` - 要推送的翻译数据，格式为 `{语言代码: {键: 值}}`
    ///
    /// # Returns
    ///
    /// 推送响应，包含 added、existed、failed 键名列表
    ///
    /// # Errors
    ///
    /// 如果请求失败，返回错误
    pub fn push_translations(&self, translations: Translations) -> Result<PushKeysResponse> {
        let url = format!("{}/cli/keys", self.base_url);
        let agent = ureq::Agent::new();

        let body = serde_json::json!({
            "project_id": self.project_id.to_string(),
            "keys": [],
            "translations": translations,
        });

        let response = agent
            .post(&url)
            .set("X-API-Key", &self.api_key)
            .set("Content-Type", "application/json")
            .send_json(body)
            .map_err(|e| anyhow::anyhow!("Request failed: {}", e))?;

        // 处理速率限制
        if response.status() == 429 {
            let retry_after = response
                .header("Retry-After")
                .and_then(|s| s.parse().ok())
                .unwrap_or(60);

            return Err(anyhow::anyhow!(
                "Rate limited. Retry after {} seconds",
                retry_after
            ));
        }

        let status = response.status();
        if status < 200 || status >= 300 {
            let error_text = response.into_string()?;
            return Err(anyhow::anyhow!("API error ({}): {}", status, error_text));
        }

        let json: serde_json::Value = response
            .into_json()
            .context("Failed to parse response as JSON")?;

        // 解析响应
        let data = json.get("data")
            .ok_or_else(|| anyhow::anyhow!("Missing 'data' field in response"))?;

        // 处理 data 为 null 的情况
        if data.is_null() {
            return Ok(PushKeysResponse::default());
        }

        Ok(PushKeysResponse {
            added: data.get("added")
                .and_then(|v| v.as_array())
                .map(|arr| arr.iter().filter_map(|s| s.as_str().map(|s| s.to_string())).collect())
                .unwrap_or_default(),
            existed: data.get("existed")
                .and_then(|v| v.as_array())
                .map(|arr| arr.iter().filter_map(|s| s.as_str().map(|s| s.to_string())).collect())
                .unwrap_or_default(),
            failed: data.get("failed")
                .and_then(|v| v.as_array())
                .map(|arr| arr.iter().filter_map(|s| s.as_str().map(|s| s.to_string())).collect())
                .unwrap_or_default(),
        })
    }

    /// 推送翻译键
    ///
    /// 创建新的翻译键（如果不存在），并可选地设置初始翻译值。
    ///
    /// # Arguments
    ///
    /// * `keys` - 要创建的键名列表
    /// * `translations` - 可选的翻译数据
    ///
    /// # Returns
    ///
    /// 推送响应
    ///
    /// # Errors
    ///
    /// 如果请求失败，返回错误
    pub fn push_keys(
        &self,
        keys: Vec<String>,
        translations: Option<Translations>,
    ) -> Result<PushKeysResponse> {
        let url = format!("{}/cli/keys", self.base_url);
        let agent = ureq::Agent::new();

        let mut body = serde_json::json!({
            "project_id": self.project_id.to_string(),
            "keys": keys,
        });

        if let Some(trans) = translations {
            body["translations"] = serde_json::to_value(trans)?;
        }

        let response = agent
            .post(&url)
            .set("X-API-Key", &self.api_key)
            .set("Content-Type", "application/json")
            .send_json(body)
            .map_err(|e| anyhow::anyhow!("Request failed: {}", e))?;

        let status = response.status();
        if status < 200 || status >= 300 {
            let error_text = response.into_string()?;
            return Err(anyhow::anyhow!("API error ({}): {}", status, error_text));
        }

        let json: serde_json::Value = response
            .into_json()
            .context("Failed to parse response as JSON")?;

        let data = json.get("data")
            .ok_or_else(|| anyhow::anyhow!("Missing 'data' field in response"))?;

        Ok(PushKeysResponse {
            added: data.get("added")
                .and_then(|v| v.as_array())
                .map(|arr| arr.iter().filter_map(|s| s.as_str().map(|s| s.to_string())).collect())
                .unwrap_or_default(),
            existed: data.get("existed")
                .and_then(|v| v.as_array())
                .map(|arr| arr.iter().filter_map(|s| s.as_str().map(|s| s.to_string())).collect())
                .unwrap_or_default(),
            failed: data.get("failed")
                .and_then(|v| v.as_array())
                .map(|arr| arr.iter().filter_map(|s| s.as_str().map(|s| s.to_string())).collect())
                .unwrap_or_default(),
        })
    }
}

/// 推送键响应
///
/// 描述批量推送操作的结果。
#[derive(Debug, Clone, Default)]
pub struct PushKeysResponse {
    /// 新创建的键
    pub added: Vec<String>,
    /// 已存在的键
    pub existed: Vec<String>,
    /// 失败的键
    pub failed: Vec<String>,
}

impl PushKeysResponse {
    /// 获取总处理数
    pub fn total(&self) -> usize {
        self.added.len() + self.existed.len() + self.failed.len()
    }

    /// 检查是否全部成功
    pub fn is_success(&self) -> bool {
        self.failed.is_empty()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_push_keys_response_default() {
        let response = PushKeysResponse::default();
        assert_eq!(response.total(), 0);
        assert!(response.is_success());
    }

    #[test]
    fn test_push_keys_response_with_data() {
        let response = PushKeysResponse {
            added: vec!["key1".to_string(), "key2".to_string()],
            existed: vec!["key3".to_string()],
            failed: Vec::new(),
        };
        assert_eq!(response.total(), 3);
        assert!(response.is_success());
    }

    #[test]
    fn test_push_keys_response_with_failures() {
        let response = PushKeysResponse {
            added: vec!["key1".to_string()],
            existed: Vec::new(),
            failed: vec!["key2".to_string()],
        };
        assert_eq!(response.total(), 2);
        assert!(!response.is_success());
    }

    #[test]
    fn test_transform_translations_format_key_centric_to_language_centric() {
        // Key-centric format (API response)
        let key_centric = serde_json::json!({
            "greeting": {
                "en": "Hello",
                "zh_CN": "你好"
            },
            "farewell": {
                "en": "Goodbye"
            },
            "user.name": {
                "en": "John",
                "zh_CN": "张三"
            }
        });

        let result = APIClient::transform_translations_format(key_centric).unwrap();

        // Verify structure
        assert_eq!(result.len(), 2);
        assert!(result.contains_key("en"));
        assert!(result.contains_key("zh_CN"));

        // Verify English translations
        let en = result.get("en").unwrap();
        assert_eq!(en.len(), 3);
        assert_eq!(en.get("greeting"), Some(&"Hello".to_string()));
        assert_eq!(en.get("farewell"), Some(&"Goodbye".to_string()));
        assert_eq!(en.get("user.name"), Some(&"John".to_string()));

        // Verify Chinese translations
        let zh = result.get("zh_CN").unwrap();
        assert_eq!(zh.len(), 2);
        assert_eq!(zh.get("greeting"), Some(&"你好".to_string()));
        assert_eq!(zh.get("user.name"), Some(&"张三".to_string()));
    }

    #[test]
    fn test_transform_translations_format_empty() {
        let empty = serde_json::json!({});
        let result = APIClient::transform_translations_format(empty).unwrap();
        assert!(result.is_empty());
    }

    // ========== URL 规范化测试 ==========

    #[test]
    fn test_api_client_new_url_normalization_trailing_slash() {
        // 测试移除 URL 末尾斜杠
        let client = APIClient::new(
            "http://localhost:8080/api/".to_string(),
            "test-key".to_string(),
            1,
        )
        .unwrap();
        assert_eq!(client.base_url(), "http://localhost:8080/api");
        assert_eq!(client.project_id(), 1);
        assert_eq!(client.api_key(), "test-key");
    }

    #[test]
    fn test_api_client_new_url_normalization_multiple_slashes() {
        // 测试移除多个末尾斜杠
        let client = APIClient::new(
            "http://localhost:8080/api///".to_string(),
            "test-key".to_string(),
            1,
        )
        .unwrap();
        assert_eq!(client.base_url(), "http://localhost:8080/api");
    }

    #[test]
    fn test_api_client_new_url_normalization_with_whitespace() {
        // 测试去除首尾空白字符
        let client = APIClient::new(
            "  http://localhost:8080/api  ".to_string(),
            "test-key".to_string(),
            1,
        )
        .unwrap();
        assert_eq!(client.base_url(), "http://localhost:8080/api");
    }

    #[test]
    fn test_api_client_new_url_with_path() {
        // 测试带有路径的 URL
        let client = APIClient::new(
            "https://api.example.com/v1/".to_string(),
            "test-key".to_string(),
            42,
        )
        .unwrap();
        assert_eq!(client.base_url(), "https://api.example.com/v1");
        assert_eq!(client.project_id(), 42);
    }

    #[test]
    fn test_api_client_new_url_invalid_empty() {
        // 测试空 URL
        let result = APIClient::new("".to_string(), "test-key".to_string(), 1);
        assert!(result.is_err());
        assert!(result.unwrap_err().to_string().contains("cannot be empty"));
    }

    #[test]
    fn test_api_client_new_url_invalid_whitespace_only() {
        // 测试只包含空白的 URL
        let result = APIClient::new("   ".to_string(), "test-key".to_string(), 1);
        assert!(result.is_err());
        assert!(result.unwrap_err().to_string().contains("cannot be empty"));
    }

    #[test]
    fn test_api_client_new_url_invalid_format_no_protocol() {
        // 测试缺少协议的 URL
        let result = APIClient::new("localhost:8080/api".to_string(), "test-key".to_string(), 1);
        assert!(result.is_err());
        let error_msg = result.unwrap_err().to_string();
        assert!(error_msg.contains("http://") || error_msg.contains("https://"));
    }

    #[test]
    fn test_api_client_new_url_invalid_format_relative_path() {
        // 测试相对路径 URL
        let result = APIClient::new("/api/v1".to_string(), "test-key".to_string(), 1);
        assert!(result.is_err());
    }

    #[test]
    fn test_api_client_new_project_id_zero() {
        // 测试无效的项目 ID
        let result = APIClient::new(
            "http://localhost:8080/api".to_string(),
            "test-key".to_string(),
            0,
        );
        assert!(result.is_err());
        assert!(result.unwrap_err().to_string().contains("positive"));
    }

    #[test]
    fn test_api_client_new_https_url() {
        // 测试 HTTPS URL
        let client = APIClient::new(
            "https://secure-api.example.com/".to_string(),
            "secure-key".to_string(),
            99,
        )
        .unwrap();
        assert_eq!(client.base_url(), "https://secure-api.example.com");
        assert_eq!(client.project_id(), 99);
    }
}
