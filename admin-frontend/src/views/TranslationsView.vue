<template>
  <div class="translations-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">翻译管理</h1>
        <p class="page-subtitle">管理所有项目的翻译内容</p>
      </div>
      <el-select
        v-model="selectedProjectId"
        placeholder="选择项目"
        filterable
        @change="handleProjectChange"
        class="project-selector"
        size="large"
      >
        <el-option
          v-for="project in projects"
          :key="project.id"
          :label="project.name"
          :value="project.id"
        />
      </el-select>
    </div>

    <div v-if="selectedProjectId" class="content-wrapper">
      <!-- 工具栏 -->
      <div class="toolbar-card">
        <div class="toolbar">
          <div class="search-input-group">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索翻译键..."
              clearable
              prefix-icon="Search"
              class="search-input"
              @input="handleSearch"
            />
          </div>
          <div class="toolbar-actions">
            <el-button type="primary" @click="showAddKeyDialog = true" icon="Plus">
              添加翻译键
            </el-button>
            <el-button @click="handleExport" :loading="loading" icon="Download">
              导出
            </el-button>
            <el-button @click="showImportDialog = true" icon="Upload">
              导入
            </el-button>
            <el-button type="success" @click="showMachineTranslationDialog = true" icon="MagicStick">
              机器翻译
            </el-button>
          </div>
        </div>
      </div>

      <!-- 翻译表格 -->
      <div class="table-card">
        <el-table
          v-loading="loading"
          :data="matrix?.rows || []"
          style="width: 100%"
          class="translation-table"
          :empty-text="'暂无翻译数据'"
        >
          <!-- 翻译键列 -->
          <el-table-column prop="key_name" label="翻译键" fixed width="220" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="key-text">{{ row.key_name }}</span>
            </template>
          </el-table-column>

          <!-- 上下文列 -->
          <el-table-column prop="context" label="上下文" width="180" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="context-text">{{ row.context || '-' }}</span>
            </template>
          </el-table-column>

          <!-- 动态语言列 -->
          <el-table-column
            v-for="lang in matrix?.languages || []"
            :key="lang.id"
            :label="lang.name"
            min-width="250"
          >
            <template #header>
              <div class="column-header">
                {{ lang.name }} <el-tag size="small" class="lang-tag">{{ lang.code }}</el-tag>
              </div>
            </template>
            <template #default="{ row }">
              <div
                class="cell-wrapper"
                @click="editCell(row.key_name, lang)"
                :class="{ 'is-empty': !row.translations[lang.code]?.value }"
              >
                <!-- 编辑模式 -->
                <div
                  v-if="editingCell?.keyName === row.key_name && editingCell?.languageId === lang.id"
                  class="cell-editing"
                  @click.stop
                >
                  <el-input
                    v-model="editingValue"
                    type="textarea"
                    :rows="2"
                    ref="editInput"
                    @blur="saveCell"
                    @keydown.enter.exact.prevent="saveCell"
                    @keydown.esc="cancelEdit"
                    placeholder="输入翻译..."
                  />
                </div>
                <!-- 显示模式 -->
                <div v-else class="cell-display">
                  <span v-if="row.translations[lang.code]?.value" class="cell-value">
                    {{ row.translations[lang.code].value }}
                  </span>
                  <span v-else class="cell-placeholder">点击添加翻译</span>
                  <div v-if="row.translations[lang.code]?.updated_at" class="cell-time">
                    {{ formatTime(row.translations[lang.code].updated_at) }}
                  </div>
                </div>
              </div>
            </template>
          </el-table-column>

          <!-- 操作列 -->
          <el-table-column label="操作" width="200" fixed="right" align="center">
            <template #default="{ row }">
              <button class="action-button action-delete" @click="handleDeleteKey(row.key_name)">
                <el-icon><Delete /></el-icon>
                <span>删除</span>
              </button>
            </template>
          </el-table-column>
        </el-table>

        <!-- 分页 -->
        <div class="pagination-container" v-if="matrix && matrix.total_count > 0">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            :total="matrix.total_count"
            @size-change="loadMatrix"
            @current-change="loadMatrix"
          />
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-card">
      <div class="empty-state">
        <el-icon :size="64" class="empty-icon"><FolderOpened /></el-icon>
        <h3 class="empty-title">请选择一个项目</h3>
        <p class="empty-description">选择项目后即可开始管理翻译内容</p>
      </div>
    </div>

    <!-- 添加翻译键对话框 -->
    <el-dialog
      v-model="showAddKeyDialog"
      title="添加翻译键"
      width="600px"
      :close-on-click-modal="false"
      destroy-on-close
      class="translation-dialog"
    >
      <el-form :model="newKey" label-width="100px" @submit.prevent="handleAddKey" class="translation-form">
        <el-form-item label="翻译键名" required>
          <el-input v-model="newKey.keyName" placeholder="例如: welcome.message" />
        </el-form-item>
        <el-form-item label="上下文">
          <el-input v-model="newKey.context" placeholder="说明这个翻译键的使用场景" />
        </el-form-item>

        <el-divider content-position="left">翻译内容</el-divider>

        <div class="language-inputs-scroll">
          <el-form-item
            v-for="lang in availableLanguages"
            :key="lang.id"
            :label="lang.name"
          >
            <el-input
              v-model="newKey.translations[lang.code]"
              :placeholder="`输入 ${lang.name} 翻译（可选）`"
              type="textarea"
              autosize
            />
          </el-form-item>
        </div>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showAddKeyDialog = false">取消</el-button>
          <el-button type="primary" @click="handleAddKey" :loading="loading">
            添加
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 导入对话框 -->
    <el-dialog
      v-model="showImportDialog"
      title="导入翻译"
      width="500px"
      class="import-dialog"
    >
      <div class="import-wrapper">
        <p class="import-tip">请选择标准 JSON 格式的翻译文件</p>
        <input
          type="file"
          ref="fileInput"
          @change="handleFileSelect"
          accept=".json"
          class="file-input"
          style="display: none"
        />
        <div class="upload-area" @click="fileInput?.click()">
            <el-icon class="upload-icon" :size="40"><UploadFilled /></el-icon>
            <div class="upload-text" v-if="!importFile">点击选择文件</div>
            <div class="file-name" v-else>
                <el-icon><Document /></el-icon> {{ importFile.name }}
            </div>
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showImportDialog = false">取消</el-button>
          <el-button type="primary" @click="handleImport" :disabled="!importFile" :loading="loading">
            导入
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 机器翻译对话框 -->
    <MachineTranslationDialog
      v-model="showMachineTranslationDialog"
      :project-id="selectedProjectId || 0"
      :show-tabs="true"
      title="机器翻译"
      @filled="loadMatrix"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, watch, computed } from 'vue'
