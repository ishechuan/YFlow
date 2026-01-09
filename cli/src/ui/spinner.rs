//! Spinner utilities
//!
//! Provides animated spinner for indicating ongoing operations.

use std::io::Write;

/// Spinner 字符集
const SPINNER_CHARS: &[&str] = &["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];

/// Spinner 实例
///
/// 用于显示正在进行的操作。
pub struct Spinner {
    message: String,
    timer: Option<std::time::Instant>,
}

impl Spinner {
    /// 创建新的 Spinner
    ///
    /// # Arguments
    ///
    /// * `message` - 要显示的消息
    pub fn new(message: &str) -> Self {
        Self {
            message: message.to_string(),
            timer: None,
        }
    }

    /// 启动 spinner
    pub fn start(&mut self) {
        self.timer = Some(std::time::Instant::now());
        self.tick(0);
    }

    /// 更新 spinner
    fn tick(&self, index: usize) {
        let spin = SPINNER_CHARS[index % SPINNER_CHARS.len()];
        print!("\r{} {}", spin, self.message);
        let _ = std::io::stdout().flush();
    }

    /// 停止 spinner
    ///
    /// # Arguments
    ///
    /// * `success` - 是否成功完成
    /// * `message` - 可选的完成消息
    pub fn stop(&self, success: bool, message: Option<&str>) {
        // 清除 spinner 行
        print!("\r{}\r", " ".repeat(50));
        let _ = std::io::stdout().flush();

        if let Some(msg) = message {
            if success {
                println!("   ✓ {}", msg);
            } else {
                println!("   ✗ {}", msg);
            }
        }
    }
}

/// 安全停止 spinner
pub fn safe_stop_spinner() {
    print!("\r{}\r", " ".repeat(50));
    let _ = std::io::stdout().flush();
}
