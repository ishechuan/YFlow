package repository

import (
	"context"
	"errors"
	"yflow/internal/domain"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TranslationRepository 翻译仓储实现
type TranslationRepository struct {
	db *gorm.DB
}

// NewTranslationRepository 创建翻译仓储实例
func NewTranslationRepository(db *gorm.DB) *TranslationRepository {
	return &TranslationRepository{db: db}
}

// GetByID 根据ID获取翻译
func (r *TranslationRepository) GetByID(ctx context.Context, id uint64) (*domain.Translation, error) {
	var translation domain.Translation
	if err := r.db.WithContext(ctx).Preload("Project").Preload("Language").First(&translation, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTranslationNotFound
		}
		return nil, err
	}
	return &translation, nil
}

// GetByProjectID 根据项目ID获取翻译（分页）
func (r *TranslationRepository) GetByProjectID(ctx context.Context, projectID uint64, limit, offset int) ([]*domain.Translation, int64, error) {
	var translations []*domain.Translation
	var total int64

	query := r.db.WithContext(ctx).Where("project_id = ?", projectID)

	// 计算总数
	if err := query.Model(&domain.Translation{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := query.Preload("Language").Limit(limit).Offset(offset).Find(&translations).Error; err != nil {
		return nil, 0, err
	}

	return translations, total, nil
}

// GetByProjectAndLanguage 根据项目和语言获取翻译
func (r *TranslationRepository) GetByProjectAndLanguage(ctx context.Context, projectID, languageID uint64) ([]*domain.Translation, error) {
	var translations []*domain.Translation
	if err := r.db.WithContext(ctx).Where("project_id = ? AND language_id = ?", projectID, languageID).Find(&translations).Error; err != nil {
		return nil, err
	}
	return translations, nil
}

// GetByProjectKeyLanguage 根据项目ID、键名和语言ID获取翻译
func (r *TranslationRepository) GetByProjectKeyLanguage(ctx context.Context, projectID uint64, keyName string, languageID uint64) (*domain.Translation, error) {
	var translation domain.Translation
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND key_name = ? AND language_id = ?", projectID, keyName, languageID).
		First(&translation).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 返回nil表示未找到，不是错误
		}
		return nil, err
	}

	return &translation, nil
}

// GetByProjectKeyLanguages 批量获取翻译（修复 N+1 查询问题）
func (r *TranslationRepository) GetByProjectKeyLanguages(ctx context.Context, keys []domain.TranslationKey) ([]*domain.Translation, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	// 构建 OR 条件查询所有匹配的翻译
	var conditions []string
	var args []interface{}

	for _, key := range keys {
		conditions = append(conditions, "(project_id = ? AND key_name = ? AND language_id = ?)")
		args = append(args, key.ProjectID, key.KeyName, key.LanguageID)
	}

	var translations []*domain.Translation
	err := r.db.WithContext(ctx).
		Where(strings.Join(conditions, " OR "), args...).
		Find(&translations).Error

	if err != nil {
		return nil, err
	}

	return translations, nil
}

// GetStats 获取全局翻译统计信息（总翻译数和唯一键数）
// 使用聚合查询避免 N+1 问题
func (r *TranslationRepository) GetStats(ctx context.Context) (totalTranslations int, totalKeys int, err error) {
	// 获取总翻译数
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.Translation{}).Count(&count).Error; err != nil {
		return 0, 0, err
	}
	totalTranslations = int(count)

	// 获取唯一键数
	if err := r.db.WithContext(ctx).Model(&domain.Translation{}).Distinct("key_name").Count(&count).Error; err != nil {
		return 0, 0, err
	}
	totalKeys = int(count)

	return totalTranslations, totalKeys, nil
}

