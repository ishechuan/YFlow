package service

import (
	"context"
	"yflow/internal/domain"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// UserService 用户服务实现
type UserService struct {
	userRepo    domain.UserRepository
	authService domain.AuthService
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo domain.UserRepository, authService domain.AuthService) *UserService {
	return &UserService{
		userRepo:    userRepo,
		authService: authService,
	}
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, params domain.LoginParams) (*domain.LoginResult, error) {
	// 查询用户
	user, err := s.userRepo.GetByUsername(ctx, params.Username)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(params.Password))
	if err != nil {
		return nil, domain.ErrInvalidPassword
	}

	// 生成JWT token
	token, err := s.authService.GenerateToken(ctx, user)
	if err != nil {
		return nil, err
	}

	// 生成刷新token
	refreshToken, err := s.authService.GenerateRefreshToken(ctx, user)
	if err != nil {
		return nil, err
	}

	// 不返回密码
	userResponse := *user
	userResponse.Password = ""

	return &domain.LoginResult{
		User:         &userResponse,
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken 刷新token
func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*domain.LoginResult, error) {
	// 验证刷新token
	userFromToken, err := s.authService.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	// 查询用户确保用户仍然存在
	user, err := s.userRepo.GetByID(ctx, userFromToken.ID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	// 生成新token
	token, err := s.authService.GenerateToken(ctx, user)
	if err != nil {
		return nil, err
	}

	// 生成新刷新token
	newRefreshToken, err := s.authService.GenerateRefreshToken(ctx, user)
	if err != nil {
		return nil, err
	}

	// 不返回密码
	userResponse := *user
	userResponse.Password = ""

	return &domain.LoginResult{
		User:         &userResponse,
		AccessToken:  token,
		RefreshToken: newRefreshToken,
	}, nil
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(ctx context.Context, userID uint64) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 不返回密码
	user.Password = ""
	return user, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, params domain.CreateUserParams) (*domain.User, error) {
	// 检查用户名是否已存在
	if _, err := s.userRepo.GetByUsername(ctx, params.Username); err == nil {
		return nil, domain.ErrUserExists
	}

	// 检查邮箱是否已存在
	if params.Email != "" {
		if _, err := s.userRepo.GetByEmail(ctx, params.Email); err == nil {
			return nil, domain.ErrEmailExists
		}
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username: params.Username,
		Email:    params.Email,
		Password: string(hashedPassword),
		Role:     params.Role,
		Status:   "active",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 不返回密码
	user.Password = ""
	return user, nil
}

// GetAllUsers 获取用户列表
func (s *UserService) GetAllUsers(ctx context.Context, limit, offset int, keyword string) ([]*domain.User, int64, error) {
	users, total, err := s.userRepo.GetAll(ctx, limit, offset, keyword)
	if err != nil {
		return nil, 0, err
	}

	// 清除密码字段
	for _, user := range users {
		user.Password = ""
	}

	return users, total, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(ctx context.Context, id uint64) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 不返回密码
	user.Password = ""
	return user, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(ctx context.Context, id uint64, params domain.UpdateUserParams) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if params.Username != "" && params.Username != user.Username {
		// 检查新用户名是否已存在
		if _, err := s.userRepo.GetByUsername(ctx, params.Username); err == nil {
			return nil, domain.ErrUserExists
		}
		user.Username = params.Username
	}

	if params.Email != "" && params.Email != user.Email {
		// 检查新邮箱是否已存在
		if _, err := s.userRepo.GetByEmail(ctx, params.Email); err == nil {
			return nil, domain.ErrEmailExists
		}
		user.Email = params.Email
	}

	if params.Role != "" {
		user.Role = params.Role
	}

	if params.Status != "" {
		user.Status = params.Status
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// 不返回密码
	user.Password = ""
	return user, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, userID uint64, params domain.ChangePasswordParams) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(params.OldPassword)); err != nil {
		return domain.ErrInvalidPassword
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	return s.userRepo.Update(ctx, user)
}

// ResetPassword 重置用户密码（管理员功能）
func (s *UserService) ResetPassword(ctx context.Context, userID uint64, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	return s.userRepo.Update(ctx, user)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, id uint64) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 不能删除管理员用户
	if strings.ToLower(user.Role) == "admin" {
		return domain.ErrCannotDeleteAdmin
	}

	return s.userRepo.Delete(ctx, id)
}
