<template>
  <div class="translations-view">
    <el-card class="box-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <h1 class="page-title">翻译管理</h1>
          <div class="header-actions">
            <!-- Project Selector -->
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
        </div>
      </template>

      <div v-if="selectedProjectId" class="content-wrapper">
        <!-- Toolbar -->
        <div class="toolbar">
          <div class="search-box">
            <el-input
              v-model="searchKeyword"
              @input="handleSearch"
              placeholder="搜索翻译键..."
              clearable
              prefix-icon="Search"
              class="search-input"
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
          </div>
        </div>

        <!-- Translation Matrix Table -->
        <el-table
          v-loading="loading"
          :data="matrix?.rows || []"
          style="width: 100%"
          border
          stripe
          height="calc(100vh - 350px)"
          class="translation-table"
        >
          <!-- Key Column -->
          <el-table-column prop="key_name" label="翻译键" fixed width="220" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="key-text">{{ row.key_name }}</span>
            </template>
          </el-table-column>

          <!-- Context Column -->
          <el-table-column prop="context" label="上下文" width="180" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="context-text">{{ row.context || '-' }}</span>
            </template>
          </el-table-column>

          <!-- Dynamic Language Columns -->
          <el-table-column
            v-for="lang in matrix?.languages || []"
            :key="lang.id"
            :label="lang.name"
            min-width="250"
          >
            <template #header>
              <div class="column-header">
                {{ lang.name }} <el-tag size="small" type="info" class="lang-tag">{{ lang.code }}</el-tag>
              </div>
            </template>
            <template #default="{ row }">
              <div
                class="cell-wrapper"
                @click="editCell(row.key_name, lang)"
                :class="{ 'is-empty': !row.translations[lang.code]?.value }"
              >
                <!-- Editing Mode -->
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
                <!-- Display Mode -->
                <div v-else class="cell-display">
                  <span v-if="row.translations[lang.code]?.value" class="cell-value">
                    {{ row.translations[lang.code].value }}
                  </span>
                  <span v-else class="cell-placeholder">点击添加翻译</span>
                </div>
              </div>
            </template>
          </el-table-column>

          <!-- Actions Column -->
          <el-table-column label="操作" width="80" fixed="right" align="center">
            <template #default="{ row }">
              <el-button
                type="danger"
                circle
                size="small"
                icon="Delete"
                @click="handleDeleteKey(row.key_name)"
                title="删除翻译键"
              />
            </template>
          </el-table-column>
        </el-table>

        <!-- Pagination -->
        <div class="pagination-wrapper" v-if="matrix && matrix.total_count > 0">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            :total="matrix.total_count"
            @size-change="loadMatrix"
            @current-change="loadMatrix"
            background
          />
        </div>
      </div>

      <el-empty v-else description="请选择一个项目以管理翻译" class="empty-state">
        <el-icon :size="60" class="empty-icon"><FolderOpened /></el-icon>
      </el-empty>
    </el-card>

    <!-- Add Key Dialog -->
    <el-dialog
      v-model="showAddKeyDialog"
      title="添加翻译键"
      width="600px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form :model="newKey" label-width="100px" @submit.prevent="handleAddKey">
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

    <!-- Import Dialog -->
    <el-dialog
      v-model="showImportDialog"
      title="导入翻译"
      width="500px"
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
import { Search, Plus, Download, Upload, UploadFilled, FolderOpened, Document, Delete } from '@element-plus/icons-vue'

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

    console.log('Translation matrix API response:', response)

    // Extract data and meta from response
    const matrixData = response.data || {}
    const meta = response.meta || {}

    // Transform backend map structure to TranslationMatrix
    // Backend returns: { "key1": { "en": "value1", "zh": "value2" }, ...}
    const rows: any[] = []

    for (const [keyName, translations] of Object.entries(matrixData)) {
      if (keyName && typeof translations === 'object') {
        const translationCells: Record<string, any> = {}

        // Transform language code -> value to our cell structure
        for (const [langCode, cellData] of Object.entries(translations as Record<string, any>)) {
          const lang = languages.find(l => l.code === langCode)
          if (lang) {
            // Backend now returns object {id: number, value: string}
            // Add backwards compatibility check just in case
            const value = typeof cellData === 'object' ? cellData.value : cellData
            const id = typeof cellData === 'object' ? cellData.id : undefined

            translationCells[langCode] = {
              language_id: lang.id,
              value: value || '',
              id: id,
            }
          }
        }

        rows.push({
          key_name: keyName,
          context: '', // Backend doesn't return context in matrix view
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

    console.log('Transformed matrix:', matrix.value)

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
        // User cancelled, revert value (reload matrix is overkill but safest)
        // Or just re-display original value
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
  // Confirmation moved inside try-catch block for ElMessageBox
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
.translations-view {
  background-color: #f3f4f6;
}

.box-card {
  border-radius: 8px;
  border: none;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
}

.project-selector {
  width: 240px;
}

.content-wrapper {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.search-box {
  flex: 1;
  max-width: 400px;
}

.toolbar-actions {
  display: flex;
  gap: 1rem;
}

.translation-table {
  width: 100%;
}

.key-text {
  font-weight: 600;
  color: #1f2937;
}

.context-text {
  color: #6b7280;
  font-style: italic;
}

.lang-tag {
  margin-left: 8px;
  font-weight: normal;
}

.column-header {
  display: flex;
  align-items: center;
}

.cell-wrapper {
  padding: 8px 4px;
  min-height: 48px;
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
}

.cell-wrapper:hover {
  background-color: #f9fafb;
}

.cell-wrapper.is-empty {
  background-color: #fff1f2;
}

.cell-wrapper.is-empty:hover {
    background-color: #ffe4e6;
}

.cell-editing {
  width: 100%;
}

.cell-display {
  width: 100%;
}

.cell-value {
    color: #374151;
    white-space: pre-wrap;
    word-break: break-word;
}

.cell-placeholder {
  color: #9ca3af;
  font-style: italic;
  font-size: 0.875rem;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 1.5rem;
}

.empty-state {
    padding: 60px 0;
}

/* Dialog Styles */
.language-inputs-scroll {
  max-height: 400px;
  overflow-y: auto;
  padding-right: 10px;
}

.import-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px;
  border: 2px dashed #e5e7eb;
  border-radius: 8px;
  transition: border-color 0.3s;
}

.import-wrapper:hover {
    border-color: #409eff;
}

.upload-area {
    display: flex;
    flex-direction: column;
    align-items: center;
    cursor: pointer;
    padding: 20px;
    width: 100%;
}

.upload-icon {
    color: #909399;
    margin-bottom: 10px;
}

.upload-text {
    color: #606266;
    font-size: 14px;
}

.file-name {
    display: flex;
    align-items: center;
    gap: 8px;
    color: #409eff;
    font-weight: 500;
}

.import-tip {
    margin-bottom: 15px;
    font-size: 14px;
    color: #606266;
}

/* Custom Scrollbar for language inputs */
.language-inputs-scroll::-webkit-scrollbar {
    width: 6px;
}

.language-inputs-scroll::-webkit-scrollbar-thumb {
    background-color: #d1d5db;
    border-radius: 3px;
}

.language-inputs-scroll::-webkit-scrollbar-track {
    background-color: #f3f4f6;
}
</style>
