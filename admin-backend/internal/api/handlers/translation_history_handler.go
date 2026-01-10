package handlers

import (
	"strconv"
	"yflow/internal/api/response"
	"yflow/internal/domain"
	"yflow/internal/dto"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TranslationHistoryHandler 翻译历史处理器
type TranslationHistoryHandler struct {
	historyRepo domain.TranslationHistoryRepository
	logger      *zap.Logger
}

// NewTranslationHistoryHandler 创建翻译历史处理器
func NewTranslationHistoryHandler(
	historyRepo domain.TranslationHistoryRepository,
	logger *zap.Logger,
) *TranslationHistoryHandler {
	return &TranslationHistoryHandler{
		historyRepo: historyRepo,
		logger:      logger,
	}
}

// GetByTranslationID 获取单个翻译的历史记录
// @Summary      获取翻译历史
// @Description  根据翻译ID获取历史记录列表
// @Tags         翻译历史
// @Accept       json
// @Produce      json
// @Param        id         path      int     true   "翻译ID"
// @Param        page       query     int     false  "页码"  default(1)
// @Param        page_size  query     int     false  "每页数量"  default(10)
// @Success      200        {object}  dto.TranslationHistoryListResponse
// @Failure      400        {object}  map[string]string
// @Failure      404        {object}  map[string]string
// @Security     BearerAuth
// @Router       /translations/{id}/history [get]
func (h *TranslationHistoryHandler) GetByTranslationID(ctx *gin.Context) {
	// 解析翻译ID
	idStr := ctx.Param("id")
	translationID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "无效的翻译ID")
		return
	}

	// 解析分页参数
	var req dto.ListTranslationHistoryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.ValidationError(ctx, err.Error())
		return
	}

	// 验证分页参数
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	offset := (req.Page - 1) * req.PageSize

	// 查询历史记录
	histories, total, err := h.historyRepo.ListByTranslationID(ctx.Request.Context(), translationID, req.PageSize, offset)
	if err != nil {
		h.logger.Error("获取翻译历史失败", zap.Error(err), zap.Uint64("translation_id", translationID))
		response.InternalServerError(ctx, "获取翻译历史失败")
		return
	}

	// 转换为响应格式
	responses := make([]*dto.TranslationHistoryResponse, len(histories))
	for i, history := range histories {
		responses[i] = h.toHistoryResponse(history)
	}

	// 构建分页元数据
	meta := &response.Meta{
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalCount: total,
		TotalPages: (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}

	response.SuccessWithMeta(ctx, dto.TranslationHistoryListResponse{
		Histories: responses,
		Meta:      meta,
	}, meta)
}

// GetByProjectID 获取项目翻译历史
// @Summary      获取项目翻译历史
// @Description  根据项目ID获取翻译历史记录列表
// @Tags         翻译历史
// @Accept       json
// @Produce      json
// @Param        project_id path      int     true   "项目ID"
// @Param        page       query     int     false  "页码"  default(1)
// @Param        page_size  query     int     false  "每页数量"  default(10)
// @Param        operation  query     string  false  "操作类型筛选"
// @Param        start_date query     string  false  "开始时间 (格式: 2006-01-02)"
// @Param        end_date   query     string  false  "结束时间 (格式: 2006-01-02)"
// @Success      200        {object}  dto.TranslationHistoryListResponse
// @Failure      400        {object}  map[string]string
// @Failure      404        {object}  map[string]string
// @Security     BearerAuth
// @Router       /projects/{project_id}/translation-history [get]
func (h *TranslationHistoryHandler) GetByProjectID(ctx *gin.Context) {
	// 解析项目ID
	projectIDStr := ctx.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "无效的项目ID")
		return
	}

	// 解析查询参数
	var req dto.ListTranslationHistoryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.ValidationError(ctx, err.Error())
		return
	}

	// 验证分页参数
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	// 转换为领域层参数
	params := req.ToDomainParams()

	// 查询历史记录
	histories, total, err := h.historyRepo.ListByProjectID(ctx.Request.Context(), projectID, params)
	if err != nil {
		h.logger.Error("获取项目翻译历史失败", zap.Error(err), zap.Uint64("project_id", projectID))
		response.InternalServerError(ctx, "获取项目翻译历史失败")
		return
	}

	// 转换为响应格式
	responses := make([]*dto.TranslationHistoryResponse, len(histories))
	for i, history := range histories {
		responses[i] = h.toHistoryResponse(history)
	}

	// 构建分页元数据
	meta := &response.Meta{
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalCount: total,
		TotalPages: (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}

	response.SuccessWithMeta(ctx, dto.TranslationHistoryListResponse{
		Histories: responses,
		Meta:      meta,
	}, meta)
}

// GetByUserID 获取用户操作历史
// @Summary      获取用户翻译操作历史
// @Description  根据用户ID获取翻译操作历史记录列表
// @Tags         翻译历史
// @Accept       json
// @Produce      json
// @Param        id         path      int     true   "用户ID"
// @Param        page       query     int     false  "页码"  default(1)
// @Param        page_size  query     int     false  "每页数量"  default(10)
// @Param        operation  query     string  false  "操作类型筛选"
// @Param        start_date query     string  false  "开始时间 (格式: 2006-01-02)"
// @Param        end_date   query     string  false  "结束时间 (格式: 2006-01-02)"
// @Success      200        {object}  dto.TranslationHistoryListResponse
// @Failure      400        {object}  map[string]string
// @Failure      404        {object}  map[string]string
// @Security     BearerAuth
// @Router       /users/{id}/translation-history [get]
func (h *TranslationHistoryHandler) GetByUserID(ctx *gin.Context) {
	// 解析用户ID
	userIDStr := ctx.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "无效的用户ID")
		return
	}

	// 解析查询参数
	var req dto.ListTranslationHistoryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.ValidationError(ctx, err.Error())
		return
	}

	// 验证分页参数
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	// 转换为领域层参数
	params := req.ToDomainParams()

	// 查询历史记录
	histories, total, err := h.historyRepo.ListByUserID(ctx.Request.Context(), userID, params)
	if err != nil {
		h.logger.Error("获取用户翻译历史失败", zap.Error(err), zap.Uint64("user_id", userID))
		response.InternalServerError(ctx, "获取用户翻译历史失败")
		return
	}

	// 转换为响应格式
	responses := make([]*dto.TranslationHistoryResponse, len(histories))
	for i, history := range histories {
		responses[i] = h.toHistoryResponse(history)
	}

	// 构建分页元数据
	meta := &response.Meta{
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalCount: total,
		TotalPages: (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}

	response.SuccessWithMeta(ctx, dto.TranslationHistoryListResponse{
		Histories: responses,
		Meta:      meta,
	}, meta)
}

// toHistoryResponse 将领域模型转换为响应DTO
func (h *TranslationHistoryHandler) toHistoryResponse(history *domain.TranslationHistory) *dto.TranslationHistoryResponse {
	return &dto.TranslationHistoryResponse{
		ID:            history.ID,
		TranslationID: history.TranslationID,
		ProjectID:     history.ProjectID,
		KeyName:       history.KeyName,
		LanguageID:    history.LanguageID,
		OldValue:      history.OldValue,
		NewValue:      history.NewValue,
		Operation:     history.Operation,
		OperatedBy:    history.OperatedBy,
		OperatedAt:    history.OperatedAt,
		Metadata:      history.Metadata,
	}
}
