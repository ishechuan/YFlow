<script setup lang="ts">
import { ref, computed } from 'vue'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getProjects,
  createProject,
  updateProject,
  deleteProject,
} from '@/services/projectService'
import type {
  Project,
  CreateProjectRequest,
  UpdateProjectRequest,
} from '@/types/api'
import {
  Plus,
  Edit,
  Delete,
  Search,
  Refresh,
} from '@element-plus/icons-vue'

// ============ 状态管理 ============
const queryClient = useQueryClient()

// 搜索和分页状态
const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(10)

// 对话框状态
const dialogVisible = ref(false)
const dialogTitle = ref('创建项目')
const isEditMode = ref(false)
const currentEditId = ref<number | null>(null)

// 表单引用和数据
const formRef = ref<FormInstance>()
const formData = ref<CreateProjectRequest | UpdateProjectRequest>({
  name: '',
  description: '',
})

// 表单验证规则
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入项目名称', trigger: 'blur' },
    { min: 2, max: 100, message: '项目名称长度在 2 到 100 个字符', trigger: 'blur' },
  ],
}

// ============ 数据获取 ============
const queryParams = computed(() => ({
  page: currentPage.value,
  page_size: pageSize.value,
  keyword: searchKeyword.value || undefined,
}))

const {
  data: projectsData,
  isLoading,
  isError,
  error,
  refetch,
} = useQuery({
  queryKey: ['projects', queryParams],
  queryFn: () => getProjects(queryParams.value),
})

const projects = computed(() => projectsData.value?.data || [])
const totalCount = computed(() => projectsData.value?.meta?.total_count || 0)

// ============ CRUD 操作 ============

// 创建项目
const createMutation = useMutation({
  mutationFn: createProject,
  onSuccess: () => {
    ElMessage.success('项目创建成功')
    dialogVisible.value = false
    queryClient.invalidateQueries({ queryKey: ['projects'] })
  },
  onError: (err: Error) => {
    ElMessage.error(err.message || '创建项目失败')
  },
})

// 更新项目
const updateMutation = useMutation({
  mutationFn: ({ id, data }: { id: number; data: UpdateProjectRequest }) =>
    updateProject(id, data),
  onSuccess: () => {
    ElMessage.success('项目更新成功')
    dialogVisible.value = false
    queryClient.invalidateQueries({ queryKey: ['projects'] })
  },
  onError: (err: Error) => {
    ElMessage.error(err.message || '更新项目失败')
  },
})

// 删除项目
const deleteMutation = useMutation({
  mutationFn: deleteProject,
  onSuccess: () => {
    ElMessage.success('项目删除成功')
    queryClient.invalidateQueries({ queryKey: ['projects'] })
  },
  onError: (err: Error) => {
    ElMessage.error(err.message || '删除项目失败')
  },
})

// ============ 事件处理 ============

// 搜索
const handleSearch = () => {
  currentPage.value = 1 // 搜索时重置到第一页
}

// 清空搜索
const handleClearSearch = () => {
  searchKeyword.value = ''
  handleSearch()
}

// 刷新列表
const handleRefresh = () => {
  refetch()
}

// 分页变化
const handlePageChange = (page: number) => {
  currentPage.value = page
}

const handlePageSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
}

// 打开创建对话框
const handleCreate = () => {
  isEditMode.value = false
  dialogTitle.value = '创建项目'
  formData.value = {
    name: '',
    description: '',
  }
  dialogVisible.value = true
}

// 打开编辑对话框
const handleEdit = (project: Project) => {
  isEditMode.value = true
  dialogTitle.value = '编辑项目'
  currentEditId.value = project.id
  formData.value = {
    name: project.name,
    description: project.description || '',
    status: project.status,
  }
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate((valid) => {
    if (valid) {
      if (isEditMode.value && currentEditId.value) {
        updateMutation.mutate({
          id: currentEditId.value,
          data: formData.value as UpdateProjectRequest,
        })
      } else {
        createMutation.mutate(formData.value as CreateProjectRequest)
      }
    }
  })
}

// 取消表单
const handleCancel = () => {
  dialogVisible.value = false
  formRef.value?.resetFields()
}

