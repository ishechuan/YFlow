package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"yflow/internal/domain"
)

func TestUserModel(t *testing.T) {
	// 创建用户模型实例
	now := time.Now()
	user := domain.User{
		ID:        1,
		Username:  "testuser",
		Password:  "hashedpassword",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 验证字段值
	assert.Equal(t, uint(1), user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "hashedpassword", user.Password)
	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
}

func TestProjectModel(t *testing.T) {
	// 创建项目模型实例
	now := time.Now()
	project := domain.Project{
		ID:          1,
		Name:        "Test Project",
		Description: "Test project description",
		Slug:        "test-project",
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   gorm.DeletedAt{},
	}

	// 验证字段值
	assert.Equal(t, uint(1), project.ID)
	assert.Equal(t, "Test Project", project.Name)
	assert.Equal(t, "Test project description", project.Description)
	assert.Equal(t, "test-project", project.Slug)
	assert.Equal(t, "active", project.Status)
	assert.Equal(t, now, project.CreatedAt)
	assert.Equal(t, now, project.UpdatedAt)
	assert.False(t, project.DeletedAt.Valid)
}

func TestLanguageModel(t *testing.T) {
	// 创建语言模型实例
	now := time.Now()
	language := domain.Language{
		ID:        1,
		Code:      "en",
		Name:      "English",
		IsDefault: true,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: gorm.DeletedAt{},
	}

	// 验证字段值
	assert.Equal(t, uint(1), language.ID)
	assert.Equal(t, "en", language.Code)
	assert.Equal(t, "English", language.Name)
	assert.True(t, language.IsDefault)
	assert.Equal(t, "active", language.Status)
	assert.Equal(t, now, language.CreatedAt)
	assert.Equal(t, now, language.UpdatedAt)
	assert.False(t, language.DeletedAt.Valid)
}

func TestTranslationModel(t *testing.T) {
	// 创建项目和语言模型实例
	now := time.Now()
	project := domain.Project{
		ID:   1,
		Name: "Test Project",
		Slug: "test-project",
	}

	language := domain.Language{
		ID:   1,
		Code: "en",
		Name: "English",
	}

	// 创建翻译模型实例
	translation := domain.Translation{
		ID:         1,
		ProjectID:  1,
		KeyName:    "welcome.message",
		Context:    "Welcome message on home page",
		LanguageID: 1,
		Value:      "Welcome to our application",
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
		DeletedAt:  gorm.DeletedAt{},
		Project:    project,
		Language:   language,
	}

	// 验证字段值
	assert.Equal(t, uint(1), translation.ID)
	assert.Equal(t, uint(1), translation.ProjectID)
	assert.Equal(t, "welcome.message", translation.KeyName)
	assert.Equal(t, "Welcome message on home page", translation.Context)
	assert.Equal(t, uint(1), translation.LanguageID)
	assert.Equal(t, "Welcome to our application", translation.Value)
	assert.Equal(t, "active", translation.Status)
	assert.Equal(t, now, translation.CreatedAt)
	assert.Equal(t, now, translation.UpdatedAt)
	assert.False(t, translation.DeletedAt.Valid)

	// 验证关联
	assert.Equal(t, project, translation.Project)
	assert.Equal(t, language, translation.Language)
}
