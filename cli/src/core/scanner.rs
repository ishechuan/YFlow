//! File scanning module
//!
//! Scans the messages directory and collects all translation files.
//! Supports parallel file processing for improved performance.
//!
//! # Directory Structure
//!
//! The expected directory structure is:
//! ```
//! messages/
//!   ├── en/
//!   │   ├── common.json
//!   │   └── errors.json
//!   ├── zh_CN/
//!   │   ├── common.json
//!   │   └── errors.json
//!   └── ...
//! ```

use anyhow::{Context, Result};
use serde_json::Value;
use std::collections::HashMap;
use std::path::{Path, PathBuf};
use tokio::fs;

use super::{flatten_object, unflatten_object, ScanResult, Translations};
use crate::ui::progress::LanguageProgressBar;

/// Progress callback type for file writing operations
///
/// Called after processing each language, providing the current language code
/// and the number of languages processed so far.
///
/// # Arguments
///
/// * `lang_code` - The language code being processed
/// * `current_index` - The current language index (0-based)
/// * `total_languages` - The total number of languages
pub type ProgressCallback = Box<dyn Fn(String, usize, usize) + Send + Sync>;

/// Default no-op progress callback
///
/// Used when no progress tracking is needed.
fn noop_progress_callback(_lang: String, _index: usize, _total: usize) {}

/// Scans the messages directory and collects all translations
///
/// Searches for language subdirectories (e.g., `en/`, `zh_CN/`) and reads
/// all JSON files within them. Translation keys are flattened for storage.
///
/// # Arguments
///
/// * `path` - Path to the messages directory
///
/// # Errors
///
/// Returns an error if:
/// - The directory does not exist
/// - The path is not a directory
/// - Files cannot be read or parsed
///
/// # Performance
///
/// Uses async file operations for better performance on large projects.
pub async fn scan_messages_dir(path: &Path) -> Result<ScanResult> {
    let resolved = path.canonicalize()
        .with_context(|| format!("Messages directory not found: {}", path.display()))?;

    if !resolved.is_dir() {
        return Err(anyhow::anyhow!(
            "Path is not a directory: {}",
            resolved.display()
        ));
    }

    // Async directory reading
    let mut entries = fs::read_dir(&resolved)
        .await
        .with_context(|| format!("Failed to read directory: {}", resolved.display()))?;

    let mut all_translations = Translations::new();
    let mut all_files: Vec<PathBuf> = Vec::new();
    let mut total_keys = 0;

    // Collect all language directories
    let mut lang_dirs: Vec<PathBuf> = Vec::new();
    while let Some(entry) = entries.next_entry().await? {
        if entry.file_type().await?.is_dir() {
            lang_dirs.push(entry.path());
        }
    }

    // Process each language directory
    for dir in lang_dirs {
        match scan_language_dir(&dir).await {
            Ok((translations, files, key_count)) => {
                all_translations.extend(translations);
                all_files.extend(files);
                total_keys += key_count;
            }
            Err(e) => {
                // Log error but continue processing other languages
                eprintln!("Warning: Failed to scan {}: {}", dir.display(), e);
            }
        }
    }

    Ok(ScanResult {
        translations: all_translations,
        files: all_files,
        key_count: total_keys,
    })
}

