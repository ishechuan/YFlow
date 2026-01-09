//! Language mapping module
//!
//! Handles translation between local language codes and backend language codes.
//! For example: "zh_CN" -> "zh", "zh_TW" -> "tw"

use std::collections::HashMap;

/// 语言映射器
///
/// 提供本地语言代码和后端语言代码之间的双向转换。
///
/// # Example
///
/// ```ignore
/// let mapper = LanguageMapper::new(Some(hashmap! {
///     "zh_CN".to_string() => "zh".to_string(),
///     "zh_TW".to_string() => "tw".to_string(),
/// }));
///
/// // 导入时：本地代码 -> 后端代码
/// let backend_code = mapper.to_backend("zh_CN");  // "zh"
///
/// // 同步时：后端代码 -> 本地代码
/// let local_code = mapper.to_local("zh");  // 如果没有映射，返回原值
/// ```
#[derive(Debug, Clone, Default)]
pub struct LanguageMapper {
    /// 本地代码 -> 后端代码
    local_to_backend: HashMap<String, String>,
    /// 后端代码 -> 本地代码（反向映射）
    backend_to_local: HashMap<String, String>,
}

impl LanguageMapper {
    /// 创建新的语言映射器
    ///
    /// # Arguments
    ///
    /// * `mapping` - 可选的语言映射表，格式为 `{"local": "backend"}`
    pub fn new(mapping: Option<HashMap<String, String>>) -> Self {
        let mapping = mapping.unwrap_or_default();
        let mut local_to_backend = HashMap::new();
        let mut backend_to_local = HashMap::new();

        for (local, backend) in &mapping {
            local_to_backend.insert(local.clone(), backend.clone());
            backend_to_local.insert(backend.clone(), local.clone());
        }

        Self {
            local_to_backend,
            backend_to_local,
        }
    }

    /// 将本地语言代码转换为后端语言代码
    ///
    /// 如果没有定义映射，返回原代码。
    ///
    /// # Arguments
    ///
    /// * `local_code` - 本地语言代码
    ///
    /// # Returns
    ///
    /// 对应的后端语言代码，如果没有映射则返回原代码
    pub fn to_backend(&self, local_code: &str) -> String {
        self.local_to_backend
            .get(local_code)
            .cloned()
            .unwrap_or_else(|| local_code.to_string())
    }

    /// 将后端语言代码转换为本地语言代码
    ///
    /// 用于同步操作时，将后端返回的语言代码转换回本地代码。
    /// 如果没有定义映射，返回原代码。
    ///
    /// # Arguments
    ///
    /// * `backend_code` - 后端语言代码
    ///
    /// # Returns
    ///
    /// 对应的本地语言代码，如果没有映射则返回原代码
    pub fn to_local(&self, backend_code: &str) -> String {
        self.backend_to_local
            .get(backend_code)
            .cloned()
            .unwrap_or_else(|| backend_code.to_string())
    }

    /// 应用语言映射：将翻译数据的语言代码转换为后端代码
    ///
    /// 导入操作时使用，将本地语言代码转换为后端期望的代码。
    ///
    /// # Arguments
    ///
    /// * `translations` - 原始翻译数据
    ///
    /// # Returns
    ///
    /// 转换后的翻译数据
    pub fn apply_to_translations(
        &self,
        translations: HashMap<String, HashMap<String, String>>,
    ) -> HashMap<String, HashMap<String, String>> {
        let mut result = HashMap::new();

        for (local_code, lang_data) in translations {
            let backend_code = self.to_backend(&local_code);

            if !result.contains_key(&backend_code) {
                result.insert(backend_code.clone(), HashMap::new());
            }

            if let Some(target) = result.get_mut(&backend_code) {
                target.extend(lang_data);
            }
        }

        result
    }

    /// 反向应用语言映射：将翻译数据的语言代码转换为本地代码
    ///
    /// 同步操作时使用，将后端返回的语言代码转换回本地代码。
    ///
    /// # Arguments
    ///
    /// * `translations` - 后端返回的翻译数据
    ///
    /// # Returns
    ///
    /// 转换后的翻译数据
    pub fn reverse_translations(
        &self,
        translations: HashMap<String, HashMap<String, String>>,
    ) -> HashMap<String, HashMap<String, String>> {
        let mut result = HashMap::new();

        for (backend_code, lang_data) in translations {
            let local_code = self.to_local(&backend_code);

            if !result.contains_key(&local_code) {
                result.insert(local_code.clone(), HashMap::new());
            }

            if let Some(target) = result.get_mut(&local_code) {
                target.extend(lang_data);
            }
        }

        result
    }

    /// 检查是否需要进行语言代码映射
    ///
    /// # Returns
    ///
    /// 如果有定义映射返回 `true`，否则返回 `false`
    pub fn needs_mapping(&self) -> bool {
        !self.local_to_backend.is_empty()
    }

