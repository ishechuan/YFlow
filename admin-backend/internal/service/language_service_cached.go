package service

import (
	"context"
	"yflow/internal/domain"
)

// CachedLanguageService 带缓存的语言服务实现
type CachedLanguageService struct {
	languageService *LanguageService
	cacheService    domain.CacheService
	mutexManager    *CacheMutexManager
}

// NewCachedLanguageService 创建带缓存的语言服务实例
func NewCachedLanguageService(
	languageService *LanguageService,
	cacheService domain.CacheService,
) *CachedLanguageService {
	return &CachedLanguageService{
		languageService: languageService,
		cacheService:    cacheService,
		mutexManager:    NewCacheMutexManager(),
	}
}

// Create 创建语言（更新缓存）
func (s *CachedLanguageService) Create(ctx context.Context, params domain.CreateLanguageParams, userID uint64) (*domain.Language, error) {
	language, err := s.languageService.Create(ctx, params, userID)
	if err != nil {
		return nil, err
	}

	// 清除语言列表缓存
	s.cacheService.Delete(ctx, s.cacheService.GetLanguagesKey())

	// 清除所有项目的翻译矩阵缓存，因为新增语言可能影响所有项目
	s.cacheService.DeleteByPattern(ctx, domain.TranslationMatrixPrefix+"*")

	// 清除仪表板缓存
	s.cacheService.Delete(ctx, s.cacheService.GetDashboardStatsKey())

	return language, nil
}

// GetByID 根据ID获取语言
func (s *CachedLanguageService) GetByID(ctx context.Context, id uint64) (*domain.Language, error) {
	// 单个语言查询不频繁，不进行缓存
	return s.languageService.GetByID(ctx, id)
}

// GetAll 获取所有语言（使用缓存）
func (s *CachedLanguageService) GetAll(ctx context.Context) ([]*domain.Language, error) {
	cacheKey := s.cacheService.GetLanguagesKey()

	// 使用互斥锁防止缓存击穿
	mutex := s.mutexManager.GetMutex(cacheKey)
	mutex.Lock()
	defer func() {
		mutex.Unlock()
		s.mutexManager.RemoveMutex(cacheKey) // 请求完成后移除锁
	}()

	// 尝试从缓存获取
	var languages []*domain.Language
	err := s.cacheService.GetJSONWithEmptyCheck(ctx, cacheKey, &languages)
	if err == nil {
		return languages, nil
	}

	// 缓存未命中，从数据库获取
	languages, err = s.languageService.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// 更新缓存，添加随机过期时间防止雪崩
	expiration := s.cacheService.AddRandomExpiration(domain.DefaultExpiration)
	if err := s.cacheService.SetJSONWithEmptyCache(ctx, cacheKey, languages, expiration); err != nil {
		// 缓存更新失败，但不影响返回结果
	}

	return languages, nil
}

// Update 更新语言（更新缓存）
func (s *CachedLanguageService) Update(ctx context.Context, id uint64, params domain.CreateLanguageParams, userID uint64) (*domain.Language, error) {
	language, err := s.languageService.Update(ctx, id, params, userID)
	if err != nil {
		return nil, err
	}

	// 清除语言列表缓存
	s.cacheService.Delete(ctx, s.cacheService.GetLanguagesKey())

	// 清除所有项目的翻译矩阵缓存，因为语言变更可能影响所有项目
	s.cacheService.DeleteByPattern(ctx, domain.TranslationMatrixPrefix+"*")

	// 清除仪表板缓存
	s.cacheService.Delete(ctx, s.cacheService.GetDashboardStatsKey())

	return language, nil
}

// Delete 删除语言（更新缓存）
func (s *CachedLanguageService) Delete(ctx context.Context, id uint64) error {
	err := s.languageService.Delete(ctx, id)
	if err != nil {
		return err
	}

	// 清除语言列表缓存
	s.cacheService.Delete(ctx, s.cacheService.GetLanguagesKey())

	// 清除所有项目的翻译矩阵缓存，因为删除语言可能影响所有项目
	s.cacheService.DeleteByPattern(ctx, domain.TranslationMatrixPrefix+"*")

	// 清除仪表板缓存
	s.cacheService.Delete(ctx, s.cacheService.GetDashboardStatsKey())

	return nil
}