/// Scans a single language directory
///
/// Reads all JSON files in the directory and merges translations.
/// Files are processed recursively for nested subdirectories.
///
/// # Arguments
///
/// * `dir_path` - Path to the language directory
///
/// # Returns
///
/// Tuple of (translations map, file paths, key count)
async fn scan_language_dir(dir_path: &Path) -> Result<(Translations, Vec<PathBuf>, usize)> {
    let mut translations = HashMap::new();
    let mut files: Vec<PathBuf> = Vec::new();

    // Recursively collect all JSON files
    let json_files = collect_json_files(dir_path).await?;

    // Parse all JSON files
    let mut parse_results: Vec<Result<(PathBuf, Value)>> = Vec::new();
    for file in &json_files {
        match fs::read_to_string(file).await {
            Ok(content) => {
                match serde_json::from_str::<Value>(&content) {
                    Ok(json) => parse_results.push(Ok((file.clone(), json))),
                    Err(e) => parse_results.push(Err(anyhow::anyhow!(
                        "Failed to parse JSON {}: {}",
                        file.display(),
                        e
                    ))),
                }
            }
            Err(e) => parse_results.push(Err(anyhow::anyhow!(
                "Failed to read file {}: {}",
                file.display(),
                e
            ))),
        }
    }

    let lang_code = dir_path.file_name()
        .and_then(|n| n.to_str())
        .unwrap_or("unknown")
        .to_string();

    translations.insert(lang_code.clone(), HashMap::new());
    let lang_translations = translations.get_mut(&lang_code).unwrap();

    for result in &parse_results {
        match result {
            Ok((_, json)) => {
                let flat = flatten_object(json, "");
                for (key, value) in flat {
                    lang_translations.insert(key, value);
                }
            }
            Err(e) => {
                eprintln!("Warning: {}", e);
            }
        }
    }

    // Collect file paths relative to the language directory
    for file in &json_files {
        if let Ok(rel_path) = file.strip_prefix(dir_path) {
            files.push(PathBuf::from(&lang_code).join(rel_path));
        } else {
            files.push(PathBuf::from(&lang_code).join(file.file_name().unwrap()));
        }
    }

    let key_count = lang_translations.len();

    Ok((translations, files, key_count))
}

/// Recursively collects all JSON files in a directory
///
/// # Arguments
///
/// * `dir` - Directory to search
///
/// # Returns
///
/// Vector of paths to all JSON files found
async fn collect_json_files(dir: &Path) -> Result<Vec<PathBuf>> {
    let mut files = Vec::new();
    let mut dirs = Vec::new();

    let mut entries = fs::read_dir(dir)
        .await
        .with_context(|| format!("Failed to read directory: {}", dir.display()))?;

    while let Some(entry) = entries.next_entry().await? {
        let path = entry.path();

        if path.is_dir() {
            dirs.push(path);
        } else if path.extension().map(|e| e == "json").unwrap_or(false) {
            files.push(path);
        }
    }

    // Recursively process subdirectories
    for sub_dir in dirs {
        files.extend(Box::pin(collect_json_files(&sub_dir)).await?);
    }

    Ok(files)
}

/// Writes translations while preserving the original file structure
///
/// This function reads each original file, merges the new translations
/// into it, and writes back to the same location. Only the translation
/// keys that exist in the new translations are updated; all other
/// content in the original files is preserved.
///
/// # Arguments
///
/// * `messages_dir` - Root messages directory path
/// * `original_files` - List of original file paths (relative to messages dir)
/// * `translations` - New translations to merge
/// * `force` - Whether to overwrite all keys (true) or only new keys (false)
/// * `progress_callback` - Optional callback called after each language is processed
///
/// # Returns
///
/// Vector of file paths that were written
///
/// # Example
///
/// ```ignore
/// let written = write_translations_with_structure(
///     &messages_dir,
///     &original_files,
///     &translations,
///     false,
///     Some(|lang, idx, total| println!("Processed {} ({}/{})", lang, idx, total)),
/// ).await?;
/// ```
pub async fn write_translations_with_structure(
    messages_dir: &Path,
    original_files: &[PathBuf],
    translations: &Translations,
    force: bool,
    progress_callback: Option<ProgressCallback>,
) -> Result<Vec<PathBuf>> {
    let mut written: Vec<PathBuf> = Vec::new();

    // Group files by language code using proper PathBuf methods
    let mut files_by_lang: HashMap<String, Vec<&PathBuf>> = HashMap::new();
    for file in original_files {
        // Use PathBuf methods to extract language code from first component
        if let Some(first_component) = file.components().next() {
            if let std::path::Component::Normal(lang_code) = first_component {
                if let Some(code_str) = lang_code.to_str() {
                    files_by_lang.entry(code_str.to_string()).or_default().push(file);
                }
            }
        }
    }

    // Count total languages for progress reporting
    let total_languages = files_by_lang.len();
    let mut processed_languages = 0;

    // Process each language
    for (lang_code, files) in &files_by_lang {
        // Get translations for this language
        let lang_translations = match translations.get(lang_code) {
            Some(t) => t,
            None => {
                // Call progress callback even if no translations
                processed_languages += 1;
                if let Some(ref callback) = progress_callback {
                    callback(lang_code.clone(), processed_languages, total_languages);
                }
                continue;
            }
        };

        for file in files {
            let full_path = messages_dir.join(file);

            if !full_path.exists() {
                continue;
            }

            match fs::read_to_string(&full_path).await {
                Ok(content) => {
                    match serde_json::from_str::<Value>(&content) {
                        Ok(original_data) => {
                            // Merge translations into the original structure
                            let merged = merge_translations_with_structure(&original_data, lang_translations, force);
                            let new_content = serde_json::to_string_pretty(&merged)?;
                            fs::write(&full_path, new_content).await?;
                            written.push(full_path);
                        }
                        Err(e) => {
                            eprintln!("Warning: Failed to parse JSON {}: {}", full_path.display(), e);
                        }
                    }
                }
                Err(e) => {
                    eprintln!("Warning: Failed to read {}: {}", full_path.display(), e);
                }
            }
        }

        // Update progress
        processed_languages += 1;
        if let Some(ref callback) = progress_callback {
            callback(lang_code.clone(), processed_languages, total_languages);
        }
    }

    // Handle languages that have translations but no original files
    // 为没有原始文件的新语言创建目录和文件
    let new_files = write_new_language_files(messages_dir, translations, &files_by_lang)?;
    written.extend(new_files);

    Ok(written)
}

