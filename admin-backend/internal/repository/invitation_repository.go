package repository

import (
	"context"
	"errors"
	"time"
	"yflow/internal/domain"

	"gorm.io/gorm"
)

// InvitationRepository 邀请码仓储实现
type InvitationRepository struct {
	db *gorm.DB
}

// NewInvitationRepository 创建邀请码仓储实例
func NewInvitationRepository(db *gorm.DB) *InvitationRepository {
	return &InvitationRepository{db: db}
}

// GetByID 根据ID获取邀请码
func (r *InvitationRepository) GetByID(ctx context.Context, id uint64) (*domain.Invitation, error) {
	var invitation domain.Invitation
	if err := r.db.WithContext(ctx).Preload("Inviter").First(&invitation, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrInvitationNotFound
		}
		return nil, err
	}
	return &invitation, nil
}

// GetByCode 根据邀请码获取邀请
func (r *InvitationRepository) GetByCode(ctx context.Context, code string) (*domain.Invitation, error) {
	var invitation domain.Invitation
	if err := r.db.WithContext(ctx).Preload("Inviter").Where("code = ?", code).First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrInvitationNotFound
		}
		return nil, err
	}
	return &invitation, nil
}

// GetByInviter 根据邀请人ID获取邀请列表
func (r *InvitationRepository) GetByInviter(ctx context.Context, inviterID uint64, limit, offset int) ([]*domain.Invitation, int64, error) {
	var invitations []*domain.Invitation
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Invitation{}).Where("inviter_id = ?", inviterID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&invitations).Error; err != nil {
		return nil, 0, err
	}

	return invitations, total, nil
}

// GetActiveInvitations 获取所有有效的邀请
func (r *InvitationRepository) GetActiveInvitations(ctx context.Context) ([]*domain.Invitation, error) {
	var invitations []*domain.Invitation
	if err := r.db.WithContext(ctx).
		Where("status = ?", domain.InvitationStatusActive).
		Where("expires_at > ?", time.Now()).
		Find(&invitations).Error; err != nil {
		return nil, err
	}
	return invitations, nil
}

// Create 创建邀请码
func (r *InvitationRepository) Create(ctx context.Context, invitation *domain.Invitation) error {
	return r.db.WithContext(ctx).Create(invitation).Error
}

// Update 更新邀请码
func (r *InvitationRepository) Update(ctx context.Context, invitation *domain.Invitation) error {
	return r.db.WithContext(ctx).Save(invitation).Error
}

// MarkAsUsed 标记邀请码已使用
func (r *InvitationRepository) MarkAsUsed(ctx context.Context, code string, userID uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Invitation{}).
		Where("code = ?", code).
		Updates(map[string]interface{}{
			"status":   domain.InvitationStatusUsed,
			"used_at":  now,
			"used_by":  userID,
		}).Error
}

// Revoke 撤销邀请码
func (r *InvitationRepository) Revoke(ctx context.Context, code string) error {
	return r.db.WithContext(ctx).Model(&domain.Invitation{}).
		Where("code = ?", code).
		Update("status", domain.InvitationStatusRevoked).Error
}

// Delete 根据邀请码删除邀请
func (r *InvitationRepository) Delete(ctx context.Context, code string) error {
	return r.db.WithContext(ctx).Where("code = ?", code).Delete(&domain.Invitation{}).Error
}

// DeleteByID 根据ID删除邀请
func (r *InvitationRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Invitation{}, id).Error
}
