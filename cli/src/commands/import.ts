/**
 * yflow CLI import 命令
 * 将前端 messages 目录的翻译导入到后端数据库
 */

import { loadConfig } from "../config.js";
import { createAPIClient } from "../api.js";
import { scanMessagesDir } from "../scanner.js";
import { createLanguageMapper } from "../language-mapping.js";
import { showSpinner, stopSpinner, createMultiProgressBar, shouldShowProgress, safeStopProgress } from "../ui.js";
import type { ImportResult, Translations } from "../types.js";

export interface ImportOptions {
  configPath?: string;
  dryRun?: boolean;
}

// 每批导入的键数量限制
const BATCH_SIZE = 50;

// 批次间延迟（毫秒）- 避免速率限制（批量操作限流每秒20个请求）
const BATCH_DELAY = 200;

// 最大重试次数
const MAX_RETRIES = 3;

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * 执行导入命令
 */
export async function runImport(options: ImportOptions = {}): Promise<ImportResult> {
  console.log("🔄 正在导入翻译到后端...\n");

  // 1. 加载配置
  const useProgress = shouldShowProgress() && !options.dryRun;
  console.log("📖 加载配置文件...");
  const config = loadConfig(options.configPath);
  console.log(`   - messages 目录: ${config.messagesDir}`);
  console.log(`   - 项目 ID: ${config.projectId}`);
  console.log(`   - API 地址: ${config.apiUrl}`);

  // 1.1 初始化语言映射
  const languageMapper = createLanguageMapper(config.languageMapping);
  if (languageMapper.needsMapping()) {
    console.log(`   - ${languageMapper.getDescription()}`);
  }
  console.log();

  // 2. 创建 API 客户端
  const api = createAPIClient({
    baseUrl: config.apiUrl,
    apiKey: config.apiKey,
    projectId: config.projectId,
  });

  // 3. 检查认证
  if (useProgress) {
    showSpinner("验证 API 认证...");
  } else {
    console.log("🔐 验证 API 认证...");
  }
  const isAuthenticated = await api.checkAuth();
  if (useProgress) {
    stopSpinner(isAuthenticated, "验证 API 认证");
  } else {
    console.log("   - 认证成功\n");
  }
  if (!isAuthenticated) {
    throw new Error("API 认证失败，请检查 apiKey 是否正确");
  }

  // 4. 扫描 messages 目录
  if (useProgress) {
    showSpinner("扫描 messages 目录...");
  } else {
    console.log("📂 扫描 messages 目录...");
  }
  const scanResult = await scanMessagesDir(config.messagesDir);
  if (useProgress) {
    stopSpinner(true);
  }
  console.log(`   - 扫描文件数: ${scanResult.files.length}`);
  console.log(`   - 翻译键数: ${scanResult.keyCount}`);
  console.log(`   - 语言: ${Object.keys(scanResult.translations).join(", ")}\n`);

  if (scanResult.keyCount === 0) {
    console.log("⚠️  没有找到翻译内容，跳过导入\n");
    return { added: 0, updated: 0, failed: 0, errors: [] };
  }

  // 5. 执行导入
  if (options.dryRun) {
    console.log("🧪 模拟运行 (dry-run)，不实际导入\n");

    // 应用语言映射
    const mappedTranslations = languageMapper.applyToTranslations(scanResult.translations);

    console.log("   将要导入的翻译预览:");
    for (const [lang, translations] of Object.entries(mappedTranslations)) {
      const keys = Object.keys(translations);
      console.log(`   - ${lang}: ${keys.length} 个键`);
      if (keys.length <= 5) {
        for (const key of keys) {
          console.log(`     - ${key}: "${translations[key]}"`);
        }
      } else {
        console.log(`     - 前5个键: ${keys.slice(0, 5).join(", ")}...`);
      }
    }
    console.log();
    return { added: scanResult.keyCount, updated: 0, failed: 0, errors: [] };
  }

  if (useProgress) {
    showSpinner("正在上传翻译到后端...");
  } else {
    console.log("📤 正在上传翻译到后端...");
  }

  // 应用语言映射
  const mappedTranslations = languageMapper.applyToTranslations(scanResult.translations);

  // 停止 spinner，显示进度条
  if (useProgress) {
    stopSpinner(true);
  }

  // 分批导入翻译
  const result = await importTranslationsInBatches(api, mappedTranslations, useProgress);

  // 6. 输出结果
  console.log("\n✅ 导入完成!");
  console.log(`   - 新增: ${result.added}`);
  console.log(`   - 更新: ${result.updated}`);
  console.log(`   - 失败: ${result.failed}`);

  if (result.errors.length > 0) {
    console.log("\n❌ 错误详情:");
    for (const error of result.errors) {
      console.log(`   - ${error}`);
    }
  }

  return result;
}

/**
 * 将翻译按批次大小拆分
 */
