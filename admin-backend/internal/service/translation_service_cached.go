package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"yflow/internal/domain"
)

// CachedTranslationService 带缓存的翻译服务实现
type CachedTranslationService struct {
	translationService *TranslationService
	cacheService       domain.CacheService
	mutexManager       *CacheMutexManager
}

// NewCachedTranslationService 创建带缓存的翻译服务实例
func NewCachedTranslationService(
	translationService *TranslationService,
	cacheService domain.CacheService,
) *CachedTranslationService {
	return &CachedTranslationService{
		translationService: translationService,
		cacheService:       cacheService,
		mutexManager:       NewCacheMutexManager(),
	}
}

// Create 创建翻译（更新缓存）
func (s *CachedTranslationService) Create(ctx context.Context, input domain.TranslationInput, userID uint64) (*domain.Translation, error) {
	translation, err := s.translationService.Create(ctx, input, userID)
	if err != nil {
		return nil, err
	}

	// 清除相关缓存
	s.invalidateProjectCache(ctx, input.ProjectID)

	return translation, nil
}

// CreateBatch 批量创建翻译（更新缓存）
func (s *CachedTranslationService) CreateBatch(ctx context.Context, inputs []domain.TranslationInput) error {
	err := s.translationService.CreateBatch(ctx, inputs)
	if err != nil {
		return err
	}

	// 清除相关缓存
	projectIDs := make(map[uint64]bool)
	for _, input := range inputs {
		projectIDs[input.ProjectID] = true
	}

	for projectID := range projectIDs {
		s.invalidateProjectCache(ctx, projectID)
	}

	return nil
}

// CreateBatchFromRequest 从批量翻译参数创建翻译（更新缓存）
func (s *CachedTranslationService) CreateBatchFromRequest(ctx context.Context, params domain.BatchTranslationParams) error {
	err := s.translationService.CreateBatchFromRequest(ctx, params)
	if err != nil {
		return err
	}

	// 清除相关缓存
	s.invalidateProjectCache(ctx, params.ProjectID)

	return nil
}

// UpsertBatch 批量创建或更新翻译（更新缓存）
func (s *CachedTranslationService) UpsertBatch(ctx context.Context, inputs []domain.TranslationInput) error {
	err := s.translationService.UpsertBatch(ctx, inputs)
	if err != nil {
		return err
	}

	// 清除相关缓存
	projectIDs := make(map[uint64]bool)
	for _, input := range inputs {
		projectIDs[input.ProjectID] = true
	}

	for projectID := range projectIDs {
		s.invalidateProjectCache(ctx, projectID)
	}

	return nil
}

// GetByID 根据ID获取翻译
func (s *CachedTranslationService) GetByID(ctx context.Context, id uint64) (*domain.Translation, error) {
	// 这个方法不缓存，因为单个翻译查询不频繁
	return s.translationService.GetByID(ctx, id)
}

// TranslationCacheResult 定义翻译缓存结果结构体
type TranslationCacheResult struct {
	Translations []*domain.Translation `json:"translations"`
	Total        int64                 `json:"total"`
}

