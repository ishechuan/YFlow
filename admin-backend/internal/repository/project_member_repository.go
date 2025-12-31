package repository

import (
	"context"
	"errors"
	"yflow/internal/domain"

	"gorm.io/gorm"
)

// ProjectMemberRepository 项目成员仓储实现
type ProjectMemberRepository struct {
	db *gorm.DB
}

// NewProjectMemberRepository 创建项目成员仓储实例
func NewProjectMemberRepository(db *gorm.DB) *ProjectMemberRepository {
	return &ProjectMemberRepository{db: db}
}

// GetByProjectAndUser 根据项目ID和用户ID获取成员关系
func (r *ProjectMemberRepository) GetByProjectAndUser(ctx context.Context, projectID, userID uint64) (*domain.ProjectMember, error) {
	var member domain.ProjectMember
	if err := r.db.WithContext(ctx).Where("project_id = ? AND user_id = ?", projectID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrMemberNotFound
		}
		return nil, err
	}
	return &member, nil
}

// GetByProjectID 根据项目ID获取所有成员
func (r *ProjectMemberRepository) GetByProjectID(ctx context.Context, projectID uint64) ([]*domain.ProjectMember, error) {
	var members []*domain.ProjectMember
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Preload("User").Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// GetByUserID 根据用户ID获取所有项目成员关系
func (r *ProjectMemberRepository) GetByUserID(ctx context.Context, userID uint64) ([]*domain.ProjectMember, error) {
	var members []*domain.ProjectMember
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Preload("Project").Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// Create 创建项目成员关系
func (r *ProjectMemberRepository) Create(ctx context.Context, member *domain.ProjectMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// Update 更新项目成员关系
func (r *ProjectMemberRepository) Update(ctx context.Context, member *domain.ProjectMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

// Delete 删除项目成员关系
func (r *ProjectMemberRepository) Delete(ctx context.Context, projectID, userID uint64) error {
	return r.db.WithContext(ctx).Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&domain.ProjectMember{}).Error
}
