package handlers

import (
	"yflow/internal/api/response"
	"yflow/internal/domain"
	"yflow/internal/dto"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ProjectMemberHandler 项目成员处理器
type ProjectMemberHandler struct {
	projectMemberService domain.ProjectMemberService
}

// NewProjectMemberHandler 创建项目成员处理器
func NewProjectMemberHandler(projectMemberService domain.ProjectMemberService) *ProjectMemberHandler {
	return &ProjectMemberHandler{
		projectMemberService: projectMemberService,
	}
}

// AddMember 添加项目成员
// @Summary      添加项目成员
// @Description  将用户添加到项目中并分配角色
// @Tags         项目成员管理
// @Accept       json
// @Produce      json
// @Param        project_id  path      int                             true  "项目ID"
// @Param        member      body      dto.AddProjectMemberRequest  true  "成员信息"
// @Success      201         {object}  domain.ProjectMember
// @Failure      400         {object}  map[string]string
// @Failure      404         {object}  map[string]string
// @Failure      409         {object}  map[string]string
// @Security     BearerAuth
// @Router       /projects/{project_id}/members [post]
func (h *ProjectMemberHandler) AddMember(ctx *gin.Context) {
	// 解析项目ID
	projectIDStr := ctx.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		response.ValidationError(ctx, "无效的项目ID")
		return
	}

	var req dto.AddProjectMemberRequest

	// 绑定请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, err.Error())
		return
	}

	// 获取当前用户ID
	currentUserID, exists := ctx.Get("userID")
	if !exists {
		response.Unauthorized(ctx, "未找到用户信息")
		return
	}

	// DTO -> Domain params
	params := domain.AddMemberParams{
		MemberUserID: req.UserID,
		Role:         req.Role,
	}

	// 调用添加成员服务
	member, err := h.projectMemberService.AddMember(ctx.Request.Context(), projectID, params, currentUserID.(uint64))
	if err != nil {
		switch err {
		case domain.ErrProjectNotFound:
			response.NotFound(ctx, "项目不存在")
		case domain.ErrUserNotFound:
			response.NotFound(ctx, "用户不存在")
		case domain.ErrMemberExists:
			response.Conflict(ctx, "用户已是项目成员")
		default:
			response.InternalServerError(ctx, "添加项目成员失败")
		}
		return
	}

	response.Created(ctx, member)
}

// GetProjectMembers 获取项目成员列表
// @Summary      获取项目成员列表
// @Description  获取指定项目的所有成员信息
// @Tags         项目成员管理
// @Accept       json
// @Produce      json
// @Param        project_id  path      int  true  "项目ID"
// @Success      200         {object}  []dto.ProjectMemberInfo
// @Failure      400         {object}  map[string]string
// @Failure      404         {object}  map[string]string
// @Security     BearerAuth
// @Router       /projects/{project_id}/members [get]
func (h *ProjectMemberHandler) GetProjectMembers(ctx *gin.Context) {
	// 解析项目ID
	projectIDStr := ctx.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		response.ValidationError(ctx, "无效的项目ID")
		return
	}

	// 获取项目成员列表
	members, err := h.projectMemberService.GetProjectMembers(ctx.Request.Context(), projectID)
	if err != nil {
		switch err {
		case domain.ErrProjectNotFound:
			response.NotFound(ctx, "项目不存在")
		default:
			response.InternalServerError(ctx, "获取项目成员失败")
		}
		return
	}

	response.Success(ctx, members)
}

// GetUserProjects 获取用户参与的项目列表
// @Summary      获取用户参与的项目列表
// @Description  获取指定用户参与的所有项目
// @Tags         项目成员管理
// @Accept       json
// @Produce      json
// @Param        user_id  path      int  true  "用户ID"
// @Success      200      {object}  []domain.Project
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Security     BearerAuth
// @Router       /users/{user_id}/projects [get]
func (h *ProjectMemberHandler) GetUserProjects(ctx *gin.Context) {
	// 解析用户ID
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.ValidationError(ctx, "无效的用户ID")
		return
	}

	// 获取用户项目列表
	projects, err := h.projectMemberService.GetUserProjects(ctx.Request.Context(), userID)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			response.NotFound(ctx, "用户不存在")
		default:
			response.InternalServerError(ctx, "获取用户项目失败")
		}
		return
	}

	response.Success(ctx, projects)
}

