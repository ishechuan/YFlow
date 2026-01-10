package repository

import (
	"context"
	"errors"
	"time"
	"yflow/internal/domain"

	"gorm.io/gorm"
)

// TranslationHistoryRepository 翻译历史仓储实现
type TranslationHistoryRepository struct {
	db *gorm.DB
}

// NewTranslationHistoryRepository 创建翻译历史仓储实例
func NewTranslationHistoryRepository(db *gorm.DB) *TranslationHistoryRepository {
	return &TranslationHistoryRepository{db: db}
}

// Create 创建翻译历史记录
func (r *TranslationHistoryRepository) Create(ctx context.Context, history *domain.TranslationHistory) error {
	if history == nil {
		return errors.New("history cannot be nil")
	}

	// 设置操作时间
	if history.OperatedAt.IsZero() {
		history.OperatedAt = time.Now()
	}

	return r.db.WithContext(ctx).Create(history).Error
}

// CreateBatch 批量创建翻译历史记录
func (r *TranslationHistoryRepository) CreateBatch(ctx context.Context, histories []*domain.TranslationHistory) error {
	if len(histories) == 0 {
		return nil
	}

	// 设置操作时间
	now := time.Now()
	for _, history := range histories {
		if history.OperatedAt.IsZero() {
			history.OperatedAt = now
		}
	}

	return r.db.WithContext(ctx).CreateInBatches(histories, 100).Error
}

// ListByTranslationID 根据翻译ID获取历史记录
func (r *TranslationHistoryRepository) ListByTranslationID(ctx context.Context, translationID uint64, limit, offset int) ([]*domain.TranslationHistory, int64, error) {
	var histories []*domain.TranslationHistory
	var total int64

	// 构建查询条件
	query := r.db.WithContext(ctx).Where("translation_id = ?", translationID)

	// 计算总数
	if err := query.Model(&domain.TranslationHistory{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// 按操作时间倒序排列
	if err := query.Order("operated_at DESC").Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// ListByProjectID 根据项目ID获取历史记录
func (r *TranslationHistoryRepository) ListByProjectID(ctx context.Context, projectID uint64, params domain.TranslationHistoryQueryParams) ([]*domain.TranslationHistory, int64, error) {
	var histories []*domain.TranslationHistory
	var total int64

	// 构建基础查询条件
	query := r.db.WithContext(ctx).Where("project_id = ?", projectID)

	// 应用筛选条件
	query = applyTranslationHistoryFilters(query, params)

	// 计算总数
	if err := query.Model(&domain.TranslationHistory{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	// 按操作时间倒序排列
	if err := query.Order("operated_at DESC").Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// ListByUserID 根据用户ID获取历史记录
func (r *TranslationHistoryRepository) ListByUserID(ctx context.Context, userID uint64, params domain.TranslationHistoryQueryParams) ([]*domain.TranslationHistory, int64, error) {
	var histories []*domain.TranslationHistory
	var total int64

	// 构建基础查询条件
	query := r.db.WithContext(ctx).Where("operated_by = ?", userID)

	// 应用筛选条件
	query = applyTranslationHistoryFilters(query, params)

	// 计算总数
	if err := query.Model(&domain.TranslationHistory{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	// 按操作时间倒序排列
	if err := query.Order("operated_at DESC").Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// applyTranslationHistoryFilters 应用翻译历史筛选条件
func applyTranslationHistoryFilters(query *gorm.DB, params domain.TranslationHistoryQueryParams) *gorm.DB {
	// 操作类型筛选
	if params.Operation != "" {
		query = query.Where("operation = ?", params.Operation)
	}

	// 时间范围筛选
	if params.StartDate != "" {
		startTime, err := time.Parse("2006-01-02", params.StartDate)
		if err == nil {
			query = query.Where("operated_at >= ?", startTime)
		}
	}

	if params.EndDate != "" {
		endTime, err := time.Parse("2006-01-02", params.EndDate)
		if err == nil {
			// 结束时间包含当天
			endTime = endTime.Add(24 * time.Hour)
			query = query.Where("operated_at < ?", endTime)
		}
	}

	return query
}
