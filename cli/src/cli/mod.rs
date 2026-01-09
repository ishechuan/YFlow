//! CLI argument parsing and command definitions
//!
//! Uses clap for command-line interface parsing.
//! Provides subcommands for import, sync, init, version, and help operations.

mod commands;

pub use commands::ImportCmd;
pub use commands::SyncCmd;

use clap::{Parser, Subcommand};
use std::path::PathBuf;

/// YFlow CLI - Translation management tool
///
/// A CLI tool for importing and syncing translations between
/// local files and the YFlow backend.
#[derive(Parser, Debug)]
#[command(name = "yflow")]
#[command(author = "YFlow Team")]
#[command(version = "1.0.0")]
#[command(about = "YFlow CLI - Import and sync translations", long_about = None)]
pub struct CliArgs {
    /// Configuration file path
    #[arg(short, long, value_name = "PATH", global = true)]
    pub config: Option<PathBuf>,

    /// Enable verbose output
    #[arg(short, long, global = true)]
    pub verbose: bool,

    #[command(subcommand)]
    pub command: Commands,
}

/// CLI 命令枚举
///
/// 包含所有可用的子命令：
/// - import: 从本地 messages 目录导入翻译到后端
/// - sync: 从后端同步翻译到本地 messages 目录
/// - init: 创建示例配置文件
/// - version: 显示版本信息
/// - help: 显示帮助信息
#[derive(Subcommand, Debug)]
pub enum Commands {
    /// Import translations from local messages directory to backend
    ///
    /// Scans the local messages directory for translation files and imports
    /// them into the YFlow backend database. Supports batching, retry logic,
    /// and progress display.
    ///
    /// Example: `yflow import --dry-run`
    #[command(name = "import")]
    Import(ImportCmd),

    /// Sync translations from backend to local messages directory
    ///
    /// Downloads translations from the YFlow backend and writes them to
    /// the local messages directory. Preserves the original file structure.
    ///
    /// Example: `yflow sync --force`
    #[command(name = "sync")]
    Sync(SyncCmd),

    /// Initialize a sample configuration file
    ///
    /// Creates a `.i18nrc.json` configuration file in the current directory
    /// or the specified output path.
    ///
    /// Example: `yflow init --output /path/to/config.json`
    #[command(name = "init")]
    Init {
        /// Output path (default: .i18nrc.json in current directory)
        #[arg(short, long, value_name = "PATH")]
        output: Option<PathBuf>,
    },

    /// Display version information
    ///
    /// Shows the version number, build information, and other details
    /// about the YFlow CLI tool.
    ///
    /// Example: `yflow version` or `yflow --version`
    #[command(name = "version")]
    Version,

    /// Show help information
    ///
    /// Displays help information for all commands or a specific command.
    /// Can be used as a standalone command or with a command name argument.
    ///
    /// Examples:
    ///   `yflow help-cmd` - Show general help
    ///   `yflow help-cmd import` - Show help for import command
    ///   `yflow import --help` - Alternative way to show command help
    #[command(name = "help-cmd")]
    HelpCmd {
        /// The command to get help for (optional)
        ///
        /// If provided, shows detailed help for the specified command.
        /// If omitted, shows general help for all commands.
        #[arg(value_name = "COMMAND")]
        command: Option<String>,
    },
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::PathBuf;

    /// 测试命令枚举的默认构造
    #[test]
    fn test_commands_import_default() {
        let cmd = Commands::Import(ImportCmd {
            config: None,
            dry_run: false,
        });
        assert!(matches!(cmd, Commands::Import(_)));
    }

    #[test]
    fn test_commands_sync_default() {
        let cmd = Commands::Sync(SyncCmd {
            config: None,
            dry_run: false,
            force: false,
        });
        assert!(matches!(cmd, Commands::Sync(_)));
    }

    #[test]
    fn test_commands_init_default() {
        let cmd = Commands::Init { output: None };
        assert!(matches!(cmd, Commands::Init { output: None }));
    }

    #[test]
    fn test_commands_init_with_output() {
        let cmd = Commands::Init {
            output: Some(PathBuf::from("/custom/path/config.json")),
        };
        if let Commands::Init { output } = cmd {
            assert_eq!(output, Some(PathBuf::from("/custom/path/config.json")));
        } else {
            panic!("Expected Init command");
        }
    }

