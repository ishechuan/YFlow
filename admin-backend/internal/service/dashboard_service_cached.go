package service

import (
	"context"
	"yflow/internal/domain"
)

// CachedDashboardService 带缓存的仪表板服务实现
type CachedDashboardService struct {
	dashboardService *DashboardService
	cacheService     domain.CacheService
	mutexManager     *CacheMutexManager
}

// NewCachedDashboardService 创建带缓存的仪表板服务实例
func NewCachedDashboardService(
	dashboardService *DashboardService,
	cacheService domain.CacheService,
) *CachedDashboardService {
	return &CachedDashboardService{
		dashboardService: dashboardService,
		cacheService:     cacheService,
		mutexManager:     NewCacheMutexManager(),
	}
}

// GetStats 获取仪表板统计信息（使用缓存）
func (s *CachedDashboardService) GetStats(ctx context.Context) (*domain.DashboardStats, error) {
	cacheKey := s.cacheService.GetDashboardStatsKey()

	// 使用互斥锁防止缓存击穿
	mutex := s.mutexManager.GetMutex(cacheKey)
	mutex.Lock()
	defer func() {
		mutex.Unlock()
		s.mutexManager.RemoveMutex(cacheKey) // 请求完成后移除锁
	}()

	// 尝试从缓存获取
	var stats *domain.DashboardStats
	err := s.cacheService.GetJSONWithEmptyCheck(ctx, cacheKey, &stats)
	if err == nil {
		return stats, nil
	}

	// 缓存未命中，从数据库获取
	stats, err = s.dashboardService.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	// 更新缓存，添加随机过期时间防止雪崩
	expiration := s.cacheService.AddRandomExpiration(domain.LongExpiration)
	if err := s.cacheService.SetJSONWithEmptyCache(ctx, cacheKey, stats, expiration); err != nil {
		// 缓存更新失败，但不影响返回结果
	}

	return stats, nil
}
