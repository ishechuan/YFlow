package handlers

import (
	"yflow/internal/api/response"
	"yflow/internal/domain"
	"yflow/internal/dto"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// LanguageHandler 语言处理器
type LanguageHandler struct {
	languageService domain.LanguageService
}

// NewLanguageHandler 创建语言处理器
func NewLanguageHandler(languageService domain.LanguageService) *LanguageHandler {
	return &LanguageHandler{
		languageService: languageService,
	}
}

// Create 创建语言
// @Summary      创建语言
// @Description  创建新的语言
// @Tags         语言管理
// @Accept       json
// @Produce      json
// @Param        language  body      dto.CreateLanguageRequest  true  "语言信息"
// @Success      201       {object}  domain.Language
// @Failure      400       {object}  map[string]string
// @Failure      409       {object}  map[string]string
// @Security     BearerAuth
// @Router       /languages [post]
func (h *LanguageHandler) Create(ctx *gin.Context) {
	var req dto.CreateLanguageRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		response.Unauthorized(ctx, "未找到用户信息")
		return
	}

	// DTO -> Domain params
	params := domain.CreateLanguageParams{
		Code:      req.Code,
		Name:      req.Name,
		IsDefault: req.IsDefault,
	}

	language, err := h.languageService.Create(ctx.Request.Context(), params, userID.(uint64))
	if err != nil {
		switch err {
		case domain.ErrLanguageExists:
			response.Conflict(ctx, err.Error())
		case domain.ErrInvalidLanguage:
			response.ValidationError(ctx, err.Error())
		default:
			response.InternalServerError(ctx, "创建语言失败")
		}
		return
	}

	response.Created(ctx, language)
}

// GetAll 获取所有语言
// @Summary      获取语言列表
// @Description  获取所有语言列表
// @Tags         语言管理
// @Accept       json
// @Produce      json
// @Success      200  {array}   domain.Language
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /languages [get]
func (h *LanguageHandler) GetAll(ctx *gin.Context) {
	languages, err := h.languageService.GetAll(ctx.Request.Context())
	if err != nil {
		response.InternalServerError(ctx, "获取语言列表失败")
		return
	}

	response.Success(ctx, languages)
}

// Update 更新语言
// @Summary      更新语言
// @Description  更新语言信息
// @Tags         语言管理
// @Accept       json
// @Produce      json
// @Param        id        path      int                            true  "语言ID"
// @Param        language  body      domain.CreateLanguageRequest  true  "语言信息"
// @Success      200       {object}  domain.Language
// @Failure      400       {object}  map[string]string
// @Failure      404       {object}  map[string]string
// @Security     BearerAuth
// @Router       /languages/{id} [put]
func (h *LanguageHandler) Update(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "无效的语言ID")
		return
	}

	var req dto.CreateLanguageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		response.Unauthorized(ctx, "未找到用户信息")
		return
	}

	// DTO -> Domain params
	params := domain.CreateLanguageParams{
		Code:      req.Code,
		Name:      req.Name,
		IsDefault: req.IsDefault,
	}

	language, err := h.languageService.Update(ctx.Request.Context(), id, params, userID.(uint64))
	if err != nil {
		switch err {
		case domain.ErrLanguageNotFound:
			response.NotFound(ctx, err.Error())
		case domain.ErrLanguageExists, domain.ErrInvalidInput:
			response.ValidationError(ctx, err.Error())
		default:
			response.InternalServerError(ctx, "更新语言失败")
		}
		return
	}

	response.Success(ctx, language)
}

// Delete 删除语言
// @Summary      删除语言
// @Description  删除指定的语言
// @Tags         语言管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "语言ID"
// @Success      204  {object}  nil
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /languages/{id} [delete]
func (h *LanguageHandler) Delete(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "无效的语言ID")
		return
	}

	err = h.languageService.Delete(ctx.Request.Context(), id)
	if err != nil {
		switch err {
		case domain.ErrLanguageNotFound:
			response.NotFound(ctx, err.Error())
		case domain.ErrInvalidInput:
			response.ValidationError(ctx, err.Error())
		default:
			response.InternalServerError(ctx, "删除语言失败")
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}