    #[test]
    fn test_commands_version() {
        let cmd = Commands::Version;
        assert!(matches!(cmd, Commands::Version));
    }

    #[test]
    fn test_commands_help_cmd_no_args() {
        let cmd = Commands::HelpCmd { command: None };
        if let Commands::HelpCmd { command } = cmd {
            assert!(command.is_none());
        } else {
            panic!("Expected HelpCmd command");
        }
    }

    #[test]
    fn test_commands_help_cmd_with_command() {
        let cmd = Commands::HelpCmd {
            command: Some("import".to_string()),
        };
        if let Commands::HelpCmd { command } = cmd {
            assert_eq!(command, Some("import".to_string()));
        } else {
            panic!("Expected HelpCmd command");
        }
    }

    /// 测试 CLI 参数解析 - 基本解析
    #[test]
    fn test_cli_args_parse_import() {
        let args = CliArgs::parse_from(&["yflow", "import"]);
        assert!(matches!(args.command, Commands::Import(_)));
        assert!(args.config.is_none());
    }

    #[test]
    fn test_cli_args_parse_sync() {
        let args = CliArgs::parse_from(&["yflow", "sync"]);
        assert!(matches!(args.command, Commands::Sync(_)));
    }

    #[test]
    fn test_cli_args_parse_version() {
        let args = CliArgs::parse_from(&["yflow", "version"]);
        assert!(matches!(args.command, Commands::Version));
    }

    #[test]
    fn test_cli_args_parse_help_cmd() {
        let args = CliArgs::parse_from(&["yflow", "help-cmd"]);
        if let Commands::HelpCmd { command } = args.command {
            assert!(command.is_none());
        } else {
            panic!("Expected HelpCmd command");
        }
    }

    #[test]
    fn test_cli_args_parse_help_cmd_with_command() {
        let args = CliArgs::parse_from(&["yflow", "help-cmd", "sync"]);
        if let Commands::HelpCmd { command } = args.command {
            assert_eq!(command, Some("sync".to_string()));
        } else {
            panic!("Expected HelpCmd command");
        }
    }

    /// 测试 CLI 参数解析 - 带选项
    #[test]
    fn test_cli_args_parse_with_config() {
        let args = CliArgs::parse_from(&["yflow", "-c", "custom.json", "import"]);
        assert_eq!(args.config, Some(PathBuf::from("custom.json")));
    }

    #[test]
    fn test_cli_args_parse_with_long_config() {
        let args = CliArgs::parse_from(&["yflow", "--config", "custom.json", "sync"]);
        assert_eq!(args.config, Some(PathBuf::from("custom.json")));
    }

    #[test]
    fn test_cli_args_parse_import_with_dry_run() {
        let args = CliArgs::parse_from(&["yflow", "import", "--dry-run"]);
        if let Commands::Import(cmd) = args.command {
            assert!(cmd.dry_run);
        } else {
            panic!("Expected Import command");
        }
    }

    #[test]
    fn test_cli_args_parse_sync_with_force() {
        let args = CliArgs::parse_from(&["yflow", "sync", "--force"]);
        if let Commands::Sync(cmd) = args.command {
            assert!(cmd.force);
        } else {
            panic!("Expected Sync command");
        }
    }

    #[test]
    fn test_cli_args_parse_init_with_output() {
        let args = CliArgs::parse_from(&["yflow", "init", "-o", "/path/to/config.json"]);
        if let Commands::Init { output } = args.command {
            assert_eq!(output, Some(PathBuf::from("/path/to/config.json")));
        } else {
            panic!("Expected Init command");
        }
    }

    /// 测试 CLI 参数解析 - 组合选项
    #[test]
    fn test_cli_args_parse_with_multiple_options() {
        let args = CliArgs::parse_from(&["yflow", "-c", "config.json", "sync", "--dry-run", "--force"]);
        if let Commands::Sync(cmd) = args.command {
            assert_eq!(args.config, Some(PathBuf::from("config.json")));
            assert!(cmd.dry_run);
            assert!(cmd.force);
        } else {
            panic!("Expected Sync command");
        }
    }

    /// 测试 CLI 参数解析 - 全局 verbose 选项
    #[test]
    fn test_cli_args_parse_with_verbose() {
        let args = CliArgs::parse_from(&["yflow", "-v", "import"]);
        assert!(args.verbose);
    }
}
