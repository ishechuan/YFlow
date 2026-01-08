<template>
  <div class="users-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">用户管理</h1>
        <p class="page-subtitle">管理系统用户和权限</p>
      </div>
      <button class="create-button" @click="openAddDialog">
        <el-icon><Plus /></el-icon>
        <span>添加用户</span>
      </button>
    </div>

    <!-- 搜索和操作栏 -->
    <div class="search-card">
      <div class="search-bar">
        <div class="search-input-group">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索用户名或邮箱"
            :prefix-icon="Search"
            clearable
            class="search-input"
            @clear="handleSearch"
            @keyup.enter="handleSearch"
          />
          <button class="search-button" @click="handleSearch">
            <el-icon><Search /></el-icon>
            <span>搜索</span>
          </button>
        </div>
        <button class="refresh-button" @click="fetchUsers">
          <el-icon><Refresh /></el-icon>
        </button>
      </div>
    </div>

    <!-- 用户列表 -->
    <div class="table-card">
      <!-- 加载中状态 -->
      <div v-if="loading" class="loading-container">
        <el-skeleton :rows="5" animated />
      </div>

      <!-- 数据表格 -->
      <div v-else class="table-container">
        <el-table
          :data="users"
          style="width: 100%"
          class="users-table"
          :empty-text="searchKeyword ? '没有找到匹配的用户' : '暂无用户数据'"
        >
          <el-table-column prop="id" label="ID" width="80" align="center" />
          <el-table-column prop="username" label="用户名" min-width="140">
            <template #default="{ row }">
              <div class="username-cell">
                <strong class="username-text">{{ row.username }}</strong>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="email" label="邮箱" min-width="200">
            <template #default="{ row }">
              <span class="email-text">{{ row.email }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="role" label="角色" width="110" align="center">
            <template #default="{ row }">
              <div :class="['role-badge', `role-${row.role}`]">
                {{ getRoleDisplayName(row.role) }}
              </div>
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
          <el-table-column prop="created_at" label="创建时间" width="170">
            <template #default="{ row }">
              <span class="date-text">{{ formatDate(row.created_at) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="260" fixed="right" align="center">
            <template #default="{ row }">
              <div class="action-buttons">
                <button class="action-button action-edit" @click="openEditDialog(row)">
                  <el-icon><Edit /></el-icon>
                  <span>编辑</span>
                </button>
                <button class="action-button action-reset" @click="openResetPasswordDialog(row)">
                  <el-icon><RefreshRight /></el-icon>
                  <span>重置</span>
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
            :total="total"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
          />
        </div>
      </div>
    </div>

    <!-- 创建/编辑用户对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑用户' : '添加用户'"
      width="500px"
      :close-on-click-modal="false"
      class="user-dialog"
      @closed="resetForm"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="80px"
        class="user-form"
      >
        <el-form-item label="用户名" prop="username">
          <el-input v-model="formData.username" placeholder="请输入用户名" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="formData.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="密码" prop="password">
          <el-input
            v-model="formData.password"
            type="password"
            placeholder="请输入密码（至少6位）"
            show-password
          />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="formData.role" placeholder="请选择角色" style="width: 100%">
            <el-option label="管理员" value="admin" />
            <el-option label="成员" value="member" />
            <el-option label="查看者" value="viewer" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="isEdit" label="状态" prop="status">
          <el-select v-model="formData.status" placeholder="请选择状态" style="width: 100%">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="disabled" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          {{ isEdit ? '更新' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 重置密码对话框 -->
    <el-dialog
      v-model="passwordDialogVisible"
      title="重置密码"
      width="400px"
      class="password-dialog"
    >
      <el-form
        ref="passwordFormRef"
        :model="passwordFormData"
        :rules="passwordRules"
        label-width="100px"
        class="password-form"
      >
        <el-form-item label="用户">
          <span>{{ editingUser?.username }}</span>
        </el-form-item>
        <el-form-item label="新密码" prop="new_password">
          <el-input
            v-model="passwordFormData.new_password"
            type="password"
            placeholder="请输入新密码（至少6位）"
            show-password
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="passwordFormData.confirmPassword"
            type="password"
            placeholder="请再次输入新密码"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleResetPassword" :loading="resettingPassword">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit, Delete, Search, Refresh, RefreshRight } from '@element-plus/icons-vue'
import { getUsers, createUser, updateUser, deleteUser, resetUserPassword } from '@/services/user'
import type { User, UpdateUserRequest } from '@/types/api'
import type { FormInstance, FormRules } from 'element-plus'

const loading = ref(false)
const users = ref<User[]>([])
const dialogVisible = ref(false)
const passwordDialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const resettingPassword = ref(false)
const formRef = ref<FormInstance>()
const passwordFormRef = ref<FormInstance>()

// 分页
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const searchKeyword = ref('')

// 编辑状态
const editingId = ref<number | null>(null)
const editingUser = ref<User | null>(null)

// 用户表单数据
interface UserFormData {
  username: string
  email: string
  password: string
  role: 'admin' | 'member' | 'viewer'
  status: 'active' | 'disabled'
}

const formData = reactive<UserFormData>({
  username: '',
  email: '',
  password: '',
  role: 'member',
  status: 'active',
})

const passwordFormData = reactive({
  new_password: '',
  confirmPassword: '',
})

const validateConfirmPassword = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  if (value !== passwordFormData.new_password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules = reactive<FormRules>({
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度在 3-50 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_]+$/, message: '用户名只能包含字母、数字和下划线', trigger: 'blur' },
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 100, message: '密码长度至少 6 个字符', trigger: 'blur' },
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' },
  ],
  status: [
    { required: true, message: '请选择状态', trigger: 'change' },
  ],
})

const passwordRules = reactive<FormRules>({
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 100, message: '密码长度至少 6 个字符', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
})

onMounted(() => {
  fetchUsers()
})

const fetchUsers = async () => {
  loading.value = true
  try {
    const response = await getUsers({
      page: currentPage.value,
      page_size: pageSize.value,
      keyword: searchKeyword.value,
    })
    users.value = response.data
    total.value = response.meta.total_count
  } catch (error) {
    const message = error instanceof Error ? error.message : '获取用户列表失败'
    ElMessage.error(message)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  fetchUsers()
}

const handleSizeChange = () => {
  currentPage.value = 1
  fetchUsers()
}

const handleCurrentChange = () => {
  fetchUsers()
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

const getRoleDisplayName = (role: string) => {
  const roleMap: Record<string, string> = {
    admin: '管理员',
    member: '成员',
    viewer: '查看者',
  }
  return roleMap[role] || role
}

const resetForm = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  formData.username = ''
  formData.email = ''
  formData.password = ''
  formData.role = 'member'
  formData.status = 'active'
  editingId.value = null
  isEdit.value = false
}

const openAddDialog = () => {
  isEdit.value = false
  dialogVisible.value = true
}

const openEditDialog = (row: User) => {
  isEdit.value = true
  editingId.value = row.id
  formData.username = row.username || ''
  formData.email = row.email || ''
  formData.role = (row.role as 'admin' | 'member' | 'viewer') || 'member'
  formData.status = (row.status as 'active' | 'disabled') || 'active'
  dialogVisible.value = true
}

const openResetPasswordDialog = (row: User) => {
  editingUser.value = row
  passwordFormData.new_password = ''
  passwordFormData.confirmPassword = ''
  passwordDialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        if (isEdit.value && editingId.value !== null) {
          const updateData: UpdateUserRequest = {
            email: formData.email,
            role: formData.role,
            status: formData.status,
          }
          await updateUser(editingId.value as number, updateData)
          ElMessage.success('更新成功')
        } else {
          await createUser({
            username: formData.username,
            email: formData.email,
            password: formData.password,
            role: formData.role,
          })
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        fetchUsers()
      } catch (error) {
        const message = error instanceof Error ? error.message : '操作失败'
        ElMessage.error(message)
      } finally {
        submitting.value = false
      }
    }
  })
}

