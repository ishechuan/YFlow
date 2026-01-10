package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"yflow/internal/domain"
)

// TranslationService 翻译服务实现
type TranslationService struct {
	translationRepo domain.TranslationRepository
	projectRepo     domain.ProjectRepository
	languageRepo    domain.LanguageRepository
	historyRepo     domain.TranslationHistoryRepository
}

// NewTranslationService 创建翻译服务实例
func NewTranslationService(
	translationRepo domain.TranslationRepository,
	projectRepo domain.ProjectRepository,
	languageRepo domain.LanguageRepository,
	historyRepo domain.TranslationHistoryRepository,
) *TranslationService {
	return &TranslationService{
		translationRepo: translationRepo,
		projectRepo:     projectRepo,
		languageRepo:    languageRepo,
		historyRepo:     historyRepo,
	}
}

// Create 创建翻译
func (s *TranslationService) Create(ctx context.Context, input domain.TranslationInput, userID uint64) (*domain.Translation, error) {
	// 验证项目是否存在
	_, err := s.projectRepo.GetByID(ctx, input.ProjectID)
	if err != nil {
		return nil, domain.ErrProjectNotFound
	}

	// 验证语言是否存在
	_, err = s.languageRepo.GetByID(ctx, input.LanguageID)
	if err != nil {
		return nil, domain.ErrLanguageNotFound
	}

	// 检查翻译是否已存在
	keyName := strings.TrimSpace(input.KeyName)
	existing, err := s.translationRepo.GetByProjectKeyLanguage(ctx, input.ProjectID, keyName, input.LanguageID)
	if err == nil && existing != nil {
		return nil, domain.NewAppErrorWithDetails(
			domain.ErrorTypeConflict,
			"TRANSLATION_EXISTS",
			"该项目中已存在相同键名和语言的翻译",
			fmt.Sprintf("项目ID: %d, 键名: %s, 语言ID: %d", input.ProjectID, keyName, input.LanguageID),
		)
	}

	// 创建翻译
	translation := &domain.Translation{
		ProjectID:  input.ProjectID,
		KeyName:    keyName,
		Context:    strings.TrimSpace(input.Context),
		LanguageID: input.LanguageID,
		Value:      strings.TrimSpace(input.Value),
		Status:     "active",
		CreatedBy:  userID,
		UpdatedBy:  userID,
	}

	if err := s.translationRepo.Create(ctx, translation); err != nil {
		// 检查是否是唯一约束冲突错误
		if isDuplicateKeyError(err) {
			return nil, domain.NewAppErrorWithDetails(
				domain.ErrorTypeConflict,
				"TRANSLATION_EXISTS",
				"该项目中已存在相同键名和语言的翻译",
				fmt.Sprintf("项目ID: %d, 键名: %s, 语言ID: %d", input.ProjectID, keyName, input.LanguageID),
			)
		}
		return nil, err
	}

	// 记录创建历史
	history := &domain.TranslationHistory{
		TranslationID: &translation.ID,
		ProjectID:     translation.ProjectID,
		KeyName:       translation.KeyName,
		LanguageID:    translation.LanguageID,
		OldValue:      nil, // 创建操作没有旧值
		NewValue:      &translation.Value,
		Operation:     "create",
		OperatedBy:    userID,
		Metadata:      "{}",
	}
	// 忽略历史记录错误，不影响主操作
	_ = s.historyRepo.Create(ctx, history)

	return translation, nil
}

