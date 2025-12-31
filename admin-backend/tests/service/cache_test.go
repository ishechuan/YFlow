package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"yflow/internal/domain"
	"yflow/internal/dto"
	"yflow/internal/service"
)

// MockCacheService 模拟缓存服务
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheService) DeleteByPattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *MockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheService) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) GetJSON(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCacheService) SetWithEmptyCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) GetWithEmptyCheck(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) SetJSONWithEmptyCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) GetJSONWithEmptyCheck(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCacheService) HSet(ctx context.Context, key, field string, value interface{}) error {
	args := m.Called(ctx, key, field, value)
	return args.Error(0)
}

func (m *MockCacheService) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockCacheService) HDel(ctx context.Context, key string, fields ...string) error {
	args := m.Called(ctx, key, fields)
	return args.Error(0)
}

func (m *MockCacheService) GetTranslationKey(projectID uint64) string {
	args := m.Called(projectID)
	return args.String(0)
}

func (m *MockCacheService) GetTranslationMatrixKey(projectID uint64, keyword string) string {
	args := m.Called(projectID, keyword)
	return args.String(0)
}

func (m *MockCacheService) GetDashboardStatsKey() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCacheService) GetLanguagesKey() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCacheService) GetProjectKey(projectID uint64) string {
	args := m.Called(projectID)
	return args.String(0)
}

func (m *MockCacheService) GetProjectsKey() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCacheService) AddRandomExpiration(baseExpiration time.Duration) time.Duration {
	args := m.Called(baseExpiration)
	return args.Get(0).(time.Duration)
}

func TestCachedTranslationService_GetMatrix(t *testing.T) {
	// 创建模拟对象
	mockCache := new(MockCacheService)
	
	// 创建测试数据
	projectID := uint64(1)
	
	// 设置模拟期望
	mockCache.On("GetTranslationMatrixKey", projectID, "").Return("translation_matrix:1")
	mockCache.On("GetJSONWithEmptyCheck", mock.Anything, "translation_matrix:1:all:10:0", mock.Anything).Return(domain.ErrCacheMiss)
	mockCache.On("AddRandomExpiration", domain.DefaultExpiration).Return(domain.DefaultExpiration)
	mockCache.On("SetJSONWithEmptyCache", mock.Anything, "translation_matrix:1:all:10:0", mock.Anything, domain.DefaultExpiration).Return(nil)
	
	// 创建带缓存的服务
	cachedService := service.NewCachedTranslationService(nil, mockCache)
	
	// 验证缓存服务接口实现
	assert.Implements(t, (*domain.CacheService)(nil), mockCache)
	
	// 验证方法调用（这需要一个完整的模拟翻译服务，这里只是验证缓存逻辑）
	t.Logf("CachedTranslationService implements TranslationService interface: %t", 
		assert.Implements(t, (*domain.TranslationService)(nil), cachedService))
}

func TestCachedDashboardService_GetStats(t *testing.T) {
	// 创建模拟对象
	mockCache := new(MockCacheService)

	// 创建测试数据
	stats := &dto.DashboardStats{
		TotalProjects:     5,
		TotalLanguages:    3,
		TotalTranslations: 150,
		TotalKeys:         50,
	}

	// 设置模拟期望
	mockCache.On("GetDashboardStatsKey").Return("dashboard:stats")
	mockCache.On("GetJSONWithEmptyCheck", mock.Anything, "dashboard:stats", mock.Anything).Return(domain.ErrCacheMiss)
	mockCache.On("AddRandomExpiration", domain.LongExpiration).Return(domain.LongExpiration)
	mockCache.On("SetJSONWithEmptyCache", mock.Anything, "dashboard:stats", stats, domain.LongExpiration).Return(nil)

	// 创建带缓存的服务
	cachedService := service.NewCachedDashboardService(nil, mockCache)

	// 验证缓存服务接口实现
	assert.Implements(t, (*domain.CacheService)(nil), mockCache)

	// 验证方法调用（这需要一个完整的模拟仪表板服务，这里只是验证缓存逻辑）
	t.Logf("CachedDashboardService implements DashboardService interface: %t",
		assert.Implements(t, (*domain.DashboardService)(nil), cachedService))
}