import { getTranslationMatrix, batchCreateTranslations, exportTranslations, importTranslations } from '@/services/translation'
import { getLanguages } from '@/services/language'
import type { TranslationMatrix, Language, BatchTranslationRequest, ImportTranslationsData } from '@/types/translation'
import type { Project } from '@/types/api'
import api from '@/services/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Plus, Download, Upload, UploadFilled, FolderOpened, Document, Delete, MagicStick } from '@element-plus/icons-vue'
import MachineTranslationDialog from '@/components/MachineTranslationDialog.vue'

// State
const projects = ref<Project[]>([])
const selectedProjectId = ref<number | null>(null)
const matrix = ref<TranslationMatrix | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
const allLanguages = ref<Language[]>([])

// Search and Pagination
const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(20)

// Editing
const editingCell = ref<{ keyName: string; languageId: number } | null>(null)
const editingValue = ref('')
const editInput = ref<HTMLTextAreaElement[] | null>(null)

// Dialogs
const showAddKeyDialog = ref(false)
const showImportDialog = ref(false)
const showMachineTranslationDialog = ref(false)
const newKey = ref({ keyName: '', context: '', translations: {} as Record<string, string> })
const importFile = ref<File | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

// Available languages (active only)
const availableLanguages = computed(() => {
  return allLanguages.value.filter(l => l.status === 'active')
})

