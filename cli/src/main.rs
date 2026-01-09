//! YFlow CLI - Main entry point
//!
//! A CLI tool for importing and syncing translations between
//! local files and the YFlow backend.

mod cli;
mod core;
mod api;
mod ui;

use anyhow::Result;
use clap::Parser;
use cli::{CliArgs, Commands};
use core::config::create_sample_config;
use std::path::PathBuf;
use std::process;
use tracing::info;
use tracing_subscriber;

/// 程序名称
const PROGRAM_NAME: &str = "yflow";

/// 程序版本（从 Cargo.toml 自动获取）
const VERSION: &str = env!("CARGO_PKG_VERSION");

/// 构建信息
const BUILD_INFO: &str = concat!(env!("CARGO_PKG_VERSION"), " (build)");

#[tokio::main]
async fn main() -> Result<()> {
    // 初始化日志
    tracing_subscriber::fmt()
        .with_max_level(tracing::Level::INFO)
        .with_target(false)
        .init();

    // 解析命令行参数
    let args = CliArgs::parse();

    // 如果启用了 verbose 模式，启用更详细的日志
    if args.verbose {
        tracing_subscriber::fmt()
            .with_max_level(tracing::Level::DEBUG)
            .with_target(true)
            .init();
    }

    // 执行命令
    let result: Result<()> = match &args.command {
        Commands::Import(cmd) => cmd.run(args.config.clone()).await.map(|_| ()),
        Commands::Sync(cmd) => cmd.run(args.config.clone()).await.map(|_| ()),
        Commands::Init { output } => {
            init_config(output.as_ref())?;
            Ok(())
        }
        Commands::Version => {
            show_version();
            Ok(())
        }
        Commands::HelpCmd { command } => {
            show_help(command.as_deref());
            Ok(())
        }
    };

    // 处理错误
    match result {
        Ok(_) => {
            info!("Done.");
            Ok(())
        }
        Err(e) => {
            eprintln!("\n❌ Error: {}", e);

            // 检查是否是配置文件错误
            if e.to_string().contains("Config") {
                println!("\n💡 Hint: Run 'yflow init' to create a sample configuration file.");
            }

            process::exit(1);
        }
    }
}

/// 显示版本信息
///
/// 输出程序名称、版本号和构建信息。
fn show_version() {
    println!("{} v{}", PROGRAM_NAME, VERSION);
    println!("Build: {}", BUILD_INFO);
    println!();
    println!("A CLI tool for importing and syncing translations between");
    println!("local files and the YFlow backend.");
}

/// 显示帮助信息
///
/// 显示全局帮助信息或特定命令的详细帮助。
///
/// # Arguments
///
/// * `command` - 可选的命令名称，如果提供则显示该命令的详细帮助
fn show_help(command: Option<&str>) {
    if let Some(cmd_name) = command {
        // 显示特定命令的帮助信息
        show_command_help(cmd_name);
    } else {
        // 显示全局帮助信息
        println!(
            r#"{PROGRAM_NAME} - YFlow Internationalization Management CLI Tool

Usage:
  {PROGRAM_NAME} <command> [options]

Commands:
  import    Import translations from local messages directory to backend
  sync      Sync translations from backend to local messages directory
  init      Create a sample configuration file
  version   Display version information
  help      Show this help message or help for a specific command

Options:
  --config <path>    Configuration file path (default: .i18nrc.json)
  --dry-run          Simulate execution without making changes
  --force            Force overwrite all translations (sync command)
  --help, -h         Show help information
  --version, -v      Show version information
  --verbose, -v      Enable verbose output

Examples:
  {PROGRAM_NAME} import                    # Import translations
  {PROGRAM_NAME} import --dry-run          # Simulate import
  {PROGRAM_NAME} sync                      # Sync translations
  {PROGRAM_NAME} sync --force              # Force sync
  {PROGRAM_NAME} init                      # Create configuration file
  {PROGRAM_NAME} help import               # Show help for import command
  {PROGRAM_NAME} version                   # Show version information
"#
        );
    }
}