    /// 获取映射描述
    ///
    /// # Returns
    ///
    /// 描述当前映射的字符串，如 `"zh_CN → zh, zh_TW → tw"`
    pub fn get_description(&self) -> String {
        if !self.needs_mapping() {
            return "No language mapping".to_string();
        }

        let mappings: Vec<String> = self
            .local_to_backend
            .iter()
            .map(|(local, backend)| format!("{} → {}", local, backend))
            .collect();

        format!("Language mapping: {}", mappings.join(", "))
    }
}

/// 创建语言映射器的便捷函数
///
/// # Example
///
/// ```ignore
/// use std::collections::HashMap;
///
/// let mapping: HashMap<String, String> = HashMap::from([
///     ("zh_CN".to_string(), "zh".to_string()),
///     ("zh_TW".to_string(), "tw".to_string()),
/// ]);
///
/// let mapper = create_language_mapper(Some(mapping));
/// ```
pub fn create_language_mapper(
    mapping: Option<HashMap<String, String>>,
) -> LanguageMapper {
    LanguageMapper::new(mapping)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_no_mapping() {
        let mapper = LanguageMapper::new(None);

        assert!(!mapper.needs_mapping());
        assert_eq!(mapper.to_backend("en"), "en");
        assert_eq!(mapper.to_local("en"), "en");
    }

    #[test]
    fn test_single_mapping() {
        let mapper = LanguageMapper::new(Some(HashMap::from([
            ("zh_CN".to_string(), "zh".to_string()),
        ])));

        assert!(mapper.needs_mapping());
        assert_eq!(mapper.to_backend("zh_CN"), "zh");
        assert_eq!(mapper.to_local("zh"), "zh_CN");
    }

    #[test]
    fn test_multiple_mappings() {
        let mapper = LanguageMapper::new(Some(HashMap::from([
            ("zh_CN".to_string(), "zh".to_string()),
            ("zh_TW".to_string(), "tw".to_string()),
            ("en_US".to_string(), "en".to_string()),
        ])));

        assert_eq!(mapper.to_backend("zh_CN"), "zh");
        assert_eq!(mapper.to_backend("zh_TW"), "tw");
        assert_eq!(mapper.to_backend("en_US"), "en");

        // 未映射的代码应该原样返回
        assert_eq!(mapper.to_backend("ja_JP"), "ja_JP");
    }

    #[test]
    fn test_apply_to_translations() {
        let mapper = LanguageMapper::new(Some(HashMap::from([
            ("zh_CN".to_string(), "zh".to_string()),
        ])));

        let translations = HashMap::from([
            ("zh_CN".to_string(), HashMap::from([
                ("hello".to_string(), "你好".to_string()),
            ])),
            ("en".to_string(), HashMap::from([
                ("hello".to_string(), "Hello".to_string()),
            ])),
        ]);

        let result = mapper.apply_to_translations(translations);

        assert_eq!(result.len(), 2);
        assert!(result.contains_key("zh"));
        assert!(result.contains_key("en"));
        assert_eq!(result.get("zh").unwrap().get("hello"), Some(&"你好".to_string()));
        assert_eq!(result.get("en").unwrap().get("hello"), Some(&"Hello".to_string()));
    }

    #[test]
    fn test_reverse_translations() {
        let mapper = LanguageMapper::new(Some(HashMap::from([
            ("zh_CN".to_string(), "zh".to_string()),
        ])));

        let translations = HashMap::from([
            ("zh".to_string(), HashMap::from([
                ("hello".to_string(), "你好".to_string()),
            ])),
            ("en".to_string(), HashMap::from([
                ("hello".to_string(), "Hello".to_string()),
            ])),
        ]);

        let result = mapper.reverse_translations(translations);

        assert_eq!(result.len(), 2);
        assert!(result.contains_key("zh_CN"));
        assert!(result.contains_key("en"));
        assert_eq!(result.get("zh_CN").unwrap().get("hello"), Some(&"你好".to_string()));
    }

    #[test]
    fn test_get_description() {
        let mapper = LanguageMapper::new(Some(HashMap::from([
            ("zh_CN".to_string(), "zh".to_string()),
            ("zh_TW".to_string(), "tw".to_string()),
        ])));

        let desc = mapper.get_description();
        assert!(desc.contains("Language mapping:"));
        assert!(desc.contains("zh_CN → zh"));
        assert!(desc.contains("zh_TW → tw"));
    }

    #[test]
    fn test_duplicate_backend_codes() {
        // 如果两个本地代码映射到同一个后端代码，应该合并
        let mapper = LanguageMapper::new(Some(HashMap::from([
            ("zh_CN".to_string(), "zh".to_string()),
            ("zh_SG".to_string(), "zh".to_string()),
        ])));

        let translations = HashMap::from([
            ("zh_CN".to_string(), HashMap::from([
                ("greeting".to_string(), "你好".to_string()),
            ])),
            ("zh_SG".to_string(), HashMap::from([
                ("greeting".to_string(), "你好".to_string()),
            ])),
        ]);

        let result = mapper.apply_to_translations(translations);

        // 应该只返回一个 "zh" 语言
        assert_eq!(result.len(), 1);
        assert!(result.contains_key("zh"));
    }
}
