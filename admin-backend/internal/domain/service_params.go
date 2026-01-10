package domain

// ========== User Service Params ==========

// LoginParams 登录参数
type LoginParams struct {
	Username string
	Password string
}

// LoginResult 登录结果
type LoginResult struct {
	User         *User
	AccessToken  string
	RefreshToken string
}

// CreateUserParams 创建用户参数
type CreateUserParams struct {
	Username string
	Email    string
	Password string
	Role     string
}

// UpdateUserParams 更新用户参数
type UpdateUserParams struct {
	Username string
	Email    string
	Role     string
	Status   string
}

// ChangePasswordParams 修改密码参数
type ChangePasswordParams struct {
	OldPassword string
	NewPassword string
}

// ========== Project Service Params ==========

// CreateProjectParams 创建项目参数
type CreateProjectParams struct {
	Name        string
	Description string
}

// UpdateProjectParams 更新项目参数
type UpdateProjectParams struct {
	Name        string
	Description string
	Status      string
}

// ========== Language Service Params ==========

// CreateLanguageParams 创建语言参数
type CreateLanguageParams struct {
	Code      string
	Name      string
	IsDefault bool
}

// ========== Translation Service Params ==========

// TranslationInput 翻译输入
type TranslationInput struct {
	ProjectID  uint64
	LanguageID uint64
	KeyName    string
	Context    string
	Value      string
}

// BatchTranslationParams 批量翻译参数
type BatchTranslationParams struct {
	ProjectID    uint64
	KeyName      string
	Context      string
	Translations map[string]string // language_code -> value
}

// ========== Dashboard Service Params ==========

// DashboardStats 仪表板统计结果
type DashboardStats struct {
	TotalProjects     int `json:"total_projects"`
	TotalLanguages    int `json:"total_languages"`
	TotalTranslations int `json:"total_translations"`
	TotalKeys         int `json:"total_keys"`
}

// ========== Project Member Service Params ==========

// AddMemberParams 添加成员参数
type AddMemberParams struct {
	MemberUserID uint64
	Role         string
}

// UpdateMemberRoleParams 更新成员角色参数
type UpdateMemberRoleParams struct {
	Role string
}

// ProjectMemberInfo 项目成员信息
type ProjectMemberInfo struct {
	ID       uint64
	UserID   uint64
	Username string
	Email    string
	Role     string
}

// ========== Translation History Service Params ==========

// TranslationHistoryQueryParams 翻译历史查询参数
type TranslationHistoryQueryParams struct {
	Limit     int
	Offset    int
	Operation string // 操作类型筛选
	StartDate string // 开始时间 (格式: 2006-01-02)
	EndDate   string // 结束时间 (格式: 2006-01-02)
}
