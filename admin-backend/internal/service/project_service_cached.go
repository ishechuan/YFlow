package service

import (
	"context"
	"fmt"
	"yflow/internal/domain"
)

// CachedProjectService 带缓存的项目服务实现
type CachedProjectService struct {
	projectService *ProjectService
	cacheService   domain.CacheService
	mutexManager   *CacheMutexManager
}

// NewCachedProjectService 创建带缓存的项目服务实例
func NewCachedProjectService(
	projectService *ProjectService,
	cacheService domain.CacheService,
) *CachedProjectService {
	return &CachedProjectService{
		projectService: projectService,
		cacheService:   cacheService,
		mutexManager:   NewCacheMutexManager(),
	}
}

// Create 创建项目（更新缓存）
func (s *CachedProjectService) Create(ctx context.Context, params domain.CreateProjectParams, userID uint64) (*domain.Project, error) {
	project, err := s.projectService.Create(ctx, params, userID)
	if err != nil {
		return nil, err
	}

	// 清除项目列表缓存（包括所有分页的缓存）
	baseKey := s.cacheService.GetProjectsKey()
	s.cacheService.DeleteByPattern(ctx, baseKey+"*") // 使用通配符删除所有相关缓存

	// 清除仪表板缓存
	s.cacheService.Delete(ctx, s.cacheService.GetDashboardStatsKey())

	return project, nil
}

// GetByID 根据ID获取项目（使用缓存）
func (s *CachedProjectService) GetByID(ctx context.Context, id uint64) (*domain.Project, error) {
	cacheKey := s.cacheService.GetProjectKey(id)

	// 使用互斥锁防止缓存击穿
	mutex := s.mutexManager.GetMutex(cacheKey)
	mutex.Lock()
	defer func() {
		mutex.Unlock()
		s.mutexManager.RemoveMutex(cacheKey) // 请求完成后移除锁
	}()

	// 尝试从缓存获取
	var project *domain.Project
	err := s.cacheService.GetJSONWithEmptyCheck(ctx, cacheKey, &project)
	if err == nil {
		return project, nil
	}

	// 缓存未命中，从数据库获取
	project, err = s.projectService.GetByID(ctx, id)
	if err != nil {
		// 对于不存在的项目，也缓存一小段时间防止缓存穿透
		if err == domain.ErrProjectNotFound {
			expiration := s.cacheService.AddRandomExpiration(domain.ShortExpiration)
			s.cacheService.SetJSONWithEmptyCache(ctx, cacheKey, nil, expiration)
		}
		return nil, err
	}

	// 更新缓存，添加随机过期时间防止雪崩
	expiration := s.cacheService.AddRandomExpiration(domain.DefaultExpiration)
	if err := s.cacheService.SetJSONWithEmptyCache(ctx, cacheKey, project, expiration); err != nil {
		// 缓存更新失败，但不影响返回结果
	}

	return project, nil
}

// GetAll 获取所有项目（使用缓存）
func (s *CachedProjectService) GetAll(ctx context.Context, limit, offset int, keyword string) ([]*domain.Project, int64, error) {
	// 生成缓存键
	cacheKey := s.cacheService.GetProjectsKey()
	if keyword != "" {
		// 如果有搜索关键词，添加到缓存键中
		cacheKey += ":search:" + keyword
	}
	cacheKey += fmt.Sprintf(":%d:%d", limit, offset)

	// 使用互斥锁防止缓存击穿
	mutex := s.mutexManager.GetMutex(cacheKey)
	mutex.Lock()
	defer func() {
		mutex.Unlock()
		s.mutexManager.RemoveMutex(cacheKey) // 请求完成后移除锁
	}()

	// 尝试从缓存获取
	type projectsCacheResult struct {
		Projects []*domain.Project `json:"projects"`
		Total    int64             `json:"total"`
	}

	var cachedResult projectsCacheResult
	err := s.cacheService.GetJSONWithEmptyCheck(ctx, cacheKey, &cachedResult)
	if err == nil {
		return cachedResult.Projects, cachedResult.Total, nil
	}

	// 缓存未命中，从数据库获取
	projects, total, err := s.projectService.GetAll(ctx, limit, offset, keyword)
	if err != nil {
		return nil, 0, err
	}

	// 更新缓存，添加随机过期时间防止雪崩
	cachedResult = projectsCacheResult{
		Projects: projects,
		Total:    total,
	}

	expiration := s.cacheService.AddRandomExpiration(domain.DefaultExpiration)
	if err := s.cacheService.SetJSONWithEmptyCache(ctx, cacheKey, cachedResult, expiration); err != nil {
		// 缓存更新失败，但不影响返回结果
	}

	return projects, total, nil
}

// Update 更新项目（更新缓存）
func (s *CachedProjectService) Update(ctx context.Context, id uint64, params domain.UpdateProjectParams, userID uint64) (*domain.Project, error) {
	project, err := s.projectService.Update(ctx, id, params, userID)
	if err != nil {
		return nil, err
	}

	// 清除该项目的缓存
	s.cacheService.Delete(ctx, s.cacheService.GetProjectKey(id))

	// 清除项目列表缓存（包括所有分页的缓存）
	baseKey := s.cacheService.GetProjectsKey()
	s.cacheService.DeleteByPattern(ctx, baseKey+"*")

	// 清除仪表板缓存
	s.cacheService.Delete(ctx, s.cacheService.GetDashboardStatsKey())

	return project, nil
}

// Delete 删除项目（更新缓存）
func (s *CachedProjectService) Delete(ctx context.Context, id uint64) error {
	err := s.projectService.Delete(ctx, id)
	if err != nil {
		return err
	}

	// 清除该项目的缓存
	s.cacheService.Delete(ctx, s.cacheService.GetProjectKey(id))

	// 清除项目列表缓存（包括所有分页的缓存）
	baseKey := s.cacheService.GetProjectsKey()
	s.cacheService.DeleteByPattern(ctx, baseKey+"*")

	// 清除仪表板缓存
	s.cacheService.Delete(ctx, s.cacheService.GetDashboardStatsKey())

	return nil
}

// GetAccessibleProjects 获取用户可访问的项目列表（不缓存，因为依赖用户权限）
func (s *CachedProjectService) GetAccessibleProjects(ctx context.Context, userID uint64, limit, offset int, keyword string) ([]*domain.Project, int64, error) {
	// 用户权限相关的查询不缓存，直接调用基础服务
	return s.projectService.GetAccessibleProjects(ctx, userID, limit, offset, keyword)
}
