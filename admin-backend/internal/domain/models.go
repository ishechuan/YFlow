package domain

import (
	"time"

	"gorm.io/gorm"
)

// User 用户领域模型
type User struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"unique;size:50;not null" json:"username"`
	Email     string    `gorm:"unique;size:100" json:"email"`
	Password  string    `gorm:"not null" json:"password"`
	Role      string    `gorm:"size:20;default:member;index:idx_user_role" json:"role"`     // admin, member, viewer
	Status    string    `gorm:"size:20;default:active;index:idx_user_status" json:"status"` // active, disabled
	CreatedBy uint64    `json:"created_by"`
	UpdatedBy uint64    `json:"updated_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Project 项目领域模型
type Project struct {
	ID           uint64         `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:100;not null;unique;index:idx_project_search" json:"name"` // 项目名称
	Description  string         `gorm:"size:500;index:idx_project_search" json:"description"`          // 项目描述
	Slug         string         `gorm:"size:100;not null;unique;index" json:"slug"`                    // 项目标识，用于URL
	Status       string         `gorm:"size:20;default:active;index:idx_project_status" json:"status"` // 项目状态：active, archived
	CreatedBy    uint64         `json:"created_by"`
	UpdatedBy    uint64         `json:"updated_by"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Translations []Translation  `gorm:"foreignKey:ProjectID" json:"-"` // 关联的翻译
}

// Language 语言领域模型
type Language struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Code      string         `gorm:"size:10;not null;unique" json:"code"`  // 语言代码，如 en, zh-CN
	Name      string         `gorm:"size:50;not null" json:"name"`         // 语言名称，如 English, 简体中文
	IsDefault bool           `gorm:"default:false" json:"is_default"`      // 是否为默认语言
	Status    string         `gorm:"size:20;default:active" json:"status"` // 状态：active, inactive
	CreatedBy uint64         `json:"created_by"`
	UpdatedBy uint64         `json:"updated_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Translation 翻译领域模型
type Translation struct {
	ID         uint64         `gorm:"primaryKey" json:"id"`
	ProjectID  uint64         `gorm:"not null;index:idx_translation_project;uniqueIndex:idx_translation_unique,priority:1" json:"project_id"`    // 关联的项目ID
	KeyName    string         `gorm:"size:255;not null;index:idx_translation_key;uniqueIndex:idx_translation_unique,priority:2" json:"key_name"` // 翻译键名
	Context    string         `gorm:"size:500" json:"context"`                                                                                   // 上下文说明
	LanguageID uint64         `gorm:"not null;index:idx_translation_language;uniqueIndex:idx_translation_unique,priority:3" json:"language_id"`  // 语言ID
	Value      string         `gorm:"type:text" json:"value"`                                                                                    // 翻译值
	Status     string         `gorm:"size:20;default:active;index:idx_translation_status" json:"status"`                                         // 状态：active, deprecated
	CreatedBy  uint64         `json:"created_by"`
	UpdatedBy  uint64         `json:"updated_by"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Project  Project  `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`  // 关联的项目
	Language Language `gorm:"foreignKey:LanguageID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"` // 关联的语言
}

// ProjectMember 项目成员关联模型
type ProjectMember struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	ProjectID uint64         `gorm:"not null;index:idx_project_member;uniqueIndex:idx_project_member_unique,priority:1" json:"project_id"`
	UserID    uint64         `gorm:"not null;index:idx_project_member;uniqueIndex:idx_project_member_unique,priority:2" json:"user_id"`
	Role      string         `gorm:"size:20;default:viewer;index:idx_project_member_role" json:"role"` // owner, editor, viewer
	CreatedBy uint64         `json:"created_by"`
	UpdatedBy uint64         `json:"updated_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Project Project `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	User    User    `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

// Invitation 邀请码领域模型
type Invitation struct {
	ID          uint64     `gorm:"primaryKey" json:"id"`
	Code        string     `gorm:"size:64;not null;uniqueIndex:idx_invitation_code" json:"code"`     // 邀请码
	InviterID   uint64     `gorm:"not null;index:idx_invitation_inviter" json:"inviter_id"`          // 邀请人ID
	Role        string     `gorm:"size:20;default:member" json:"role"`                               // 赋予被邀请人的角色: admin, member, viewer
	Status      string     `gorm:"size:20;default:active;index:idx_invitation_status" json:"status"` // 状态: active, used, revoked, expired
	ExpiresAt   time.Time  `gorm:"not null;index:idx_invitation_expires" json:"expires_at"`          // 过期时间
	UsedAt      *time.Time `json:"used_at,omitempty"`                                                // 使用时间
	UsedBy      *uint64    `json:"used_by,omitempty"`                                                // 被邀请人ID
	Description string     `gorm:"size:255" json:"description,omitempty"`                            // 邀请描述
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	Inviter *User `gorm:"foreignKey:InviterID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"inviter,omitempty"`
}

// TranslationHistory 翻译历史记录
type TranslationHistory struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	TranslationID *uint64   `gorm:"index:idx_translation_history_translation" json:"translation_id,omitempty"` // 关联翻译ID（删除后可为空）
	ProjectID     uint64    `gorm:"not null;index:idx_translation_history_project" json:"project_id"`          // 冗余字段，便于按项目查询
	KeyName       string    `gorm:"size:255;index:idx_translation_history_key" json:"key_name"`                // 翻译键名
	LanguageID    uint64    `gorm:"index:idx_translation_history_language" json:"language_id"`                 // 语言ID
	OldValue      *string   `gorm:"type:text" json:"old_value,omitempty"`                                      // 旧值（create操作为空）
	NewValue      *string   `gorm:"type:text" json:"new_value,omitempty"`                                      // 新值（delete操作为空）
	Operation     string    `gorm:"size:20;index:idx_translation_history_operation" json:"operation"`          // 操作类型：create|update|delete|import|export|machine_translate
	OperatedBy    uint64    `gorm:"not null;index:idx_translation_history_user" json:"operated_by"`            // 操作者用户ID
	OperatedAt    time.Time `gorm:"not null;index:idx_translation_history_time" json:"operated_at"`            // 操作时间
	Metadata      string    `gorm:"type:json" json:"metadata,omitempty"`                                       // 额外信息（JSON格式）

	Translation *Translation `gorm:"foreignKey:TranslationID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"-"`
}

// InvitationStatus 邀请状态常量
const (
	InvitationStatusActive  = "active"
	InvitationStatusUsed    = "used"
	InvitationStatusRevoked = "revoked"
	InvitationStatusExpired = "expired"
)

// IsValid 检查邀请是否有效
func (i *Invitation) IsValid() bool {
	if i.Status != InvitationStatusActive {
		return false
	}
	if time.Now().After(i.ExpiresAt) {
		return false
	}
	return true
}