// UpdateMemberRole 更新成员角色
// @Summary      更新项目成员角色
// @Description  更新项目成员的角色权限
// @Tags         项目成员管理
// @Accept       json
// @Produce      json
// @Param        project_id  path      int                                true  "项目ID"
// @Param        user_id     path      int                                true  "用户ID"
// @Param        role        body      dto.UpdateProjectMemberRequest  true  "角色信息"
// @Success      200         {object}  domain.ProjectMember
// @Failure      400         {object}  map[string]string
// @Failure      404         {object}  map[string]string
// @Security     BearerAuth
// @Router       /projects/{project_id}/members/{user_id} [put]
func (h *ProjectMemberHandler) UpdateMemberRole(ctx *gin.Context) {
	// 解析项目ID
	projectIDStr := ctx.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		response.ValidationError(ctx, "无效的项目ID")
		return
	}

	// 解析用户ID
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.ValidationError(ctx, "无效的用户ID")
		return
	}

	var req dto.UpdateProjectMemberRequest

	// 绑定请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(ctx, err.Error())
		return
	}

	// DTO -> Domain params
	params := domain.UpdateMemberRoleParams{
		Role: req.Role,
	}

	// 调用更新成员角色服务
	member, err := h.projectMemberService.UpdateMemberRole(ctx.Request.Context(), projectID, userID, params)
	if err != nil {
		switch err {
		case domain.ErrMemberNotFound:
			response.NotFound(ctx, "项目成员不存在")
		default:
			response.InternalServerError(ctx, "更新成员角色失败")
		}
		return
	}

	response.Success(ctx, member)
}

// RemoveMember 移除项目成员
// @Summary      移除项目成员
// @Description  从项目中移除指定成员
// @Tags         项目成员管理
// @Accept       json
// @Produce      json
// @Param        project_id  path      int  true  "项目ID"
// @Param        user_id     path      int  true  "用户ID"
// @Success      200         {object}  map[string]string
// @Failure      400         {object}  map[string]string
// @Failure      403         {object}  map[string]string
// @Failure      404         {object}  map[string]string
// @Security     BearerAuth
// @Router       /projects/{project_id}/members/{user_id} [delete]
func (h *ProjectMemberHandler) RemoveMember(ctx *gin.Context) {
	// 解析项目ID
	projectIDStr := ctx.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		response.ValidationError(ctx, "无效的项目ID")
		return
	}

	// 解析用户ID
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.ValidationError(ctx, "无效的用户ID")
		return
	}

	// 调用移除成员服务
	if err := h.projectMemberService.RemoveMember(ctx.Request.Context(), projectID, userID); err != nil {
		switch err {
		case domain.ErrMemberNotFound:
			response.NotFound(ctx, "项目成员不存在")
		case domain.ErrCannotRemoveOwner:
			response.Forbidden(ctx, "不能移除项目所有者")
		default:
			response.InternalServerError(ctx, "移除项目成员失败")
		}
		return
	}

	response.Success(ctx, map[string]string{"message": "项目成员移除成功"})
}

// CheckPermission 检查用户权限
// @Summary      检查用户项目权限
// @Description  检查用户在指定项目中是否具有所需权限
// @Tags         项目成员管理
// @Accept       json
// @Produce      json
// @Param        project_id     path      int     true   "项目ID"
// @Param        user_id        path      int     true   "用户ID"
// @Param        required_role  query     string  true   "所需角色" Enums(viewer, editor, owner)
// @Success      200            {object}  map[string]bool
// @Failure      400            {object}  map[string]string
// @Failure      404            {object}  map[string]string
// @Security     BearerAuth
// @Router       /projects/{project_id}/members/{user_id}/permission [get]
func (h *ProjectMemberHandler) CheckPermission(ctx *gin.Context) {
	// 解析项目ID
	projectIDStr := ctx.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		response.ValidationError(ctx, "无效的项目ID")
		return
	}

	// 解析用户ID
	userIDStr := ctx.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.ValidationError(ctx, "无效的用户ID")
		return
	}

	// 获取所需角色
	requiredRole := ctx.Query("required_role")
	if requiredRole == "" {
		response.ValidationError(ctx, "缺少required_role参数")
		return
	}

	// 检查权限
	hasPermission, err := h.projectMemberService.CheckPermission(ctx.Request.Context(), userID, projectID, requiredRole)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			response.NotFound(ctx, "用户不存在")
		default:
			response.InternalServerError(ctx, "检查权限失败")
		}
		return
	}

	response.Success(ctx, map[string]bool{"has_permission": hasPermission})
}