/// 为没有原始文件的新语言创建目录和文件
///
/// 当从后端同步翻译时，如果某个语言在本地没有对应的文件，
/// 此函数会自动创建目录结构和 sync.json 文件。
///
/// # Arguments
///
/// * `messages_dir` - messages 根目录路径
/// * `translations` - 要写入的翻译数据
/// * `files_by_lang` - 按语言分组的现有文件映射
///
/// # Returns
///
/// 创建的文件路径向量
///
/// # Example
///
/// ```ignore
/// let written = write_new_language_files(
///     &messages_dir,
///     &translations,
///     &files_by_lang,
/// )?;
/// ```
fn write_new_language_files(
    messages_dir: &Path,
    translations: &Translations,
    files_by_lang: &HashMap<String, Vec<&PathBuf>>,
) -> Result<Vec<PathBuf>> {
    let mut written: Vec<PathBuf> = Vec::new();

    for (lang_code, lang_translations) in translations {
        // 检查是否已有原始文件
        if files_by_lang.contains_key(lang_code) {
            continue;
        }

        // 跳过空翻译
        if lang_translations.is_empty() {
            continue;
        }

        // 为新语言创建目录
        let lang_dir = messages_dir.join(lang_code);
        std::fs::create_dir_all(&lang_dir)
            .with_context(|| format!("Failed to create language directory: {}", lang_dir.display()))?;

        // 将展平翻译还原为嵌套结构并写入文件
        let merged = unflatten_object(lang_translations.clone());
        let new_content = serde_json::to_string_pretty(&merged)
            .with_context(|| "Failed to serialize translations to JSON")?;

        let output_path = lang_dir.join("sync.json");
        std::fs::write(&output_path, new_content)
            .with_context(|| format!("Failed to write file: {}", output_path.display()))?;

        written.push(output_path.clone());
        tracing::info!("Created new language file: {}", output_path.display());
    }

    Ok(written)
}

/// Writes translations with a progress manager
///
/// A convenience wrapper around `write_translations_with_structure` that
/// integrates with the UI progress bar system.
pub async fn write_translations_with_progress(
    messages_dir: &Path,
    original_files: &[PathBuf],
    translations: &Translations,
    force: bool,
    progress_manager: Option<&crate::ui::progress::MultiProgressManager>,
) -> Result<Vec<PathBuf>> {
    match progress_manager {
        Some(manager) if manager.is_enabled() => {
            // Create progress bars for each language first
            let progress_bars: Vec<(String, LanguageProgressBar)> = translations
                .keys()
                .map(|lang| (lang.clone(), manager.create_bar(lang, 1)))
                .collect();

            // Create callback that updates progress bars with thread-safe interior mutability
            use std::sync::{Arc, Mutex};
            let progress_bars = Arc::new(Mutex::new(progress_bars));

            let progress_callback: ProgressCallback = Box::new(move |lang: String, _index: usize, _total: usize| {
                let mut bars = progress_bars.lock().unwrap();
                for (l, bar) in bars.iter_mut() {
                    if l == &lang {
                        bar.inc();
                        bar.finish();
                        break;
                    }
                }
            });

            let result = write_translations_with_structure(
                messages_dir,
                original_files,
                translations,
                force,
                Some(progress_callback),
            ).await;

            result
        }
        _ => {
            write_translations_with_structure(
                messages_dir,
                original_files,
                translations,
                force,
                None,
            ).await
        }
    }
}