/// 显示特定命令的帮助信息
///
/// # Arguments
///
/// * `command` - 命令名称
fn show_command_help(command: &str) {
    match command.to_lowercase().as_str() {
        "import" => {
            println!(
                r#"Import translations from local messages directory to backend

Usage: {PROGRAM_NAME} import [options]

Options:
  --config <path>    Configuration file path (default: .i18nrc.json)
  --dry-run          Simulate import without making changes
  --help, -h         Show this help message

Examples:
  {PROGRAM_NAME} import                    # Import translations
  {PROGRAM_NAME} import --dry-run          # Preview what would be imported
  {PROGRAM_NAME} import --config .i18nrc   # Use custom config file
"#
            );
        }
        "sync" => {
            println!(
                r#"Sync translations from backend to local messages directory

Usage: {PROGRAM_NAME} sync [options]

Options:
  --config <path>    Configuration file path (default: .i18nrc.json)
  --dry-run          Simulate sync without making changes
  --force            Force overwrite all existing translations
  --help, -h         Show this help message

Examples:
  {PROGRAM_NAME} sync                      # Sync translations
  {PROGRAM_NAME} sync --dry-run            # Preview what would be synced
  {PROGRAM_NAME} sync --force              # Force overwrite all
  {PROGRAM_NAME} sync --config .i18nrc     # Use custom config file
"#
            );
        }
        "init" => {
            println!(
                r#"Create a sample configuration file

Usage: {PROGRAM_NAME} init [options]

Options:
  --output <path>    Output path (default: .i18nrc.json)
  --help, -h         Show this help message

Examples:
  {PROGRAM_NAME} init                      # Create .i18nrc.json in current directory
  {PROGRAM_NAME} init --output /path/to/config.json  # Custom output path
"#
            );
        }
        "version" => {
            show_version();
        }
        "help" => {
            println!(
                r#"Show help information

Usage: {PROGRAM_NAME} help [command]

Options:
  command            Command to get help for (optional)
  --help, -h         Show this help message

Examples:
  {PROGRAM_NAME} help              # Show general help
  {PROGRAM_NAME} help import       # Show help for import command
  {PROGRAM_NAME} help sync         # Show help for sync command
"#
            );
        }
        _ => {
            eprintln!("Unknown command: {}", command);
            eprintln!("Run '{} help' for available commands.", PROGRAM_NAME);
        }
    }
}