// GetMatrix 获取翻译矩阵（key-language映射），支持分页和搜索
func (r *TranslationRepository) GetMatrix(ctx context.Context, projectID uint64, limit, offset int, keyword string) (map[string]map[string]domain.TranslationCell, int64, error) {
	// 优化：使用单个查询获取总数和键名
	var totalCount int64
	var keyNames []string

	// 构建基础查询条件，添加状态过滤提高性能
	baseWhere := "project_id = ? AND status = ?"
	baseArgs := []interface{}{projectID, "active"}

	// 优化关键词搜索查询
	var countQuery *gorm.DB
	if keyword != "" {
		// 优化搜索策略：先尝试精确匹配，再尝试模糊匹配
		// 这样可以更好地利用索引
		searchWhere := baseWhere + " AND (key_name LIKE ? OR value LIKE ?)"
		searchArgs := append(baseArgs, "%"+keyword+"%", "%"+keyword+"%")
		countQuery = r.db.WithContext(ctx).Model(&domain.Translation{}).
			Select("DISTINCT key_name").
			Where(searchWhere, searchArgs...)
	} else {
		countQuery = r.db.WithContext(ctx).Model(&domain.Translation{}).
			Select("DISTINCT key_name").
			Where(baseWhere, baseArgs...)
	}

	// 使用子查询优化计数性能
	var uniqueKeys []string
	if err := countQuery.Pluck("key_name", &uniqueKeys).Error; err != nil {
		return nil, 0, err
	}
	totalCount = int64(len(uniqueKeys))

	// 如果没有数据，直接返回
	if totalCount == 0 {
		return make(map[string]map[string]domain.TranslationCell), 0, nil
	}

	// 应用分页获取实际需要的键名
	if limit > 0 && offset >= 0 {
		end := offset + limit
		if end > len(uniqueKeys) {
			end = len(uniqueKeys)
		}
		if offset < len(uniqueKeys) {
			keyNames = uniqueKeys[offset:end]
		}
	} else {
		keyNames = uniqueKeys
	}

	// 如果分页后没有数据，返回空矩阵
	if len(keyNames) == 0 {
		return make(map[string]map[string]domain.TranslationCell), totalCount, nil
	}

	// 优化：使用JOIN查询避免N+1问题，只查询必要字段
	var results []struct {
		ID           uint64    `gorm:"column:id"`
		KeyName      string    `gorm:"column:key_name"`
		LanguageCode string    `gorm:"column:language_code"`
		Value        string    `gorm:"column:value"`
		UpdatedAt    time.Time `gorm:"column:updated_at"`
	}

	err := r.db.WithContext(ctx).
		Table("translations t").
		Select("t.id, t.key_name, l.code as language_code, t.value, t.updated_at").
		Joins("INNER JOIN languages l ON t.language_id = l.id AND l.status = ?", "active").
		Where("t.project_id = ? AND t.key_name IN ? AND t.status = ?", projectID, keyNames, "active").
		Find(&results).Error

	if err != nil {
		return nil, 0, err
	}

	// 构建矩阵
	matrix := make(map[string]map[string]domain.TranslationCell)
	for _, result := range results {
		if matrix[result.KeyName] == nil {
			matrix[result.KeyName] = make(map[string]domain.TranslationCell)
		}
		matrix[result.KeyName][result.LanguageCode] = domain.TranslationCell{
			ID:        result.ID,
			Value:     result.Value,
			UpdatedAt: result.UpdatedAt,
		}
	}

	return matrix, totalCount, nil
}

// Create 创建翻译
func (r *TranslationRepository) Create(ctx context.Context, translation *domain.Translation) error {
	return r.db.WithContext(ctx).Create(translation).Error
}

// CreateBatch 批量创建翻译
func (r *TranslationRepository) CreateBatch(ctx context.Context, translations []*domain.Translation) error {
	if len(translations) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(translations, 100).Error
}

// Update 更新翻译
func (r *TranslationRepository) Update(ctx context.Context, translation *domain.Translation) error {
	return r.db.WithContext(ctx).Save(translation).Error
}

// Delete 删除翻译
func (r *TranslationRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Translation{}, id).Error
}

// DeleteBatch 批量删除翻译
func (r *TranslationRepository) DeleteBatch(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Delete(&domain.Translation{}, ids).Error
}

// UpsertBatch 批量创建或更新翻译
// 如果翻译已存在（基于唯一索引：project_id + key_name + language_id），则更新
// 如果不存在，则创建
// 使用数据库原生的 UPSERT 能力（MySQL: ON DUPLICATE KEY UPDATE, PostgreSQL: ON CONFLICT DO UPDATE）
func (r *TranslationRepository) UpsertBatch(ctx context.Context, translations []*domain.Translation) error {
	if len(translations) == 0 {
		return nil
	}

	// 使用 GORM 的 OnConflict 子句实现 Upsert
	// 这会根据不同数据库自动生成对应的 SQL：
	// - MySQL: INSERT ... ON DUPLICATE KEY UPDATE
	// - PostgreSQL: INSERT ... ON CONFLICT ... DO UPDATE
	// - SQLite: INSERT ... ON CONFLICT ... DO UPDATE
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			// 基于唯一索引 idx_translation_unique (project_id, key_name, language_id)
			Columns: []clause.Column{
				{Name: "project_id"},
				{Name: "key_name"},
				{Name: "language_id"},
			},
			// 冲突时更新这些字段
			DoUpdates: clause.AssignmentColumns([]string{"value", "context", "updated_at"}),
		}).
		Create(&translations).Error
}
