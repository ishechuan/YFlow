package service

import (
	"context"
	"yflow/internal/domain"
)

// ProjectMemberService 项目成员服务实现
type ProjectMemberService struct {
	memberRepo  domain.ProjectMemberRepository
	userRepo    domain.UserRepository
	projectRepo domain.ProjectRepository
}

// NewProjectMemberService 创建项目成员服务实例
func NewProjectMemberService(
	memberRepo domain.ProjectMemberRepository,
	userRepo domain.UserRepository,
	projectRepo domain.ProjectRepository,
) *ProjectMemberService {
	return &ProjectMemberService{
		memberRepo:  memberRepo,
		userRepo:    userRepo,
		projectRepo: projectRepo,
	}
}

// AddMember 添加项目成员
func (s *ProjectMemberService) AddMember(ctx context.Context, projectID uint64, params domain.AddMemberParams, createdBy uint64) (*domain.ProjectMember, error) {
	// 检查项目是否存在
	if _, err := s.projectRepo.GetByID(ctx, projectID); err != nil {
		return nil, err
	}

	// 检查用户是否存在
	if _, err := s.userRepo.GetByID(ctx, params.MemberUserID); err != nil {
		return nil, err
	}

	// 检查用户是否已是项目成员
	if _, err := s.memberRepo.GetByProjectAndUser(ctx, projectID, params.MemberUserID); err == nil {
		return nil, domain.ErrMemberExists
	}

	member := &domain.ProjectMember{
		ProjectID: projectID,
		UserID:    params.MemberUserID,
		Role:      params.Role,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}

	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// GetProjectMembers 获取项目成员列表
func (s *ProjectMemberService) GetProjectMembers(ctx context.Context, projectID uint64) ([]*domain.ProjectMemberInfo, error) {
	// 检查项目是否存在
	if _, err := s.projectRepo.GetByID(ctx, projectID); err != nil {
		return nil, err
	}

	members, err := s.memberRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return []*domain.ProjectMemberInfo{}, nil
	}

	// 批量获取用户信息 (修复 N+1 查询)
	userIDs := make([]uint64, len(members))
	for i, member := range members {
		userIDs[i] = member.UserID
	}
	users, err := s.userRepo.GetByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	// 构建用户ID到用户的映射
	userMap := make(map[uint64]*domain.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// 构建成员信息
	var memberInfos []*domain.ProjectMemberInfo
	for _, member := range members {
		user, exists := userMap[member.UserID]
		if !exists {
			continue // 跳过不存在的用户
		}

		memberInfo := &domain.ProjectMemberInfo{
			ID:       member.ID,
			UserID:   member.UserID,
			Username: user.Username,
			Email:    user.Email,
			Role:     member.Role,
		}
		memberInfos = append(memberInfos, memberInfo)
	}

	return memberInfos, nil
}

// GetUserProjects 获取用户参与的项目列表
func (s *ProjectMemberService) GetUserProjects(ctx context.Context, userID uint64) ([]*domain.Project, error) {
	// 检查用户是否存在
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return nil, err
	}

	members, err := s.memberRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return []*domain.Project{}, nil
	}

	// 批量获取项目信息 (修复 N+1 查询)
	projectIDs := make([]uint64, len(members))
	for i, member := range members {
		projectIDs[i] = member.ProjectID
	}
	projects, err := s.projectRepo.GetByIDs(ctx, projectIDs)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// UpdateMemberRole 更新成员角色
func (s *ProjectMemberService) UpdateMemberRole(ctx context.Context, projectID, userID uint64, params domain.UpdateMemberRoleParams) (*domain.ProjectMember, error) {
	member, err := s.memberRepo.GetByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	member.Role = params.Role
	if err := s.memberRepo.Update(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// RemoveMember 移除项目成员
func (s *ProjectMemberService) RemoveMember(ctx context.Context, projectID, userID uint64) error {
	// 检查成员是否存在
	member, err := s.memberRepo.GetByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		return err
	}

	// 不能移除项目所有者
	if member.Role == "owner" {
		return domain.ErrCannotRemoveOwner
	}

	return s.memberRepo.Delete(ctx, projectID, userID)
}

// CheckPermission 检查用户权限
func (s *ProjectMemberService) CheckPermission(ctx context.Context, userID, projectID uint64, requiredRole string) (bool, error) {
	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}

	// 管理员拥有所有权限
	if user.Role == "admin" {
		return true, nil
	}

	// 获取用户在项目中的角色
	member, err := s.memberRepo.GetByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		return false, nil // 用户不是项目成员
	}

	// 角色权限层级：owner > editor > viewer
	roleLevel := map[string]int{
		"viewer": 1,
		"editor": 2,
		"owner":  3,
	}

	userLevel, exists := roleLevel[member.Role]
	if !exists {
		return false, nil
	}

	requiredLevel, exists := roleLevel[requiredRole]
	if !exists {
		return false, nil
	}

	return userLevel >= requiredLevel, nil
}

// GetMemberRole 获取用户在项目中的角色
func (s *ProjectMemberService) GetMemberRole(ctx context.Context, userID, projectID uint64) (string, error) {
	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}

	// 管理员默认为owner权限
	if user.Role == "admin" {
		return "owner", nil
	}

	// 获取用户在项目中的角色
	member, err := s.memberRepo.GetByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		return "", err
	}

	return member.Role, nil
}
