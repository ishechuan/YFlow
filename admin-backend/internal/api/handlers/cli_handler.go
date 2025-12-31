package handlers

import (
	"yflow/internal/api/response"
	"yflow/internal/domain"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CLIHandler CLI处理器
type CLIHandler struct {
	translationService domain.TranslationService
	projectService     domain.ProjectService
	languageService    domain.LanguageService
}

// NewCLIHandler 创建CLI处理器
func NewCLIHandler(
	translationService domain.TranslationService,
	projectService domain.ProjectService,
	languageService domain.LanguageService,
) *CLIHandler {
	return &CLIHandler{
		translationService: translationService,
		projectService:     projectService,
		languageService:    languageService,
	}
}

// Auth CLI身份验证
// @Summary      CLI身份验证
// @Description  验证CLI API Key
// @Tags         CLI
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.APIResponse
// @Failure      401  {object}  response.APIResponse
// @Security     ApiKeyAuth
// @Router       /cli/auth [get]
func (h *CLIHandler) Auth(ctx *gin.Context) {
	// API Key认证由中间件处理，能到这里说明认证成功
	response.Success(ctx, gin.H{
		"status":  "ok",
		"message": "CLI authentication successful",
	})
}

// GetTranslations 获取翻译数据
// @Summary      获取翻译数据
// @Description  获取项目翻译数据供CLI使用
// @Tags         CLI
// @Accept       json
// @Produce      json
// @Param        project_id  query     string  false  "项目ID"
// @Param        locale      query     string  false  "语言代码"
// @Success      200         {object}  response.APIResponse
// @Failure      400         {object}  response.APIResponse
// @Failure      404         {object}  response.APIResponse
// @Security     ApiKeyAuth
// @Router       /cli/translations [get]
func (h *CLIHandler) GetTranslations(ctx *gin.Context) {
	projectIDStr := ctx.Query("project_id")
	locale := ctx.Query("locale")

	// 如果没有指定项目ID，返回错误
	if projectIDStr == "" {
		response.BadRequest(ctx, "project_id is required")
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid project_id")
		return
	}

	// 验证项目是否存在
	_, err = h.projectService.GetByID(ctx.Request.Context(), projectID)
	if err != nil {
		switch err {
		case domain.ErrProjectNotFound:
			response.NotFound(ctx, err.Error())
		default:
			response.InternalServerError(ctx, "获取项目失败")
		}
		return
	}

	// 获取翻译矩阵数据（不分页，获取所有数据）
	matrix, _, err := h.translationService.GetMatrix(ctx.Request.Context(), projectID, -1, 0, "")
	if err != nil {
		response.InternalServerError(ctx, "获取翻译数据失败")
		return
	}

	// 转换为简单格式 (key -> language -> value)
	simpleMatrix := make(map[string]map[string]string)
	for key, langs := range matrix {
		simpleMatrix[key] = make(map[string]string)
		for lang, cell := range langs {
			simpleMatrix[key][lang] = cell.Value
		}
	}

	// 如果指定了locale，只返回该语言的数据
	if locale != "" {
		filteredMatrix := make(map[string]map[string]string)
		for key, translations := range simpleMatrix {
			if value, exists := translations[locale]; exists {
				filteredMatrix[key] = map[string]string{locale: value}
			}
		}
		response.Success(ctx, filteredMatrix)
		return
	}

	// 返回完整的翻译矩阵
	response.Success(ctx, simpleMatrix)
}

// PushKeysRequest 推送键请求
type PushKeysRequest struct {
	ProjectID    string                       `json:"project_id" binding:"required"`
	Keys         []string                     `json:"keys"`                  // 可选：如果为空且提供了 Translations，则执行批量导入
	Defaults     map[string]string            `json:"defaults"`              // 已废弃，保持向后兼容
	Translations map[string]map[string]string `json:"translations"`          // 语言代码 -> 键值对映射
}

// PushKeysResponse 推送键响应
type PushKeysResponse struct {
	Added   []string `json:"added"`
	Existed []string `json:"existed"`
	Failed  []string `json:"failed"`
}

// PushKeys 推送翻译键
// @Summary      推送翻译键或批量导入翻译
// @Description  从CLI推送新的翻译键，或批量导入/更新翻译数据
// @Tags         CLI
// @Accept       json
// @Produce      json
// @Param        request  body      PushKeysRequest  true  "推送键请求"
// @Success      200      {object}  response.APIResponse
// @Failure      400      {object}  response.APIResponse
// @Failure      404      {object}  response.APIResponse
// @Security     ApiKeyAuth
// @Router       /cli/keys [post]
func (h *CLIHandler) PushKeys(ctx *gin.Context) {
	var req PushKeysRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, err.Error())
		return
	}

	projectID, err := strconv.ParseUint(req.ProjectID, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid project_id")
		return
	}

	// 验证项目是否存在
	_, err = h.projectService.GetByID(ctx.Request.Context(), projectID)
	if err != nil {
		switch err {
		case domain.ErrProjectNotFound:
			response.NotFound(ctx, err.Error())
		default:
			response.InternalServerError(ctx, "获取项目失败")
		}
		return
	}

	// 获取所有语言
	languages, err := h.languageService.GetAll(ctx.Request.Context())
	if err != nil {
		response.InternalServerError(ctx, "获取语言列表失败")
		return
	}

	// 创建语言代码到ID的映射
	languageCodeToID := make(map[string]uint64)
	for _, lang := range languages {
		languageCodeToID[lang.Code] = lang.ID
	}

	// 判断操作类型：批量导入或推送键
	if len(req.Keys) == 0 && req.Translations != nil && len(req.Translations) > 0 {
		// 批量导入模式
		h.handleBulkImport(ctx, projectID, req.Translations, languageCodeToID)
		return
	}

	// 推送键模式（原逻辑）
	h.handlePushKeys(ctx, projectID, req, languages, languageCodeToID)
}

