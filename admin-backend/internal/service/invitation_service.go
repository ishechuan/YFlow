package service

import (
	"context"
	"fmt"
	"time"
	"yflow/internal/domain"
	"yflow/internal/utils"
)

// InvitationService 邀请码服务实现
type InvitationService struct {
	invitationRepo domain.InvitationRepository
	userRepo       domain.UserRepository
	securityUtils  *utils.SecurityUtils
	frontendURL    string
}

// NewInvitationService 创建邀请码服务实例
func NewInvitationService(
	invitationRepo domain.InvitationRepository,
	userRepo domain.UserRepository,
	frontendURL string,
) *InvitationService {
	return &InvitationService{
		invitationRepo: invitationRepo,
		userRepo:       userRepo,
		securityUtils:  utils.NewSecurityUtils(),
		frontendURL:    frontendURL,
	}
}

// CreateInvitation 创建邀请码
func (s *InvitationService) CreateInvitation(ctx context.Context, inviterID uint64, params domain.CreateInvitationParams) (*domain.Invitation, string, error) {
	// 验证角色
	role := params.Role
	if role == "" {
		role = "member"
	}
	if role != "admin" && role != "member" && role != "viewer" {
		return nil, "", domain.ErrInvalidRole
	}

	// 验证过期天数
	expiresInDays := params.ExpiresInDays
	if expiresInDays <= 0 {
		expiresInDays = 7 // 默认7天
	}
	if expiresInDays > 365 {
		expiresInDays = 365 // 最多365天
	}

	// 生成邀请码
	code, err := s.generateInvitationCode()
	if err != nil {
		return nil, "", err
	}

	// 创建邀请记录
	invitation := &domain.Invitation{
		Code:        code,
		InviterID:   inviterID,
		Role:        role,
		Status:      domain.InvitationStatusActive,
		ExpiresAt:   time.Now().AddDate(0, 0, expiresInDays),
		Description: params.Description,
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, "", err
	}

	// 生成邀请链接
	invitationURL := s.generateInvitationURL(code)

	return invitation, invitationURL, nil
}

// GetInvitation 获取邀请码详情
func (s *InvitationService) GetInvitation(ctx context.Context, code string) (*domain.Invitation, error) {
	invitation, err := s.invitationRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// 清除敏感信息
	if invitation.Inviter != nil {
		invitation.Inviter.Password = ""
	}

	return invitation, nil
}

// GetInvitationsByInviter 获取邀请人创建的邀请列表
func (s *InvitationService) GetInvitationsByInviter(ctx context.Context, inviterID uint64, limit, offset int) ([]*domain.Invitation, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return s.invitationRepo.GetByInviter(ctx, inviterID, limit, offset)
}

// ValidateInvitation 验证邀请码是否有效
func (s *InvitationService) ValidateInvitation(ctx context.Context, code string) (*domain.Invitation, error) {
	invitation, err := s.invitationRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// 检查状态
	if invitation.Status == domain.InvitationStatusUsed {
		return nil, domain.ErrInvitationUsed
	}
	if invitation.Status == domain.InvitationStatusRevoked {
		return nil, domain.ErrInvitationRevoked
	}

	// 检查是否过期
	if time.Now().After(invitation.ExpiresAt) {
		return nil, domain.ErrInvitationExpired
	}

	// 清除敏感信息
	if invitation.Inviter != nil {
		invitation.Inviter.Password = ""
	}

	return invitation, nil
}

// UseInvitation 使用邀请码（创建用户并标记邀请为已使用）
func (s *InvitationService) UseInvitation(ctx context.Context, code string, userID uint64) error {
	invitation, err := s.invitationRepo.GetByCode(ctx, code)
	if err != nil {
		return err
	}

	// 检查状态
	if invitation.Status == domain.InvitationStatusUsed {
		return domain.ErrInvitationUsed
	}
	if invitation.Status == domain.InvitationStatusRevoked {
		return domain.ErrInvitationRevoked
	}

	// 检查是否过期
	if time.Now().After(invitation.ExpiresAt) {
		return domain.ErrInvitationExpired
	}

	// 标记为已使用
	return s.invitationRepo.MarkAsUsed(ctx, code, userID)
}

// RevokeInvitation 撤销邀请码
func (s *InvitationService) RevokeInvitation(ctx context.Context, code string) error {
	invitation, err := s.invitationRepo.GetByCode(ctx, code)
	if err != nil {
		return err
	}

	// 如果已经使用或撤销，不能再次撤销
	if invitation.Status != domain.InvitationStatusActive {
		return domain.ErrInvalidInvitation
	}

	return s.invitationRepo.Revoke(ctx, code)
}

// DeleteInvitation 删除邀请码
func (s *InvitationService) DeleteInvitation(ctx context.Context, code string) error {
	return s.invitationRepo.Delete(ctx, code)
}

// generateInvitationCode 生成邀请码
func (s *InvitationService) generateInvitationCode() (string, error) {
	return s.securityUtils.GenerateSecureToken(32)
}

// generateInvitationURL 生成邀请链接
func (s *InvitationService) generateInvitationURL(code string) string {
	if s.frontendURL == "" {
		s.frontendURL = "http://localhost:3000"
	}
	return fmt.Sprintf("%s/register?code=%s", s.frontendURL, code)
}
