/**
 * yflow CLI messages 目录扫描模块
 */

import { readFileSync, existsSync, promises as fs } from "fs";
import { join, resolve, relative } from "path";
import type { Translations, ScanResult } from "./types.js";

const JSON_EXTENSIONS = [".json"];

/**
 * 扫描 messages 目录并收集所有翻译
 * @param messagesDir - messages 目录路径
 * @returns ScanResult 扫描结果
 */
export async function scanMessagesDir(messagesDir: string): Promise<ScanResult> {
  const resolvedPath = resolve(messagesDir);

  if (!existsSync(resolvedPath)) {
    throw new Error(`messages 目录不存在: ${messagesDir}`);
  }

  const translations: Translations = {};
  const files: string[] = [];
  let keyCount = 0;

  // 扫描语言子目录
  const entries = await fs.readdir(resolvedPath, { withFileTypes: true });

  for (const entry of entries) {
    if (!entry.isDirectory()) {
      continue;
    }

    const langCode = entry.name;
    const langDir = join(resolvedPath, langCode);

    // 扫描该语言目录下的 JSON 文件
    await scanLanguageDir(langCode, langDir, translations, files);
  }

  // 统计总键数
  for (const lang of Object.values(translations)) {
    keyCount += Object.keys(lang).length;
  }

  return {
    translations,
    files,
    keyCount,
  };
}

/**
 * 扫描单个语言目录
 */
async function scanLanguageDir(
  langCode: string,
  dirPath: string,
  translations: Translations,
  files: string[]
): Promise<void> {
  // 确保该语言的对象存在
  if (!translations[langCode]) {
    translations[langCode] = {};
  }

  const entries = await fs.readdir(dirPath, { withFileTypes: true });

  for (const entry of entries) {
    const fullPath = join(dirPath, entry.name);

    if (entry.isDirectory()) {
      // 递归扫描子目录
      await scanLanguageDir(langCode, fullPath, translations, files);
    } else if (isJSONFile(entry.name)) {
      // 解析 JSON 文件
      try {
        const content = readFileSync(fullPath, "utf-8");
        const jsonData = JSON.parse(content);

        // 展平并合并翻译
        const flatTranslations = flattenObject("", jsonData);

        for (const [key, value] of Object.entries(flatTranslations)) {
          if (typeof value === "string") {
            translations[langCode][key] = value;
          }
        }

        files.push(relative(process.cwd(), fullPath));
      } catch (error) {
        throw new Error(`解析文件失败 ${fullPath}: ${error}`);
      }
    }
  }
}

/**
 * 递归展平嵌套对象
 * { "a": { "b": "value" } } -> { "a.b": "value" }
 */
function flattenObject(
  prefix: string,
  obj: Record<string, unknown>
): Record<string, string> {
  const result: Record<string, string> = {};

  for (const [key, value] of Object.entries(obj)) {
    const newKey = prefix ? `${prefix}.${key}` : key;

    if (value === null || value === undefined) {
      continue;
    }

    if (typeof value === "string") {
      result[newKey] = value;
    } else if (typeof value === "object") {
      // 递归处理嵌套对象
      const nested = flattenObject(newKey, value as Record<string, unknown>);
      Object.assign(result, nested);
    }
    // 忽略其他类型（数字、数组等）
  }

  return result;
}

/**
 * 检查文件是否为 JSON 文件
 */
function isJSONFile(filename: string): boolean {
  const ext = filename.slice(filename.lastIndexOf(".")).toLowerCase();
  return JSON_EXTENSIONS.includes(ext);
}

/**
 * 将展平的翻译写回文件
 * @param messagesDir - messages 目录路径
 * @param translations - 按语言分组的翻译
 */
