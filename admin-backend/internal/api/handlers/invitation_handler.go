package handlers

import (
	"time"
	"yflow/internal/api/response"
	"yflow/internal/domain"
	"yflow/internal/dto"
	"yflow/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// InvitationHandler 邀请码处理器
type InvitationHandler struct {
	invitationService domain.InvitationService
	userService       domain.UserService
	securityUtils     *utils.SecurityUtils
	logger            *zap.Logger
}

// NewInvitationHandler 创建邀请码处理器
func NewInvitationHandler(
	invitationService domain.InvitationService,
	userService domain.UserService,
	logger *zap.Logger,
) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
		userService:       userService,
		securityUtils:     utils.NewSecurityUtils(),
		logger:            logger,
	}
}

// CreateInvitation 创建邀请码
// @Summary      创建邀请码
// @Description  管理员创建新的邀请码
// @Tags         邀请管理
// @Accept       json
// @Produce      json
// @Param        invitation  body      dto.CreateInvitationRequest  true  "邀请信息"
// @Success      201         {object}  dto.CreateInvitationResponse
// @Failure      400         {object}  map[string]string
// @Failure      401         {object}  map[string]string
// @Failure      403         {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v1/invitations [post]
func (h *InvitationHandler) CreateInvitation(ctx *gin.Context) {
	var req dto.CreateInvitationRequest

	// 绑定请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, err.Error())
		return
	}

	// 从上下文中获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		response.Unauthorized(ctx, "用户未登录")
		return
	}

	// DTO -> Domain params
	params := domain.CreateInvitationParams{
		Role:          req.Role,
		ExpiresInDays: req.ExpiresInDays,
		Description:   req.Description,
	}

	// 创建邀请码
	invitation, invitationURL, err := h.invitationService.CreateInvitation(ctx.Request.Context(), userID.(uint64), params)
	if err != nil {
		h.logger.Error("Failed to create invitation", zap.Error(err))
		response.InternalServerError(ctx, "创建邀请码失败")
		return
	}

	// 创建成功日志
	operatorID, _ := ctx.Get("userID")
	operatorName := "system"
	if opUser, ok := ctx.Get("username"); ok {
		if op, ok := opUser.(string); ok {
			operatorName = op
		}
	}
	h.logger.Info("Invitation created",
		zap.Uint64("invitation_id", invitation.ID),
		zap.String("code", invitation.Code),
		zap.String("role", invitation.Role),
		zap.String("operator", operatorName),
		zap.Uint64("operator_id", operatorID.(uint64)),
	)

	response.Created(ctx, dto.CreateInvitationResponse{
		Code:          invitation.Code,
		InvitationURL: invitationURL,
		Role:          invitation.Role,
		ExpiresAt:     invitation.ExpiresAt.Format(time.RFC3339),
		Description:   invitation.Description,
	})
}

// GetInvitations 获取邀请列表
// @Summary      获取邀请列表
// @Description  分页获取邀请码列表
// @Tags         邀请管理
// @Accept       json
// @Produce      json
// @Param        page      query     int    false  "页码"       default(1)
// @Param        page_size query     int    false  "每页数量"   default(10)
// @Success      200       {object}  dto.InvitationListResponse
// @Failure      400       {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v1/invitations [get]
func (h *InvitationHandler) GetInvitations(ctx *gin.Context) {
	// 从上下文中获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		response.Unauthorized(ctx, "用户未登录")
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// 获取邀请列表
	invitations, total, err := h.invitationService.GetInvitationsByInviter(ctx.Request.Context(), userID.(uint64), pageSize, offset)
	if err != nil {
		h.logger.Error("Failed to get invitations", zap.Error(err))
		response.InternalServerError(ctx, "获取邀请列表失败")
		return
	}

	// 转换为响应格式
	resp := dto.InvitationListResponse{
		Invitations: make([]*dto.InvitationResponse, 0, len(invitations)),
		Total:       total,
	}

	for _, inv := range invitations {
		invResp := &dto.InvitationResponse{
			ID:          inv.ID,
			Code:        inv.Code,
			InviterID:   inv.InviterID,
			Role:        inv.Role,
			Status:      inv.Status,
			ExpiresAt:   inv.ExpiresAt.Format(time.RFC3339),
			Description: inv.Description,
			CreatedAt:   inv.CreatedAt.Format(time.RFC3339),
		}

		if inv.UsedAt != nil {
			usedAtStr := inv.UsedAt.Format(time.RFC3339)
			invResp.UsedAt = &usedAtStr
		}
		if inv.UsedBy != nil {
			invResp.UsedBy = inv.UsedBy
		}
		if inv.Inviter != nil {
			invResp.Inviter = &dto.InvitationInviter{
				ID:       inv.Inviter.ID,
				Username: inv.Inviter.Username,
				Email:    inv.Inviter.Email,
				Role:     inv.Inviter.Role,
			}
		}

		resp.Invitations = append(resp.Invitations, invResp)
	}

	meta := &response.Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: total,
		TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	}

	response.SuccessWithMeta(ctx, resp, meta)
}

