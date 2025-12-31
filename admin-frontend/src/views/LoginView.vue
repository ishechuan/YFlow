<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useMutation } from '@tanstack/vue-query'
import { useAuthStore } from '@/stores/auth'
import { login as loginApi } from '@/services/authService'
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()

const formRef = ref<FormInstance>()

const form = reactive({
  username: '',
  password: '',
})

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度在 3-50 个字符', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 100, message: '密码长度至少 6 个字符', trigger: 'blur' },
  ],
}

// 使用 vue-query 的 useMutation 管理登录
const loginMutation = useMutation({
  mutationFn: () => loginApi({ username: form.username, password: form.password }),
  onSuccess: (data) => {
    // 验证响应数据完整性
    if (!data || !data.token || !data.refresh_token || !data.user) {
      ElMessage.error('登录响应数据不完整，请联系管理员')
      return
    }

    authStore.setAuth(data)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  },
  onError: (error: Error) => {
    // 显示错误消息
    const errorMessage = error.message || '登录失败，请检查用户名和密码'
    ElMessage.error(errorMessage)
  },
})

const loading = computed(() => loginMutation.isPending.value)
const error = computed(() => loginMutation.error.value)

const handleLogin = async () => {
  if (!formRef.value) return
  
  // 防止重复提交
  if (loading.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return
    loginMutation.mutate()
  })
}

// 支持回车登录
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' && !loading.value) {
    handleLogin()
  }
}
</script>

<template>
  <div class="login-container" @keydown="handleKeydown">
    <el-card class="login-card">
      <template #header>
        <div class="card-header">
          <h2>yflow 管理后台</h2>
          <p class="subtitle">用户登录</p>
        </div>
      </template>

      <el-alert
        v-if="error"
        :title="error.message"
        type="error"
        show-icon
        class="login-error"
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
            placeholder="请输入用户名"
            prefix-icon="User"
            clearable
          />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            prefix-icon="Lock"
            show-password
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            @click="handleLogin"
            class="login-btn"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.card-header {
  text-align: center;
}

.card-header h2 {
  margin: 0;
  font-size: 24px;
  color: #303133;
}

.subtitle {
  margin: 8px 0 0;
  font-size: 14px;
  color: #909399;
}

.login-error {
  margin-bottom: 20px;
}

.login-btn {
  width: 100%;
  height: 44px;
  font-size: 16px;
}
</style>