// CreateBatch 批量创建翻译
func (s *TranslationService) CreateBatch(ctx context.Context, inputs []domain.TranslationInput) error {
	if len(inputs) == 0 {
		return nil
	}

	// 收集所有请求中的项目和语言ID
	projectIDSet := make(map[uint64]bool)
	languageIDSet := make(map[uint64]bool)

	for _, input := range inputs {
		projectIDSet[input.ProjectID] = true
		languageIDSet[input.LanguageID] = true
	}

	// 转换为切片
	projectIDs := make([]uint64, 0, len(projectIDSet))
	for id := range projectIDSet {
		projectIDs = append(projectIDs, id)
	}
	languageIDs := make([]uint64, 0, len(languageIDSet))
	for id := range languageIDSet {
		languageIDs = append(languageIDs, id)
	}

	// 批量验证项目 (修复 N+1 查询)
	projects, err := s.projectRepo.GetByIDs(ctx, projectIDs)
	if err != nil {
		return err
	}
	if len(projects) != len(projectIDs) {
		return domain.ErrProjectNotFound
	}

	// 批量验证语言 (修复 N+1 查询)
	languages, err := s.languageRepo.GetByIDs(ctx, languageIDs)
	if err != nil {
		return err
	}
	if len(languages) != len(languageIDs) {
		return domain.ErrLanguageNotFound
	}

	// 构建所有要查询的键（修复 N+1 查询问题）
	keys := make([]domain.TranslationKey, 0, len(inputs))
	for _, input := range inputs {
		keys = append(keys, domain.TranslationKey{
			ProjectID:  input.ProjectID,
			KeyName:    strings.TrimSpace(input.KeyName),
			LanguageID: input.LanguageID,
		})
	}

	// 批量查询已存在的翻译
	existingTranslations, err := s.translationRepo.GetByProjectKeyLanguages(ctx, keys)
	if err != nil {
		return err
	}

	// 构建已存在翻译的 map 用于快速查找
	existingMap := make(map[string]*domain.Translation)
	for _, t := range existingTranslations {
		key := fmt.Sprintf("%d:%s:%d", t.ProjectID, t.KeyName, t.LanguageID)
		existingMap[key] = t
	}

	// 检查重复翻译并转换为domain对象
	translations := make([]*domain.Translation, 0, len(inputs))
	duplicates := make([]string, 0)

	for _, input := range inputs {
		keyName := strings.TrimSpace(input.KeyName)
		mapKey := fmt.Sprintf("%d:%s:%d", input.ProjectID, keyName, input.LanguageID)

		// 使用 map 快速查找
		if existing, exists := existingMap[mapKey]; exists {
			duplicates = append(duplicates, fmt.Sprintf("项目ID:%d, 键名:%s, 语言ID:%d", existing.ProjectID, existing.KeyName, existing.LanguageID))
			continue
		}

		translations = append(translations, &domain.Translation{
			ProjectID:  input.ProjectID,
			KeyName:    keyName,
			Context:    strings.TrimSpace(input.Context),
			LanguageID: input.LanguageID,
			Value:      strings.TrimSpace(input.Value),
			Status:     "active",
		})
	}

	// 如果有重复项，返回错误
	if len(duplicates) > 0 {
		return domain.NewAppErrorWithDetails(
			domain.ErrorTypeConflict,
			"TRANSLATION_EXISTS",
			"批量创建中存在重复的翻译",
			fmt.Sprintf("重复项: %s", strings.Join(duplicates, "; ")),
		)
	}

	// 如果没有有效的翻译需要创建
	if len(translations) == 0 {
		return nil
	}

	return s.translationRepo.CreateBatch(ctx, translations)
}

// UpsertBatch 批量创建或更新翻译
// 如果翻译已存在（基于 project_id + key_name + language_id），则更新
// 如果不存在，则创建
func (s *TranslationService) UpsertBatch(ctx context.Context, inputs []domain.TranslationInput) error {
	if len(inputs) == 0 {
		return nil
	}

	// 收集所有请求中的项目和语言ID
	projectIDSet := make(map[uint64]bool)
	languageIDSet := make(map[uint64]bool)

	for _, input := range inputs {
		projectIDSet[input.ProjectID] = true
		languageIDSet[input.LanguageID] = true
	}

	// 转换为切片
	projectIDs := make([]uint64, 0, len(projectIDSet))
	for id := range projectIDSet {
		projectIDs = append(projectIDs, id)
	}
	languageIDs := make([]uint64, 0, len(languageIDSet))
	for id := range languageIDSet {
		languageIDs = append(languageIDs, id)
	}

	// 批量验证项目 (修复 N+1 查询)
	projects, err := s.projectRepo.GetByIDs(ctx, projectIDs)
	if err != nil {
		return err
	}
	if len(projects) != len(projectIDs) {
		return domain.ErrProjectNotFound
	}

	// 批量验证语言 (修复 N+1 查询)
	languages, err := s.languageRepo.GetByIDs(ctx, languageIDs)
	if err != nil {
		return err
	}
	if len(languages) != len(languageIDs) {
		return domain.ErrLanguageNotFound
	}

	// 转换为 domain 对象
	translations := make([]*domain.Translation, 0, len(inputs))
	for _, input := range inputs {
		translations = append(translations, &domain.Translation{
			ProjectID:  input.ProjectID,
			KeyName:    strings.TrimSpace(input.KeyName),
			Context:    strings.TrimSpace(input.Context),
			LanguageID: input.LanguageID,
			Value:      strings.TrimSpace(input.Value),
			Status:     "active",
		})
	}

	// 使用 UpsertBatch 而不是 CreateBatch
	return s.translationRepo.UpsertBatch(ctx, translations)
}