// Load projects on mount
onMounted(async () => {
  await loadProjects()
  await loadAllLanguages()
})

// Load all languages
const loadAllLanguages = async () => {
  try {
    allLanguages.value = await getLanguages()
  } catch (err: any) {
    console.error('Failed to load languages:', err)
  }
}

// Load projects list
const loadProjects = async () => {
  try {
    const response = await api.get('/projects', { params: { page: 1, page_size: 100 } })
    projects.value = response.data || []
  } catch (err: any) {
    console.error('Failed to load projects:', err)
  }
}

// Handle project change
const handleProjectChange = () => {
  currentPage.value = 1
  searchKeyword.value = ''
  loadMatrix()
}

// Load translation matrix
const loadMatrix = async () => {
  if (!selectedProjectId.value) return

  loading.value = true
  error.value = null

  try {
    // Get languages first
    const languages = await getLanguages()

    // Get matrix data from backend (returns map[string]map[string]string)
    const response = await api.get(`/translations/matrix/by-project/${selectedProjectId.value}`, {
      params: {
        page: currentPage.value,
        page_size: pageSize.value,
        keyword: searchKeyword.value || undefined
      }
    }) as any

    // Extract data and meta from response
    const matrixData = response.data || {}
    const meta = response.meta || {}

    // Transform backend map structure to TranslationMatrix
    const rows: any[] = []

    for (const [keyName, translations] of Object.entries(matrixData)) {
      if (keyName && typeof translations === 'object') {
        const translationCells: Record<string, any> = {}

        // Transform language code -> value to our cell structure
        for (const [langCode, cellData] of Object.entries(translations as Record<string, any>)) {
          const lang = languages.find(l => l.code === langCode)
          if (lang) {
            const value = typeof cellData === 'object' ? cellData.value : cellData
            const id = typeof cellData === 'object' ? cellData.id : undefined

            translationCells[langCode] = {
              language_id: lang.id,
              value: value || '',
              id: id,
              updated_at: typeof cellData === 'object' ? cellData.updated_at : undefined,
            }
          }
        }

        rows.push({
          key_name: keyName,
          context: '',
          translations: translationCells
        })
      }
    }

    matrix.value = {
      languages: languages.filter(l => l.status === 'active'),
      rows: rows,
      total_count: meta.total_count || rows.length,
      page: meta.page || currentPage.value,
      page_size: meta.page_size || pageSize.value,
      total_pages: meta.total_pages || 1
    }

  } catch (err: any) {
    error.value = err.message || '加载翻译数据失败'
    console.error('Failed to load translation matrix:', err)
    matrix.value = null
  } finally {
    loading.value = false
  }
}

// Search handler with debounce
let searchTimeout: number | null = null
const handleSearch = () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = window.setTimeout(() => {
    currentPage.value = 1
    loadMatrix()
  }, 300)
}

// Pagination
const changePage = (page: number) => {
  currentPage.value = page
  loadMatrix()
}

// Edit cell
const editCell = (keyName: string, lang: Language) => {
  const row = matrix.value?.rows.find((r) => r.key_name === keyName)
  if (!row) return

  editingCell.value = { keyName, languageId: lang.id }
  editingValue.value = row.translations[lang.code]?.value || ''

  nextTick(() => {
    editInput.value?.[0]?.focus()
  })
}

// Cancel edit
const cancelEdit = () => {
  editingCell.value = null
  editingValue.value = ''
}

// Format time display
const formatTime = (timeStr: string) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}

const isSaving = ref(false)