// GetByProjectID 根据项目ID获取翻译（使用缓存）
func (s *CachedTranslationService) GetByProjectID(ctx context.Context, projectID uint64, limit, offset int) ([]*domain.Translation, int64, error) {
	// 生成缓存键
	cacheKey := fmt.Sprintf("%s:%d:%d", s.cacheService.GetTranslationKey(projectID), limit, offset)

	// 使用互斥锁防止缓存击穿
	mutex := s.mutexManager.GetMutex(cacheKey)
	mutex.Lock()
	defer func() {
		mutex.Unlock()
		s.mutexManager.RemoveMutex(cacheKey) // 请求完成后移除锁
	}()

	// 尝试从缓存获取
	var cachedResult TranslationCacheResult
	err := s.cacheService.GetJSONWithEmptyCheck(ctx, cacheKey, &cachedResult)
	if err == nil {
		return cachedResult.Translations, cachedResult.Total, nil
	}

	// 缓存未命中，从数据库获取
	translations, total, err := s.translationService.GetByProjectID(ctx, projectID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// 更新缓存，添加随机过期时间防止雪崩
	cachedResult = TranslationCacheResult{
		Translations: translations,
		Total:        total,
	}

	expiration := s.cacheService.AddRandomExpiration(domain.DefaultExpiration)
	if err := s.cacheService.SetJSONWithEmptyCache(ctx, cacheKey, cachedResult, expiration); err != nil {
		// 缓存更新失败，但不影响返回结果
	}

	return translations, total, nil
}

// MatrixCacheResult 定义缓存结果结构体
type MatrixCacheResult struct {
	Matrix map[string]map[string]domain.TranslationCell `json:"matrix"`
	Total  int64                                        `json:"total"`
}

// GetMatrix 获取翻译矩阵（使用缓存）
func (s *CachedTranslationService) GetMatrix(ctx context.Context, projectID uint64, limit, offset int, keyword string) (map[string]map[string]domain.TranslationCell, int64, error) {
	// 优化缓存键生成，区分搜索和非搜索查询
	var cacheKey string
	if keyword != "" {
		// 搜索查询使用较短的缓存时间
		cacheKey = fmt.Sprintf("%s:search:%s:%d:%d", s.cacheService.GetTranslationMatrixKey(projectID, ""), s.hashKeyword(keyword), limit, offset)
	} else {
		// 非搜索查询使用较长的缓存时间
		cacheKey = fmt.Sprintf("%s:all:%d:%d", s.cacheService.GetTranslationMatrixKey(projectID, ""), limit, offset)
	}

	// 使用互斥锁防止缓存击穿
	mutex := s.mutexManager.GetMutex(cacheKey)
	mutex.Lock()
	defer func() {
		mutex.Unlock()
		s.mutexManager.RemoveMutex(cacheKey) // 请求完成后移除锁
	}()

	// 尝试从缓存获取
	var cachedResult MatrixCacheResult
	err := s.cacheService.GetJSONWithEmptyCheck(ctx, cacheKey, &cachedResult)
	if err == nil {
		return cachedResult.Matrix, cachedResult.Total, nil
	}

	// 缓存未命中，从数据库获取
	matrix, total, err := s.translationService.GetMatrix(ctx, projectID, limit, offset, keyword)
	if err != nil {
		return nil, 0, err
	}

	// 更新缓存，添加随机过期时间防止雪崩
	cachedResult = MatrixCacheResult{
		Matrix: matrix,
		Total:  total,
	}

	// 根据查询类型设置不同的缓存时间
	var expiration time.Duration
	if keyword != "" {
		// 搜索查询缓存较短时间
		expiration = s.cacheService.AddRandomExpiration(5 * time.Minute)
	} else {
		// 非搜索查询缓存较长时间
		expiration = s.cacheService.AddRandomExpiration(domain.DefaultExpiration)
	}

	if err := s.cacheService.SetJSONWithEmptyCache(ctx, cacheKey, cachedResult, expiration); err != nil {
		// 缓存更新失败，但不影响返回结果
	}

	return matrix, total, nil
}

// Update 更新翻译（更新缓存）
func (s *CachedTranslationService) Update(ctx context.Context, id uint64, input domain.TranslationInput, userID uint64) (*domain.Translation, error) {
	// 先获取原始翻译，用于后续清除缓存
	oldTranslation, err := s.translationService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	translation, err := s.translationService.Update(ctx, id, input, userID)
	if err != nil {
		return nil, err
	}

	// 清除相关缓存
	s.invalidateProjectCache(ctx, oldTranslation.ProjectID)
	if input.ProjectID != 0 && input.ProjectID != oldTranslation.ProjectID {
		s.invalidateProjectCache(ctx, input.ProjectID)
	}

	return translation, nil
}

// Delete 删除翻译（更新缓存）
func (s *CachedTranslationService) Delete(ctx context.Context, id uint64, userID uint64) error {
	// 先获取翻译，用于后续清除缓存
	translation, err := s.translationService.GetByID(ctx, id)
	if err != nil {
		return err
	}

	err = s.translationService.Delete(ctx, id, userID)
	if err != nil {
		return err
	}

	// 清除相关缓存
	s.invalidateProjectCache(ctx, translation.ProjectID)

	return nil
}

// DeleteBatch 批量删除翻译（更新缓存）
func (s *CachedTranslationService) DeleteBatch(ctx context.Context, ids []uint64) error {
	// 这里需要先查询所有翻译，获取相关的项目ID
	projectIDs := make(map[uint64]bool)
	for _, id := range ids {
		translation, err := s.translationService.GetByID(ctx, id)
		if err == nil {
			projectIDs[translation.ProjectID] = true
		}
	}

	err := s.translationService.DeleteBatch(ctx, ids)
	if err != nil {
		return err
	}

	// 清除相关缓存
	for projectID := range projectIDs {
		s.invalidateProjectCache(ctx, projectID)
	}

	// 清除仪表板缓存
	s.cacheService.Delete(ctx, s.cacheService.GetDashboardStatsKey())

	return nil
}

// Export 导出翻译
func (s *CachedTranslationService) Export(ctx context.Context, projectID uint64, format string) ([]byte, error) {
	// 使用缓存的矩阵数据
	matrix, _, err := s.GetMatrix(ctx, projectID, -1, 0, "")
	if err != nil {
		return nil, err
	}

	// 转换为简单格式 (key -> language -> value)
	simpleMatrix := make(map[string]map[string]string)
	for key, langs := range matrix {
		simpleMatrix[key] = make(map[string]string)
		for lang, cell := range langs {
			simpleMatrix[key][lang] = cell.Value
		}
	}

	switch format {
	case "json":
		return json.MarshalIndent(simpleMatrix, "", "  ")
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// Import 导入翻译（更新缓存）
func (s *CachedTranslationService) Import(ctx context.Context, projectID uint64, data []byte, format string) error {
	err := s.translationService.Import(ctx, projectID, data, format)
	if err != nil {
		return err
	}

	// 清除相关缓存
	s.invalidateProjectCache(ctx, projectID)

	return nil
}

// invalidateProjectCache 清除项目相关的所有缓存
func (s *CachedTranslationService) invalidateProjectCache(ctx context.Context, projectID uint64) {
	// 使用管道操作提高性能
	// 清除翻译列表缓存
	s.cacheService.DeleteByPattern(ctx, s.cacheService.GetTranslationKey(projectID)+"*")

	// 清除翻译矩阵缓存
	s.cacheService.DeleteByPattern(ctx, s.cacheService.GetTranslationMatrixKey(projectID, "")+"*")

	// 清除仪表板缓存
	s.cacheService.Delete(ctx, s.cacheService.GetDashboardStatsKey())
}

// invalidateLanguageCache 清除语言相关的缓存（当语言被修改时调用）
func (s *CachedTranslationService) invalidateLanguageCache(ctx context.Context) {
	// 清除所有项目的翻译矩阵缓存，因为语言变更可能影响所有项目
	s.cacheService.DeleteByPattern(ctx, domain.TranslationMatrixPrefix+"*")

	// 清除仪表板缓存
	s.cacheService.Delete(ctx, s.cacheService.GetDashboardStatsKey())
}

// invalidateSpecificTranslationCache 清除特定翻译键的缓存
func (s *CachedTranslationService) invalidateSpecificTranslationCache(ctx context.Context, projectID uint64, keyName string) {
	// 清除包含特定键名的翻译矩阵缓存
	pattern := fmt.Sprintf("%s%d:*%s*", domain.TranslationMatrixPrefix, projectID, keyName)
	s.cacheService.DeleteByPattern(ctx, pattern)

	// 清除仪表板缓存
	s.cacheService.Delete(ctx, s.cacheService.GetDashboardStatsKey())
}

// hashKeyword 对关键词进行简单哈希，避免缓存键过长
func (s *CachedTranslationService) hashKeyword(keyword string) string {
	// 简单的哈希函数，生产环境可以使用更复杂的哈希
	hash := 0
	for _, char := range keyword {
		hash = 31*hash + int(char)
	}
	return strconv.Itoa(hash)
}
