<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import type { FormInstance, FormRules } from 'element-plus'
import {
  createInvitation,
  getInvitations,
  revokeInvitation,
} from '@/services/invitation'
import type { CreateInvitationParams, Invitation } from '@/types/api'
import {
  Plus,
  Delete,
  Search,
  Refresh,
  Link,
} from '@element-plus/icons-vue'

// ============ 状态管理 ============
const queryClient = useQueryClient()

// 搜索和分页状态
const searchKeyword = ref('')
const currentPage = ref(1)
const pageSize = ref(10)

// 对话框状态
const dialogVisible = ref(false)
const formRef = ref<FormInstance>()
const formData = ref({
  role: 'member' as 'admin' | 'member' | 'viewer',
  expiresInDays: 7,
  description: '',
})

// 表单验证规则
const formRules: FormRules = {
  role: [
    { required: true, message: '请选择角色', trigger: 'change' },
  ],
  expiresInDays: [
    { required: true, message: '请输入有效期', trigger: 'blur' },
    { type: 'number', min: 1, max: 365, message: '有效期必须在 1-365 天之间', trigger: 'blur' },
  ],
}

// 角色选项
const roleOptions = [
  { label: '管理员 (admin)', value: 'admin' },
  { label: '成员 (member)', value: 'member' },
  { label: '查看者 (viewer)', value: 'viewer' },
]

// ============ 数据获取 ============
const queryParams = computed(() => ({
  page: currentPage.value,
  page_size: pageSize.value,
  keyword: searchKeyword.value || undefined,
}))

const {
  data: listData,
  isLoading,
  isError,
  error,
  refetch,
} = useQuery({
  queryKey: ['invitations', queryParams],
  queryFn: () => getInvitations(currentPage.value, pageSize.value),
})

const invitations = computed(() => listData.value?.data?.invitations || [])
const totalCount = computed(() => listData.value?.data?.total || 0)

// ============ CRUD 操作 ============

// 创建邀请 Mutation
const createMutation = useMutation({
  mutationFn: (params: CreateInvitationParams) => createInvitation(params),
  onSuccess: (data) => {
    ElMessage.success('邀请码创建成功')
    dialogVisible.value = false
    resetForm()
    queryClient.invalidateQueries({ queryKey: ['invitations'] })
    // 显示创建的邀请信息
    ElMessageBox.alert(
      `
        <div style="margin-bottom: 16px;">
          <strong>邀请码：</strong>${data.code}
        </div>
        <div style="margin-bottom: 16px;">
          <strong>邀请链接：</strong>
          <a href="${data.invitation_url}" target="_blank" style="color: #06b6d4;">${data.invitation_url}</a>
        </div>
        <div style="margin-bottom: 16px;">
          <strong>角色：</strong>${getRoleDisplayName(data.role)}
        </div>
        <div>
          <strong>过期时间：</strong>${formatDateTime(data.expires_at)}
        </div>
      `,
      '邀请码已创建',
      {
        confirmButtonText: '复制链接',
        cancelButtonText: '关闭',
        dangerouslyUseHTMLString: true,
        type: 'success',
      }
    ).then(async () => {
      try {
        await navigator.clipboard.writeText(data.invitation_url)
        ElMessage.success('链接已复制到剪贴板')
      } catch {
        ElMessage.warning('无法自动复制，请手动复制')
      }
    }).catch(() => {
      // 用户点击关闭，不做任何操作
    })
  },
  onError: (error: Error) => {
    ElMessage.error(error.message || '创建邀请码失败')
  },
})

