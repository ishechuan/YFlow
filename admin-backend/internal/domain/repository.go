package domain

import "context"
import "time"

// UserRepository 用户数据访问接口
type UserRepository interface {
	GetByID(ctx context.Context, id uint64) (*User, error)
	GetByIDs(ctx context.Context, ids []uint64) ([]*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetAll(ctx context.Context, limit, offset int, keyword string) ([]*User, int64, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint64) error
}

// ProjectRepository 项目数据访问接口
type ProjectRepository interface {
	GetByID(ctx context.Context, id uint64) (*Project, error)
	GetByIDs(ctx context.Context, ids []uint64) ([]*Project, error)
	GetBySlug(ctx context.Context, slug string) (*Project, error)
	GetAll(ctx context.Context, limit, offset int, keyword string) ([]*Project, int64, error)
	Create(ctx context.Context, project *Project) error
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id uint64) error
}

// LanguageRepository 语言数据访问接口
type LanguageRepository interface {
	GetByID(ctx context.Context, id uint64) (*Language, error)
	GetByIDs(ctx context.Context, ids []uint64) ([]*Language, error)
	GetByCode(ctx context.Context, code string) (*Language, error)
	GetAll(ctx context.Context) ([]*Language, error)
	Create(ctx context.Context, language *Language) error
	Update(ctx context.Context, language *Language) error
	Delete(ctx context.Context, id uint64) error
	GetDefault(ctx context.Context) (*Language, error)
}

// TranslationRepository 翻译数据访问接口
type TranslationRepository interface {
	GetByID(ctx context.Context, id uint64) (*Translation, error)
	GetByProjectID(ctx context.Context, projectID uint64, limit, offset int) ([]*Translation, int64, error)
	GetByProjectAndLanguage(ctx context.Context, projectID, languageID uint64) ([]*Translation, error)
	GetByProjectKeyLanguage(ctx context.Context, projectID uint64, keyName string, languageID uint64) (*Translation, error)
	GetByProjectKeyLanguages(ctx context.Context, keys []TranslationKey) ([]*Translation, error)
	GetMatrix(ctx context.Context, projectID uint64, limit, offset int, keyword string) (map[string]map[string]TranslationCell, int64, error)
	GetStats(ctx context.Context) (totalTranslations int, totalKeys int, err error)
	Create(ctx context.Context, translation *Translation) error
	CreateBatch(ctx context.Context, translations []*Translation) error
	UpsertBatch(ctx context.Context, translations []*Translation) error
	Update(ctx context.Context, translation *Translation) error
	Delete(ctx context.Context, id uint64) error
	DeleteBatch(ctx context.Context, ids []uint64) error
}

// TranslationKey 用于批量查询的翻译键
type TranslationKey struct {
	ProjectID  uint64
	KeyName    string
	LanguageID uint64
}

// TranslationCell 翻译矩阵单元格数据
type TranslationCell struct {
	ID        uint64    `json:"id"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProjectMemberRepository 项目成员数据访问接口
type ProjectMemberRepository interface {
	GetByProjectAndUser(ctx context.Context, projectID, userID uint64) (*ProjectMember, error)
	GetByProjectID(ctx context.Context, projectID uint64) ([]*ProjectMember, error)
	GetByUserID(ctx context.Context, userID uint64) ([]*ProjectMember, error)
	Create(ctx context.Context, member *ProjectMember) error
	Update(ctx context.Context, member *ProjectMember) error
	Delete(ctx context.Context, projectID, userID uint64) error
}

// InvitationRepository 邀请码数据访问接口
type InvitationRepository interface {
	GetByID(ctx context.Context, id uint64) (*Invitation, error)
	GetByCode(ctx context.Context, code string) (*Invitation, error)
	GetByInviter(ctx context.Context, inviterID uint64, limit, offset int) ([]*Invitation, int64, error)
	GetActiveInvitations(ctx context.Context) ([]*Invitation, error)
	Create(ctx context.Context, invitation *Invitation) error
	Update(ctx context.Context, invitation *Invitation) error
	MarkAsUsed(ctx context.Context, code string, userID uint64) error
	Revoke(ctx context.Context, code string) error
	Delete(ctx context.Context, code string) error
	DeleteByID(ctx context.Context, id uint64) error
}

// TranslationHistoryRepository 翻译历史数据访问接口
type TranslationHistoryRepository interface {
	Create(ctx context.Context, history *TranslationHistory) error
	CreateBatch(ctx context.Context, histories []*TranslationHistory) error
	ListByTranslationID(ctx context.Context, translationID uint64, limit, offset int) ([]*TranslationHistory, int64, error)
	ListByProjectID(ctx context.Context, projectID uint64, params TranslationHistoryQueryParams) ([]*TranslationHistory, int64, error)
	ListByUserID(ctx context.Context, userID uint64, params TranslationHistoryQueryParams) ([]*TranslationHistory, int64, error)
}
