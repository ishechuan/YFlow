<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useQuery, useMutation } from '@tanstack/vue-query'
import { useAuthStore } from '@/stores/auth'
import { validateInvitation, registerWithInvitation } from '@/services/invitation'
import type { FormInstance, FormRules } from 'element-plus'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const formRef = ref<FormInstance>()

const invitationCode = ref<string>('')
const invitationValid = ref<boolean>(false)
const invitationInvalid = ref<boolean>(false)
const validationMessage = ref<string>('')

const form = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
})

// 验证密码一致性
const validateConfirmPassword = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
  if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules: FormRules = {
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
    { min: 8, max: 100, message: '密码长度至少 8 个字符', trigger: 'blur' },
    { pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/, message: '密码必须包含大小写字母和数字', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
}

// 使用 vue-query 的 useQuery 验证邀请码
const {
  data: validationData,
  isLoading: isValidating,
  isError: validationError,
  error: validationErrorMsg,
} = useQuery({
  queryKey: ['invitation', invitationCode],
  queryFn: () => validateInvitation(invitationCode.value),
  enabled: computed(() => invitationCode.value.length > 0),
  retry: false,
})

// 监听验证结果
const validationResult = computed(() => validationData.value)
const isValidInvitation = computed(() => validationResult.value?.valid === true)

// 验证成功/失败的处理
onMounted(() => {
  const code = route.query.code as string
  if (code) {
    invitationCode.value = code
  } else {
    invitationInvalid.value = true
    validationMessage.value = '无效的邀请链接，请确认链接是否正确'
  }
})

// 监听验证结果变化
import { watch } from 'vue'
watch(isValidInvitation, (valid) => {
  if (valid && validationResult.value) {
    invitationValid.value = true
    invitationInvalid.value = false
  }
})
watch(validationError, (error) => {
  if (error) {
    invitationValid.value = false
    invitationInvalid.value = true
    validationMessage.value = (validationErrorMsg.value as Error)?.message || '邀请码验证失败'
  }
})

// 注册 Mutation
const registerMutation = useMutation({
  mutationFn: () =>
    registerWithInvitation({
      code: invitationCode.value,
      username: form.username,
      email: form.email,
      password: form.password,
    }),
  onSuccess: (data) => {
    if (!data || !data.user) {
      ElMessage.error('注册响应数据不完整')
      return
    }

    authStore.setAuth({
      token: (data as { token?: string }).token || '',
      refresh_token: (data as { refresh_token?: string }).refresh_token || '',
      user: data.user,
    })

    ElMessage.success('注册成功，欢迎加入！')
    router.push('/dashboard')
  },
  onError: (error: Error) => {
    ElMessage.error(error.message || '注册失败，请稍后重试')
  },
})

const loading = computed(() => registerMutation.isPending.value)
const registerError = computed(() => registerMutation.error.value)

const handleRegister = async () => {
  if (!formRef.value) return

  if (loading.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return
    registerMutation.mutate()
  })
}

// 支持回车注册
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' && !loading.value) {
    handleRegister()
  }
}

// 格式化日期
const formatDate = (dateStr: string) => {
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
</script>

<template>
  <div class="register-container" @keydown="handleKeydown">
    <div class="register-content">
      <div class="register-header">
        <h1>yflow</h1>
        <p class="subtitle">加入我们一起管理国际化资源</p>
      </div>

      <!-- 验证中状态 -->
      <el-card v-if="isValidating" class="loading-card">
        <div class="loading-state">
          <el-icon class="is-loading" :size="24"><Loading /></el-icon>
          <p>正在验证邀请码...</p>
        </div>
      </el-card>

      <!-- 邀请信息卡片（验证成功后显示） -->
      <el-card v-if="invitationValid" class="invitation-card">
        <template #header>
          <div class="card-header">
            <el-icon><Check /></el-icon>
            <span>邀请信息</span>
          </div>
        </template>

        <div class="invitation-info">
          <div class="info-item">
            <span class="label">邀请人</span>
            <span class="value">{{ validationResult?.inviter?.username || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="label">将授予角色</span>
            <el-tag>{{ getRoleDisplayName(validationResult?.role || '') }}</el-tag>
          </div>
          <div class="info-item">
            <span class="label">过期时间</span>
            <span class="value">{{ formatDate(validationResult?.expires_at || '') }}</span>
          </div>
        </div>
      </el-card>

      <!-- 邀请码无效 -->
      <el-card v-if="invitationInvalid" class="error-card">
        <div class="error-state">
          <el-icon :size="48"><CircleCloseFilled /></el-icon>
          <h3>邀请码无效</h3>
          <p>{{ validationMessage }}</p>
          <el-button type="primary" @click="router.push('/login')">
            返回登录
          </el-button>
        </div>
      </el-card>

      <!-- 注册表单 -->
      <el-card v-if="invitationValid" class="form-card">
        <template #header>
          <div class="card-header">
            <h2>用户注册</h2>
          </div>
        </template>

        <el-alert
          v-if="registerError"
          :title="registerError.message"
          type="error"
          show-icon
          class="register-error"
        />

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          size="large"
        >
          <el-form-item label="用户名" prop="username">
            <el-input
              v-model="form.username"
              placeholder="请输入用户名（3-50个字符，仅字母、数字和下划线）"
              prefix-icon="User"
              clearable
            />
          </el-form-item>

          <el-form-item label="邮箱" prop="email">
            <el-input
              v-model="form.email"
              type="email"
              placeholder="请输入邮箱地址"
              prefix-icon="Message"
              clearable
            />
          </el-form-item>

          <el-form-item label="密码" prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码（至少8位，包含大小写字母和数字）"
              prefix-icon="Lock"
              show-password
            />
          </el-form-item>

          <el-form-item label="确认密码" prop="confirmPassword">
            <el-input
              v-model="form.confirmPassword"
              type="password"
              placeholder="请再次输入密码"
              prefix-icon="Lock"
              show-password
            />
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              :loading="loading"
              @click="handleRegister"
              class="register-btn"
            >
              注册
            </el-button>
          </el-form-item>

          <div class="login-link">
            已有账号？<el-link type="primary" @click="router.push('/login')">立即登录</el-link>
          </div>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<style scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #06b6d4 0%, #14b8a6 100%);
  padding: 20px;
}

.register-content {
  width: 100%;
  max-width: 460px;
}

.register-header {
  text-align: center;
  margin-bottom: 24px;
  color: #fff;
}

.register-header h1 {
  margin: 0;
  font-size: 32px;
  font-weight: 700;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.register-header .subtitle {
  margin: 8px 0 0;
  font-size: 14px;
  opacity: 0.9;
}

.loading-card,
.error-card,
.form-card,
.invitation-card {
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
  margin-bottom: 10px;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 32px;
  color: #06b6d4;
}

.loading-state p {
  margin-top: 12px;
  font-size: 14px;
}

.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 32px;
  color: #f56c6c;
}

.error-state h3 {
  margin: 16px 0 8px;
  font-size: 18px;
}

.error-state p {
  margin: 0 0 20px;
  color: #909399;
  font-size: 14px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.invitation-info {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-item .label {
  color: #909399;
  font-size: 14px;
}

.info-item .value {
  color: #303133;
  font-weight: 500;
}

.register-error {
  margin-bottom: 20px;
}

.register-btn {
  width: 100%;
  height: 44px;
  font-size: 16px;
}

.login-link {
  text-align: center;
  margin-top: 8px;
  font-size: 14px;
  color: #909399;
}
</style>