// 撤销邀请 Mutation
const revokeMutation = useMutation({
  mutationFn: (code: string) => revokeInvitation(code),
  onSuccess: () => {
    ElMessage.success('邀请码已撤销')
    queryClient.invalidateQueries({ queryKey: ['invitations'] })
  },
  onError: (error: Error) => {
    ElMessage.error(error.message || '撤销邀请码失败')
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

// 计算序号
const getIndex = (index: number) => {
  return (currentPage.value - 1) * pageSize.value + index + 1
}

// 重置表单
const resetForm = () => {
  formData.value = {
    role: 'member',
    expiresInDays: 7,
    description: '',
  }
  formRef.value?.clearValidate()
}

// 打开创建对话框
const handleCreate = () => {
  resetForm()
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    const params: CreateInvitationParams = {
      role: formData.value.role,
      expires_in_days: formData.value.expiresInDays,
    }

    if (formData.value.description) {
      params.description = formData.value.description
    }

    createMutation.mutate(params)
  })
}

// 取消表单
const handleCancel = () => {
  dialogVisible.value = false
  resetForm()
}

// 撤销邀请确认
const handleRevoke = (row: Invitation) => {
  ElMessageBox.confirm(
    `确定要撤销邀请码 ${row.code} 吗？撤销后该邀请码将无法继续使用。`,
    '撤销邀请码',
    {
      confirmButtonText: '确定撤销',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(() => {
    revokeMutation.mutate(row.code)
  }).catch(() => {
    // 用户取消，不做任何操作
  })
}

// 复制邀请链接
const handleCopyLink = async (row: Invitation) => {
  const link = `${window.location.origin}/register?code=${row.code}`
  try {
    await navigator.clipboard.writeText(link)
    ElMessage.success('邀请链接已复制到剪贴板')
  } catch {
    ElMessage.warning('无法自动复制，请手动复制')
  }
}

// 格式化日期时间
const formatDateTime = (dateStr: string) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// 获取角色显示名称
const getRoleDisplayName = (role: string) => {
  const roleMap: Record<string, string> = {
    admin: '管理员',
    member: '成员',
    viewer: '查看者',
  }
  return roleMap[role] || role
}

// 状态显示配置
const getStatusConfig = (status: string) => {
  const statusMap: Record<string, { type: 'success' | 'info' | 'warning' | 'danger'; text: string }> = {
    active: { type: 'success', text: '有效' },
    used: { type: 'info', text: '已使用' },
    revoked: { type: 'danger', text: '已撤销' },
    expired: { type: 'warning', text: '已过期' },
  }
  return statusMap[status] || { type: 'info', text: status }
}

</script>

<template>
  <div class="invitations-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">邀请管理</h1>
        <p class="page-subtitle">创建和管理用户邀请码</p>
      </div>
      <button class="create-button" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        <span>创建邀请码</span>
      </button>
    </div>

    <!-- 搜索和操作栏 -->
    <div class="search-card">
      <div class="search-bar">
        <div class="search-input-group">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索邀请码或描述"
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

    <!-- 邀请列表 -->
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
        <h3 class="error-title">加载邀请列表失败</h3>
        <p class="error-message">{{ error?.message || '请检查网络连接或稍后重试' }}</p>
      </div>

      <!-- 数据表格 -->
      <div v-else class="table-container">
        <el-table
          :data="invitations"
          style="width: 100%"
          class="invitations-table"
          :empty-text="searchKeyword ? '没有找到匹配的邀请' : '暂无邀请数据'"
        >
          <el-table-column type="index" label="序号" width="80" :index="getIndex" />

          <el-table-column prop="code" label="邀请码" min-width="180">
            <template #default="{ row }">
              <span class="code-badge">{{ row.code }}</span>
            </template>
          </el-table-column>

          <el-table-column label="邀请人" width="120">
            <template #default="{ row }">
              <span class="inviter-text">{{ row.inviter?.username || '-' }}</span>
            </template>
          </el-table-column>

          <el-table-column label="角色" width="110" align="center">
            <template #default="{ row }">
              <div :class="['role-badge', `role-${row.role}`]">
                {{ getRoleDisplayName(row.role) }}
              </div>
            </template>
          </el-table-column>

          <el-table-column label="状态" width="90" align="center">
            <template #default="{ row }">
              <span
                :class="['status-badge', `status-${row.status}`]"
              >
                {{ getStatusConfig(row.status).text }}
              </span>
            </template>
          </el-table-column>

          <el-table-column label="过期时间" width="170">
            <template #default="{ row }">
              <span class="date-text">{{ formatDateTime(row.expires_at) }}</span>
            </template>
          </el-table-column>

          <el-table-column label="创建时间" width="170">
            <template #default="{ row }">
              <span class="date-text">{{ formatDateTime(row.created_at) }}</span>
            </template>
          </el-table-column>

          <el-table-column label="操作" width="180" fixed="right" align="center">
            <template #default="{ row }">
              <div class="action-buttons">
                <button class="action-button action-copy" @click="handleCopyLink(row)">
                  <el-icon><Link /></el-icon>
                  <span>复制链接</span>
                </button>
                <button
                  class="action-button action-revoke"
                  :disabled="row.status !== 'active'"
                  @click="handleRevoke(row)"
                >
                  <el-icon><Delete /></el-icon>
                  <span>撤销</span>
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

    <!-- 创建邀请对话框 -->
    <el-dialog
      v-model="dialogVisible"
      title="创建邀请码"
      width="480px"
      :close-on-click-modal="false"
      @close="handleCancel"
      class="invitation-dialog"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
        class="invitation-form"
      >
        <el-form-item label="角色" prop="role">
          <el-select v-model="formData.role" placeholder="请选择角色" style="width: 100%;">
            <el-option
              v-for="item in roleOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="有效期" prop="expiresInDays">
          <el-input-number
            v-model="formData.expiresInDays"
            :min="1"
            :max="365"
            placeholder="请输入有效期"
            style="width: 100%;"
          />
          <span class="form-tip">天（范围：1-365天）</span>
        </el-form-item>

        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入描述（可选）"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="handleCancel">取消</el-button>
        <el-button
          type="primary"
          :loading="createMutation.isPending.value"
          @click="handleSubmit"
        >
          创建
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.invitations-page {
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

.table-container {
  padding: 4px;
}

/* 表格样式 */
:deep(.invitations-table) {
  border: none;
}

:deep(.invitations-table .el-table__header-wrapper) {
  background: #f8fafc;
}

:deep(.invitations-table th.el-table__cell) {
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  color: #475569;
  font-weight: 600;
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 16px 12px;
}

:deep(.invitations-table td.el-table__cell) {
  border-bottom: 1px solid #f1f5f9;
  padding: 16px 12px;
}

:deep(.invitations-table tr:hover > td) {
  background: #f8fafc;
}

:deep(.invitations-table .el-table__empty-block) {
  padding: 40px 0;
}

/* 邀请码徽章 */
.code-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  color: #0284c7;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 600;
  font-family: 'Fira Code', monospace;
}

.inviter-text {
  font-size: 14px;
  color: #64748b;
}

/* 角色徽章 */
.role-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.role-admin {
  background: linear-gradient(135deg, #fff1f2 0%, #ffe4e6 100%);
  color: #e11d48;
}

.role-member {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  color: #2563eb;
}

.role-viewer {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
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
  white-space: nowrap;
}

.status-active {
  background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
  color: #059669;
}

.status-used {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  color: #2563eb;
}

.status-revoked {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  color: #dc2626;
}

.status-expired {
  background: linear-gradient(135deg, #fffbeb 0%, #fed7aa 100%);
  color: #d97706;
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

.action-copy {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  color: #2563eb;
}

.action-copy:hover {
  background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
  transform: translateY(-1px);
}

.action-revoke {
  background: linear-gradient(135deg, #fff1f2 0%, #ffe4e6 100%);
  color: #e11d48;
}

.action-revoke:hover:not(:disabled) {
  background: linear-gradient(135deg, #ffe4e6 0%, #fecdd3 100%);
  transform: translateY(-1px);
}

.action-revoke:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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
:deep(.invitation-dialog .el-dialog) {
  border-radius: 16px;
}

:deep(.invitation-dialog .el-dialog__header) {
  padding: 24px 24px 16px;
  border-bottom: 1px solid #f1f5f9;
}

:deep(.invitation-dialog .el-dialog__title) {
  font-size: 18px;
  font-weight: 700;
  color: #0f172a;
}

:deep(.invitation-dialog .el-dialog__body) {
  padding: 24px;
}

:deep(.invitation-dialog .el-dialog__footer) {
  padding: 16px 24px 24px;
  border-top: 1px solid #f1f5f9;
}

/* 表单样式 */
:deep(.invitation-form .el-form-item__label) {
  font-weight: 600;
  color: #475569;
}

:deep(.invitation-form .el-input__wrapper) {
  border-radius: 10px;
  box-shadow: none;
  border: 1px solid #e2e8f0;
  transition: all 0.2s ease;
}

:deep(.invitation-form .el-input__wrapper:hover) {
  border-color: #cbd5e1;
}

:deep(.invitation-form .el-input__wrapper.is-focus) {
  border-color: #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

:deep(.invitation-form .el-textarea__inner) {
  border-radius: 10px;
  box-shadow: none;
  border: 1px solid #e2e8f0;
  transition: all 0.2s ease;
}

:deep(.invitation-form .el-textarea__inner:hover) {
  border-color: #cbd5e1;
}

:deep(.invitation-form .el-textarea__inner:focus) {
  border-color: #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

.form-tip {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: #64748b;
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

  :deep(.invitations-table) {
    font-size: 13px;
  }

  .action-button span {
    display: none;
  }
}
</style>