// CreateBatchFromRequest 从批量翻译参数创建或更新翻译
// 现在使用 UpsertBatch，支持创建和更新操作
func (s *TranslationService) CreateBatchFromRequest(ctx context.Context, params domain.BatchTranslationParams) error {
	// 获取所有语言
	languages, err := s.languageRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	// 创建语言代码到ID的映射
	languageCodeToID := make(map[string]uint64)
	for _, lang := range languages {
		languageCodeToID[lang.Code] = lang.ID
	}

	// 转换为标准翻译请求
	var inputs []domain.TranslationInput
	for langCode, value := range params.Translations {
		// 跳过空值
		if value == "" {
			continue
		}

		if langID, exists := languageCodeToID[langCode]; exists {
			inputs = append(inputs, domain.TranslationInput{
				ProjectID:  params.ProjectID,
				KeyName:    params.KeyName,
				Context:    params.Context,
				LanguageID: langID,
				Value:      value,
			})
		}
	}

	if len(inputs) == 0 {
		return fmt.Errorf("no valid translations to create")
	}

	// 使用 UpsertBatch 而不是 CreateBatch，支持创建和更新
	return s.UpsertBatch(ctx, inputs)
}

// GetByID 根据ID获取翻译
func (s *TranslationService) GetByID(ctx context.Context, id uint64) (*domain.Translation, error) {
	return s.translationRepo.GetByID(ctx, id)
}

// GetByProjectID 根据项目ID获取翻译
func (s *TranslationService) GetByProjectID(ctx context.Context, projectID uint64, limit, offset int) ([]*domain.Translation, int64, error) {
	// 验证项目是否存在
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, 0, domain.ErrProjectNotFound
	}

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.translationRepo.GetByProjectID(ctx, projectID, limit, offset)
}

// GetMatrix 获取翻译矩阵
func (s *TranslationService) GetMatrix(ctx context.Context, projectID uint64, limit, offset int, keyword string) (map[string]map[string]domain.TranslationCell, int64, error) {
	// 验证项目是否存在
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, 0, domain.ErrProjectNotFound
	}

	return s.translationRepo.GetMatrix(ctx, projectID, limit, offset, keyword)
}

// Update 更新翻译
func (s *TranslationService) Update(ctx context.Context, id uint64, input domain.TranslationInput, userID uint64) (*domain.Translation, error) {
	// 获取现有翻译
	translation, err := s.translationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 保存旧值用于历史记录
	oldValue := translation.Value

	// 如果项目ID改变，验证新项目
	if input.ProjectID != 0 && input.ProjectID != translation.ProjectID {
		_, err := s.projectRepo.GetByID(ctx, input.ProjectID)
		if err != nil {
			return nil, domain.ErrProjectNotFound
		}
		translation.ProjectID = input.ProjectID
	}

	// 如果语言ID改变，验证新语言
	if input.LanguageID != 0 && input.LanguageID != translation.LanguageID {
		_, err := s.languageRepo.GetByID(ctx, input.LanguageID)
		if err != nil {
			return nil, domain.ErrLanguageNotFound
		}
		translation.LanguageID = input.LanguageID
	}

	// 更新其他字段
	if input.KeyName != "" {
		translation.KeyName = strings.TrimSpace(input.KeyName)
	}

	if input.Context != "" {
		translation.Context = strings.TrimSpace(input.Context)
	}

	if input.Value != "" {
		translation.Value = strings.TrimSpace(input.Value)
	}

	// 更新UpdatedBy字段
	translation.UpdatedBy = userID

	// 保存更新
	if err := s.translationRepo.Update(ctx, translation); err != nil {
		return nil, err
	}

	// 记录更新历史
	newValue := translation.Value
	history := &domain.TranslationHistory{
		TranslationID: &translation.ID,
		ProjectID:     translation.ProjectID,
		KeyName:       translation.KeyName,
		LanguageID:    translation.LanguageID,
		OldValue:      &oldValue,
		NewValue:      &newValue,
		Operation:     "update",
		OperatedBy:    userID,
		Metadata:      "{}", // 可以记录变更的字段
	}
	// 忽略历史记录错误，不影响主操作
	_ = s.historyRepo.Create(ctx, history)

	return translation, nil
}

// Delete 删除翻译
func (s *TranslationService) Delete(ctx context.Context, id uint64, userID uint64) error {
	// 检查翻译是否存在并获取详情
	translation, err := s.translationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 记录删除历史
	oldValue := translation.Value
	history := &domain.TranslationHistory{
		TranslationID: &translation.ID,
		ProjectID:     translation.ProjectID,
		KeyName:       translation.KeyName,
		LanguageID:    translation.LanguageID,
		OldValue:      &oldValue,
		NewValue:      nil,
		Operation:     "delete",
		OperatedBy:    userID,
		Metadata:      "{}",
	}
	// 忽略历史记录错误，不影响主操作
	_ = s.historyRepo.Create(ctx, history)

	return s.translationRepo.Delete(ctx, id)
}

// DeleteBatch 批量删除翻译
func (s *TranslationService) DeleteBatch(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}

	return s.translationRepo.DeleteBatch(ctx, ids)
}

