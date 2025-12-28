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
          <a href="${data.invitation_url}" target="_blank" style="color: #409eff;">${data.invitation_url}</a>
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
      <div>
        <h1>邀请管理</h1>
        <p class="subtitle">创建和管理用户邀请码</p>
      </div>
      <el-button type="primary" :icon="Plus" @click="handleCreate">
        创建邀请码
      </el-button>
    </div>

    <!-- 搜索和操作栏 -->
    <el-card class="search-card" shadow="never">
      <div class="search-bar">
        <div class="search-input-group">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索邀请码或描述"
            :prefix-icon="Search"
            clearable
            @clear="handleClearSearch"
            @keyup.enter="handleSearch"
          />
          <el-button type="primary" :icon="Search" @click="handleSearch">
            搜索
          </el-button>
        </div>
        <el-button :icon="Refresh" @click="handleRefresh">刷新</el-button>
      </div>
    </el-card>

    <!-- 邀请列表 -->
    <el-card class="table-card" shadow="never">
      <!-- 加载中状态 -->
      <div v-if="isLoading" class="loading-container">
        <el-skeleton :rows="5" animated />
      </div>

      <!-- 错误状态 -->
      <el-alert
        v-else-if="isError"
        type="error"
        :title="error?.message || '加载邀请列表失败'"
        :description="'请检查网络连接或稍后重试'"
        show-icon
        :closable="false"
      />

      <!-- 数据表格 -->
      <div v-else>
        <el-table
          :data="invitations"
          style="width: 100%"
          :empty-text="searchKeyword ? '没有找到匹配的邀请' : '暂无邀请数据'"
        >
          <el-table-column type="index" label="序号" width="80" :index="getIndex" />

          <el-table-column prop="code" label="邀请码" min-width="180">
            <template #default="{ row }">
              <el-tag size="small" type="info">{{ row.code }}</el-tag>
            </template>
          </el-table-column>

          <el-table-column label="邀请人" width="120">
            <template #default="{ row }">
              {{ row.inviter?.username || '-' }}
            </template>
          </el-table-column>

          <el-table-column label="角色" width="100">
            <template #default="{ row }">
              <el-tag
                size="small"
                :type="row.role === 'admin' ? 'danger' : row.role === 'member' ? 'primary' : 'info'"
              >
                {{ getRoleDisplayName(row.role) }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusConfig(row.status).type" size="small">
                {{ getStatusConfig(row.status).text }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column label="过期时间" width="160">
            <template #default="{ row }">
              {{ formatDateTime(row.expires_at) }}
            </template>
          </el-table-column>

          <el-table-column label="创建时间" width="160">
            <template #default="{ row }">
              {{ formatDateTime(row.created_at) }}
            </template>
          </el-table-column>

          <el-table-column label="操作" width="160" fixed="right">
            <template #default="{ row }">
              <el-button
                type="primary"
                size="small"
                :icon="Link"
                link
                @click="handleCopyLink(row)"
              >
                复制链接
              </el-button>
              <el-button
                type="danger"
                size="small"
                :icon="Delete"
                link
                :disabled="row.status !== 'active'"
                @click="handleRevoke(row)"
              >
                撤销
              </el-button>
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
    </el-card>

    <!-- 创建邀请对话框 -->
    <el-dialog
      v-model="dialogVisible"
      title="创建邀请码"
      width="480px"
      :close-on-click-modal="false"
      @close="handleCancel"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
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

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0;
  font-size: 28px;
  color: #303133;
  font-weight: 600;
}

.subtitle {
  margin: 8px 0 0;
  font-size: 14px;
  color: #909399;
}

.search-card {
  margin-bottom: 16px;
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

.table-card {
  border-radius: 8px;
}

.loading-container {
  padding: 20px;
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.form-tip {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .page-header h1 {
    font-size: 24px;
  }

  .search-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .search-input-group {
    max-width: 100%;
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
