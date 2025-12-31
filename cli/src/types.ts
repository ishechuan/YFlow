/**
 * yflow CLI 类型定义
 */

// 配置文件结构
export interface I18nConfig {
  messagesDir: string;
  projectId: number;
  apiUrl: string;
  apiKey: string;
  // 语言代码映射：将本地语言代码映射到后端语言代码
  // 例如: { "zh_CN": "zh", "zh_TW": "tw" }
  languageMapping?: Record<string, string>;
}

// 翻译数据格式：语言代码 -> 键值对
export type Translations = Record<string, Record<string, string>>;

// 翻译键的详细信息
export interface TranslationKey {
  key: string;
  path: string[]; // 用于追踪嵌套路径
  value: string;
}

// 扫描结果
export interface ScanResult {
  translations: Translations; // 按语言分组的翻译
  files: string[]; // 扫描的文件列表
  keyCount: number; // 总键数
}

// 导入结果
export interface ImportResult {
  added: number;
  updated: number;
  failed: number;
  errors: string[];
}

// 同步结果
export interface SyncResult {
  downloaded: number;
  written: number;
  skipped: number;
  errors: string[];
}

// API 错误详情
export interface APIError {
  code: string;
  message: string;
  details?: string;
}

// API 响应类型
export interface APIResponse<T> {
  success: boolean;
  data?: T | null;
  message?: string;
  error?: APIError;
}

// 后端推送键响应
export interface PushKeysResponse {
  added: string[];
  existed: string[];
  failed: string[];
}
