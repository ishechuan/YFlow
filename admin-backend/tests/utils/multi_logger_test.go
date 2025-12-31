package utils_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"yflow/internal/config"
	"yflow/utils"
)

func TestNewLoggerManager(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "logger_manager_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 测试配置
	cfg := config.LogConfig{
		Level:      "info",
		Format:     "json",
		Output:     "file", // Use file output for testing file creation
		LogDir:     tempDir,
		DateFormat: "2006-01-02",
		MaxSize:    1,
		MaxAge:     1,
		MaxBackups: 1,
		Compress:   false,
	}

	// 创建日志管理器
	loggerManager, err := utils.NewLoggerManager(cfg)
	require.NoError(t, err)
	require.NotNil(t, loggerManager)

	// 验证获取应用日志器
	appLogger := loggerManager.GetAppLogger()
	assert.NotNil(t, appLogger)

	// Write a log entry to ensure file is created
	appLogger.Info("test", zap.String("type", "app"))

	// 验证日志文件是否创建
	files, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	assert.NotEmpty(t, files)
}