const handleResetPassword = async () => {
  if (!passwordFormRef.value || editingId.value === null) return

  await passwordFormRef.value.validate(async (valid) => {
    if (valid) {
      resettingPassword.value = true
      try {
        await resetUserPassword(editingId.value as number, {
          new_password: passwordFormData.new_password,
        })
        ElMessage.success('密码重置成功')
        passwordDialogVisible.value = false
      } catch (error) {
        const message = error instanceof Error ? error.message : '重置密码失败'
        ElMessage.error(message)
      } finally {
        resettingPassword.value = false
      }
    }
  })
}

const handleDelete = (row: User) => {
  ElMessageBox.confirm(
    `确定要删除用户 "${row.username}" 吗？此操作不可恢复。`,
    '删除确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await deleteUser(row.id)
      ElMessage.success('删除成功')
      fetchUsers()
    } catch (error) {
      const message = error instanceof Error ? error.message : '删除失败'
      ElMessage.error(message)
    }
  }).catch(() => {})
}
</script>

<style scoped>
.users-page {
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

.table-container {
  padding: 4px;
}

/* 表格样式 */
:deep(.users-table) {
  border: none;
}

:deep(.users-table .el-table__header-wrapper) {
  background: #f8fafc;
}

:deep(.users-table th.el-table__cell) {
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
  color: #475569;
  font-weight: 600;
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 16px 12px;
}

:deep(.users-table td.el-table__cell) {
  border-bottom: 1px solid #f1f5f9;
  padding: 16px 12px;
}

:deep(.users-table tr:hover > td) {
  background: #f8fafc;
}

:deep(.users-table .el-table__empty-block) {
  padding: 40px 0;
}

/* 用户名和邮箱 */
.username-text {
  font-size: 15px;
  font-weight: 600;
  color: #0f172a;
}

.email-text {
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

.status-inactive {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  color: #64748b;
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

.action-reset {
  background: linear-gradient(135deg, #fffbeb 0%, #fed7aa 100%);
  color: #d97706;
}

.action-reset:hover {
  background: linear-gradient(135deg, #fed7aa 0%, #fdba74 100%);
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
:deep(.user-dialog .el-dialog),
:deep(.password-dialog .el-dialog) {
  border-radius: 16px;
}

:deep(.user-dialog .el-dialog__header),
:deep(.password-dialog .el-dialog__header) {
  padding: 24px 24px 16px;
  border-bottom: 1px solid #f1f5f9;
}

:deep(.user-dialog .el-dialog__title),
:deep(.password-dialog .el-dialog__title) {
  font-size: 18px;
  font-weight: 700;
  color: #0f172a;
}

:deep(.user-dialog .el-dialog__body),
:deep(.password-dialog .el-dialog__body) {
  padding: 24px;
}

:deep(.user-dialog .el-dialog__footer),
:deep(.password-dialog .el-dialog__footer) {
  padding: 16px 24px 24px;
  border-top: 1px solid #f1f5f9;
}

/* 表单样式 */
:deep(.user-form .el-form-item__label),
:deep(.password-form .el-form-item__label) {
  font-weight: 600;
  color: #475569;
}

:deep(.user-form .el-input__wrapper),
:deep(.password-form .el-input__wrapper) {
  border-radius: 10px;
  box-shadow: none;
  border: 1px solid #e2e8f0;
  transition: all 0.2s ease;
}

:deep(.user-form .el-input__wrapper:hover),
:deep(.password-form .el-input__wrapper:hover) {
  border-color: #cbd5e1;
}

:deep(.user-form .el-input__wrapper.is-focus),
:deep(.password-form .el-input__wrapper.is-focus) {
  border-color: #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

:deep(.user-form .el-select .el-input__wrapper),
:deep(.password-form .el-select .el-input__wrapper) {
  cursor: pointer;
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

  :deep(.users-table) {
    font-size: 13px;
  }

  .action-button span {
    display: none;
  }
}
</style>
