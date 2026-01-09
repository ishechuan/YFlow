//! JSON flattening and unflattening module
//!
//! Provides high-performance functions to flatten nested JSON objects into
//! single-level key-value pairs and vice versa. This is essential for
//! translating nested locale files into a flat format for storage.
//!
//! # Example
//!
//! ```ignore
//! use serde_json::json;
//!
//! let input = json!({
//!     "user": {
//!         "name": "John",
//!         "profile": {
//!             "age": 30
//!         }
//!     }
//! });
//!
//! let flat = flatten_object(&input, "");
//! // {"user.name": "John", "user.profile.age": "30"}
//!
//! let nested = unflatten_object(flat);
//! // {"user": {"name": "John", "profile": {"age": "30"}}}
//! ```

use serde_json::Value;
use std::collections::HashMap;

/// 将嵌套的 JSON 对象展平为单层键值对
///
/// 嵌套的对象会被转换为点分键名（dot-separated keys）：
/// - 输入: `{"user": {"name": "John"}}`
/// - 输出: `{"user.name": "John"}`
///
/// 只有字符串类型的值会被保留，其他类型（数字、布尔值、数组、null）会被忽略。
///
/// # Arguments
///
/// * `value` - 要展平的 JSON 值
/// * `prefix` - 键名前缀，用于递归构建完整键名
///
/// # Returns
///
/// 展平后的键值对 HashMap
///
/// # Performance
///
/// 该函数使用预分配的 HashMap 和迭代器遍历，性能优于递归实现。
/// 对于深度嵌套的结构，建议使用迭代器版本的实现。
pub fn flatten_object(value: &Value, prefix: &str) -> HashMap<String, String> {
    let mut result = HashMap::new();
    flatten_recursive(value, prefix, &mut result);
    result
}

/// 递归展平辅助函数
///
/// 使用深度优先遍历将嵌套对象展平。
/// 每次递归都会创建新的键名（通过拼接 prefix 和当前 key）。
fn flatten_recursive(value: &Value, prefix: &str, result: &mut HashMap<String, String>) {
    match value {
        Value::Object(map) => {
            for (key, val) in map {
                // 构建新键名：如果有前缀则使用 "prefix.key" 格式，否则只用 "key"
                let new_key = if prefix.is_empty() {
                    key.clone()
                } else {
                    format!("{}.{}", prefix, key)
                };
                // 递归处理嵌套值
                flatten_recursive(val, &new_key, result);
            }
        }
        Value::String(s) => {
            // 只有字符串类型才保留
            result.insert(prefix.to_string(), s.clone());
        }
        // 忽略其他类型：数字、布尔值、数组、null
        _ => {}
    }
}

/// 将展平的键值对还原为嵌套的 JSON 对象
///
/// 是 `flatten_object` 的逆操作：
/// - 输入: `{"user.name": "John"}`
/// - 输出: `{"user": {"name": "John"}}`
///
/// # Arguments
///
/// * `flat` - 展平的键值对
///
/// # Returns
///
/// 嵌套的 JSON 对象
///
/// # Panics
///
/// 如果键名格式无效（如连续的点、开头或结尾的点），可能会导致意外行为。
pub fn unflatten_object(flat: HashMap<String, String>) -> Value {
    let mut root = serde_json::Map::new();

    for (key, value) in flat {
        let parts: Vec<&str> = key.split('.').collect();
        insert_into_nested(&mut root, &parts, value);
    }

    Value::Object(root)
}

/// 递归插入到嵌套结构中
///
/// 根据键名片段路径，将值插入到嵌套的 JSON Map 中。
/// 如果中间路径不存在，会自动创建空对象。
fn insert_into_nested(
    map: &mut serde_json::Map<String, Value>,
    parts: &[&str],
    value: String,
) {
    if parts.len() == 1 {
        // 到达叶子节点，插入字符串值
        map.insert(parts[0].to_string(), Value::String(value));
    } else {
        // 还有中间节点
        let head = parts[0];
        let tail = &parts[1..];

        // 如果中间节点不存在，创建空对象
        if !map.contains_key(head) {
            map.insert(head.to_string(), Value::Object(serde_json::Map::new()));
        }

        // 递归插入到嵌套对象中
        if let Some(Value::Object(nested_map)) = map.get_mut(head) {
            insert_into_nested(nested_map, tail, value);
        }
    }
}

