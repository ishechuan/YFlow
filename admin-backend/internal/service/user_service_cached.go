package service

import (
	"context"
	"fmt"
	"yflow/internal/domain"
)

// CachedUserService 带缓存的用户服务实现
type CachedUserService struct {
	userService  *UserService
	cacheService domain.CacheService
	mutexManager *CacheMutexManager
}

// NewCachedUserService 创建带缓存的用户服务实例
func NewCachedUserService(
	userService *UserService,
	cacheService domain.CacheService,
) *CachedUserService {
	return &CachedUserService{
		userService:  userService,
		cacheService: cacheService,
		mutexManager: NewCacheMutexManager(),
	}
}

// Login 用户登录
func (s *CachedUserService) Login(ctx context.Context, params domain.LoginParams) (*domain.LoginResult, error) {
	// 登录操作不缓存，直接调用基础服务
	return s.userService.Login(ctx, params)
}

// RefreshToken 刷新token
func (s *CachedUserService) RefreshToken(ctx context.Context, refreshToken string) (*domain.LoginResult, error) {
	// 刷新token操作不缓存，直接调用基础服务
	return s.userService.RefreshToken(ctx, refreshToken)
}

// GetUserInfo 获取用户信息（使用缓存）
func (s *CachedUserService) GetUserInfo(ctx context.Context, userID uint64) (*domain.User, error) {
	cacheKey := fmt.Sprintf("user:%d", userID)

	// 使用互斥锁防止缓存击穿
	mutex := s.mutexManager.GetMutex(cacheKey)
	mutex.Lock()
	defer func() {
		mutex.Unlock()
		s.mutexManager.RemoveMutex(cacheKey) // 请求完成后移除锁
	}()

	// 尝试从缓存获取
	var user *domain.User
	err := s.cacheService.GetJSONWithEmptyCheck(ctx, cacheKey, &user)
	if err == nil {
		return user, nil
	}

	// 缓存未命中，从数据库获取
	user, err = s.userService.GetUserInfo(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 更新缓存，添加随机过期时间防止雪崩
	expiration := s.cacheService.AddRandomExpiration(domain.DefaultExpiration)
	if err := s.cacheService.SetJSONWithEmptyCache(ctx, cacheKey, user, expiration); err != nil {
		// 缓存更新失败，但不影响返回结果
	}

	return user, nil
}

// CreateUser 创建用户（不缓存）
func (s *CachedUserService) CreateUser(ctx context.Context, params domain.CreateUserParams) (*domain.User, error) {
	return s.userService.CreateUser(ctx, params)
}

// GetAllUsers 获取用户列表（不缓存）
func (s *CachedUserService) GetAllUsers(ctx context.Context, limit, offset int, keyword string) ([]*domain.User, int64, error) {
	return s.userService.GetAllUsers(ctx, limit, offset, keyword)
}

// GetUserByID 根据ID获取用户（使用缓存）
func (s *CachedUserService) GetUserByID(ctx context.Context, id uint64) (*domain.User, error) {
	// 复用GetUserInfo的缓存逻辑
	return s.GetUserInfo(ctx, id)
}

// UpdateUser 更新用户（清除缓存）
func (s *CachedUserService) UpdateUser(ctx context.Context, id uint64, params domain.UpdateUserParams) (*domain.User, error) {
	user, err := s.userService.UpdateUser(ctx, id, params)
	if err != nil {
		return nil, err
	}

	// 清除用户缓存
	cacheKey := fmt.Sprintf("user:%d", id)
	s.cacheService.Delete(ctx, cacheKey)

	return user, nil
}

// ChangePassword 修改密码（不缓存）
func (s *CachedUserService) ChangePassword(ctx context.Context, userID uint64, params domain.ChangePasswordParams) error {
	return s.userService.ChangePassword(ctx, userID, params)
}

// ResetPassword 重置密码（不缓存）
func (s *CachedUserService) ResetPassword(ctx context.Context, userID uint64, newPassword string) error {
	return s.userService.ResetPassword(ctx, userID, newPassword)
}

// DeleteUser 删除用户（清除缓存）
func (s *CachedUserService) DeleteUser(ctx context.Context, id uint64) error {
	err := s.userService.DeleteUser(ctx, id)
	if err != nil {
		return err
	}

	// 清除用户缓存
	cacheKey := fmt.Sprintf("user:%d", id)
	s.cacheService.Delete(ctx, cacheKey)

	return nil
}
