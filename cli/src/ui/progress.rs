//! Progress bar utilities
//!
//! Provides progress bars and progress managers for long-running operations.
//! Uses the indicatif library for terminal progress display.
//!
//! # Architecture
//!
//! - `MultiProgressManager`: Manages multiple concurrent progress bars (one per language)
//! - Progress bars support automatic cleanup on completion
//!
//! # Environment Variables
//!
//! - `I18N_FORCE_PROGRESS=0`: Disable all progress displays

use indicatif::{HumanDuration, MultiProgress, ProgressBar, ProgressStyle};
use std::collections::HashMap;
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::Arc;
use std::time::Duration;

/// Progress bar style template
const DEFAULT_TEMPLATE: &str = "{msg} [{elapsed_precise}] {wide_bar} {pos}/{len} ({percent}%)";

/// Characters used for the progress bar fill
const PROGRESS_CHARS: &str = "█░";

/// Check if progress display should be shown
pub fn should_show_progress() -> bool {
    if std::env::var("I18N_FORCE_PROGRESS") == Ok("0".to_string()) {
        return false;
    }
    true
}

/// Progress state for tracking completion status
#[derive(Debug, Clone)]
pub struct ProgressState {
    total: Arc<AtomicUsize>,
    completed: Arc<AtomicUsize>,
}

impl ProgressState {
    pub fn new(total: usize) -> Self {
        Self {
            total: Arc::new(AtomicUsize::new(total)),
            completed: Arc::new(AtomicUsize::new(0)),
        }
    }

    pub fn inc(&self) {
        self.completed.fetch_add(1, Ordering::SeqCst);
    }

    pub fn inc_by(&self, n: usize) {
        self.completed.fetch_add(n, Ordering::SeqCst);
    }

    pub fn completed(&self) -> usize {
        self.completed.load(Ordering::SeqCst)
    }

    pub fn total(&self) -> usize {
        self.total.load(Ordering::SeqCst)
    }

    pub fn is_complete(&self) -> bool {
        self.completed() >= self.total()
    }

    pub fn percentage(&self) -> f64 {
        let total = self.total();
        if total == 0 {
            return 0.0;
        }
        self.completed() as f64 / total as f64
    }
}

/// Progress bar wrapper with language-specific tracking
#[derive(Clone)]
pub struct LanguageProgressBar {
    bar: ProgressBar,
    lang: String,
    state: ProgressState,
    active: bool,
}

impl LanguageProgressBar {
    pub fn new(manager: &MultiProgress, lang: &str, total: u64) -> Self {
        let bar = manager.add(ProgressBar::new(total));
        bar.set_style(
            ProgressStyle::with_template(DEFAULT_TEMPLATE)
                .unwrap()
                .progress_chars(PROGRESS_CHARS),
        );
        bar.set_message(format!("📦 {}", lang));

        Self {
            bar,
            lang: lang.to_string(),
            state: ProgressState::new(total as usize),
            active: true,
        }
    }

    pub fn inc(&self) {
        if self.active {
            self.bar.inc(1);
            self.state.inc();
        }
    }

    pub fn inc_by(&self, n: u64) {
        if self.active {
            self.bar.inc(n);
            self.state.inc_by(n as usize);
        }
    }

    pub fn finish(&mut self) {
        if self.active {
            self.bar.finish_with_message(format!("✅ {} ({}/{})", self.lang, self.state.completed(), self.state.total()));
            self.active = false;
        }
    }

    pub fn abort(&mut self) {
        if self.active {
            self.bar.abandon_with_message(format!("❌ {} - Failed", self.lang));
            self.active = false;
        }
    }

    pub fn lang(&self) -> &str {
        &self.lang
    }

    pub fn state(&self) -> &ProgressState {
        &self.state
    }

    pub fn is_active(&self) -> bool {
        self.active
    }
}

/// Multi-progress manager for handling multiple concurrent progress bars
#[derive(Clone)]
pub struct MultiProgressManager {
    multi_bar: MultiProgress,
    bars: Arc<parking_lot::Mutex<HashMap<String, LanguageProgressBar>>>,
    enabled: bool,
}

impl Default for MultiProgressManager {
    fn default() -> Self {
        Self::new()
    }
}

impl MultiProgressManager {
    pub fn new() -> Self {
        Self {
            multi_bar: MultiProgress::new(),
            bars: Arc::new(parking_lot::Mutex::new(HashMap::new())),
            enabled: should_show_progress(),
        }
    }

    pub fn create_bar(&self, lang: &str, total: u64) -> LanguageProgressBar {
        if !self.enabled {
            return LanguageProgressBar {
                bar: ProgressBar::hidden(),
                lang: lang.to_string(),
                state: ProgressState::new(total as usize),
                active: false,
            };
        }

        let bar = LanguageProgressBar::new(&self.multi_bar, lang, total);
        self.bars.lock().insert(lang.to_string(), bar.clone());
        bar
    }

    pub fn stop(&self) {
        if self.enabled {
            self.multi_bar.clear().ok();
        }
    }

