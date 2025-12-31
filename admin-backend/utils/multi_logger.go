package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"yflow/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LoggerManager 日志管理器（简化版：只保留单一日志器）
type LoggerManager struct {
	config config.LogConfig
	logger *zap.Logger
}

// NewLoggerManager 创建日志管理器
func NewLoggerManager(cfg config.LogConfig) (*LoggerManager, error) {
	// 确保日志目录存在
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %v", err)
	}

	logger, err := createLogger(cfg)
	if err != nil {
		return nil, err
	}

	return &LoggerManager{
		config: cfg,
		logger: logger,
	}, nil
}

// createLogger 创建日志器（统一处理）
func createLogger(cfg config.LogConfig) (*zap.Logger, error) {
	level := parseLogLevel(cfg.Level)

	// 创建编码器配置
	encoderConfig := getEncoderConfig()

	var cores []zapcore.Core

	// 控制台输出
	if cfg.Output == "stdout" || cfg.Output == "both" {
		consoleEncoder := getConsoleEncoder(encoderConfig, cfg.Format)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// 文件输出
	if cfg.Output == "file" || cfg.Output == "both" {
		filename := getLogFilename(cfg.LogDir, "app", cfg.DateFormat)
		fileWriter := &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
			Compress:   cfg.Compress,
		}

		// 根据配置选择文件编码器
		var fileEncoder zapcore.Encoder
		if cfg.Format == "json" {
			fileEncoder = zapcore.NewJSONEncoder(encoderConfig)
		} else {
			fileEncoder = zapcore.NewConsoleEncoder(encoderConfig)
		}
		fileCore := zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(fileWriter),
			level,
		)
		cores = append(cores, fileCore)
	}

	// 额外写入错误日志文件（始终创建，用于独立收集错误）
	if cfg.Output == "file" || cfg.Output == "both" {
		errorFilename := getLogFilename(cfg.LogDir, "error", cfg.DateFormat)
		errorWriter := &lumberjack.Logger{
			Filename:   errorFilename,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
			Compress:   cfg.Compress,
		}

		errorCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(errorWriter),
			zapcore.ErrorLevel,
		)
		cores = append(cores, errorCore)
	}

	core := zapcore.NewTee(cores...)
	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}

// parseLogLevel 解析日志级别
func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// getEncoderConfig 获取编码器配置
func getEncoderConfig() zapcore.EncoderConfig {
	config := zap.NewProductionEncoderConfig()
	config.TimeKey = "timestamp"
	config.LevelKey = "level"
	config.NameKey = "logger"
	config.CallerKey = "caller"
	config.MessageKey = "message"
	config.StacktraceKey = "stacktrace"
	config.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	config.EncodeLevel = zapcore.LowercaseLevelEncoder
	config.EncodeCaller = zapcore.ShortCallerEncoder
	return config
}

// getConsoleEncoder 获取控制台编码器
func getConsoleEncoder(config zapcore.EncoderConfig, format string) zapcore.Encoder {
	if format == "json" {
		return zapcore.NewJSONEncoder(config)
	}
	// 控制台使用彩色输出
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(config)
}

// getLogFilename 获取日志文件名
func getLogFilename(logDir, logType, dateFormat string) string {
	dateStr := time.Now().Format(dateFormat)
	filename := fmt.Sprintf("%s-%s.log", logType, dateStr)
	return filepath.Join(logDir, filename)
}

// GetAppLogger 获取应用日志器
func (lm *LoggerManager) GetAppLogger() *zap.Logger {
	return lm.logger
}

// SyncAll 同步日志缓冲区
func (lm *LoggerManager) SyncAll() {
	if lm.logger != nil {
		lm.logger.Sync()
	}
}

// ========== 安全日志函数（保持为包级函数，因为与日志器无关）==========

// SanitizeLogValue 清理日志值，防止日志注入
// 移除或替换可能导致日志伪造的字符（换行符、回车符等）
func SanitizeLogValue(value string) string {
	if value == "" {
		return value
	}
	// 移除回车符和换行符，替换为转义序列
	result := strings.ReplaceAll(value, "\r", "")
	result = strings.ReplaceAll(result, "\n", "\\n")
	// 移除制表符（可能影响日志格式）
	result = strings.ReplaceAll(result, "\t", " ")
	// 限制长度，防止过长的值
	if len(result) > 1000 {
		result = result[:1000] + "..."
	}
	return result
}