// handleBulkImport 处理批量导入翻译
func (h *CLIHandler) handleBulkImport(
	ctx *gin.Context,
	projectID uint64,
	translations map[string]map[string]string,
	languageCodeToID map[string]uint64,
) {
	// 获取现有的翻译键，用于判断新增或更新
	matrix, _, err := h.translationService.GetMatrix(ctx.Request.Context(), projectID, -1, 0, "")
	if err != nil {
		response.InternalServerError(ctx, "获取现有翻译失败")
		return
	}

	var added []string
	var existed []string
	var failed []string

	// 收集所有要导入的翻译
	var inputs []domain.TranslationInput

	for langCode, langTranslations := range translations {
		langID, exists := languageCodeToID[langCode]
		if !exists {
			// 忽略未知语言
			continue
		}

		for key, value := range langTranslations {
			// 跳过空值
			if value == "" {
				continue
			}

			// 判断是新增还是更新
			if _, keyExists := matrix[key]; keyExists {
				if !containsString(existed, key) {
					existed = append(existed, key)
				}
			} else {
				if !containsString(added, key) && !containsString(existed, key) {
					added = append(added, key)
				}
			}

			inputs = append(inputs, domain.TranslationInput{
				ProjectID:  projectID,
				KeyName:    key,
				LanguageID: langID,
				Value:      value,
			})
		}
	}

	if len(inputs) == 0 {
		response.Success(ctx, PushKeysResponse{
			Added:   []string{},
			Existed: existed,
			Failed:  []string{},
		})
		return
	}

	// 使用 UpsertBatch 进行批量导入/更新
	err = h.translationService.UpsertBatch(ctx.Request.Context(), inputs)
	if err != nil {
		// 如果失败，标记所有键为失败
		for _, key := range added {
			failed = append(failed, key)
		}
		added = []string{}
	}

	result := PushKeysResponse{
		Added:   added,
		Existed: existed,
		Failed:  failed,
	}

	response.Success(ctx, result)
}

// handlePushKeys 处理推送键（原逻辑）
func (h *CLIHandler) handlePushKeys(
	ctx *gin.Context,
	projectID uint64,
	req PushKeysRequest,
	languages []*domain.Language,
	languageCodeToID map[string]uint64,
) {
	// 获取现有的翻译键
	matrix, _, err := h.translationService.GetMatrix(ctx.Request.Context(), projectID, -1, 0, "")
	if err != nil {
		response.InternalServerError(ctx, "获取现有翻译失败")
		return
	}

	// 找到默认语言
	var defaultLanguage *domain.Language
	for _, lang := range languages {
		if lang.IsDefault {
			defaultLanguage = lang
			break
		}
	}
	if defaultLanguage == nil && len(languages) > 0 {
		defaultLanguage = languages[0]
	}

	var added []string
	var existed []string
	var failed []string

	// 处理每个键
	for _, key := range req.Keys {
		if _, exists := matrix[key]; exists {
			existed = append(existed, key)
			continue
		}

		// 为所有语言创建新的翻译记录
		keyAdded := false
		keyFailed := false

		for _, language := range languages {
			// 确定翻译值
			var value string

			// 优先使用新的多语言数据结构
			if req.Translations != nil {
				if langTranslations, exists := req.Translations[language.Code]; exists {
					value = langTranslations[key]
				}
			} else {
				// 向后兼容：使用旧的 Defaults 字段
				if language.Code == defaultLanguage.Code {
					value = req.Defaults[key]
				}
			}

			input := domain.TranslationInput{
				ProjectID:  projectID,
				KeyName:    key,
				LanguageID: language.ID,
				Value:      value,
			}

			_, err := h.translationService.Create(ctx.Request.Context(), input, 1)
			if err != nil {
				keyFailed = true
			} else if !keyAdded {
				keyAdded = true
			}
		}

		if keyFailed && !keyAdded {
			failed = append(failed, key)
		} else if keyAdded {
			added = append(added, key)
		}
	}

	result := PushKeysResponse{
		Added:   added,
		Existed: existed,
		Failed:  failed,
	}

	response.Success(ctx, result)
}

// containsString 检查字符串是否在切片中
func containsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
