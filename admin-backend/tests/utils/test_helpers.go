package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"yflow/internal/domain"
)

// CreateTempDir 创建临时目录并在测试结束后删除
func CreateTempDir(t *testing.T, prefix string) string {
	tempDir, err := os.MkdirTemp("", prefix)
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})
	return tempDir
}

// SetupTestLogger 设置测试日志器并返回观察者
func SetupTestLogger(t *testing.T) (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	return logger, logs
}

// SetupTestDB 创建测试数据库
func SetupTestDB(t *testing.T) *gorm.DB {
	// 使用环境变量或默认测试配置
	dbUser := getEnvOrDefault("TEST_DB_USER", "root")
	dbPass := getEnvOrDefault("TEST_DB_PASS", "")
	dbHost := getEnvOrDefault("TEST_DB_HOST", "localhost")
	dbPort := getEnvOrDefault("TEST_DB_PORT", "3306")
	dbName := getEnvOrDefault("TEST_DB_NAME", fmt.Sprintf("i18n_flow_test_%d", os.Getpid()))

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// 先连接到MySQL服务器
	rootDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", dbUser, dbPass, dbHost, dbPort)
	rootDB, err := gorm.Open(mysql.Open(rootDsn), &gorm.Config{})
	if err != nil {
		t.Skipf("跳过测试：无法连接到MySQL服务器: %v", err)
		return nil
	}

	// 创建测试数据库
	err = rootDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)).Error
	require.NoError(t, err)

	err = rootDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName)).Error
	require.NoError(t, err)

	// 连接到测试数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移模型
	err = db.AutoMigrate(
		&domain.User{},
		&domain.Project{},
		&domain.Language{},
		&domain.Translation{},
	)
	require.NoError(t, err)

	// 在测试结束时清理
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}

		// 删除测试数据库
		rootDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
		sqlDB, _ = rootDB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	})

	return db
}

// getEnvOrDefault 获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// CreateTestFile 创建测试文件并返回路径
func CreateTestFile(t *testing.T, dir, name, content string) string {
	if dir == "" {
		dir = t.TempDir()
	}

	filePath := filepath.Join(dir, name)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	return filePath
}

// CreateTestUser 创建测试用户
func CreateTestUser(t *testing.T, db *gorm.DB, username, password string) *domain.User {
	user := &domain.User{
		Username: username,
		Password: password,
	}

	err := db.Create(user).Error
	require.NoError(t, err)

	return user
}

// CreateTestProject 创建测试项目
func CreateTestProject(t *testing.T, db *gorm.DB, name, description, slug string) *domain.Project {
	project := &domain.Project{
		Name:        name,
		Description: description,
		Slug:        slug,
		Status:      "active",
	}

	err := db.Create(project).Error
	require.NoError(t, err)

	return project
}

// CreateTestLanguage 创建测试语言
func CreateTestLanguage(t *testing.T, db *gorm.DB, code, name string, isDefault bool) *domain.Language {
	language := &domain.Language{
		Code:      code,
		Name:      name,
		IsDefault: isDefault,
		Status:    "active",
	}

	err := db.Create(language).Error
	require.NoError(t, err)

	return language
}

// CreateTestTranslation 创建测试翻译
func CreateTestTranslation(t *testing.T, db *gorm.DB, projectID uint64, keyName, context string, languageID uint64, value string) *domain.Translation {
	translation := &domain.Translation{
		ProjectID:  projectID,
		KeyName:    keyName,
		Context:    context,
		LanguageID: languageID,
		Value:      value,
		Status:     "active",
	}

	err := db.Create(translation).Error
	require.NoError(t, err)

	return translation
}

// AssertFileExists 断言文件存在
func AssertFileExists(t *testing.T, path string) {
	_, err := os.Stat(path)
	require.NoError(t, err)
}

// AssertFileNotExists 断言文件不存在
func AssertFileNotExists(t *testing.T, path string) {
	_, err := os.Stat(path)
	require.True(t, os.IsNotExist(err))
}

// ReadTestFile 读取测试文件内容
func ReadTestFile(t *testing.T, path string) string {
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}