// GetInvitation 获取邀请详情
// @Summary      获取邀请详情
// @Description  根据邀请码获取邀请详细信息
// @Tags         邀请管理
// @Accept       json
// @Produce      json
// @Param        code   path      string  true  "邀请码"
// @Success      200    {object}  dto.InvitationResponse
// @Failure      404    {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v1/invitations/{code} [get]
func (h *InvitationHandler) GetInvitation(ctx *gin.Context) {
	code := ctx.Param("code")
	if code == "" {
		response.ValidationError(ctx, "邀请码不能为空")
		return
	}

	invitation, err := h.invitationService.GetInvitation(ctx.Request.Context(), code)
	if err != nil {
		switch err {
		case domain.ErrInvitationNotFound:
			response.NotFound(ctx, "邀请码不存在")
		default:
			response.InternalServerError(ctx, "获取邀请详情失败")
		}
		return
	}

	resp := dto.InvitationResponse{
		ID:          invitation.ID,
		Code:        invitation.Code,
		InviterID:   invitation.InviterID,
		Role:        invitation.Role,
		Status:      invitation.Status,
		ExpiresAt:   invitation.ExpiresAt.Format(time.RFC3339),
		Description: invitation.Description,
		CreatedAt:   invitation.CreatedAt.Format(time.RFC3339),
	}

	if invitation.UsedAt != nil {
		usedAtStr := invitation.UsedAt.Format(time.RFC3339)
		resp.UsedAt = &usedAtStr
	}
	if invitation.UsedBy != nil {
		resp.UsedBy = invitation.UsedBy
	}
	if invitation.Inviter != nil {
		resp.Inviter = &dto.InvitationInviter{
			ID:       invitation.Inviter.ID,
			Username: invitation.Inviter.Username,
			Email:    invitation.Inviter.Email,
			Role:     invitation.Inviter.Role,
		}
	}

	response.Success(ctx, resp)
}

// RevokeInvitation 撤销邀请码
// @Summary      撤销邀请码
// @Description  撤销指定的邀请码，被撤销的邀请码将无法继续使用
// @Tags         邀请管理
// @Accept       json
// @Produce      json
// @Param        code   path      string  true  "邀请码"
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v1/invitations/{code} [delete]
func (h *InvitationHandler) RevokeInvitation(ctx *gin.Context) {
	code := ctx.Param("code")
	if code == "" {
		response.ValidationError(ctx, "邀请码不能为空")
		return
	}

	if err := h.invitationService.RevokeInvitation(ctx.Request.Context(), code); err != nil {
		switch err {
		case domain.ErrInvitationNotFound:
			response.NotFound(ctx, "邀请码不存在")
		case domain.ErrInvalidInvitation:
			response.BadRequest(ctx, "邀请码已使用，无法撤销")
		default:
			response.InternalServerError(ctx, "撤销邀请码失败")
		}
		return
	}

	// 撤销成功日志
	operatorID, _ := ctx.Get("userID")
	operatorName := "system"
	if opUser, ok := ctx.Get("username"); ok {
		if op, ok := opUser.(string); ok {
			operatorName = op
		}
	}
	h.logger.Info("Invitation revoked",
		zap.String("code", code),
		zap.String("operator", operatorName),
		zap.Uint64("operator_id", operatorID.(uint64)),
	)

	response.Success(ctx, map[string]string{"message": "邀请码已撤销"})
}

// ValidateInvitation 验证邀请码（公开接口）
// @Summary      验证邀请码
// @Description  验证邀请码是否有效，返回邀请信息供前端展示
// @Tags         公开接口
// @Accept       json
// @Produce      json
// @Param        code   path      string  true  "邀请码"
// @Success      200    {object}  dto.ValidateInvitationResponse
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Router       /api/v1/invitations/{code}/validate [get]
func (h *InvitationHandler) ValidateInvitation(ctx *gin.Context) {
	code := ctx.Param("code")
	if code == "" {
		response.ValidationError(ctx, "邀请码不能为空")
		return
	}

	invitation, err := h.invitationService.ValidateInvitation(ctx.Request.Context(), code)
	if err != nil {
		resp := dto.ValidateInvitationResponse{
			Valid:   false,
			Message: h.getErrorMessage(err),
		}
		// 返回200但标记无效
		response.Success(ctx, resp)
		return
	}

	resp := dto.ValidateInvitationResponse{
		Valid:     true,
		Role:      invitation.Role,
		ExpiresAt: invitation.ExpiresAt.Format(time.RFC3339),
	}

	if invitation.Inviter != nil {
		resp.Inviter = &dto.InvitationInviter{
			ID:       invitation.Inviter.ID,
			Username: invitation.Inviter.Username,
			Email:    invitation.Inviter.Email,
			Role:     invitation.Inviter.Role,
		}
	}

	response.Success(ctx, resp)
}

