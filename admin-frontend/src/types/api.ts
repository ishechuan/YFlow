/**
 * 后端统一 API 响应结构
 */
export interface APIResponse<T = any> {
  success: boolean
  data?: T
  error?: ErrorInfo
  meta?: PaginationMeta
}

/**
 * 错误信息结构
 */
export interface ErrorInfo {
  code: string
  message: string
  details?: string
}

/**
 * 分页元数据
 */
export interface PaginationMeta {
  page: number
  page_size: number
  total_count: number
  total_pages: number
}

/**
 * 用户信息
 */
export interface User {
  id: number
  username: string
  email?: string
  role?: string
  status?: string
  created_at: string
  updated_at: string
}

/**
 * 登录响应数据
 */
export interface LoginResponse {
  token: string
  refresh_token: string
  user: User
}

/**
 * 登录请求参数
 */
export interface LoginParams {
  username: string
  password: string
}

/**
 * 刷新 Token 请求参数
 */
export interface RefreshTokenParams {
  refresh_token: string
}

/**
 * 仪表板统计信息
 */
export interface DashboardStats {
  TotalProjects: number
  TotalLanguages: number
  TotalKeys: number
  TotalTranslations: number
}

/**
 * 项目实体
 */
export interface Project {
  id: number
  name: string
  slug: string
  description?: string
  status: 'active' | 'archived'
  created_at: string
  updated_at: string
}

/**
 * 创建项目请求参数
 */
export interface CreateProjectRequest {
  name: string
  description?: string
}

/**
 * 更新项目请求参数
 */
export interface UpdateProjectRequest {
  name?: string
  description?: string
  status?: 'active' | 'archived'
}

/**
 * 项目列表查询参数
 */
export interface ProjectListParams {
  page?: number
  page_size?: number
  keyword?: string
}

/**
 * 项目列表响应数据
 */
export interface ProjectListResponse {
  data: Project[]
  meta: PaginationMeta
}

// ==================== 邀请相关类型 ====================

/**
 * 邀请人信息
 */
export interface InvitationInviter {
  id: number
  username: string
  email: string
  role: string
}

/**
 * 验证邀请码响应
 */
export interface ValidateInvitationResponse {
  valid: boolean
  inviter?: InvitationInviter
  role: string
  expires_at: string
  message?: string
}

/**
 * 使用邀请码注册请求参数
 */
export interface RegisterParams {
  code: string
  username: string
  email: string
  password: string
}

/**
 * 使用邀请码注册响应
 */
export interface RegisterResponse {
  message: string
  user: User
}

/**
 * 创建邀请请求参数
 */
export interface CreateInvitationParams {
  role?: 'admin' | 'member' | 'viewer'
  expires_in_days?: number
  description?: string
}

/**
 * 创建邀请响应
 */
export interface CreateInvitationResponse {
  code: string
  invitation_url: string
  role: string
  expires_at: string
  description?: string
}

/**
 * 邀请详情
 */
export interface Invitation {
  id: number
  code: string
  inviter_id: number
  inviter?: InvitationInviter
  role: string
  status: 'active' | 'used' | 'revoked' | 'expired'
  expires_at: string
  used_at?: string
  used_by?: number
  description?: string
  created_at: string
}

/**
 * 邀请列表响应
 */
export interface InvitationListResponse {
  data: {
    invitations: Invitation[]
    total: number
  }
  meta?: PaginationMeta
}

// ==================== 用户管理相关类型 ====================

/**
 * 创建用户请求参数
 */
export interface CreateUserRequest {
  username: string
  email: string
  password: string
  role: 'admin' | 'member' | 'viewer'
}

/**
 * 更新用户请求参数
 */
export interface UpdateUserRequest {
  username?: string
  email?: string
  role?: 'admin' | 'member' | 'viewer'
  status?: 'active' | 'disabled'
}

/**
 * 修改密码请求参数
 */
export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

/**
 * 重置密码请求参数
 */
export interface ResetPasswordRequest {
  new_password: string
}

/**
 * 用户列表查询参数
 */
export interface UserListParams {
  page?: number
  page_size?: number
  keyword?: string
}

/**
 * 用户列表响应数据
 */
export interface UserListResponse {
  data: User[]
  meta: PaginationMeta
}
