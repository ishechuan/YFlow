package utils

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	log_utils "yflow/utils"
	"go.uber.org/zap"
)

// DBSecurityConfig 数据库安全配置
type DBSecurityConfig struct {
	EnableQueryLogging    bool            // 是否启用查询日志
	EnableSlowQueryLog    bool            // 是否启用慢查询日志
	SlowQueryThreshold    time.Duration   // 慢查询阈值
	MaxQueryLength        int             // 最大查询长度
	EnableSuspiciousCheck bool            // 是否启用可疑查询检查
	LogLevel              logger.LogLevel // 日志级别
}

// DefaultDBSecurityConfig 默认数据库安全配置
func DefaultDBSecurityConfig() DBSecurityConfig {
	return DBSecurityConfig{
		EnableQueryLogging:    true,
		EnableSlowQueryLog:    true,
		SlowQueryThreshold:    time.Second * 2, // 2秒
		MaxQueryLength:        10000,           // 10KB
		EnableSuspiciousCheck: true,
		LogLevel:              logger.Warn,
	}
}

// SecurityLogger 安全日志记录器
type SecurityLogger struct {
	config DBSecurityConfig
	logger logger.Interface
	zapLogger *zap.Logger
}

// NewSecurityLogger 创建安全日志记录器
func NewSecurityLogger(config DBSecurityConfig, zapLogger *zap.Logger) *SecurityLogger {
	return &SecurityLogger{
		config: config,
		logger: logger.Default.LogMode(config.LogLevel),
		zapLogger: zapLogger,
	}
}

// LogMode 设置日志模式
func (l *SecurityLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.config.LogLevel = level
	return &newLogger
}

// Info 信息日志
func (l *SecurityLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.config.EnableQueryLogging && l.zapLogger != nil {
		l.zapLogger.Info("DB: "+msg, zap.Any("data", data))
	}
}

// Warn 警告日志
func (l *SecurityLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.zapLogger != nil {
		l.zapLogger.Warn("DB: "+msg, zap.Any("data", data))
	}
}

// Error 错误日志
func (l *SecurityLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.zapLogger != nil {
		l.zapLogger.Error("DB: "+msg, zap.Any("data", data))
	}
}

// Trace 跟踪日志（查询日志）
func (l *SecurityLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.zapLogger == nil {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 检查查询长度
	if len(sql) > l.config.MaxQueryLength {
		l.zapLogger.Warn("DB: Oversized query detected",
			zap.String("sql", log_utils.SanitizeLogValue(sql[:min(100, len(sql))])),
			zap.Int("length", len(sql)),
			zap.Duration("elapsed", elapsed),
		)
		return
	}

	// 检查可疑查询
	if l.config.EnableSuspiciousCheck && l.isSuspiciousQuery(sql) {
		l.zapLogger.Warn("DB: Suspicious query detected",
			zap.String("sql", log_utils.SanitizeLogValue(sql)),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.Error(err),
		)
	}

	// 慢查询日志
	if l.config.EnableSlowQueryLog && elapsed > l.config.SlowQueryThreshold {
		l.zapLogger.Warn("DB: Slow query detected",
			zap.String("sql", log_utils.SanitizeLogValue(sql)),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.Error(err),
		)
	}

	// 错误查询日志
	if err != nil && err != gorm.ErrRecordNotFound {
		l.zapLogger.Error("DB: Query execution error",
			zap.String("sql", log_utils.SanitizeLogValue(sql)),
			zap.Duration("elapsed", elapsed),
			zap.Error(err),
		)
	}

	// 正常查询日志（仅在调试模式下）
	if l.config.EnableQueryLogging && l.config.LogLevel == logger.Info {
		l.zapLogger.Info("DB: Query executed",
			zap.String("sql", log_utils.SanitizeLogValue(sql)),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
		)
	}
}

// isSuspiciousQuery 检查是否为可疑查询
func (l *SecurityLogger) isSuspiciousQuery(sql string) bool {
	sqlLower := strings.ToLower(sql)

	// 检查危险关键词
	suspiciousPatterns := []string{
		// SQL注入常见模式
		`union\s+select`,
		`or\s+1\s*=\s*1`,
		`or\s+true`,
		`and\s+1\s*=\s*1`,
		`'.*or.*'.*'`,
		`'.*and.*'.*'`,

		// 系统函数调用
		`load_file\s*\(`,
		`into\s+outfile`,
		`into\s+dumpfile`,
		`benchmark\s*\(`,
		`sleep\s*\(`,
		`waitfor\s+delay`,

		// 信息收集
		`information_schema`,
		`mysql\.user`,
		`pg_user`,
		`sys\.`,

		// 危险操作
		`drop\s+table`,
		`drop\s+database`,
		`truncate\s+table`,
		`delete\s+from.*where\s+1\s*=\s*1`,
		`update.*set.*where\s+1\s*=\s*1`,

		// 注释绕过
		`/\*.*\*/`,
		`--\s`,
		`#.*`,
	}

	for _, pattern := range suspiciousPatterns {
		if matched, _ := regexp.MatchString(pattern, sqlLower); matched {
			return true
		}
	}

	return false
}

// QueryWhitelist 查询白名单
type QueryWhitelist struct {
	AllowedTables    []string // 允许的表名
	AllowedColumns   []string // 允许的列名
	AllowedFunctions []string // 允许的函数
}

// DefaultQueryWhitelist 默认查询白名单
func DefaultQueryWhitelist() QueryWhitelist {
	return QueryWhitelist{
		AllowedTables: []string{
			"users", "projects", "languages", "translations",
			"project_members", "translation_keys",
		},
		AllowedColumns: []string{
			"id", "name", "username", "email", "password", "created_at", "updated_at",
			"project_id", "language_id", "key_name", "value", "context", "status",
			"role", "description", "code", "is_default",
		},
		AllowedFunctions: []string{
			"COUNT", "SUM", "AVG", "MAX", "MIN", "NOW", "DATE", "TIME",
		},
	}
}

// ValidateQuery 验证查询是否符合白名单
func (w *QueryWhitelist) ValidateQuery(sql string) error {
	sqlUpper := strings.ToUpper(sql)

	// 提取表名
	tables := w.extractTables(sql)
	for _, table := range tables {
		if !w.isAllowedTable(table) {
			return fmt.Errorf("不允许访问表: %s", table)
		}
	}

	// 提取列名
	columns := w.extractColumns(sql)
	for _, column := range columns {
		if !w.isAllowedColumn(column) {
			return fmt.Errorf("不允许访问列: %s", column)
		}
	}

	// 检查函数调用
	functions := w.extractFunctions(sql)
	for _, function := range functions {
		if !w.isAllowedFunction(function) {
			return fmt.Errorf("不允许使用函数: %s", function)
		}
	}

	// 检查是否包含危险操作
	dangerousOps := []string{"DROP", "TRUNCATE", "ALTER", "CREATE", "GRANT", "REVOKE"}
	for _, op := range dangerousOps {
		if strings.Contains(sqlUpper, op) {
			return fmt.Errorf("不允许执行危险操作: %s", op)
		}
	}

	return nil
}

// extractTables 提取SQL中的表名
func (w *QueryWhitelist) extractTables(sql string) []string {
	var tables []string

	// 简单的表名提取（可以根据需要改进）
	patterns := []string{
		`FROM\s+(\w+)`,
		`JOIN\s+(\w+)`,
		`UPDATE\s+(\w+)`,
		`INSERT\s+INTO\s+(\w+)`,
		`DELETE\s+FROM\s+(\w+)`,
	}

	sqlUpper := strings.ToUpper(sql)
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(sqlUpper, -1)
		for _, match := range matches {
			if len(match) > 1 {
				tables = append(tables, strings.ToLower(match[1]))
			}
		}
	}

	return tables
}

