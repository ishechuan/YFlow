/**
 * yflow CLI 后端 API 客户端
 */

import type { Translations, APIResponse, PushKeysResponse } from "./types.js";

export interface APIClientOptions {
  baseUrl: string;
  apiKey: string;
  projectId: number;
}

/**
 * API 客户端类
 */
export class APIClient {
  private baseUrl: string;
  private apiKey: string;
  private projectId: number;

  constructor(options: APIClientOptions) {
    this.baseUrl = options.baseUrl.replace(/\/$/, ""); // 移除末尾斜杠
    this.apiKey = options.apiKey;
    this.projectId = options.projectId;
  }

  /**
   * 获取认证状态
   */
  async checkAuth(): Promise<boolean> {
    try {
      const response = await this.request<{ status: string }>("/cli/auth", {
        method: "GET",
      });
      return response.success && response.data?.status === "ok";
    } catch {
      return false;
    }
  }

  /**
   * 获取所有翻译数据
   */
  async getTranslations(): Promise<Translations> {
    const response = await this.request<Record<string, Record<string, string>>>(
      `/cli/translations?project_id=${this.projectId}`,
      {
        method: "GET",
      }
    );

    if (!response.success || !response.data) {
      throw new Error(response.error || "获取翻译数据失败");
    }

    return response.data;
  }

  /**
   * 获取指定语言的翻译数据
   */
  async getTranslationsByLocale(locale: string): Promise<Record<string, string>> {
    const response = await this.request<Record<string, Record<string, string>>>(
      `/cli/translations?project_id=${this.projectId}&locale=${locale}`,
      {
        method: "GET",
      }
    );

    if (!response.success || !response.data) {
      throw new Error(response.error || "获取翻译数据失败");
    }

    return response.data;
  }

  /**
   * 批量导入/更新翻译
   * 当 translations 不为空时，执行批量导入/更新
   */
  async pushTranslations(translations: Translations): Promise<PushKeysResponse> {
    const response = await this.request<PushKeysResponse>("/cli/keys", {
      method: "POST",
      body: {
        project_id: String(this.projectId),
        keys: [], // 空数组表示只使用 translations 字段
        translations,
      },
    });

    if (!response.success) {
      const errorMsg = response.error?.message || response.error || "未知错误";
      throw new Error(errorMsg);
    }

    // 处理 data 为 null 的情况
    if (response.data === null || response.data === undefined) {
      // 如果成功但 data 为 null，返回空的默认响应
      return { added: [], existed: [], failed: [] };
    }

    return response.data;
  }

  /**
   * 推送翻译键（如果不存在则创建）
   */
  async pushKeys(keys: string[], translations: Translations): Promise<PushKeysResponse> {
    const response = await this.request<PushKeysResponse>("/cli/keys", {
      method: "POST",
      body: {
        project_id: String(this.projectId),
        keys,
        translations,
      },
    });

    if (!response.success || !response.data) {
      throw new Error(response.error || "推送翻译键失败");
    }

    return response.data;
  }

  /**
 * 发送 API 请求
 */
  private async request<T>(
    path: string,
    options: {
      method: string;
      body?: Record<string, unknown>;
    }
  ): Promise<APIResponse<T>> {
    const url = `${this.baseUrl}${path}`;

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      "X-API-Key": this.apiKey,
    };

    const fetchOptions: RequestInit = {
      method: options.method,
      headers,
    };

    if (options.body) {
      fetchOptions.body = JSON.stringify(options.body);
    }

    const response = await fetch(url, fetchOptions);

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`HTTP ${response.status}: ${errorText}`);
    }

    return response.json() as Promise<APIResponse<T>>;
  }

  /**
   * 获取项目 ID
   */
  getProjectId(): number {
    return this.projectId;
  }

  /**
   * 获取基础 URL
   */
  getBaseUrl(): string {
    return this.baseUrl;
  }
}

/**
 * 创建 API 客户端
 */
export function createAPIClient(options: APIClientOptions): APIClient {
  return new APIClient(options);
}