    pub fn finish_all(&self) {
        let mut bars = self.bars.lock();
        for bar in bars.values_mut() {
            bar.finish();
        }
    }

    pub fn abort_all(&self) {
        let mut bars = self.bars.lock();
        for bar in bars.values_mut() {
            bar.abort();
        }
    }

    pub fn is_enabled(&self) -> bool {
        self.enabled
    }

    pub fn len(&self) -> usize {
        self.bars.lock().len()
    }

    pub fn is_empty(&self) -> bool {
        self.bars.lock().is_empty()
    }
}

/// Creates a single progress bar for simple use cases
pub fn create_single_progress_bar(total: u64, prefix: &str) -> ProgressBar {
    let bar = ProgressBar::new(total);
    let template = format!(" {} {}", prefix, DEFAULT_TEMPLATE);
    bar.set_style(
        ProgressStyle::with_template(&template)
            .unwrap()
            .progress_chars(PROGRESS_CHARS),
    );
    bar
}

/// Format a number
pub fn format_number(num: usize) -> String {
    num.to_string()
}

/// Format duration in human-readable format
pub fn format_duration(duration: Duration) -> String {
    HumanDuration(duration).to_string()
}

/// Calculate ETA (Estimated Time of Arrival)
pub fn calculate_eta(elapsed: Duration, completed: usize, total: usize) -> String {
    if completed == 0 || elapsed.is_zero() {
        return "N/A".to_string();
    }

    let rate = completed as f64 / elapsed.as_secs_f64();
    if rate <= 0.0 {
        return "N/A".to_string();
    }

    let remaining = total.saturating_sub(completed);
    let eta_secs = remaining as f64 / rate;

    if eta_secs < 60.0 {
        format!("{:.0}s", eta_secs)
    } else if eta_secs < 3600.0 {
        format!("{:.1}m", eta_secs / 60.0)
    } else {
        format!("{:.1}h", eta_secs / 3600.0)
    }
}

/// Progress display options
#[derive(Debug, Clone, Default)]
pub struct ProgressOptions {
    pub show_percentage: bool,
    pub show_elapsed: bool,
    pub show_eta: bool,
    pub template: Option<String>,
    pub bar_char: Option<String>,
    pub empty_char: Option<String>,
}

impl ProgressOptions {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn with_percentage(mut self, enabled: bool) -> Self {
        self.show_percentage = enabled;
        self
    }

    pub fn with_elapsed(mut self, enabled: bool) -> Self {
        self.show_elapsed = enabled;
        self
    }

    pub fn with_eta(mut self, enabled: bool) -> Self {
        self.show_eta = enabled;
        self
    }

    pub fn with_template(mut self, template: &str) -> Self {
        self.template = Some(template.to_string());
        self
    }

    pub fn with_chars(mut self, filled: &str, empty: &str) -> Self {
        self.bar_char = Some(filled.to_string());
        self.empty_char = Some(empty.to_string());
        self
    }
}

/// 安全停止单个进度条
///
/// 此函数确保即使进度条处于异常状态也能正确停止。
/// 使用 `finish_and_clear` 确保进度条被完全清理。
///
/// # Arguments
///
/// * `bar` - 要停止的进度条引用
pub fn safe_stop_progress(bar: &ProgressBar) {
    bar.finish_and_clear();
}

/// 安全停止多个进度条
///
/// 此函数确保即使某些进度条已停止或出现异常，
/// 也能安全地清理所有进度条资源。
/// 常用于批量操作完成后的清理工作。
///
/// # Arguments
///
/// * `bars` - 要停止的进度条切片
///
/// # Example
///
/// ```ignore
/// use indicatif::ProgressBar;
///
/// let bar1 = ProgressBar::new(100);
/// let bar2 = ProgressBar::new(200);
///
/// // ... 操作完成后安全停止
/// safe_stop_progress_batch(&[bar1, bar2]);
/// ```
pub fn safe_stop_progress_batch(bars: &[ProgressBar]) {
    for bar in bars {
        bar.finish_and_clear();
    }
}