// Export 导出翻译
func (s *TranslationService) Export(ctx context.Context, projectID uint64, format string) ([]byte, error) {
	// 验证项目是否存在
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, domain.ErrProjectNotFound
	}

	// 获取翻译矩阵（导出所有数据，不分页）
	matrix, _, err := s.translationRepo.GetMatrix(ctx, projectID, -1, 0, "")
	if err != nil {
		return nil, err
	}

	// 转换为简单格式 (key -> language -> value)
	simpleMatrix := make(map[string]map[string]string)
	for key, langs := range matrix {
		simpleMatrix[key] = make(map[string]string)
		for lang, cell := range langs {
			simpleMatrix[key][lang] = cell.Value
		}
	}

	switch format {
	case "json":
		return json.MarshalIndent(simpleMatrix, "", "  ")
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// Import 导入翻译
func (s *TranslationService) Import(ctx context.Context, projectID uint64, data []byte, format string) error {
	// 验证项目是否存在
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return domain.ErrProjectNotFound
	}

	switch format {
	case "json":
		return s.importFromJSON(ctx, projectID, data)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// importFromJSON 从JSON导入翻译
func (s *TranslationService) importFromJSON(ctx context.Context, projectID uint64, data []byte) error {
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	// 获取所有语言
	languages, err := s.languageRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	// 创建语言代码到ID的映射
	languageCodeToID := make(map[string]uint64)
	for _, lang := range languages {
		languageCodeToID[lang.Code] = lang.ID
	}

	// 转换为翻译请求
	var inputs []domain.TranslationInput

	// 检测数据格式并转换
	matrix := s.normalizeImportData(rawData)

	for key, translations := range matrix {
		for langCode, value := range translations {
			if langID, exists := languageCodeToID[langCode]; exists {
				inputs = append(inputs, domain.TranslationInput{
					ProjectID:  projectID,
					KeyName:    key,
					LanguageID: langID,
					Value:      value,
				})
			}
		}
	}

	if len(inputs) == 0 {
		return fmt.Errorf("no valid translations found in import data")
	}

	return s.CreateBatch(ctx, inputs)
}

// normalizeImportData 标准化导入数据格式
// 支持两种格式：
// 1. key -> {language: value} (标准格式)
// 2. language -> {key: value} (前端格式)
func (s *TranslationService) normalizeImportData(rawData map[string]interface{}) map[string]map[string]string {
	matrix := make(map[string]map[string]string)

	// 检测数据格式
	if s.isLanguageToKeyFormat(rawData) {
		// 前端格式: language -> {key: value}
		for langCode, keysInterface := range rawData {
			if keys, ok := keysInterface.(map[string]interface{}); ok {
				for key, valueInterface := range keys {
					if value, ok := valueInterface.(string); ok {
						if matrix[key] == nil {
							matrix[key] = make(map[string]string)
						}
						matrix[key][langCode] = value
					}
				}
			}
		}
	} else {
		// 标准格式: key -> {language: value}
		for key, languagesInterface := range rawData {
			if languages, ok := languagesInterface.(map[string]interface{}); ok {
				matrix[key] = make(map[string]string)
				for langCode, valueInterface := range languages {
					if value, ok := valueInterface.(string); ok {
						matrix[key][langCode] = value
					}
				}
			}
		}
	}

	return matrix
}

// isLanguageToKeyFormat 检测是否为 language -> {key: value} 格式
func (s *TranslationService) isLanguageToKeyFormat(rawData map[string]interface{}) bool {
	// 检查第一层的键是否看起来像语言代码
	for key := range rawData {
		// 如果键是短的字符串（1-5个字符），可能是语言代码
		if len(key) <= 5 && isLikelyLanguageCode(key) {
			return true
		}
		// 如果键包含点号，更可能是翻译键而不是语言代码
		if strings.Contains(key, ".") {
			return false
		}
	}
	return false
}

// isLikelyLanguageCode 判断字符串是否像语言代码
func isLikelyLanguageCode(code string) bool {
	// 常见的语言代码模式
	commonLanguageCodes := []string{
		"en", "zh", "ja", "ko", "fr", "de", "es", "pt", "ru", "ar", "hi", "th", "vi", "id", "ms", "tr", "it", "pl", "nl", "sv", "da", "no", "fi",
		"zh_CN", "zh_TW", "en_US", "en_GB", "pt_BR", "es_ES", "fr_FR", "de_DE",
	}

	for _, lang := range commonLanguageCodes {
		if code == lang {
			return true
		}
	}

	// 简单的启发式规则：长度为2-5的字符串，只包含字母、数字和连字符
	if len(code) >= 2 && len(code) <= 5 {
		for _, c := range code {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-') {
				return false
			}
		}
		return true
	}

	return false
}

// isDuplicateKeyError 检查是否是重复键错误
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	// MySQL重复键错误模式
	return strings.Contains(errStr, "duplicate entry") ||
		strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "idx_translation_unique")
}
