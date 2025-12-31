/**
 * yflow CLI 语言映射处理模块
 */

import type { Translations } from "./types.js";

/**
 * 创建语言映射表（双向映射）
 */
export function createLanguageMapper(
  mapping?: Record<string, string>,
): LanguageMapper {
  return new LanguageMapper(mapping);
}

/**
 * 语言映射器
 */
export class LanguageMapper {
  // 本地代码 -> 后端代码
  private localToBackend: Map<string, string>;
  // 后端代码 -> 本地代码（用于同步时反向映射）
  private backendToLocal: Map<string, string>;

  constructor(mapping?: Record<string, string>) {
    this.localToBackend = new Map(Object.entries(mapping || {}));
    this.backendToLocal = new Map();

    // 构建反向映射
    for (const [local, backend] of this.localToBackend) {
      this.backendToLocal.set(backend, local);
    }
  }

  /**
   * 将本地语言代码转换为后端语言代码
   */
  toBackend(localCode: string): string {
    return this.localToBackend.get(localCode) || localCode;
  }

  /**
   * 将后端语言代码转换为本地语言代码
   */
  toLocal(backendCode: string): string {
    return this.backendToLocal.get(backendCode) || backendCode;
  }

  /**
   * 应用语言映射：将翻译数据的语言代码转换为后端代码
   */
  applyToTranslations(translations: Translations): Translations {
    const result: Translations = {};

    for (const [localCode, langData] of Object.entries(translations)) {
      const backendCode = this.toBackend(localCode);

      if (!result[backendCode]) {
        result[backendCode] = {};
      }

      // 合并翻译数据
      for (const [key, value] of Object.entries(langData)) {
        result[backendCode][key] = value;
      }
    }

    return result;
  }

  /**
   * 反向应用语言映射：将翻译数据的语言代码转换为本地代码（用于同步）
   */
  reverseTranslations(translations: Translations): Translations {
    const result: Translations = {};

    for (const [backendCode, langData] of Object.entries(translations)) {
      const localCode = this.toLocal(backendCode);

      if (!result[localCode]) {
        result[localCode] = {};
      }

      // 合并翻译数据
      for (const [key, value] of Object.entries(langData)) {
        result[localCode][key] = value;
      }
    }

    return result;
  }

  /**
   * 检查是否需要进行语言代码映射
   */
  needsMapping(): boolean {
    return this.localToBackend.size > 0;
  }

  /**
   * 获取映射描述
   */
  getDescription(): string {
    if (!this.needsMapping()) {
      return "无语言映射";
    }

    const mappings = Array.from(this.localToBackend.entries())
      .map(([local, backend]) => `${local} → ${backend}`)
      .join(", ");

    return `语言映射: ${mappings}`;
  }
}