// 删除项目
const handleDelete = (project: Project) => {
  ElMessageBox.confirm(
    `确定要删除项目 "${project.name}" 吗？此操作不可恢复。`,
    '删除确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  )
    .then(() => {
      deleteMutation.mutate(project.id)
    })
    .catch(() => {
      // 用户取消删除
    })
}

// 切换项目状态
const handleToggleStatus = (project: Project) => {
  const newStatus = project.status === 'active' ? 'archived' : 'active'
  updateMutation.mutate({
    id: project.id,
    data: { status: newStatus },
  })
}

// 格式化日期
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <div class="projects-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">项目管理</h1>
        <p class="page-subtitle">管理所有翻译项目</p>
      </div>
      <button class="create-button" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        <span>创建项目</span>
      </button>
    </div>

    <!-- 搜索和操作栏 -->
    <div class="search-card">
      <div class="search-bar">
        <div class="search-input-group">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索项目名称或描述"
            :prefix-icon="Search"
            clearable
            class="search-input"
            @clear="handleClearSearch"
            @keyup.enter="handleSearch"
          />
          <button class="search-button" @click="handleSearch">
            <el-icon><Search /></el-icon>
            <span>搜索</span>
          </button>
        </div>
        <button class="refresh-button" @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
        </button>
      </div>
    </div>

    <!-- 项目列表 -->
    <div class="table-card">
      <!-- 加载中状态 -->
      <div v-if="isLoading" class="loading-container">
        <el-skeleton :rows="5" animated />
      </div>

      <!-- 错误状态 -->
      <div v-else-if="isError" class="error-container">
        <svg class="error-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <h3 class="error-title">加载项目列表失败</h3>
        <p class="error-message">{{ error?.message || '请检查网络连接或稍后重试' }}</p>
      </div>

      <!-- 数据表格 -->
      <div v-else class="table-container">
        <el-table
          :data="projects"
          style="width: 100%"
          :empty-text="searchKeyword ? '没有找到匹配的项目' : '暂无项目数据'"
          class="projects-table"
        >
          <el-table-column prop="id" label="ID" width="80" align="center" />
          <el-table-column prop="name" label="项目名称" min-width="180">
            <template #default="{ row }">
              <div class="project-name-cell">
                <div class="project-name-wrapper">
                  <strong class="project-name">{{ row.name }}</strong>
                  <span class="project-slug">{{ row.slug }}</span>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column
            prop="description"
            label="描述"
            min-width="220"
            show-overflow-tooltip
          >
            <template #default="{ row }">
              <span class="description-text">{{ row.description || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100" align="center">
            <template #default="{ row }">
              <span
                :class="['status-badge', row.status === 'active' ? 'status-active' : 'status-archived']"
                @click="handleToggleStatus(row)"
              >
                {{ row.status === 'active' ? '活跃' : '已归档' }}
              </span>
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
                <button class="action-button action-edit" @click="handleEdit(row)">
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

        <!-- 分页 -->
        <div class="pagination-container">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="totalCount"
            layout="total, sizes, prev, pager, next, jumper"
            @current-change="handlePageChange"
            @size-change="handlePageSizeChange"
          />
        </div>
      </div>
    </div>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
      :close-on-click-modal="false"
      class="project-dialog"
      @close="handleCancel"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="80px"
        class="project-form"
      >
        <el-form-item label="项目名称" prop="name">
          <el-input
            v-model="formData.name"
            placeholder="请输入项目名称"
            clearable
          />
        </el-form-item>
        <el-form-item label="项目描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="4"
            placeholder="请输入项目描述（可选）"
            clearable
          />
        </el-form-item>
        <el-form-item v-if="isEditMode" label="项目状态" prop="status">
          <el-radio-group v-model="(formData as UpdateProjectRequest).status">
            <el-radio value="active">活跃</el-radio>
            <el-radio value="archived">已归档</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="handleCancel">取消</el-button>
        <el-button
          type="primary"
          :loading="createMutation.isPending.value || updateMutation.isPending.value"
          @click="handleSubmit"
        >
          {{ isEditMode ? '更新' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.projects-page {
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

/* 搜索卡片 */
.search-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  border: 1px solid #f1f5f9;
}

.search-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.search-input-group {
  display: flex;
  gap: 12px;
  flex: 1;
  max-width: 600px;
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

.search-button {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 20px;
  background: linear-gradient(135deg, #06b6d4 0%, #14b8a6 100%);
  color: #ffffff;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.search-button:hover {
  opacity: 0.9;
}

.refresh-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s ease;
}

.refresh-button:hover {
  background: #f1f5f9;
  color: #475569;
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

/* 错误状态 */
.error-container {
  padding: 60px 40px;
  text-align: center;
}

.error-icon {
  width: 64px;
  height: 64px;
  color: #f43f5e;
  margin: 0 auto 20px;
}

.error-title {
  margin: 0 0 8px;
  font-size: 18px;
  font-weight: 700;
  color: #0f172a;
}

.error-message {
  margin: 0;
  font-size: 14px;
  color: #64748b;
}

/* 表格容器 */
.table-container {
  padding: 4px;
}

/* 表格样式 */
:deep(.projects-table) {
  border: none;
}

:deep(.projects-table .el-table__header-wrapper) {
  background: #f8fafc;
}

:deep(.projects-table th.el-table__cell) {
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  color: #475569;
  font-weight: 600;
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 16px 12px;
}

:deep(.projects-table td.el-table__cell) {
  border-bottom: 1px solid #f1f5f9;
  padding: 16px 12px;
}

:deep(.projects-table tr:hover > td) {
  background: #f8fafc;
}

:deep(.projects-table .el-table__empty-block) {
  padding: 40px 0;
}

/* 项目名称单元格 */
.project-name-cell {
  display: flex;
  align-items: center;
}

.project-name-wrapper {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.project-name {
  font-size: 15px;
  font-weight: 600;
  color: #0f172a;
}

.project-slug {
  font-size: 12px;
  color: #94a3b8;
  font-family: 'Fira Code', monospace;
}

.description-text {
  font-size: 14px;
  color: #64748b;
}

/* 状态标签 */
.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.status-active {
  background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
  color: #059669;
}

.status-active:hover {
  background: linear-gradient(135deg, #d1fae5 0%, #a7f3d0 100%);
}

.status-archived {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  color: #64748b;
}

.status-archived:hover {
  background: linear-gradient(135deg, #f1f5f9 0%, #e2e8f0 100%);
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

/* 对话框样式 */
:deep(.project-dialog .el-dialog) {
  border-radius: 16px;
}

:deep(.project-dialog .el-dialog__header) {
  padding: 24px 24px 16px;
  border-bottom: 1px solid #f1f5f9;
}

:deep(.project-dialog .el-dialog__title) {
  font-size: 18px;
  font-weight: 700;
  color: #0f172a;
}

:deep(.project-dialog .el-dialog__body) {
  padding: 24px;
}

:deep(.project-dialog .el-dialog__footer) {
  padding: 16px 24px 24px;
  border-top: 1px solid #f1f5f9;
}

/* 表单样式 */
:deep(.project-form .el-form-item__label) {
  font-weight: 600;
  color: #475569;
}

:deep(.project-form .el-input__wrapper) {
  border-radius: 10px;
  box-shadow: none;
  border: 1px solid #e2e8f0;
  transition: all 0.2s ease;
}

:deep(.project-form .el-input__wrapper:hover) {
  border-color: #cbd5e1;
}

:deep(.project-form .el-input__wrapper.is-focus) {
  border-color: #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

:deep(.project-form .el-textarea__inner) {
  border-radius: 10px;
  box-shadow: none;
  border: 1px solid #e2e8f0;
  transition: all 0.2s ease;
}

:deep(.project-form .el-textarea__inner:hover) {
  border-color: #cbd5e1;
}

:deep(.project-form .el-textarea__inner:focus) {
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

  .search-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .search-input-group {
    max-width: 100%;
  }

  .search-button {
    flex: 1;
    justify-content: center;
  }

  .pagination-container {
    justify-content: center;
  }

  :deep(.el-pagination) {
    flex-wrap: wrap;
    justify-content: center;
  }

  :deep(.projects-table) {
    font-size: 13px;
  }

  .action-button span {
    display: none;
  }
}
</style>
