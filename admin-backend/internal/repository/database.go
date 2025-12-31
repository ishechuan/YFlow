package repository

import (
	"fmt"
	"yflow/internal/config"
	"yflow/internal/domain"
	internal_utils "yflow/internal/utils"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config, zapLogger *zap.Logger, monitor *internal_utils.DBSecurityMonitor) (*gorm.DB, error) {
	// 优化DSN配置，添加连接参数
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s&interpolateParams=true",
		cfg.DB.Username,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.DBName)

	// GORM配置优化
	gormConfig := &gorm.Config{
		// 禁用默认事务以提高性能
		SkipDefaultTransaction: true,
		// 批量插入优化
		CreateBatchSize: 1000,
		// 准备语句缓存
		PrepareStmt: true,
	}

	// 配置安全日志记录器
	if os.Getenv("GO_ENV") == "production" {
		gormConfig.Logger = monitor.GetLogger().LogMode(logger.Warn)
	} else {
		gormConfig.Logger = monitor.GetLogger().LogMode(logger.Info)
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	// 获取底层的sql.DB对象进行连接池优化
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 连接池优化配置
	sqlDB.SetMaxIdleConns(10)                  // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)                 // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour)        // 连接最大生存时间
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // 连接最大空闲时间

	// 自动迁移表结构
	err = db.AutoMigrate(
		&domain.User{},
		&domain.Project{},
		&domain.Language{},
		&domain.Translation{},
		&domain.ProjectMember{},
		&domain.Invitation{},
	)
	if err != nil {
		return nil, fmt.Errorf("自动迁移表结构失败: %w", err)
	}

	// 创建额外的性能优化索引
	if err := createOptimizationIndexes(db, zapLogger); err != nil {
		zapLogger.Warn("Warning during index creation", zap.Error(err))
	}

	// 初始化种子数据
	if err := initSeedData(db, zapLogger); err != nil {
		return nil, fmt.Errorf("初始化种子数据失败: %w", err)
	}

	return db, nil
}

// initSeedData 初始化种子数据
func initSeedData(db *gorm.DB, zapLogger *zap.Logger) error {
	// 创建管理员用户
	if err := createAdminUser(db, zapLogger); err != nil {
		return err
	}

	// 创建常见语言
	if err := createDefaultLanguages(db, zapLogger); err != nil {
		return err
	}

	return nil
}

// createAdminUser 创建默认管理员用户
func createAdminUser(db *gorm.DB, zapLogger *zap.Logger) error {
	var count int64
	if err := db.Model(&domain.User{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		adminPassword := os.Getenv("ADMIN_PASSWORD")
		if adminPassword == "" {
			adminPassword = "admin123" // 默认密码
		}

		password, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("生成密码哈希失败: %w", err)
		}

		adminUsername := os.Getenv("ADMIN_USERNAME")
		if adminUsername == "" {
			adminUsername = "admin"
		}

		admin := domain.User{
			Username:  adminUsername,
			Email:     "admin@yflow.com", // 默认管理员邮箱
			Password:  string(password),
			Role:      "admin",
			Status:    "active",
			CreatedBy: 1, // 系统管理员创建
			UpdatedBy: 1,
		}

		if err := db.Create(&admin).Error; err != nil {
			return fmt.Errorf("创建管理员用户失败: %w", err)
		}

		zapLogger.Info("Default admin user created", zap.String("username", adminUsername))
	} else {
		// 检查现有用户是否需要更新角色和邮箱
		var admin domain.User
		if err := db.Where("username = ?", "admin").First(&admin).Error; err == nil {
			needUpdate := false
			if admin.Role != "admin" {
				admin.Role = "admin"
				needUpdate = true
			}
			if admin.Email == "" {
				admin.Email = "admin@yflow.com"
				needUpdate = true
			}
			if admin.Status == "" {
				admin.Status = "active"
				needUpdate = true
			}

			if needUpdate {
				admin.UpdatedBy = 1 // 系统管理员更新
				if err := db.Save(&admin).Error; err != nil {
					zapLogger.Warn("Failed to update admin user info", zap.Error(err))
				} else {
					zapLogger.Info("Admin user info updated")
				}
			}
		}
		zapLogger.Info("Admin user already exists, skipping creation")
	}

	return nil
}

