/**
 * yflow CLI UI 工具模块
 * 提供进度条、spinner 等 UI 组件
 */

import cliProgress from "cli-progress";

// Spinner 字符
const SPINNER_CHARS = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];

// 缓存 spinner 状态
let spinnerTimer: ReturnType<typeof setInterval> | null = null;
let spinnerIndex = 0;
let spinnerMessage = "";

/**
 * 检查是否应该显示进度条
 * 默认开启，除非明确禁用（更好地支持 WSL + tmux）
 */
export function shouldShowProgress(): boolean {
  // 1. 环境变量明确禁用
  if (process.env.I18N_FORCE_PROGRESS === "0") {
    return false;
  }
  // 2. 默认开启，支持 WSL + tmux
  return true;
}

/**
 * 检查是否在 TTY 环境中
 * @deprecated 使用 shouldShowProgress() 代替
 */
export function isTTY(): boolean {
  return process.stdout.isTTY === true;
}

/**
 * 显示带 spinner 的提示信息（非阻塞）
 */
export function showSpinner(message: string): void {
  if (!shouldShowProgress()) {
    console.log(message);
    return;
  }

  spinnerMessage = message;
  spinnerIndex = 0;

  // 清除之前的 spinner
  if (spinnerTimer) {
    clearInterval(spinnerTimer);
  }

  const updateSpinner = () => {
    const spin = SPINNER_CHARS[spinnerIndex];
    process.stdout.write(`\r${spin} ${spinnerMessage}`);
    spinnerIndex = (spinnerIndex + 1) % SPINNER_CHARS.length;
  };

  updateSpinner();
  spinnerTimer = setInterval(updateSpinner, 80);
}

/**
 * 停止 spinner 并显示完成状态
 */
export function stopSpinner(success: boolean = true, message?: string): void {
  if (spinnerTimer) {
    clearInterval(spinnerTimer);
    spinnerTimer = null;
  }

  // 清除 spinner 行
  if (shouldShowProgress()) {
    process.stdout.write("\r" + " ".repeat(50) + "\r");
  }

  if (message) {
    console.log(success ? `   ✓ ${message}` : `   ✗ ${message}`);
  } else if (success) {
    // 默认不输出，由调用方决定
  }
}

/**
 * 创建多语言进度条管理器
 */
export function createMultiProgressBar() {
  // 创建多栏进度条
  const multiBar = new cliProgress.MultiBar(
    {
      clearOnComplete: false,
      hideCursor: true,
      format: "  {lang} [{bar}] {value}/{total}  ▸ {percentage}%",
      barCompleteChar: "█",
      barIncompleteChar: "░",
      fps: 10,
      stream: process.stdout,
      newlineOnComplete: false,
    },
    cliProgress.Presets.shades_grey
  );

  // 缓存每个语言的进度条
  const bars: Map<string, cliProgress.SingleBar> = new Map();

  /**
   * 为指定语言创建或获取进度条
   */
  function getOrCreateBar(lang: string, total: number): cliProgress.SingleBar | null {
    if (!shouldShowProgress()) {
      return null;
    }

    if (bars.has(lang)) {
      return bars.get(lang)!;
    }

    const bar = multiBar.create(total, 0, { lang });
    bars.set(lang, bar);
    return bar;
  }

  /**
   * 更新指定语言的进度条
   */
  function update(lang: string, value: number, total: number): void {
    const bar = bars.get(lang);
    if (bar) {
      if (value >= total) {
        bar.update(total, { lang });
        bar.stop();
        bars.delete(lang);
      } else {
        bar.update(value, { lang });
      }
    }
  }

  /**
   * 完成单个进度条
   */
  function complete(lang: string, value: number, total: number, extra?: string): void {
    const bar = bars.get(lang);
    if (bar) {
      bar.update(total, { lang });
      bar.stop();
      bars.delete(lang);
    }
  }

  /**
   * 停止所有进度条
   */
  function stop(): void {
    for (const bar of bars.values()) {
      bar.stop();
    }
    bars.clear();
    multiBar.stop();
  }

  /**
   * 检查是否还有活动的进度条
   */
  function hasActive(): boolean {
    return bars.size > 0;
  }

  return {
    getOrCreateBar,
    update,
    complete,
    stop,
    hasActive,
    underlying: multiBar,
  };
}

/**
 * 创建整体进度条（单进度条）
 */
export function createSingleProgressBar(options?: {
  total?: number;
  prefix?: string;
}): {
  bar: cliProgress.SingleBar | null;
  update: (value: number) => void;
  increment: (n?: number) => void;
  stop: () => void;
} {
  let bar: cliProgress.SingleBar | null = null;

  if (shouldShowProgress()) {
    bar = new cliProgress.SingleBar(
      {
        clearOnComplete: false,
        hideCursor: true,
        format: ` ${options?.prefix || "progress"} [{bar}] {value}/{total} ▸ {percentage}%`,
        barCompleteChar: "█",
        barIncompleteChar: "░",
        fps: 10,
        stream: process.stdout,
      },
      cliProgress.Presets.shades_grey
    );

    if (options?.total !== undefined) {
      bar.start(options.total, 0);
    }
  }

  return {
    bar,
    update: (value: number) => {
      if (bar) {
        bar.update(value);
        if (bar.getTotal() === value) {
          bar.stop();
          bar = null;
        }
      }
    },
    increment: (n = 1) => {
      if (bar) {
        const current = bar.getValue();
        const total = bar.getTotal();
        if (current + n >= total) {
          bar.update(total);
          bar.stop();
          bar = null;
        } else {
          bar.increment(n);
        }
      }
    },
    stop: () => {
      if (bar) {
        bar.stop();
        bar = null;
      }
    },
  };
}

/**
 * 格式化数字（添加千位分隔符）
 */
export function formatNumber(num: number): string {
  return num.toLocaleString("en-US");
}

/**
 * 安全停止进度条（确保清理）
 */
export function safeStopProgress(bars: (cliProgress.SingleBar | cliProgress.MultiBar | null)[]): void {
  for (const bar of bars) {
    if (bar) {
      try {
        bar.stop();
      } catch {
        // 忽略停止错误
      }
    }
  }
}
