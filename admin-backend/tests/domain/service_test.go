package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"yflow/internal/domain"
	"yflow/internal/dto"
)

func TestLoginRequest(t *testing.T) {
	// 创建登录请求
	req := dto.LoginRequest{
		Username: "admin",
		Password: "password",
	}

	// 验证字段值
	assert.Equal(t, "admin", req.Username)
	assert.Equal(t, "password", req.Password)
}

func TestLoginResponse(t *testing.T) {
	// 创建用户
	user := domain.User{
		ID:       1,
		Username: "admin",
	}

	// 创建登录响应
	resp := dto.LoginResponse{
		Token:        "jwt-token",
		RefreshToken: "refresh-token",
		User:         user,
	}

	// 验证字段值
	assert.Equal(t, "jwt-token", resp.Token)
	assert.Equal(t, "refresh-token", resp.RefreshToken)
	assert.Equal(t, user, resp.User)
}

func TestRefreshRequest(t *testing.T) {
	// 创建刷新令牌请求
	req := dto.RefreshRequest{
		RefreshToken: "refresh-token",
	}

	// 验证字段值
	assert.Equal(t, "refresh-token", req.RefreshToken)
}

func TestCreateProjectRequest(t *testing.T) {
	// 创建项目请求
	req := dto.CreateProjectRequest{
		Name:        "Test Project",
		Description: "Test project description",
	}

	// 验证字段值
	assert.Equal(t, "Test Project", req.Name)
	assert.Equal(t, "Test project description", req.Description)
}

func TestUpdateProjectRequest(t *testing.T) {
	// 创建更新项目请求
	req := dto.UpdateProjectRequest{
		Name:        "Updated Project",
		Description: "Updated description",
		Status:      "archived",
	}

	// 验证字段值
	assert.Equal(t, "Updated Project", req.Name)
	assert.Equal(t, "Updated description", req.Description)
	assert.Equal(t, "archived", req.Status)
}

func TestCreateLanguageRequest(t *testing.T) {
	// 创建语言请求
	req := dto.CreateLanguageRequest{
		Code:      "en",
		Name:      "English",
		IsDefault: true,
	}

	// 验证字段值
	assert.Equal(t, "en", req.Code)
	assert.Equal(t, "English", req.Name)
	assert.True(t, req.IsDefault)
}

func TestCreateTranslationRequest(t *testing.T) {
	// 创建翻译请求
	req := dto.CreateTranslationRequest{
		ProjectID:  1,
		KeyName:    "welcome.message",
		Context:    "Welcome message on home page",
		LanguageID: 1,
		Value:      "Welcome to our application",
	}

	// 验证字段值
	assert.Equal(t, uint(1), req.ProjectID)
	assert.Equal(t, "welcome.message", req.KeyName)
	assert.Equal(t, "Welcome message on home page", req.Context)
	assert.Equal(t, uint(1), req.LanguageID)
	assert.Equal(t, "Welcome to our application", req.Value)
}

func TestBatchTranslationRequest(t *testing.T) {
	// 创建批量翻译请求
	translations := map[string]string{
		"en": "Welcome",
		"fr": "Bienvenue",
		"es": "Bienvenido",
	}

	req := dto.BatchTranslationRequest{
		ProjectID:    1,
		KeyName:      "welcome.message",
		Context:      "Welcome message on home page",
		Translations: translations,
	}

	// 验证字段值
	assert.Equal(t, uint(1), req.ProjectID)
	assert.Equal(t, "welcome.message", req.KeyName)
	assert.Equal(t, "Welcome message on home page", req.Context)
	assert.Equal(t, translations, req.Translations)
	assert.Equal(t, "Welcome", req.Translations["en"])
	assert.Equal(t, "Bienvenue", req.Translations["fr"])
	assert.Equal(t, "Bienvenido", req.Translations["es"])
}

func TestDashboardStats(t *testing.T) {
	// 创建仪表板统计数据
	stats := dto.DashboardStats{
		TotalProjects:     5,
		TotalLanguages:    3,
		TotalTranslations: 150,
		TotalKeys:         50,
	}

	// 验证字段值
	assert.Equal(t, 5, stats.TotalProjects)
	assert.Equal(t, 3, stats.TotalLanguages)
	assert.Equal(t, 150, stats.TotalTranslations)
	assert.Equal(t, 50, stats.TotalKeys)
}
