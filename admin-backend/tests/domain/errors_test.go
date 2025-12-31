package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"yflow/internal/domain"
)

func TestDomainErrors(t *testing.T) {
	// 用户相关错误
	assert.Equal(t, "用户不存在", domain.ErrUserNotFound.Error())
	assert.Equal(t, "密码错误", domain.ErrInvalidPassword.Error())
	assert.Equal(t, "用户已存在", domain.ErrUserExists.Error())
	assert.Equal(t, "无效的令牌", domain.ErrInvalidToken.Error())

	// 项目相关错误
	assert.Equal(t, "项目不存在", domain.ErrProjectNotFound.Error())
	assert.Equal(t, "项目已存在", domain.ErrProjectExists.Error())
	assert.Equal(t, "无效的项目标识", domain.ErrInvalidSlug.Error())

	// 语言相关错误
	assert.Equal(t, "语言不存在", domain.ErrLanguageNotFound.Error())
	assert.Equal(t, "语言已存在", domain.ErrLanguageExists.Error())
	assert.Equal(t, "无效的语言代码", domain.ErrInvalidLanguage.Error())

	// 翻译相关错误
	assert.Equal(t, "翻译不存在", domain.ErrTranslationNotFound.Error())
	assert.Equal(t, "翻译已存在", domain.ErrTranslationExists.Error())
	assert.Equal(t, "无效的翻译键", domain.ErrInvalidKey.Error())

	// 通用错误
	assert.Equal(t, "无效的输入参数", domain.ErrInvalidInput.Error())
	assert.Equal(t, "内部服务器错误", domain.ErrInternalError.Error())
	assert.Equal(t, "未授权访问", domain.ErrUnauthorized.Error())
	assert.Equal(t, "禁止访问", domain.ErrForbidden.Error())
}
