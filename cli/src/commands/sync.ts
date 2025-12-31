/**
 * yflow CLI sync 命令
 * 从后端同步翻译到前端 messages 目录
 */

import { loadConfig } from "../config.js";
import { createAPIClient } from "../api.js";
import { scanMessagesDir, writeTranslationsWithStructure } from "../scanner.js";
import { createLanguageMapper } from "../language-mapping.js";
import { showSpinner, stopSpinner, createMultiProgressBar, shouldShowProgress } from "../ui.js";
import type { SyncResult, Translations } from "../types.js";

export interface SyncOptions {
  configPath?: string;
  dryRun?: boolean;
  force?: boolean; // 是否强制覆盖所有翻译
}

/**
 * 执行同步命令
 */
export async function runSync(options: SyncOptions = {}): Promise<SyncResult> {
  console.log("🔄 正在从后端同步翻译...\n");

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

  // 4. 从后端获取翻译
  if (useProgress) {
    showSpinner("从后端获取翻译...");
  } else {
    console.log("📥 正在从后端获取翻译...");
  }
  let backendTranslations: Translations;
  try {
    backendTranslations = await api.getTranslations();
  } catch (error) {
    throw new Error(`获取翻译失败: ${error}`);
  }
  if (useProgress) {
    stopSpinner(true);
  }

  // 4.1 应用反向语言映射（后端代码 -> 本地代码）
  const localTranslations = languageMapper.reverseTranslations(backendTranslations);

  const totalKeys = Object.values(localTranslations).reduce(
    (sum, lang) => sum + Object.keys(lang).length,
    0
  );
  console.log(`   - 获取翻译键数: ${totalKeys}`);
  console.log(`   - 语言: ${Object.keys(localTranslations).join(", ")}\n`);

  if (totalKeys === 0) {
    console.log("⚠️  后端没有翻译内容，跳过同步\n");
    return { downloaded: 0, written: 0, skipped: 0, errors: [] };
  }

  // 5. 扫描本地 messages 目录（获取原始文件结构）
  if (useProgress) {
    showSpinner("扫描本地 messages 目录...");
  } else {
    console.log("📂 扫描本地 messages 目录...");
  }
  let localScanResult;
  try {
    localScanResult = await scanMessagesDir(config.messagesDir);
  } catch {
    // 如果目录不存在，创建一个空的结构
    localScanResult = {
      translations: {},
      files: [],
      keyCount: 0,
    };
  }
  if (useProgress) {
    stopSpinner(true);
  }
  console.log(`   - 本地文件数: ${localScanResult.files.length}`);
  console.log(`   - 本地键数: ${localScanResult.keyCount}\n`);

  // 6. 计算差异并同步
  if (options.dryRun) {
    console.log("🧪 模拟运行 (dry-run)，不实际写入文件\n");
    return showSyncDiff(localTranslations, localScanResult.translations);
  }

  // 创建进度条
  const totalLanguages = Object.keys(localTranslations).length;
  const multiBar = useProgress ? createMultiProgressBar() : null;

  // 写入翻译
  if (useProgress) {
    showSpinner(`写入翻译到本地文件 (0/${totalLanguages} 语言)...`);
  } else {
    console.log("📝 正在写入翻译到本地文件...");
  }

  const writtenFiles = await writeTranslationsWithStructure(
    config.messagesDir,
    localScanResult.files,
    localTranslations,
    // 进度回调
    (currentLang, langIndex) => {
      if (multiBar) {
        multiBar.update(currentLang, langIndex + 1, totalLanguages);
      }
      if (useProgress) {
        showSpinner(`写入翻译到本地文件 (${langIndex + 1}/${totalLanguages} 语言)...`);
      }
    }
  );

  // 停止 spinner
  if (useProgress) {
    stopSpinner(true, `写入翻译到本地文件 (${totalLanguages}/${totalLanguages})`);
  }

  // 计算统计
  let downloaded = 0;
  let skipped = 0;
  const errors: string[] = [];

  for (const [lang, translations] of Object.entries(localTranslations)) {
    const localLang = localScanResult.translations[lang] || {};
    for (const [key, value] of Object.entries(translations)) {
      if (options.force || !(key in localLang)) {
        downloaded++;
      } else {
        skipped++;
      }
    }
  }

  // 停止进度条
  if (multiBar) {
    multiBar.stop();
  }

  // 7. 输出结果
  console.log("\n✅ 同步完成!");
  console.log(`   - 已下载: ${downloaded}`);
  console.log(`   - 已跳过: ${skipped}`);
  console.log(`   - 已写入文件: ${writtenFiles.length}`);

  return { downloaded, written: writtenFiles.length, skipped, errors };
}

/**
 * 显示同步差异（dry-run 模式）
 */
function showSyncDiff(
  backend: Translations,
  local: Record<string, Record<string, string>>
): SyncResult {
  let downloaded = 0;
  let skipped = 0;

  console.log("📊 同步差异预览:\n");

  for (const [lang, translations] of Object.entries(backend)) {
    const localLang = local[lang] || {};
    const newKeys: string[] = [];
    const existingKeys: string[] = [];

    for (const key of Object.keys(translations)) {
      if (key in localLang) {
        existingKeys.push(key);
        skipped++;
      } else {
        newKeys.push(key);
        downloaded++;
      }
    }

    console.log(`  ${lang}:`);
    if (newKeys.length > 0) {
      console.log(`    新增 (${newKeys.length}): ${newKeys.slice(0, 5).join(", ")}${newKeys.length > 5 ? "..." : ""}`);
    }
    if (existingKeys.length > 0) {
      console.log(`    已存在 (${existingKeys.length}): ${existingKeys.slice(0, 3).join(", ")}${existingKeys.length > 3 ? "..." : ""}`);
    }
    console.log();
  }

  console.log("📈 统计:");
  console.log(`   - 将下载: ${downloaded}`);
  console.log(`   - 将跳过: ${skipped}`);

  return { downloaded, written: 0, skipped, errors: [] };
}