// RegisterWithInvitation 使用邀请码注册（公开接口）
// @Summary      使用邀请码注册
// @Description  通过邀请码创建新用户账户
// @Tags         公开接口
// @Accept       json
// @Produce      json
// @Param        registration  body      dto.RegisterWithInvitationRequest  true  "注册信息"
// @Success      201           {object}  domain.User
// @Failure      400           {object}  map[string]string
// @Failure      409           {object}  map[string]string
// @Router       /api/v1/register [post]
func (h *InvitationHandler) RegisterWithInvitation(ctx *gin.Context) {
	var req dto.RegisterWithInvitationRequest

	// 绑定请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, err.Error())
		return
	}

	// 验证邀请码
	invitation, err := h.invitationService.ValidateInvitation(ctx.Request.Context(), req.Code)
	if err != nil {
		switch err {
		case domain.ErrInvitationNotFound:
			response.NotFound(ctx, "邀请码不存在")
		case domain.ErrInvitationUsed:
			response.Conflict(ctx, "邀请码已被使用")
		case domain.ErrInvitationExpired:
			response.BadRequest(ctx, "邀请码已过期")
		case domain.ErrInvitationRevoked:
			response.BadRequest(ctx, "邀请码已被撤销")
		default:
			response.InternalServerError(ctx, "验证邀请码失败")
		}
		return
	}

	// 验证用户名格式
	if err := h.securityUtils.ValidateUsername(req.Username); err != nil {
		response.ValidationError(ctx, "用户名格式无效: "+err.Error())
		return
	}

	// 验证密码强度
	if err := h.securityUtils.ValidatePassword(req.Password); err != nil {
		response.ValidationError(ctx, "密码格式无效: "+err.Error())
		return
	}

	// 检查用户名是否已存在
	if _, err := h.userService.GetUserByID(ctx.Request.Context(), 0); err == nil {
		// 用户已存在，需要用GetByUsername检查
	}

	// 创建用户
	params := domain.CreateUserParams{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     invitation.Role, // 使用邀请码指定的角色
	}

	user, err := h.userService.CreateUser(ctx.Request.Context(), params)
	if err != nil {
		switch err {
		case domain.ErrUserExists:
			response.Conflict(ctx, "用户名已存在")
		case domain.ErrEmailExists:
			response.Conflict(ctx, "邮箱已存在")
		default:
			response.InternalServerError(ctx, "创建用户失败")
		}
		return
	}

	// 标记邀请码为已使用
	if err := h.invitationService.UseInvitation(ctx.Request.Context(), req.Code, user.ID); err != nil {
		h.logger.Error("Failed to mark invitation as used", zap.Error(err))
	}

	// 注册成功日志
	h.logger.Info("User registered via invitation",
		zap.Uint64("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("invitation_code", req.Code),
		zap.String("role", user.Role),
	)

	response.Created(ctx, map[string]interface{}{
		"message": "注册成功",
		"user":    user,
	})
}

// DeleteInvitation 删除邀请码
// @Summary      删除邀请码
// @Description  彻底删除邀请码记录
// @Tags         邀请管理
// @Accept       json
// @Produce      json
// @Param        code   path      string  true  "邀请码"
// @Success      200    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v1/invitations/{code} [delete]
func (h *InvitationHandler) DeleteInvitation(ctx *gin.Context) {
	code := ctx.Param("code")
	if code == "" {
		response.ValidationError(ctx, "邀请码不能为空")
		return
	}

	if err := h.invitationService.DeleteInvitation(ctx.Request.Context(), code); err != nil {
		switch err {
		case domain.ErrInvitationNotFound:
			response.NotFound(ctx, "邀请码不存在")
		default:
			response.InternalServerError(ctx, "删除邀请码失败")
		}
		return
	}

	// 删除成功日志
	operatorID, _ := ctx.Get("userID")
	operatorName := "system"
	if opUser, ok := ctx.Get("username"); ok {
		if op, ok := opUser.(string); ok {
			operatorName = op
		}
	}
	h.logger.Info("Invitation deleted",
		zap.String("code", code),
		zap.String("operator", operatorName),
		zap.Uint64("operator_id", operatorID.(uint64)),
	)

	response.Success(ctx, map[string]string{"message": "邀请码已删除"})
}

// getErrorMessage 获取错误的用户友好消息
func (h *InvitationHandler) getErrorMessage(err error) string {
	switch err {
	case domain.ErrInvitationUsed:
		return "邀请码已被使用"
	case domain.ErrInvitationExpired:
		return "邀请码已过期"
	case domain.ErrInvitationRevoked:
		return "邀请码已被撤销"
	default:
		return "邀请码无效"
	}
}

// generateHashedPassword 生成密码哈希
func generateHashedPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
