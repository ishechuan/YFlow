/**
 * yflow CLI 入口文件
 */

import { existsSync } from "fs";
import { runImport } from "./commands/import.js";
import { runSync } from "./commands/sync.js";
import { createSampleConfig, getDefaultConfigPath } from "./config.js";

const PROGRAM_NAME = "yflow";

/**
 * 显示帮助信息
 */
function showHelp(): void {
  console.log(`
${PROGRAM_NAME} - yflow 国际化管理 CLI 工具

用法:
  ${PROGRAM_NAME} <命令> [选项]

命令:
  import    将前端 messages 目录的翻译导入到后端数据库
  sync      从后端同步翻译到前端 messages 目录
  init      创建示例配置文件
  help      显示帮助信息

选项:
  --config <path>    配置文件路径 (默认: .i18nrc.json)
  --dry-run          模拟运行，不实际执行修改
  --force            强制覆盖所有翻译 (sync 命令)
  --help, -h         显示帮助信息

示例:
  ${PROGRAM_NAME} import                    # 导入翻译
  ${PROGRAM_NAME} import --dry-run          # 模拟导入
  ${PROGRAM_NAME} sync                      # 同步翻译
  ${PROGRAM_NAME} sync --force              # 强制同步
  ${PROGRAM_NAME} init                      # 创建配置文件
`);
}

/**
 * 显示版本信息
 */
function showVersion(): void {
  console.log(`${PROGRAM_NAME} v1.0.0`);
}

/**
 * 创建示例配置文件
 */
async function initConfig(): Promise<void> {
  const configPath = getDefaultConfigPath();

  if (existsSync(configPath)) {
    console.log(`⚠️  配置文件已存在: ${configPath}`);
    console.log("   如需重新创建，请先删除现有文件。");
    return;
  }

  const sampleConfig = createSampleConfig();

  try {
    // 使用 Bun 写入文件
    await Bun.write(configPath, sampleConfig);
    console.log(`✅ 已创建示例配置文件: ${configPath}`);
    console.log("\n请编辑配置文件，设置正确的项目 ID 和 API 密钥。");
  } catch (error) {
    throw new Error(`创建配置文件失败: ${error}`);
  }
}

/**
 * 解析命令行参数
 */
interface ParsedArgs {
  command: string;
  config?: string;
  dryRun: boolean;
  force: boolean;
  help: boolean;
  version: boolean;
}

function parseArgs(): ParsedArgs {
  const result: ParsedArgs = {
    command: "",
    dryRun: false,
    force: false,
    help: false,
    version: false,
  };

  const rawArgs = Bun.argv.slice(2);

  for (let i = 0; i < rawArgs.length; i++) {
    const arg = rawArgs[i];
    if (!arg) continue;

    if (arg.startsWith("--")) {
      // 长选项
      switch (arg.toLowerCase()) {
        case "--help":
        case "-h":
          result.help = true;
          break;
        case "--version":
        case "-v":
          result.version = true;
          break;
        case "--dry-run":
          result.dryRun = true;
          break;
        case "--force":
          result.force = true;
          break;
        case "--config":
          if (i + 1 < rawArgs.length) {
            result.config = rawArgs[i + 1];
            i++;
          }
          break;
        default:
          console.log(`⚠️  未知选项: ${arg}`);
      }
    } else if (!arg?.startsWith("-")) {
      // 命令
      if (!result.command) {
        result.command = arg.toLowerCase();
      }
    }
  }

  return result;
}

/**
 * 主函数
 */
async function main(): Promise<void> {
  const parsedArgs = parseArgs();

  // 显示帮助或版本
  if (parsedArgs.help) {
    showHelp();
    return;
  }

  if (parsedArgs.version) {
    showVersion();
    return;
  }

  // 执行命令
  try {
    switch (parsedArgs.command) {
      case "import":
        await runImport({
          configPath: parsedArgs.config,
          dryRun: parsedArgs.dryRun,
        });
        break;

      case "sync":
        await runSync({
          configPath: parsedArgs.config,
          dryRun: parsedArgs.dryRun,
          force: parsedArgs.force,
        });
        break;

      case "init":
        await initConfig();
        break;

      case "":
        showHelp();
        break;

      default:
        console.log(`⚠️  未知命令: ${parsedArgs.command}`);
        console.log(`   运行 '${PROGRAM_NAME} --help' 查看可用命令。`);
        process.exit(1);
    }
  } catch (error) {
    console.error(`\n❌ 错误: ${error}`);

    if (error instanceof Error && error.message.includes("配置文件")) {
      console.log("\n💡 提示: 运行 'yflow init' 创建示例配置文件。");
    }

    process.exit(1);
  }
}

// 执行主函数
main();