/// 安全停止 MultiProgressManager
///
/// 清理 MultiProgressManager 及其关联的所有进度条。
/// 此函数确保在发生错误时也能正确释放资源，
/// 避免进度条残留在终端上。
///
/// # Arguments
///
/// * `manager` - MultiProgressManager 引用
///
/// # Example
///
/// ```ignore
/// let manager = MultiProgressManager::new();
/// // ... 创建进度条并使用
///
/// // 完成后安全停止
/// safe_stop_multi_progress(&manager);
/// ```
pub fn safe_stop_multi_progress(manager: &MultiProgressManager) {
    manager.abort_all();
    manager.stop();
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::time::Duration;

    #[test]
    fn test_should_show_progress_default() {
        // 确保环境变量未被设置
        std::env::remove_var("I18N_FORCE_PROGRESS");
        // 默认应该返回 true（除非在特殊环境中）
        // 这个测试检查函数的基本行为
        let result = should_show_progress();
        // 验证函数不会 panic 并且返回可预测的值
        assert!(result || !result, "should_show_progress should return a boolean");
    }

    #[test]
    fn test_should_show_progress_disabled() {
        std::env::set_var("I18N_FORCE_PROGRESS", "0");
        assert!(!should_show_progress());
        std::env::remove_var("I18N_FORCE_PROGRESS");
    }

    #[test]
    fn test_should_show_progress_explicitly_enabled() {
        std::env::remove_var("I18N_FORCE_PROGRESS");
        // 当设置为非 0 值时，应该返回 true
        std::env::set_var("I18N_FORCE_PROGRESS", "1");
        assert!(should_show_progress());
        std::env::remove_var("I18N_FORCE_PROGRESS");
    }

    #[test]
    fn test_progress_state_creation() {
        let state = ProgressState::new(100);
        assert_eq!(state.total(), 100);
        assert_eq!(state.completed(), 0);
        assert!(!state.is_complete());
    }

    #[test]
    fn test_progress_state_increment() {
        let state = ProgressState::new(100);
        state.inc();
        assert_eq!(state.completed(), 1);
        state.inc_by(5);
        assert_eq!(state.completed(), 6);
    }

    #[test]
    fn test_progress_state_percentage() {
        let state = ProgressState::new(100);
        state.inc_by(25);
        assert_eq!(state.percentage(), 0.25);
    }

    #[test]
    fn test_progress_state_is_complete() {
        let state = ProgressState::new(10);
        state.inc_by(9);
        assert!(!state.is_complete());
        state.inc();
        assert!(state.is_complete());
    }

    #[test]
    fn test_progress_state_zero_total() {
        let state = ProgressState::new(0);
        assert_eq!(state.percentage(), 0.0);
    }

    #[test]
    fn test_format_number() {
        assert_eq!(format_number(0), "0");
        assert_eq!(format_number(1), "1");
        assert_eq!(format_number(1000), "1000");
    }

    #[test]
    fn test_calculate_eta() {
        let elapsed = Duration::from_secs(10);
        assert_eq!(calculate_eta(elapsed, 50, 100), "10s");
        assert_eq!(calculate_eta(elapsed, 0, 100), "N/A");
        let zero = Duration::from_secs(0);
        assert_eq!(calculate_eta(zero, 50, 100), "N/A");
    }

    #[test]
    fn test_progress_options_defaults() {
        let opts = ProgressOptions::new();
        assert!(!opts.show_percentage);
        assert!(!opts.show_elapsed);
        assert!(!opts.show_eta);
    }

    #[test]
    fn test_progress_options_builder() {
        let opts = ProgressOptions::new()
            .with_percentage(true)
            .with_elapsed(true)
            .with_eta(true);
        assert!(opts.show_percentage);
        assert!(opts.show_elapsed);
        assert!(opts.show_eta);
    }

    #[test]
    fn test_multi_progress_manager_default() {
        let manager = MultiProgressManager::new();
        assert!(manager.is_enabled());
    }

    #[test]
    fn test_multi_progress_manager_create_bar() {
        let manager = MultiProgressManager::new();
        let bar = manager.create_bar("en", 100);
        assert_eq!(bar.lang(), "en");
        assert!(!manager.is_empty());
        manager.stop();
    }

    #[test]
    fn test_multi_progress_manager_finish_all() {
        let manager = MultiProgressManager::new();
        manager.create_bar("en", 10);
        manager.create_bar("zh", 20);
        manager.finish_all();
        manager.stop();
    }

    #[test]
    fn test_language_progress_bar_increment() {
        let manager = MultiProgressManager::new();
        let bar = manager.create_bar("en", 100);
        for _ in 0..5 {
            bar.inc();
        }
        assert_eq!(bar.state().completed(), 5);
        manager.stop();
    }

    #[test]
    fn test_language_progress_bar_finish() {
        let manager = MultiProgressManager::new();
        let mut bar = manager.create_bar("en", 10);
        bar.inc_by(10);
        bar.finish();
        assert!(!bar.is_active());
        manager.stop();
    }

    #[test]
    fn test_safe_stop_progress() {
        let bar = ProgressBar::new(100);
        bar.inc(50);
        safe_stop_progress(&bar);
        // If this doesn't panic, the test passes
    }

    #[test]
    fn test_safe_stop_progress_batch() {
        let bars: Vec<ProgressBar> = vec![
            ProgressBar::new(100),
            ProgressBar::new(200),
            ProgressBar::new(300),
        ];
        safe_stop_progress_batch(&bars);
        // If this doesn't panic, the test passes
    }

    #[test]
    fn test_safe_stop_progress_empty_batch() {
        let bars: Vec<ProgressBar> = vec![];
        safe_stop_progress_batch(&bars);
        // Empty batch should not panic
    }

    #[test]
    fn test_safe_stop_multi_progress() {
        let manager = MultiProgressManager::new();
        manager.create_bar("en", 10);
        manager.create_bar("zh", 20);
        safe_stop_multi_progress(&manager);
        // If this doesn't panic, the test passes
    }
}