// extractColumns 提取SQL中的列名
func (w *QueryWhitelist) extractColumns(sql string) []string {
	var columns []string

	// 简单的列名提取（可以根据需要改进）
	re := regexp.MustCompile(`SELECT\s+(.*?)\s+FROM`)
	matches := re.FindStringSubmatch(strings.ToUpper(sql))
	if len(matches) > 1 {
		columnStr := matches[1]
		if columnStr != "*" {
			parts := strings.Split(columnStr, ",")
			for _, part := range parts {
				column := strings.TrimSpace(part)
				// 移除表前缀
				if dotIndex := strings.LastIndex(column, "."); dotIndex != -1 {
					column = column[dotIndex+1:]
				}
				columns = append(columns, strings.ToLower(column))
			}
		}
	}

	return columns
}

// extractFunctions 提取SQL中的函数调用
func (w *QueryWhitelist) extractFunctions(sql string) []string {
	var functions []string

	re := regexp.MustCompile(`(\w+)\s*\(`)
	matches := re.FindAllStringSubmatch(strings.ToUpper(sql), -1)
	for _, match := range matches {
		if len(match) > 1 {
			functions = append(functions, match[1])
		}
	}

	return functions
}

// isAllowedTable 检查表名是否在白名单中
func (w *QueryWhitelist) isAllowedTable(table string) bool {
	for _, allowed := range w.AllowedTables {
		if strings.EqualFold(table, allowed) {
			return true
		}
	}
	return false
}

// isAllowedColumn 检查列名是否在白名单中
func (w *QueryWhitelist) isAllowedColumn(column string) bool {
	for _, allowed := range w.AllowedColumns {
		if strings.EqualFold(column, allowed) {
			return true
		}
	}
	return false
}

// isAllowedFunction 检查函数是否在白名单中
func (w *QueryWhitelist) isAllowedFunction(function string) bool {
	for _, allowed := range w.AllowedFunctions {
		if strings.EqualFold(function, allowed) {
			return true
		}
	}
	return false
}

// DBSecurityMonitor 数据库安全监控器
type DBSecurityMonitor struct {
	config    DBSecurityConfig
	whitelist QueryWhitelist
	logger    *SecurityLogger
}

// NewDBSecurityMonitor 创建数据库安全监控器
func NewDBSecurityMonitor(zapLogger *zap.Logger) *DBSecurityMonitor {
	config := DefaultDBSecurityConfig()
	return &DBSecurityMonitor{
		config:    config,
		whitelist: DefaultQueryWhitelist(),
		logger:    NewSecurityLogger(config, zapLogger),
	}
}

// GetLogger 获取安全日志记录器
func (m *DBSecurityMonitor) GetLogger() logger.Interface {
	return m.logger
}

// ValidateQuery 验证查询
func (m *DBSecurityMonitor) ValidateQuery(sql string) error {
	return m.whitelist.ValidateQuery(sql)
}