// Save cell
const saveCell = async () => {
  if (!editingCell.value || !selectedProjectId.value || isSaving.value) return

  const { keyName, languageId } = editingCell.value
  const row = matrix.value?.rows.find((r) => r.key_name === keyName)
  if (!row) return

  const lang = matrix.value?.languages.find((l) => l.id === languageId)
  if (!lang) return

  const existingCell = row.translations[lang.code]

  try {
    isSaving.value = true
    if (existingCell?.id) {
      await ElMessageBox.confirm('确定要修改这个翻译吗？', '修改确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      })
      // Update existing translation
      await api.put(`/translations/${existingCell.id}`, {
        project_id: selectedProjectId.value,
        language_id: languageId,
        key_name: keyName,
        value: editingValue.value,
        context: row.context,
      })
    } else {
      await ElMessageBox.confirm('确定要添加这个新翻译吗？', '添加确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info',
      })
      // Create new translation
      await api.post('/translations', {
        project_id: selectedProjectId.value,
        language_id: languageId,
        key_name: keyName,
        value: editingValue.value,
        context: row.context,
      })
    }

    ElMessage.success('保存成功')
    // Refresh matrix
    await loadMatrix()
  } catch (err: any) {
    if (err !== 'cancel') {
        ElMessage.error('保存失败: ' + (err.message || '未知错误'))
    } else {
        // User cancelled, revert value
    }
  } finally {
    isSaving.value = false
    cancelEdit()
  }
}


// Add Key Logic with Confirmation
const handleAddKey = async () => {
    if (!selectedProjectId.value || !newKey.value.keyName) return

    try {
        await ElMessageBox.confirm('确定要添加这个新翻译键吗？', '添加确认', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'info',
        })

        const request: BatchTranslationRequest = {
            project_id: selectedProjectId.value,
            key_name: newKey.value.keyName,
            context: newKey.value.context || undefined,
            translations: newKey.value.translations
        }

        await batchCreateTranslations(request)

        ElMessage.success('添加成功')
        // Reset form and refresh
        newKey.value = { keyName: '', context: '', translations: {} }
        showAddKeyDialog.value = false
        await loadMatrix()
    } catch (err: any) {
        if (err !== 'cancel') {
            ElMessage.error('添加失败: ' + (err.message || '未知错误'))
        }
    }
}

// Delete translation key
const handleDeleteKey = async (keyName: string) => {
  if (!selectedProjectId.value) return

  try {
    // Find all translation IDs for this key
    const row = matrix.value?.rows.find((r) => r.key_name === keyName)
    if (!row) return

    const ids: number[] = []
    Object.values(row.translations).forEach((cell) => {
      if (cell.id) ids.push(cell.id)
    })



    if (ids.length > 0) {
      // Use ElMessageBox for delete confirmation
      try {
        await ElMessageBox.confirm(
          `确定要删除翻译键 "${keyName}" 吗？这将删除所有语言的该键翻译。`,
          '删除确认',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning',
          }
        )
        await api.post('/translations/batch-delete', ids)
        ElMessage.success('删除成功')
        await loadMatrix()
      } catch (err: any) {
         if (err !== 'cancel') {
            ElMessage.error('删除失败: ' + (err.message || '未知错误'))
         }
      }
    }
  } catch (err: any) {
    alert('删除失败: ' + (err.message || '未知错误'))
  }
}

// Export translations
const handleExport = async () => {
  if (!selectedProjectId.value) return

  try {
    const data = await exportTranslations(selectedProjectId.value)

    // Download as JSON file
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `translations-project-${selectedProjectId.value}.json`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (err: any) {
    ElMessage.error('导出失败: ' + (err.message || '未知错误'))
  }
}

// File select for import
const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    const file = target.files[0]
    if (file) {
      importFile.value = file
    }
  }
}

// Import translations
const handleImport = async () => {
  if (!selectedProjectId.value || !importFile.value) return

  try {
    const text = await importFile.value.text()
    const data: ImportTranslationsData = JSON.parse(text)

    await importTranslations(selectedProjectId.value, data)

    // Reset and refresh
    importFile.value = null
    showImportDialog.value = false
    await loadMatrix()
    ElMessage.success('导入成功')
  } catch (err: any) {
    ElMessage.error('导入失败: ' + (err.message || '未知错误'))
  }
}
</script>