/// 初始化配置文件
///
/// 创建示例配置文件，如果文件已存在则提示用户。
///
/// # Arguments
///
/// * `output` - 可选的输出路径
fn init_config(output: Option<&PathBuf>) -> Result<()> {
    let path = output
        .map(|p| p.to_path_buf())
        .unwrap_or_else(|| PathBuf::from(".i18nrc.json"));

    // 检查文件是否已存在
    if path.exists() {
        println!("⚠️  Configuration file already exists: {}", path.display());
        println!("   To re-create, please delete the existing file first.");
        return Ok(());
    }

    let sample = create_sample_config();
    std::fs::write(&path, &sample)?;

    info!("Created sample configuration file: {}", path.display());
    println!("✅ Created sample configuration file: {}", path.display());
    println!();
    println!("Please edit the configuration file to set the correct project ID and API key.");
    println!("Required fields:");
    println!("  - messagesDir: Path to your messages directory");
    println!("  - projectId: Your YFlow project ID");
    println!("  - apiUrl: Your YFlow API URL");
    println!("  - apiKey: Your API key");

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::PathBuf;
    use tempfile::TempDir;

    // ========== 常量测试 ==========

    #[test]
    fn test_program_name() {
        assert_eq!(PROGRAM_NAME, "yflow");
    }

    #[test]
    fn test_version_not_empty() {
        assert!(!VERSION.is_empty());
        // 版本号应该符合语义化版本格式 (x.y.z)
        let parts: Vec<&str> = VERSION.split('.').collect();
        assert!(parts.len() >= 2, "Version should have at least major.minor");
    }

    // ========== show_version 测试 ==========

    #[test]
    fn test_show_version_no_panic() {
        // 测试 show_version 不 panic
        let result = std::panic::catch_unwind(|| {
            show_version();
        });
        // 如果 panic，测试失败
        assert!(result.is_ok());
    }

    // ========== show_help 测试 ==========

    #[test]
    fn test_show_help_no_command() {
        // 应该不 panic
        let result = std::panic::catch_unwind(|| {
            show_help(None);
        });
        assert!(result.is_ok());
    }

    #[test]
    fn test_show_help_with_import() {
        let result = std::panic::catch_unwind(|| {
            show_help(Some("import"));
        });
        assert!(result.is_ok());
    }

    #[test]
    fn test_show_help_with_sync() {
        let result = std::panic::catch_unwind(|| {
            show_help(Some("sync"));
        });
        assert!(result.is_ok());
    }

    #[test]
    fn test_show_help_with_init() {
        let result = std::panic::catch_unwind(|| {
            show_help(Some("init"));
        });
        assert!(result.is_ok());
    }

    #[test]
    fn test_show_help_with_version() {
        let result = std::panic::catch_unwind(|| {
            show_help(Some("version"));
        });
        assert!(result.is_ok());
    }

    #[test]
    fn test_show_help_with_help() {
        let result = std::panic::catch_unwind(|| {
            show_help(Some("help"));
        });
        assert!(result.is_ok());
    }

    #[test]
    fn test_show_help_with_unknown_command() {
        // 未知命令应该打印错误信息到 stderr，但不 panic
        let result = std::panic::catch_unwind(|| {
            show_help(Some("unknown_command"));
        });
        assert!(result.is_ok());
    }

    #[test]
    fn test_show_command_help_case_insensitive() {
        // 测试命令名大小写不敏感
        let result = std::panic::catch_unwind(|| {
            show_command_help("IMPORT");
        });
        assert!(result.is_ok());

        let result = std::panic::catch_unwind(|| {
            show_command_help("Sync");
        });
        assert!(result.is_ok());
    }

    // ========== init_config 测试 ==========

    #[test]
    fn test_init_config_new_file() {
        let temp_dir = TempDir::new().unwrap();
        let config_path = temp_dir.path().join(".i18nrc.json");

        // 确保文件不存在
        assert!(!config_path.exists());

        let result = init_config(Some(&config_path));

        assert!(result.is_ok());
        assert!(config_path.exists());

        // 验证配置文件内容
        let content = std::fs::read_to_string(&config_path).unwrap();
        // 检查 JSON 格式是否正确（使用 snake_case 因为 SampleConfig 没有 serde rename）
        assert!(content.contains("messages_dir"), "Content: {}", content);
        assert!(content.contains("project_id"), "Content: {}", content);
        assert!(content.contains("api_url"), "Content: {}", content);
        assert!(content.contains("api_key"), "Content: {}", content);
    }

    #[test]
    fn test_init_config_already_exists() {
        let temp_dir = TempDir::new().unwrap();
        let config_path = temp_dir.path().join(".i18nrc.json");

        // 创建已有文件
        std::fs::write(&config_path, "existing content").unwrap();

        // 应该成功但不覆盖文件
        let result = init_config(Some(&config_path));

        assert!(result.is_ok());
        assert_eq!(std::fs::read_to_string(&config_path).unwrap(), "existing content");
    }

    #[test]
    fn test_init_config_default_path() {
        let temp_dir = TempDir::new().unwrap();
        let original_cwd = std::env::current_dir().unwrap();

        // 切换到临时目录
        std::env::set_current_dir(&temp_dir).unwrap();

        // 使用默认路径
        let result = init_config(None);

        // 恢复原始目录
        std::env::set_current_dir(&original_cwd).unwrap();

        assert!(result.is_ok());

        // 验证默认配置文件已创建
        let default_path = temp_dir.path().join(".i18nrc.json");
        assert!(default_path.exists());
    }

    #[test]
    fn test_init_config_invalid_path() {
        // 使用不存在的路径（会导致父目录不存在）
        let invalid_path = PathBuf::from("/nonexistent/path/.i18nrc.json");

        let result = init_config(Some(&invalid_path));

        // 应该返回错误
        assert!(result.is_err());
    }

    // ========== 集成测试 ==========

    #[test]
    fn test_cli_args_parse_version_command() {
        // 测试 version 子命令解析
        let args = CliArgs::parse_from(&["yflow", "version"]);
        assert!(matches!(args.command, Commands::Version));
    }

    #[test]
    fn test_cli_args_parse_help_cmd_command() {
        // 测试 help-cmd 子命令解析（无参数）
        let args = CliArgs::parse_from(&["yflow", "help-cmd"]);
        if let Commands::HelpCmd { command } = args.command {
            assert!(command.is_none());
        } else {
            panic!("Expected HelpCmd command");
        }
    }

    #[test]
    fn test_cli_args_parse_help_cmd_with_command() {
        // 测试 help-cmd 子命令解析（带参数）
        let args = CliArgs::parse_from(&["yflow", "help-cmd", "import"]);
        if let Commands::HelpCmd { command } = args.command {
            assert_eq!(command, Some("import".to_string()));
        } else {
            panic!("Expected HelpCmd command");
        }
    }
}
