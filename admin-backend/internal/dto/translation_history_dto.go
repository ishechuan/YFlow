package dto

import (
	"time"
	"yflow/internal/api/response"
	"yflow/internal/domain"
)

// TranslationHistoryResponse 翻译历史响应
type TranslationHistoryResponse struct {
	ID            uint64    `json:"id"`
	TranslationID *uint64   `json:"translation_id,omitempty"`
	ProjectID     uint64    `json:"project_id"`
	KeyName       string    `json:"key_name"`
	LanguageID    uint64    `json:"language_id"`
	OldValue      *string   `json:"old_value,omitempty"`
	NewValue      *string   `json:"new_value,omitempty"`
	Operation     string    `json:"operation"`
	OperatedBy    uint64    `json:"operated_by"`
	OperatedAt    time.Time `json:"operated_at"`
	Metadata      string    `json:"metadata,omitempty"`
}

// ListTranslationHistoryRequest 翻译历史列表请求
type ListTranslationHistoryRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1" default:"1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100" default:"10"`
	Operation string `form:"operation" binding:"omitempty"`
	StartDate string `form:"start_date" binding:"omitempty,datetime=2006-01-02"`
	EndDate   string `form:"end_date" binding:"omitempty,datetime=2006-01-02"`
}

// ToDomainParams 转换为领域层查询参数
func (r *ListTranslationHistoryRequest) ToDomainParams() domain.TranslationHistoryQueryParams {
	return domain.TranslationHistoryQueryParams{
		Limit:     r.PageSize,
		Offset:    (r.Page - 1) * r.PageSize,
		Operation: r.Operation,
		StartDate: r.StartDate,
		EndDate:   r.EndDate,
	}
}

// TranslationHistoryListResponse 翻译历史列表响应
type TranslationHistoryListResponse struct {
	Histories []*TranslationHistoryResponse `json:"histories"`
	Meta      *response.Meta                `json:"meta"`
}