<style scoped>
.translations-page {
  max-width: 1400px;
  margin: 0 auto;
}

/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  gap: 20px;
}

.header-content {
  flex: 1;
}

.page-title {
  margin: 0;
  font-size: 32px;
  font-weight: 700;
  background: linear-gradient(135deg, #06b6d4 0%, #14b8a6 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  letter-spacing: -0.5px;
}

.page-subtitle {
  margin: 8px 0 0;
  font-size: 15px;
  color: #64748b;
  font-weight: 500;
}

.project-selector {
  width: 240px;
}

:deep(.project-selector .el-input__wrapper) {
  border-radius: 12px;
  box-shadow: none;
  border: 1px solid #e2e8f0;
  transition: all 0.2s ease;
}

:deep(.project-selector .el-input__wrapper:hover) {
  border-color: #cbd5e1;
}

:deep(.project-selector .el-input__wrapper.is-focus) {
  border-color: #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

/* 工具栏卡片 */
.toolbar-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  border: 1px solid #f1f5f9;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.search-input-group {
  display: flex;
  gap: 12px;
  flex: 1;
  max-width: 600px;
  min-width: 280px;
}

.search-input {
  flex: 1;
}

:deep(.search-input .el-input__wrapper) {
  border-radius: 10px;
  padding: 8px 16px;
  box-shadow: none;
  border: 1px solid #e2e8f0;
  transition: all 0.2s ease;
}

:deep(.search-input .el-input__wrapper:hover) {
  border-color: #cbd5e1;
}

:deep(.search-input .el-input__wrapper.is-focus) {
  border-color: #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

.toolbar-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

/* 表格卡片 */
.table-card {
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  border: 1px solid #f1f5f9;
  overflow: hidden;
}

/* 表格样式 */
:deep(.translation-table) {
  border: none;
}

:deep(.translation-table .el-table__header-wrapper) {
  background: #f8fafc;
}

:deep(.translation-table th.el-table__cell) {
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  color: #475569;
  font-weight: 600;
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 16px 12px;
}

:deep(.translation-table td.el-table__cell) {
  border-bottom: 1px solid #f1f5f9;
  padding: 16px 12px;
}

:deep(.translation-table tr:hover > td) {
  background: #f8fafc;
}

:deep(.translation-table .el-table__empty-block) {
  padding: 40px 0;
}

/* 翻译键样式 */
.key-text {
  font-weight: 600;
  color: #0f172a;
  font-family: 'Fira Code', monospace;
}

.context-text {
  color: #64748b;
  font-style: italic;
}

.lang-tag {
  margin-left: 8px;
  background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
  color: #059669;
  border: none;
  font-weight: 600;
}

.column-header {
  display: flex;
  align-items: center;
}

/* 单元格样式 */
.cell-wrapper {
  padding: 8px 4px;
  min-height: 48px;
  cursor: pointer;
  border-radius: 8px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
}

.cell-wrapper:hover {
  background-color: #f8fafc;
}

.cell-wrapper.is-empty {
  background: linear-gradient(135deg, #fff1f2 0%, #ffe4e6 100%);
}

.cell-wrapper.is-empty:hover {
  background: linear-gradient(135deg, #ffe4e6 0%, #fecdd3 100%);
}

.cell-editing {
  width: 100%;
}

.cell-display {
  width: 100%;
}

.cell-value {
  color: #0f172a;
  white-space: pre-wrap;
  word-break: break-word;
}

.cell-placeholder {
  color: #94a3b8;
  font-style: italic;
  font-size: 13px;
}

.cell-time {
  color: #94a3b8;
  font-size: 11px;
  margin-top: 4px;
}

/* 操作按钮 */
.action-button {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.action-delete {
  background: linear-gradient(135deg, #fff1f2 0%, #ffe4e6 100%);
  color: #e11d48;
}

.action-delete:hover {
  background: linear-gradient(135deg, #ffe4e6 0%, #fecdd3 100%);
  transform: translateY(-1px);
}

/* 分页 */
.pagination-container {
  display: flex;
  justify-content: flex-end;
  padding: 20px 24px;
  border-top: 1px solid #f1f5f9;
}

:deep(.el-pagination) {
  font-weight: 500;
}

:deep(.el-pagination .el-pager li) {
  border-radius: 8px;
  margin: 0 2px;
}

:deep(.el-pagination .el-pager li.is-active) {
  background: linear-gradient(135deg, #06b6d4 0%, #14b8a6 100%);
  color: #ffffff;
}

:deep(.el-pagination button) {
  border-radius: 8px;
}

/* 空状态 */
.empty-card {
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  border: 1px solid #f1f5f9;
}

.empty-state {
  padding: 80px 40px;
  text-align: center;
}

.empty-icon {
  color: #cbd5e1;
  margin-bottom: 20px;
}

.empty-title {
  margin: 0 0 8px;
  font-size: 20px;
  font-weight: 700;
  color: #0f172a;
}

.empty-description {
  margin: 0;
  font-size: 14px;
  color: #64748b;
}

/* 对话框样式 */
:deep(.translation-dialog .el-dialog) {
  border-radius: 16px;
}

:deep(.translation-dialog .el-dialog__header) {
  padding: 24px 24px 16px;
  border-bottom: 1px solid #f1f5f9;
}

:deep(.translation-dialog .el-dialog__title) {
  font-size: 18px;
  font-weight: 700;
  color: #0f172a;
}

:deep(.translation-dialog .el-dialog__body) {
  padding: 24px;
}

:deep(.translation-dialog .el-dialog__footer) {
  padding: 16px 24px 24px;
  border-top: 1px solid #f1f5f9;
}

.translation-form :deep(.el-form-item__label) {
  font-weight: 600;
  color: #475569;
}

:deep(.translation-form .el-input__wrapper) {
  border-radius: 10px;
  box-shadow: none;
  border: 1px solid #e2e8f0;
  transition: all 0.2s ease;
}

:deep(.translation-form .el-input__wrapper:hover) {
  border-color: #cbd5e1;
}

:deep(.translation-form .el-input__wrapper.is-focus) {
  border-color: #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

/* 语言输入滚动区域 */
.language-inputs-scroll {
  max-height: 400px;
  overflow-y: auto;
  padding-right: 10px;
}

.language-inputs-scroll::-webkit-scrollbar {
  width: 6px;
}

.language-inputs-scroll::-webkit-scrollbar-thumb {
  background: linear-gradient(180deg, #22d3ee 0%, #14b8a6 100%);
  border-radius: 3px;
}

.language-inputs-scroll::-webkit-scrollbar-track {
  background: #f1f5f9;
}

/* 导入对话框 */
.import-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.import-tip {
  margin-bottom: 15px;
  font-size: 14px;
  color: #64748b;
}

.upload-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
  padding: 20px;
  width: 100%;
  border: 2px dashed #e2e8f0;
  border-radius: 12px;
  transition: all 0.3s;
}

.upload-area:hover {
  border-color: #06b6d4;
  background: #f0fdfa;
}

.upload-icon {
  color: #94a3b8;
  margin-bottom: 10px;
}

.upload-text {
  color: #64748b;
  font-size: 14px;
}

.file-name {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #06b6d4;
  font-weight: 500;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
    gap: 16px;
  }

  .page-title {
    font-size: 24px;
  }

  .project-selector {
    width: 100%;
  }

  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .search-input-group {
    max-width: 100%;
  }

  .toolbar-actions {
    justify-content: stretch;
  }

  .toolbar-actions .el-button {
    flex: 1;
  }

  .pagination-container {
    justify-content: center;
  }

  :deep(.el-pagination) {
    flex-wrap: wrap;
    justify-content: center;
  }
}
</style>
