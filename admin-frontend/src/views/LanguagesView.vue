<template>
  <div class="languages-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">语言管理</h1>
        <p class="page-subtitle">管理所有支持的语言</p>
      </div>
      <button class="create-button" @click="openAddDialog">
        <el-icon><Plus /></el-icon>
        <span>添加语言</span>
      </button>
    </div>

    <!-- 语言列表 -->
    <div class="table-card">
      <!-- 加载中状态 -->
      <div v-if="loading" class="loading-container">
        <el-skeleton :rows="5" animated />
      </div>

      <!-- 数据表格 -->
      <div v-else class="table-container">
        <el-table
          :data="languages"
          style="width: 100%"
          class="languages-table"
          :empty-text="'暂无语言数据'"
        >
          <el-table-column prop="id" label="ID" width="80" align="center" />
          <el-table-column prop="name" label="语言名称" min-width="150">
            <template #default="{ row }">
              <div class="language-name">
                <strong>{{ row.name }}</strong>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="code" label="语言代码" width="140" align="center">
            <template #default="{ row }">
              <span class="code-badge">{{ row.code }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100" align="center">
            <template #default="{ row }">
              <span
                :class="['status-badge', row.status === 'active' ? 'status-active' : 'status-inactive']"
              >
                {{ row.status === 'active' ? '启用' : '禁用' }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="is_default" label="默认语言" width="120" align="center">
            <template #default="{ row }">
              <div v-if="row.is_default" class="default-badge">
                <el-icon :size="16"><Check /></el-icon>
                <span>默认</span>
              </div>
              <span v-else class="non-default">-</span>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="170">
            <template #default="{ row }">
              <span class="date-text">{{ formatDate(row.created_at) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right" align="center">
            <template #default="{ row }">
              <div class="action-buttons">
                <button class="action-button action-edit" @click="openEditDialog(row)">
                  <el-icon><Edit /></el-icon>
                  <span>编辑</span>
                </button>
                <button class="action-button action-delete" @click="handleDelete(row)">
                  <el-icon><Delete /></el-icon>
                  <span>删除</span>
                </button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <!-- 添加/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑语言' : '添加语言'"
      width="500px"
      :close-on-click-modal="false"
      class="language-dialog"
      @closed="resetForm"
    >
      <el-form ref="formRef" :model="formData" :rules="rules" label-width="100px" class="language-form">
        <el-form-item label="语言名称" prop="name">
          <el-input v-model="formData.name" placeholder="例如: 简体中文" />
        </el-form-item>
        <el-form-item label="语言代码" prop="code">
          <el-input v-model="formData.code" placeholder="例如: zh-CN" />
        </el-form-item>
        <el-form-item label="默认语言" prop="is_default">
          <el-switch v-model="formData.is_default" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          {{ isEdit ? '更新' : '添加' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check, Plus, Edit, Delete } from '@element-plus/icons-vue'
import { getLanguages, createLanguage, updateLanguage, deleteLanguage } from '@/services/language'
import type { Language, CreateLanguageRequest } from '@/types/translation'
import type { FormInstance, FormRules } from 'element-plus'

const loading = ref(false)
const languages = ref<Language[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()

const formData = reactive<CreateLanguageRequest>({
  name: '',
  code: '',
  is_default: false,
})

const rules = reactive<FormRules>({
  name: [{ required: true, message: '请输入语言名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入语言代码', trigger: 'blur' }],
})

const editingId = ref<number | null>(null)

onMounted(() => {
  fetchLanguages()
})

const fetchLanguages = async () => {
  loading.value = true
  try {
    languages.value = await getLanguages()
  } catch {
    ElMessage.error('获取语言列表失败')
  } finally {
    loading.value = false
  }
}

const formatDate = (dateString: string) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const resetForm = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  formData.name = ''
  formData.code = ''
  formData.is_default = false
  editingId.value = null
  isEdit.value = false
}

const openAddDialog = () => {
  isEdit.value = false
  dialogVisible.value = true
}

const openEditDialog = (row: Language) => {
  isEdit.value = true
  editingId.value = row.id
  formData.name = row.name
  formData.code = row.code
  formData.is_default = row.is_default
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        if (isEdit.value && editingId.value) {
          await updateLanguage(editingId.value, formData)
          ElMessage.success('更新成功')
        } else {
          await createLanguage(formData)
          ElMessage.success('添加成功')
        }
        dialogVisible.value = false
        fetchLanguages()
      } catch (error: unknown) {
        ElMessage.error(error instanceof Error ? error.message : '操作失败')
      } finally {
        submitting.value = false
      }
    }
  })
}

const handleDelete = (row: Language) => {
  ElMessageBox.confirm(`确定要删除语言 "${row.name}" 吗？`, '删除确认', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      await deleteLanguage(row.id)
      ElMessage.success('删除成功')
      fetchLanguages()
    } catch (error: unknown) {
      ElMessage.error(error instanceof Error ? error.message : '删除失败')
    }
  })
}
</script>

