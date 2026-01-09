//! Core modules for YFlow CLI
//!
//! This module contains the core business logic components including
//! configuration management, file scanning, JSON flattening, and language mapping.

#![allow(dead_code)]

pub mod config;
pub mod scanner;
pub mod flatten;
pub mod language_mapping;

pub use flatten::{flatten_object, unflatten_object};

use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::path::PathBuf;

/// 配置文件结构
///
/// 对应原 TypeScript 的 I18nConfig 接口
#[derive(Debug, Clone, PartialEq, Eq, Deserialize, Serialize)]
pub struct I18nConfig {
    /// messages 目录路径
    #[serde(rename = "messagesDir")]
    pub messages_dir: PathBuf,
    /// 项目 ID
    #[serde(rename = "projectId")]
    pub project_id: u64,
    /// API 地址
    #[serde(rename = "apiUrl")]
    pub api_url: String,
    /// API 密钥
    #[serde(rename = "apiKey")]
    pub api_key: String,
    /// 语言代码映射
    #[serde(rename = "languageMapping", default)]
    pub language_mapping: HashMap<String, String>,
}

/// 翻译数据格式：语言代码 -> 键值对
pub type Translations = HashMap<String, HashMap<String, String>>;

/// 扫描结果
#[derive(Debug, Clone)]
pub struct ScanResult {
    /// 按语言分组的翻译
    pub translations: Translations,
    /// 扫描的文件列表
    pub files: Vec<PathBuf>,
    /// 总键数
    pub key_count: usize,
}

/// 导入结果
#[derive(Debug, Clone, Default)]
pub struct ImportResult {
    /// 新增的键数
    pub added: usize,
    /// 更新的键数
    pub updated: usize,
    /// 失败的键数
    pub failed: usize,
    /// 错误列表
    pub errors: Vec<String>,
}

/// 同步结果
#[derive(Debug, Clone, Default)]
pub struct SyncResult {
    /// 下载的键数
    pub downloaded: usize,
    /// 写入的文件数
    pub written: usize,
    /// 跳过的键数
    pub skipped: usize,
    /// 错误列表
    pub errors: Vec<String>,
}