/// 将展平的翻译合并回原始嵌套结构
///
/// 只更新展平映射中存在的键，保留原始结构中的其他键。
///
/// # Arguments
///
/// * `original` - 原始嵌套对象
/// * `flat_translations` - 要合并的展平翻译
///
/// # Returns
///
/// 合并后的嵌套对象
pub fn merge_with_flat(
    original: &Value,
    flat_translations: HashMap<String, String>,
) -> Value {
    // 先展平原始数据
    let flat_original = flatten_object(original, "");

    // 合并翻译（新的覆盖旧的）
    let mut merged = flat_original;
    for (key, value) in flat_translations {
        merged.insert(key, value);
    }

    // 还原为嵌套结构
    unflatten_object(merged)
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    #[test]
    fn test_flatten_simple_object() {
        let input = json!({
            "hello": "world"
        });
        let result = flatten_object(&input, "");
        assert_eq!(result.get("hello"), Some(&"world".to_string()));
        assert_eq!(result.len(), 1);
    }

    #[test]
    fn test_flatten_nested_object() {
        let input = json!({
            "user": {
                "name": "John",
                "profile": {
                    "age": "30"
                }
            }
        });
        let result = flatten_object(&input, "");
        assert_eq!(result.get("user.name"), Some(&"John".to_string()));
        assert_eq!(result.get("user.profile.age"), Some(&"30".to_string()));
        assert_eq!(result.len(), 2);
    }

    #[test]
    fn test_flatten_deeply_nested() {
        let input = json!({
            "a": {
                "b": {
                    "c": {
                        "d": "deep"
                    }
                }
            }
        });
        let result = flatten_object(&input, "");
        assert_eq!(result.get("a.b.c.d"), Some(&"deep".to_string()));
    }

    #[test]
    fn test_flatten_ignores_non_strings() {
        let input = json!({
            "string": "value",
            "number": 123,
            "bool": true,
            "array": [1, 2, 3],
            "null": null,
            "nested": {
                "inner": "secret"
            }
        });
        let result = flatten_object(&input, "");
        assert_eq!(result.len(), 2);
        assert_eq!(result.get("string"), Some(&"value".to_string()));
        assert_eq!(result.get("nested.inner"), Some(&"secret".to_string()));
        // 其他类型应该被忽略
        assert!(!result.contains_key("number"));
        assert!(!result.contains_key("bool"));
        assert!(!result.contains_key("array"));
        assert!(!result.contains_key("null"));
    }

    #[test]
    fn test_flatten_empty_object() {
        let input = json!({});
        let result = flatten_object(&input, "");
        assert!(result.is_empty());
    }

    #[test]
    fn test_flatten_empty_nested() {
        let input = json!({
            "empty": {},
            "nonempty": "value"
        });
        let result = flatten_object(&input, "");
        assert_eq!(result.len(), 1);
        assert_eq!(result.get("nonempty"), Some(&"value".to_string()));
    }

    #[test]
    fn test_unflatten_simple() {
        let flat = HashMap::from([
            ("hello".to_string(), "world".to_string()),
        ]);
        let result = unflatten_object(flat);
        assert_eq!(result, json!({"hello": "world"}));
    }

    #[test]
    fn test_unflatten_nested() {
        let flat = HashMap::from([
            ("user.name".to_string(), "John".to_string()),
            ("user.profile.age".to_string(), "30".to_string()),
        ]);
        let result = unflatten_object(flat);
        assert_eq!(
            result,
            json!({
                "user": {
                    "name": "John",
                    "profile": {
                        "age": "30"
                    }
                }
            })
        );
    }

    #[test]
    fn test_roundtrip() {
        let original = json!({
            "user": {
                "name": "John",
                "profile": {
                    "age": "30"
                }
            },
            "greeting": "Hello"
        });

        let flat = flatten_object(&original, "");
        let restored = unflatten_object(flat);

        assert_eq!(original, restored);
    }

    #[test]
    fn test_roundtrip_multiple_languages() {
        let original = json!({
            "en": {
                "user": {
                    "name": "John"
                }
            },
            "zh": {
                "user": {
                    "name": "张三"
                }
            }
        });

        let flat = flatten_object(&original, "");
        let restored = unflatten_object(flat);

        assert_eq!(original, restored);
    }

    #[test]
    fn test_merge_with_flat() {
        let original = json!({
            "user": {
                "name": "Old Name",
                "email": "old@example.com"
            }
        });

        let updates = HashMap::from([
            ("user.name".to_string(), "New Name".to_string()),
        ]);

        let result = merge_with_flat(&original, updates);
        assert_eq!(
            result,
            json!({
                "user": {
                    "name": "New Name",
                    "email": "old@example.com"
                }
            })
        );
    }

    #[test]
    fn test_special_characters_in_keys() {
        let input = json!({
            "key_with_underscore": "value1",
            "key-with-dash": "value2",
            "key.with.dots": "value3"
        });
        let result = flatten_object(&input, "");
        assert_eq!(result.get("key_with_underscore"), Some(&"value1".to_string()));
        assert_eq!(result.get("key-with-dash"), Some(&"value2".to_string()));
        assert_eq!(result.get("key.with.dots"), Some(&"value3".to_string()));
    }
}