<style scoped>
.languages-page {
  max-width: 1400px;
  margin: 0 auto;
}

/* 页面头部 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
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

.create-button {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background: linear-gradient(135deg, #06b6d4 0%, #14b8a6 100%);
  color: #ffffff;
  border: none;
  border-radius: 12px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 4px 12px rgba(6, 182, 212, 0.3);
}

.create-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(6, 182, 212, 0.4);
}

.create-button:active {
  transform: translateY(0);
}

/* 表格卡片 */
.table-card {
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  border: 1px solid #f1f5f9;
  overflow: hidden;
}

.loading-container {
  padding: 40px;
}

.table-container {
  padding: 4px;
}

/* 表格样式 */
:deep(.languages-table) {
  border: none;
}

:deep(.languages-table .el-table__header-wrapper) {
  background: #f8fafc;
}

:deep(.languages-table th.el-table__cell) {
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  color: #475569;
  font-weight: 600;
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 16px 12px;
}

:deep(.languages-table td.el-table__cell) {
  border-bottom: 1px solid #f1f5f9;
  padding: 16px 12px;
}

:deep(.languages-table tr:hover > td) {
  background: #f8fafc;
}

:deep(.languages-table .el-table__empty-block) {
  padding: 40px 0;
}

/* 语言名称 */
.language-name strong {
  font-size: 15px;
  font-weight: 600;
  color: #0f172a;
}

/* 代码徽章 */
.code-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  color: #2563eb;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  font-family: 'Fira Code', monospace;
}

/* 状态标签 */
.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}

.status-active {
  background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
  color: #059669;
}

.status-inactive {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  color: #64748b;
}

/* 默认语言徽章 */
.default-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
  color: #d97706;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.non-default {
  color: #94a3b8;
}

.date-text {
  font-size: 14px;
  color: #64748b;
  font-family: 'Fira Code', monospace;
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

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

.action-edit {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  color: #2563eb;
}

.action-edit:hover {
  background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
  transform: translateY(-1px);
}

.action-delete {
  background: linear-gradient(135deg, #fff1f2 0%, #ffe4e6 100%);
  color: #e11d48;
}

.action-delete:hover {
  background: linear-gradient(135deg, #ffe4e6 0%, #fecdd3 100%);
  transform: translateY(-1px);
}

/* 对话框样式 */
:deep(.language-dialog .el-dialog) {
  border-radius: 16px;
}

:deep(.language-dialog .el-dialog__header) {
  padding: 24px 24px 16px;
  border-bottom: 1px solid #f1f5f9;
}

:deep(.language-dialog .el-dialog__title) {
  font-size: 18px;
  font-weight: 700;
  color: #0f172a;
}

:deep(.language-dialog .el-dialog__body) {
  padding: 24px;
}

:deep(.language-dialog .el-dialog__footer) {
  padding: 16px 24px 24px;
  border-top: 1px solid #f1f5f9;
}

/* 表单样式 */
:deep(.language-form .el-form-item__label) {
  font-weight: 600;
  color: #475569;
}

:deep(.language-form .el-input__wrapper) {
  border-radius: 10px;
  box-shadow: none;
  border: 1px solid #e2e8f0;
  transition: all 0.2s ease;
}

:deep(.language-form .el-input__wrapper:hover) {
  border-color: #cbd5e1;
}

:deep(.language-form .el-input__wrapper.is-focus) {
  border-color: #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
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

  .create-button {
    width: 100%;
    justify-content: center;
  }

  :deep(.languages-table) {
    font-size: 13px;
  }

  .action-button span {
    display: none;
  }
}
</style>