function chunkTranslations(
  translations: Record<string, string>,
  chunkSize: number
): Record<string, Record<string, string>>[] {
  const entries = Object.entries(translations);
  const chunks: Record<string, Record<string, string>>[] = [];

  for (let i = 0; i < entries.length; i += chunkSize) {
    const chunk: Record<string, Record<string, string>> = {};
    const batch = entries.slice(i, i + chunkSize);

    for (const [key, value] of batch) {
      chunk[key] = value;
    }

    chunks.push(chunk);
  }

  return chunks;
}

/**
 * 检查是否为速率限制错误
 */
function isRateLimitError(error: unknown): boolean {
  if (error instanceof Error && error.message.includes("429")) {
    return true;
  }
  return false;
}

/**
 * 安全获取数组长度（处理 null/undefined）
 */
function getArrayLength(arr: unknown[] | null | undefined): number {
  return Array.isArray(arr) ? arr.length : 0;
}

/**
 * 分批导入翻译
 */
async function importTranslationsInBatches(
  api: ReturnType<typeof createAPIClient>,
  translations: Translations,
  useProgress: boolean = false
): Promise<ImportResult> {
  let added = 0;
  let updated = 0;
  let failed = 0;
  const errors: string[] = [];

  // 创建进度条管理器
  const multiBar = useProgress ? createMultiProgressBar() : null;

  for (const [langCode, langTranslations] of Object.entries(translations)) {
    const totalKeys = Object.keys(langTranslations).length;
    if (totalKeys === 0) {
      continue;
    }

    if (useProgress) {
      console.log(`   - 正在导入 ${langCode} (${totalKeys} 键)...`);
    } else {
      console.log(`   - 正在导入 ${langCode} (${totalKeys} 键)...`);
    }

    // 拆分该语言的翻译为多个批次
    const chunks = chunkTranslations(langTranslations, BATCH_SIZE);
    let langAdded = 0;
    let langUpdated = 0;
    let langFailed = 0;

    // 为该语言创建进度条
    const langBar = multiBar?.getOrCreateBar(langCode, totalKeys);

    for (let i = 0; i < chunks.length; i++) {
      const chunk = chunks[i];
      const batchNum = i + 1;
      const isLastBatch = batchNum === chunks.length;

      let retryCount = 0;
      let success = false;

      while (!success && retryCount < MAX_RETRIES) {
        try {
          const result = await api.pushTranslations({
            [langCode]: chunk,
          });

          // 安全获取数组长度
          const addedCount = getArrayLength(result.added);
          const existedCount = getArrayLength(result.existed);
          const failedCount = getArrayLength(result.failed);

          // 统计结果
          langAdded += addedCount;
          langUpdated += existedCount;
          langFailed += failedCount;

          if (failedCount > 0) {
            const failedKeys = Array.isArray(result.failed) ? result.failed : [];
            errors.push(`${langCode}[${batchNum}]: 失败的键 - ${failedKeys.join(", ")}`);
          }

          // 更新进度条
          const processedKeys = langAdded + langUpdated + langFailed;
          if (langBar) {
            multiBar?.update(langCode, processedKeys, totalKeys);
          } else if (chunks.length > 1) {
            console.log(`     批次 ${batchNum}/${chunks.length}: +${addedCount}, ~${existedCount}`);
          } else {
            console.log(`     ✓ ${langCode}: +${addedCount}, ~${existedCount}`);
          }

          success = true;
        } catch (error) {
          if (isRateLimitError(error) && retryCount < MAX_RETRIES - 1) {
            retryCount++;
            const waitTime = BATCH_DELAY * retryCount * 2;
            if (!useProgress) {
              console.log(`     ⚠ 速率限制，等待 ${waitTime}ms 后重试 (${retryCount}/${MAX_RETRIES})`);
            }
            await sleep(waitTime);
          } else {
            langFailed += Object.keys(chunk).length;
            errors.push(`${langCode}[${batchNum}]: ${error}`);
            if (!useProgress) {
              console.log(`     ✗ 批次 ${batchNum}/${chunks.length}: 失败 - ${error}`);
            }
            success = true; // 即使失败也继续下一个批次
          }
        }
      }

      if (!isLastBatch) {
        await sleep(BATCH_DELAY);
      }
    }

    // 完成该语言的进度条
    const processedKeys = langAdded + langUpdated + langFailed;
    if (langBar) {
      multiBar?.complete(langCode, processedKeys, totalKeys);
    } else if (chunks.length > 1) {
      console.log(`     ✓ ${langCode} 完成: +${langAdded}, ~${langUpdated}, ✗${langFailed}`);
    }

    added += langAdded;
    updated += langUpdated;
    failed += langFailed;
  }

  // 停止进度条
  if (multiBar) {
    multiBar.stop();
  }

  return { added, updated, failed, errors };
}
