/**
 * yflow CLI 配置文件读取模块
 */

import { readFileSync, existsSync } from "fs";
import { join, resolve } from "path";
import type { I18nConfig } from "./types.js";

const CONFIG_FILENAME = ".i18nrc.json";

/**
 * 加载配置文件
 * @param configPath - 配置文件路径，默认当前目录的 .i18nrc.json
 * @returns I18nConfig 配置对象
 */
export function loadConfig(configPath?: string): I18nConfig {
  const resolvedPath = resolve(configPath || CONFIG_FILENAME);

  if (!existsSync(resolvedPath)) {
    throw new Error(`配置文件不存在: ${resolvedPath}`);
  }

  try {
    const content = readFileSync(resolvedPath, "utf-8");
    const config = JSON.parse(content) as I18nConfig;

    // 验证必需字段
    validateConfig(config);

    // 应用环境变量覆盖
    return applyEnvOverrides(config);
  } catch (error) {
    if (error instanceof SyntaxError) {
      throw new Error(`配置文件 JSON 格式错误: ${error.message}`);
    }
    throw error;
  }
}

/**
 * 验证配置文件必需字段
 */
function validateConfig(config: I18nConfig): void {
  const errors: string[] = [];

  if (!config.messagesDir) {
    errors.push("messagesDir (messages 目录路径) 是必需字段");
  }

  if (typeof config.projectId !== "number" || config.projectId <= 0) {
    errors.push("projectId (项目 ID) 必须是正整数");
  }

  if (!config.apiUrl) {
    errors.push("apiUrl (API 地址) 是必需字段");
  }

  if (!config.apiKey) {
    errors.push("apiKey (API 密钥) 是必需字段");
  }

  if (errors.length > 0) {
    throw new Error(`配置文件验证失败:\n${errors.join("\n")}`);
  }
}

/**
 * 应用环境变量覆盖
 */
function applyEnvOverrides(config: I18nConfig): I18nConfig {
  return {
    ...config,
    messagesDir: process.env.I18N_MESSAGES_DIR || config.messagesDir,
    projectId: parseInt(process.env.I18N_PROJECT_ID || String(config.projectId), 10),
    apiUrl: process.env.I18N_API_URL || config.apiUrl,
    apiKey: process.env.I18N_API_KEY || config.apiKey,
  };
}

/**
 * 获取默认配置文件搜索路径
 */
export function getDefaultConfigPath(): string {
  return join(process.cwd(), CONFIG_FILENAME);
}

/**
 * 检查配置文件是否存在
 */
export function configExists(configPath?: string): boolean {
  const resolvedPath = resolve(configPath || CONFIG_FILENAME);
  return existsSync(resolvedPath);
}

/**
 * 创建示例配置文件
 */
export function createSampleConfig(): string {
  return JSON.stringify(
    {
      messagesDir: "./src/locales",
      projectId: 1,
      apiUrl: "http://localhost:8080/api",
      apiKey: "your-api-key-here",
    },
    null,
    2
  );
}