// createDefaultLanguages 创建默认语言
func createDefaultLanguages(db *gorm.DB, zapLogger *zap.Logger) error {
	var count int64
	if err := db.Model(&domain.Language{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		// 定义常见语言列表
		languages := []domain.Language{
			{Code: "en", Name: "English", IsDefault: true, CreatedBy: 1, UpdatedBy: 1},
			{Code: "zh_CN", Name: "简体中文", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "zh_TW", Name: "繁體中文", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "ja", Name: "日本語", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "ko", Name: "한국어", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "fr", Name: "Français", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "de", Name: "Deutsch", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "es", Name: "Español", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "it", Name: "Italiano", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "ru", Name: "Русский", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "pt", Name: "Português", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "nl", Name: "Nederlands", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "ar", Name: "العربية", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "hi", Name: "हिन्दी", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "tr", Name: "Türkçe", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "pl", Name: "Polski", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "vi", Name: "Tiếng Việt", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "th", Name: "ไทย", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "id", Name: "Bahasa Indonesia", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
			{Code: "sv", Name: "Svenska", IsDefault: false, CreatedBy: 1, UpdatedBy: 1},
		}

		if err := db.CreateInBatches(languages, len(languages)).Error; err != nil {
			return fmt.Errorf("创建默认语言失败: %w", err)
		}

		zapLogger.Info("Default language list created")
	} else {
		zapLogger.Info("Language list already exists, skipping creation")
	}

	return nil
}

// IndexDefinition 索引定义
type IndexDefinition struct {
	Name      string
	TableName string
	Columns   []string
	Unique    bool
}

// createOptimizationIndexes 创建额外的性能优化索引
func createOptimizationIndexes(db *gorm.DB, zapLogger *zap.Logger) error {
	// 定义需要创建的索引
	indexes := []IndexDefinition{
		{
			Name:      "idx_translations_project_status",
			TableName: "translations",
			Columns:   []string{"project_id", "status"},
			Unique:    false,
		},
		{
			Name:      "idx_translations_search_key",
			TableName: "translations",
			Columns:   []string{"project_id", "key_name", "status"},
			Unique:    false,
		},
		{
			Name:      "idx_translations_project_lang",
			TableName: "translations",
			Columns:   []string{"project_id", "language_id"},
			Unique:    false,
		},
		{
			Name:      "idx_projects_status_name",
			TableName: "projects",
			Columns:   []string{"status", "name"},
			Unique:    false,
		},
		{
			Name:      "idx_languages_code_status",
			TableName: "languages",
			Columns:   []string{"code", "status"},
			Unique:    false,
		},
		// 添加翻译唯一约束索引（如果GORM没有自动创建）
		{
			Name:      "idx_translation_unique",
			TableName: "translations",
			Columns:   []string{"project_id", "key_name", "language_id"},
			Unique:    true,
		},
		// 项目成员相关索引
		{
			Name:      "idx_project_members_project",
			TableName: "project_members",
			Columns:   []string{"project_id"},
			Unique:    false,
		},
		{
			Name:      "idx_project_members_user",
			TableName: "project_members",
			Columns:   []string{"user_id"},
			Unique:    false,
		},
		{
			Name:      "idx_project_members_role",
			TableName: "project_members",
			Columns:   []string{"project_id", "role"},
			Unique:    false,
		},
	}

	for _, idx := range indexes {
		if err := createIndexIfNotExists(db, idx, zapLogger); err != nil {
			zapLogger.Warn("Warning during index creation", zap.String("index", idx.Name), zap.Error(err))
		}
	}

	return nil
}

// createIndexIfNotExists 如果索引不存在则创建
func createIndexIfNotExists(db *gorm.DB, idx IndexDefinition, zapLogger *zap.Logger) error {
	// 检查索引是否已存在
	exists, err := indexExists(db, idx.TableName, idx.Name)
	if err != nil {
		return fmt.Errorf("检查索引是否存在时出错: %w", err)
	}

	if exists {
		return nil // 索引已存在，跳过创建
	}

	// 构建创建索引的SQL
	indexType := "INDEX"
	if idx.Unique {
		indexType = "UNIQUE INDEX"
	}

	columnList := strings.Join(idx.Columns, ", ")
	sql := fmt.Sprintf("CREATE %s %s ON %s (%s)", indexType, idx.Name, idx.TableName, columnList)

	// 执行创建索引
	if err := db.Exec(sql).Error; err != nil {
		// 检查是否是索引已存在的错误
		if strings.Contains(err.Error(), "Duplicate key name") ||
			strings.Contains(err.Error(), "already exists") {
			return nil // 索引已存在，不是错误
		}
		return fmt.Errorf("创建索引失败: %w", err)
	}

	zapLogger.Info("Index created successfully", zap.String("index", idx.Name))
	return nil
}

// indexExists 检查索引是否存在
func indexExists(db *gorm.DB, tableName, indexName string) (bool, error) {
	var count int64
	err := db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.statistics 
		WHERE table_schema = DATABASE() 
		AND table_name = ? 
		AND index_name = ?
	`, tableName, indexName).Scan(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
