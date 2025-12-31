package repository

import (
	"context"
	"errors"
	"yflow/internal/domain"

	"gorm.io/gorm"
)

// LanguageRepository 语言仓储实现
type LanguageRepository struct {
	db *gorm.DB
}

// NewLanguageRepository 创建语言仓储实例
func NewLanguageRepository(db *gorm.DB) *LanguageRepository {
	return &LanguageRepository{db: db}
}

// GetByID 根据ID获取语言
func (r *LanguageRepository) GetByID(ctx context.Context, id uint64) (*domain.Language, error) {
	var language domain.Language
	if err := r.db.WithContext(ctx).First(&language, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrLanguageNotFound
		}
		return nil, err
	}
	return &language, nil
}

// GetByIDs 批量获取语言
func (r *LanguageRepository) GetByIDs(ctx context.Context, ids []uint64) ([]*domain.Language, error) {
	if len(ids) == 0 {
		return []*domain.Language{}, nil
	}

	var languages []*domain.Language
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&languages).Error; err != nil {
		return nil, err
	}
	return languages, nil
}

// GetByCode 根据代码获取语言
func (r *LanguageRepository) GetByCode(ctx context.Context, code string) (*domain.Language, error) {
	var language domain.Language
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&language).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrLanguageNotFound
		}
		return nil, err
	}
	return &language, nil
}

// GetAll 获取所有语言
func (r *LanguageRepository) GetAll(ctx context.Context) ([]*domain.Language, error) {
	var languages []*domain.Language
	if err := r.db.WithContext(ctx).Find(&languages).Error; err != nil {
		return nil, err
	}
	return languages, nil
}

// Create 创建语言
func (r *LanguageRepository) Create(ctx context.Context, language *domain.Language) error {
	return r.db.WithContext(ctx).Create(language).Error
}

// Update 更新语言
func (r *LanguageRepository) Update(ctx context.Context, language *domain.Language) error {
	return r.db.WithContext(ctx).Save(language).Error
}

// Delete 删除语言
func (r *LanguageRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Language{}, id).Error
}

// GetDefault 获取默认语言
func (r *LanguageRepository) GetDefault(ctx context.Context) (*domain.Language, error) {
	var language domain.Language
	if err := r.db.WithContext(ctx).Where("is_default = ?", true).First(&language).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrLanguageNotFound
		}
		return nil, err
	}
	return &language, nil
}