export async function writeTranslations(
  messagesDir: string,
  translations: Translations
): Promise<string[]> {
  const resolvedPath = resolve(messagesDir);
  const writtenFiles: string[] = [];

  if (!existsSync(resolvedPath)) {
    throw new Error(`messages 目录不存在: ${messagesDir}`);
  }

  for (const [langCode, langTranslations] of Object.entries(translations)) {
    const langDir = join(resolvedPath, langCode);

    // 确保语言目录存在
    if (!existsSync(langDir)) {
      await fs.mkdir(langDir, { recursive: true });
    }

    // 写入该语言的所有翻译到单个文件
    // 注意：这里采用简单方式，将所有翻译合并到一个文件
    // 也可以考虑保持原有的文件结构，但需要更复杂的逻辑
    const outputPath = join(langDir, "sync.json");
    await fs.writeFile(outputPath, JSON.stringify(langTranslations, null, 2), "utf-8");
    writtenFiles.push(relative(process.cwd(), outputPath));
  }

  return writtenFiles;
}

/**
 * 按原始文件结构写入翻译
 * 保持原有的目录结构，只更新已有文件中的键
 */
export async function writeTranslationsWithStructure(
  messagesDir: string,
  originalFiles: string[],
  translations: Translations,
  onProgress?: (currentLang: string, langIndex: number) => void
): Promise<string[]> {
  const resolvedPath = resolve(messagesDir);
  const writtenFiles: string[] = [];

  // 按语言分组文件
  const filesByLang = new Map<string, string[]>();
  for (const file of originalFiles) {
    const langCode = file.split("/")[0];
    if (!filesByLang.has(langCode)) {
      filesByLang.set(langCode, []);
    }
    filesByLang.get(langCode)!.push(file);
  }

  let langIndex = 0;
  for (const [langCode, files] of filesByLang) {
    for (const file of files) {
      const filePath = join(resolvedPath, file);

      if (!existsSync(filePath)) {
        continue;
      }

      try {
        const content = readFileSync(filePath, "utf-8");
        const originalData = JSON.parse(content);

        // 收集该语言的所有翻译
        const langTranslations = translations[langCode] || {};

        // 合并翻译数据，保持原始结构
        const mergedData = mergeTranslationsWithStructure(originalData, langTranslations);

        await fs.writeFile(filePath, JSON.stringify(mergedData, null, 2), "utf-8");
        writtenFiles.push(file);
      } catch (error) {
        throw new Error(`写入文件失败 ${file}: ${error}`);
      }
    }

    // 回调进度（每种语言完成后）
    langIndex++;
    onProgress?.(langCode, langIndex);
  }

  return writtenFiles;
}

/**
 * 递归获取目录下所有 JSON 文件
 */
async function getAllJSONFiles(dirPath: string): Promise<string[]> {
  const files: string[] = [];
  const entries = await fs.readdir(dirPath, { withFileTypes: true });

  for (const entry of entries) {
    const fullPath = join(dirPath, entry.name);

    if (entry.isDirectory()) {
      files.push(...(await getAllJSONFiles(fullPath)));
    } else if (isJSONFile(entry.name)) {
      files.push(relative(dirPath, fullPath));
    }
  }

  return files;
}

/**
 * 将展平的翻译合并回原始嵌套结构
 */
function mergeTranslationsWithStructure(
  original: Record<string, unknown>,
  flatTranslations: Record<string, string>
): Record<string, unknown> {
  // 首先展平原始数据
  const flatOriginal = flattenObject("", original);

  // 合并翻译（新的覆盖旧的）
  for (const [key, value] of Object.entries(flatTranslations)) {
    flatOriginal[key] = value;
  }

  // 重新组合成嵌套结构
  return unflattenObject(flatOriginal);
}

/**
 * 将展平的对象还原为嵌套结构
 * { "a.b": "value" } -> { "a": { "b": "value" } }
 */
function unflattenObject(flat: Record<string, string>): Record<string, unknown> {
  const result: Record<string, unknown> = {};

  for (const [key, value] of Object.entries(flat)) {
    const parts = key.split(".");
    let current = result;

    for (let i = 0; i < parts.length - 1; i++) {
      const part = parts[i];

      if (!(part in current)) {
        current[part] = {};
      }

      current = current as Record<string, unknown>;
    }

    current[parts[parts.length - 1]] = value;
  }

  return result;
}