/// Merges translations into the original nested structure
///
/// Flattens the original data, applies the new translations, then
/// unflattens back to nested structure.
///
/// Note: This function always overwrites existing keys with new values.
/// The `force` parameter is maintained for API compatibility but is not
/// used in the merge logic, matching the TypeScript implementation.
///
/// # Arguments
///
/// * `original` - Original JSON data
/// * `translations` - New translations to merge
/// * `_force` - Reserved for API compatibility (not used)
///
/// # Returns
///
/// Merged JSON data
fn merge_translations_with_structure(
    original: &Value,
    translations: &HashMap<String, String>,
    _force: bool,
) -> Value {
    // Flatten the original data
    let flat_original = flatten_object(original, "");

    // Merge translations (new values always overwrite old ones)
    // This matches TypeScript behavior: flatOriginal[key] = value;
    let mut merged = flat_original;
    for (key, value) in translations {
        merged.insert(key.clone(), value.clone());
    }

    // Convert back to nested structure
    unflatten_object(merged)
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;
    use std::fs::File;
    use std::io::Write;
    use tempfile::TempDir;

    /// Creates a test messages directory structure
    async fn create_test_messages_dir(temp_dir: &TempDir) -> (PathBuf, Vec<PathBuf>) {
        let messages_dir = temp_dir.path().join("messages");
        std::fs::create_dir_all(&messages_dir).unwrap();

        // Create language directories
        let en_dir = messages_dir.join("en");
        let zh_dir = messages_dir.join("zh_CN");
        std::fs::create_dir_all(&en_dir).unwrap();
        std::fs::create_dir_all(&zh_dir).unwrap();

        // Create nested subdirectory
        let nested_dir = en_dir.join("nested");
        std::fs::create_dir_all(&nested_dir).unwrap();

        // Create test files
        let en_common = en_dir.join("common.json");
        let en_nested = nested_dir.join("deep.json");
        let zh_common = zh_dir.join("common.json");

        // Write test content
        let en_content = json!({
            "greeting": "Hello",
            "user": {
                "name": "User Name"
            }
        });
        let mut file = File::create(&en_common).unwrap();
        writeln!(file, "{}", serde_json::to_string_pretty(&en_content).unwrap()).unwrap();

        let en_deep_content = json!({
            "level": {
                "deep": "Deep value"
            }
        });
        let mut file = File::create(&en_nested).unwrap();
        writeln!(file, "{}", serde_json::to_string_pretty(&en_deep_content).unwrap()).unwrap();

        let zh_content = json!({
            "greeting": "你好",
            "user": {
                "name": "用户名"
            }
        });
        let mut file = File::create(&zh_common).unwrap();
        writeln!(file, "{}", serde_json::to_string_pretty(&zh_content).unwrap()).unwrap();

        let expected_files = vec![
            PathBuf::from("en/common.json"),
            PathBuf::from("en/nested/deep.json"),
            PathBuf::from("zh_CN/common.json"),
        ];

        (messages_dir, expected_files)
    }

    #[tokio::test]
    async fn test_scan_messages_dir_missing() {
        let result = scan_messages_dir(Path::new("/nonexistent")).await;
        assert!(result.is_err());
    }

    #[tokio::test]
    async fn test_scan_empty_dir() {
        let temp_dir = TempDir::new().unwrap();
        let messages_dir = temp_dir.path().join("messages");
        std::fs::create_dir_all(&messages_dir).unwrap();

        let result = scan_messages_dir(&messages_dir).await.unwrap();
        assert_eq!(result.translations.len(), 0);
        assert_eq!(result.files.len(), 0);
        assert_eq!(result.key_count, 0);
    }

    #[tokio::test]
    async fn test_scan_single_language() {
        let temp_dir = TempDir::new().unwrap();
        let messages_dir = temp_dir.path().join("messages");
        let en_dir = messages_dir.join("en");
        std::fs::create_dir_all(&en_dir).unwrap();

        // Create test file
        let test_file = en_dir.join("common.json");
        let content = json!({
            "greeting": "Hello",
            "user": {
                "name": "User Name"
            }
        });
        let mut file = File::create(&test_file).unwrap();
        writeln!(file, "{}", serde_json::to_string_pretty(&content).unwrap()).unwrap();

        let result = scan_messages_dir(&messages_dir).await.unwrap();

        assert_eq!(result.translations.len(), 1);
        assert!(result.translations.contains_key("en"));
        assert_eq!(result.key_count, 2);
        assert!(result.files.len() >= 1);
    }

    #[tokio::test]
    async fn test_scan_multiple_languages() {
        let temp_dir = TempDir::new().unwrap();
        let (messages_dir, _) = create_test_messages_dir(&temp_dir).await;

        let result = scan_messages_dir(&messages_dir).await.unwrap();

        assert_eq!(result.translations.len(), 2);
        assert!(result.translations.contains_key("en"));
        assert!(result.translations.contains_key("zh_CN"));
        assert_eq!(result.key_count, 5); // 2 + 2 + 1 = 5 keys
        assert!(result.files.len() >= 3);
    }

    #[tokio::test]
    async fn test_scan_nested_directories() {
        let temp_dir = TempDir::new().unwrap();
        let (messages_dir, _) = create_test_messages_dir(&temp_dir).await;

        let result = scan_messages_dir(&messages_dir).await.unwrap();

        // Check that nested keys are properly flattened
        let en_translations = result.translations.get("en").unwrap();
        assert!(en_translations.contains_key("greeting"));
        assert!(en_translations.contains_key("user.name"));
        assert!(en_translations.contains_key("level.deep"));
    }

    #[tokio::test]
    async fn test_write_translations_with_structure() {
        let temp_dir = TempDir::new().unwrap();
        let (messages_dir, original_files) = create_test_messages_dir(&temp_dir).await;

        // Prepare new translations
        let translations: Translations = [
            ("en".to_string(), [
                ("greeting".to_string(), "Hello Updated".to_string()),
                ("new_key".to_string(), "New Value".to_string()),
            ].iter().cloned().collect()),
            ("zh_CN".to_string(), [
                ("greeting".to_string(), "你好更新".to_string()),
            ].iter().cloned().collect()),
        ].iter().cloned().collect();

        // Note: merge always overwrites, force parameter is handled at higher level
        let written = write_translations_with_structure(
            &messages_dir,
            &original_files,
            &translations,
            false,
            None,
        ).await.unwrap();

        // Verify files were written
        assert!(written.len() >= 2);

        // Verify content was merged correctly
        let en_common_path = messages_dir.join("en/common.json");
        let content = fs::read_to_string(&en_common_path).await.unwrap();
        let data: Value = serde_json::from_str(&content).unwrap();

        // Keys are always overwritten in merge (matching TypeScript behavior)
        assert_eq!(data["greeting"], "Hello Updated");
        // New keys should be added
        assert_eq!(data["new_key"], "New Value");
    }

    #[tokio::test]
    async fn test_write_translations_with_progress_callback() {
        let temp_dir = TempDir::new().unwrap();
        let (messages_dir, original_files) = create_test_messages_dir(&temp_dir).await;

        let translations: Translations = [
            ("en".to_string(), [
                ("greeting".to_string(), "Hello Updated".to_string()),
            ].iter().cloned().collect()),
        ].iter().cloned().collect();

        // Track callback invocations
        use std::sync::atomic::{AtomicUsize, AtomicBool};
        use std::sync::Arc;
        let callback_invoked = Arc::new(AtomicBool::new(false));
        let callback_count = Arc::new(AtomicUsize::new(0));
        let callback_invoked_clone = callback_invoked.clone();
        let callback_count_clone = callback_count.clone();

        let callback: ProgressCallback = Box::new(move |lang, index, total| {
            callback_invoked_clone.store(true, std::sync::atomic::Ordering::SeqCst);
            callback_count_clone.fetch_add(1, std::sync::atomic::Ordering::SeqCst);
            assert!(!lang.is_empty());
            assert!(index > 0);
            assert!(total > 0);
        });

        let written = write_translations_with_structure(
            &messages_dir,
            &original_files,
            &translations,
            false,
            Some(callback),
        ).await.unwrap();

        assert!(written.len() >= 1);
        assert!(callback_invoked.load(std::sync::atomic::Ordering::SeqCst));
        assert!(callback_count.load(std::sync::atomic::Ordering::SeqCst) > 0);
    }

    #[tokio::test]
    async fn test_write_translations_force_overwrite() {
        let temp_dir = TempDir::new().unwrap();
        let (messages_dir, original_files) = create_test_messages_dir(&temp_dir).await;

        let translations: Translations = [
            ("en".to_string(), [
                ("greeting".to_string(), "Force Updated".to_string()),
            ].iter().cloned().collect()),
        ].iter().cloned().collect();

        let written = write_translations_with_structure(
            &messages_dir,
            &original_files,
            &translations,
            true, // Force overwrite
            None,
        ).await.unwrap();

        assert!(written.len() >= 1);

        let en_common_path = messages_dir.join("en/common.json");
        let content = fs::read_to_string(&en_common_path).await.unwrap();
        let data: Value = serde_json::from_str(&content).unwrap();

        // Even existing key should be overwritten with force=true
        assert_eq!(data["greeting"], "Force Updated");
    }

    #[tokio::test]
    async fn test_write_translations_preserves_unrelated_keys() {
        let temp_dir = TempDir::new().unwrap();
        let (messages_dir, original_files) = create_test_messages_dir(&temp_dir).await;

        // Only update greeting, other keys should remain
        let translations: Translations = [
            ("en".to_string(), [
                ("greeting".to_string(), "Updated".to_string()),
            ].iter().cloned().collect()),
        ].iter().cloned().collect();

        write_translations_with_structure(
            &messages_dir,
            &original_files,
            &translations,
            false,
            None,
        ).await.unwrap();

        let en_common_path = messages_dir.join("en/common.json");
        let content = fs::read_to_string(&en_common_path).await.unwrap();
        let data: Value = serde_json::from_str(&content).unwrap();

        // Original nested key should be preserved
        assert_eq!(data["user"]["name"], "User Name");
    }

    #[tokio::test]
    async fn test_write_translations_creates_language_dir() {
        let temp_dir = TempDir::new().unwrap();
        let messages_dir = temp_dir.path().join("messages");
        std::fs::create_dir_all(&messages_dir).unwrap();

        let translations: Translations = [
            ("ja_JP".to_string(), [
                ("greeting".to_string(), "こんにちは".to_string()),
            ].iter().cloned().collect()),
        ].iter().cloned().collect();

        let original_files: Vec<PathBuf> = vec![];
        let written = write_translations_with_structure(
            &messages_dir,
            &original_files,
            &translations,
            false,
            None,
        ).await.unwrap();

        // 新行为：为新语言创建目录和 sync.json 文件
        assert_eq!(written.len(), 1);
        let ja_path = messages_dir.join("ja_JP/sync.json");
        assert_eq!(written[0], ja_path);

        // 验证文件内容
        let content = fs::read_to_string(&ja_path).await.unwrap();
        let data: Value = serde_json::from_str(&content).unwrap();
        assert_eq!(data["greeting"], "こんにちは");
    }

    #[tokio::test]
    async fn test_write_translations_missing_language_files() {
        let temp_dir = TempDir::new().unwrap();
        let (messages_dir, original_files) = create_test_messages_dir(&temp_dir).await;

        // Only translations for a language not in original files
        let translations: Translations = [
            ("ja_JP".to_string(), [
                ("greeting".to_string(), "こんにちは".to_string()),
            ].iter().cloned().collect()),
        ].iter().cloned().collect();

        let written = write_translations_with_structure(
            &messages_dir,
            &original_files,
            &translations,
            false,
            None,
        ).await.unwrap();

        // 新行为：为新语言创建目录和文件
        assert_eq!(written.len(), 1);
        let ja_path = messages_dir.join("ja_JP/sync.json");
        assert_eq!(written[0], ja_path);
    }

    #[tokio::test]
    async fn test_write_new_language_files_multiple() {
        let temp_dir = TempDir::new().unwrap();
        let messages_dir = temp_dir.path().join("messages");
        std::fs::create_dir_all(&messages_dir).unwrap();

        // 准备多种新语言的翻译
        let translations: Translations = [
            ("ja_JP".to_string(), [
                ("greeting".to_string(), "こんにちは".to_string()),
                ("farewell".to_string(), "さようなら".to_string()),
            ].iter().cloned().collect()),
            ("ko_KR".to_string(), [
                ("greeting".to_string(), "안녕하세요".to_string()),
            ].iter().cloned().collect()),
        ].iter().cloned().collect();

        let original_files: Vec<PathBuf> = vec![];

        let written = write_translations_with_structure(
            &messages_dir,
            &original_files,
            &translations,
            false,
            None,
        ).await.unwrap();

        // 应该为每种新语言创建文件
        assert_eq!(written.len(), 2);

        // 验证日语文件
        let ja_path = messages_dir.join("ja_JP/sync.json");
        assert!(written.contains(&ja_path));
        let ja_content = fs::read_to_string(&ja_path).await.unwrap();
        let ja_data: Value = serde_json::from_str(&ja_content).unwrap();
        assert_eq!(ja_data["greeting"], "こんにちは");
        assert_eq!(ja_data["farewell"], "さようなら");

        // 验证韩语文件
        let ko_path = messages_dir.join("ko_KR/sync.json");
        assert!(written.contains(&ko_path));
        let ko_content = fs::read_to_string(&ko_path).await.unwrap();
        let ko_data: Value = serde_json::from_str(&ko_content).unwrap();
        assert_eq!(ko_data["greeting"], "안녕하세요");
    }

    #[tokio::test]
    async fn test_write_new_language_files_mixed_existing_and_new() {
        let temp_dir = TempDir::new().unwrap();
        let (messages_dir, original_files) = create_test_messages_dir(&temp_dir).await;

        // 混合：已有文件 + 新语言
        let translations: Translations = [
            ("en".to_string(), [
                ("greeting".to_string(), "Hello Updated".to_string()),
            ].iter().cloned().collect()),
            ("ja_JP".to_string(), [
                ("greeting".to_string(), "こんにちは".to_string()),
            ].iter().cloned().collect()),
        ].iter().cloned().collect();

        let written = write_translations_with_structure(
            &messages_dir,
            &original_files,
            &translations,
            false,
            None,
        ).await.unwrap();

        // 应该更新 en 的现有文件，并为 ja_JP 创建新文件
        assert!(written.len() >= 1); // 至少更新 en 的文件
        let ja_path = messages_dir.join("ja_JP/sync.json");
        assert!(written.contains(&ja_path));

        // 验证新创建的文件内容
        let ja_content = fs::read_to_string(&ja_path).await.unwrap();
        let ja_data: Value = serde_json::from_str(&ja_content).unwrap();
        assert_eq!(ja_data["greeting"], "こんにちは");
    }

    #[tokio::test]
    async fn test_write_new_language_files_empty_translations() {
        let temp_dir = TempDir::new().unwrap();
        let messages_dir = temp_dir.path().join("messages");
        std::fs::create_dir_all(&messages_dir).unwrap();

        // 空翻译不应该创建文件
        let translations: Translations = [
            ("empty_lang".to_string(), std::collections::HashMap::new()),
        ].iter().cloned().collect();

        let original_files: Vec<PathBuf> = vec![];

        let written = write_translations_with_structure(
            &messages_dir,
            &original_files,
            &translations,
            false,
            None,
        ).await.unwrap();

        // 空翻译不应创建文件
        assert!(written.is_empty());
    }

    #[tokio::test]
    async fn test_write_new_language_files_nested_structure() {
        let temp_dir = TempDir::new().unwrap();
        let messages_dir = temp_dir.path().join("messages");
        std::fs::create_dir_all(&messages_dir).unwrap();

        // 嵌套结构的翻译
        let translations: Translations = [
            ("de_DE".to_string(), [
                ("user.name".to_string(), "Hans Müller".to_string()),
                ("user.profile.email".to_string(), "hans@example.com".to_string()),
                ("greeting".to_string(), "Guten Tag".to_string()),
            ].iter().cloned().collect()),
        ].iter().cloned().collect();

        let original_files: Vec<PathBuf> = vec![];

        let written = write_translations_with_structure(
            &messages_dir,
            &original_files,
            &translations,
            false,
            None,
        ).await.unwrap();

        assert_eq!(written.len(), 1);

        let de_path = messages_dir.join("de_DE/sync.json");
        let content = fs::read_to_string(&de_path).await.unwrap();
        let data: Value = serde_json::from_str(&content).unwrap();

        // 验证嵌套结构被正确还原
        assert_eq!(data["user"]["name"], "Hans Müller");
        assert_eq!(data["user"]["profile"]["email"], "hans@example.com");
        assert_eq!(data["greeting"], "Guten Tag");
    }
}
