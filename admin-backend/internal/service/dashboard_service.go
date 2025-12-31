package service

import (
	"context"
	"yflow/internal/domain"
)

// DashboardService 仪表板服务实现
type DashboardService struct {
	projectRepo     domain.ProjectRepository
	languageRepo    domain.LanguageRepository
	translationRepo domain.TranslationRepository
}

// NewDashboardService 创建仪表板服务实例
func NewDashboardService(
	projectRepo domain.ProjectRepository,
	languageRepo domain.LanguageRepository,
	translationRepo domain.TranslationRepository,
) *DashboardService {
	return &DashboardService{
		projectRepo:     projectRepo,
		languageRepo:    languageRepo,
		translationRepo: translationRepo,
	}
}

// GetStats 获取仪表板统计信息
func (s *DashboardService) GetStats(ctx context.Context) (*domain.DashboardStats, error) {
	stats := &domain.DashboardStats{}

	// 获取项目总数
	_, totalProjects, err := s.projectRepo.GetAll(ctx, 1000000, 0, "") // 大数获取全部，无关键词过滤
	if err != nil {
		return nil, err
	}
	stats.TotalProjects = int(totalProjects)

	// 获取语言总数
	languages, err := s.languageRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalLanguages = len(languages)

	// 使用聚合查询获取翻译统计 (修复 N+1 查询)
	totalTranslations, totalKeys, err := s.translationRepo.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	stats.TotalTranslations = totalTranslations
	stats.TotalKeys = totalKeys

	return stats, nil
}
