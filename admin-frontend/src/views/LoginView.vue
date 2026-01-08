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
  <div class="min-h-screen w-full flex items-center justify-center relative overflow-hidden" @keydown="handleKeydown">
    <!-- 深海极光背景 -->
    <div class="absolute inset-0 bg-gradient-ocean"></div>

    <!-- 动态装饰层 -->
    <div class="absolute inset-0 overflow-hidden">
      <!-- 极光波浪 -->
      <div class="aurora-wave aurora-wave-1"></div>
      <div class="aurora-wave aurora-wave-2"></div>
      <div class="aurora-wave aurora-wave-3"></div>

      <!-- 漂浮粒子 -->
      <div class="particle particle-1"></div>
      <div class="particle particle-2"></div>
      <div class="particle particle-3"></div>
      <div class="particle particle-4"></div>
      <div class="particle particle-5"></div>
      <div class="particle particle-6"></div>

      <!-- 光晕效果 -->
      <div class="glow-circle glow-circle-1"></div>
      <div class="glow-circle glow-circle-2"></div>
      <div class="glow-circle glow-circle-3"></div>
    </div>

    <!-- 登录卡片 -->
    <div class="relative z-10 w-full max-w-md px-6 animate-fade-in-up">
      <div class="glass-card rounded-3xl p-10 shadow-2xl">
        <!-- Logo 和标题 -->
        <div class="text-center mb-10">
          <div class="inline-flex items-center justify-center w-20 h-20 bg-gradient-aurora rounded-3xl shadow-lg shadow-cyan-500/30 mb-6 animate-glow">
            <span class="text-4xl font-bold text-white">Y</span>
          </div>
          <h1 class="text-4xl font-bold text-white mb-3 text-gradient-aurora">YFlow</h1>
          <p class="text-cyan-100/70 text-sm font-medium tracking-wide">国际化翻译管理平台</p>
        </div>

        <!-- 错误提示 -->
        <div v-if="error" class="mb-6">
          <div class="bg-red-500/20 backdrop-blur-sm border border-red-500/30 rounded-2xl p-4 flex items-start gap-3">
            <svg class="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
            </svg>
            <span class="text-red-300 text-sm font-medium">{{ error.message }}</span>
          </div>
        </div>

        <!-- 登录表单 -->
        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          size="large"
          class="login-form space-y-5"
        >
          <el-form-item prop="username">
            <el-input
              v-model="form.username"
              placeholder="请输入用户名"
              prefix-icon="User"
              clearable
              class="custom-input"
            />
          </el-form-item>

          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              prefix-icon="Lock"
              show-password
              class="custom-input"
            />
          </el-form-item>

          <el-form-item class="!mb-2">
            <button
              type="button"
              :disabled="loading"
              @click="handleLogin"
              class="login-button w-full py-4 bg-gradient-aurora hover:opacity-90 text-white font-semibold rounded-2xl shadow-xl shadow-cyan-500/30 transition-all duration-300 hover:shadow-2xl hover:shadow-cyan-500/40 hover:-translate-y-1 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:translate-y-0 disabled:shadow-xl disabled:scale-100 flex items-center justify-center gap-3 text-base"
            >
              <svg v-if="loading" class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span>{{ loading ? '登录中...' : '登录' }}</span>
            </button>
          </el-form-item>
        </el-form>

        <!-- 底部提示 -->
        <div class="mt-8 text-center">
          <p class="text-white/40 text-xs font-medium">
            登录即表示您同意我们的服务条款和隐私政策
          </p>
        </div>
      </div>

      <!-- 版权信息 -->
      <div class="text-center mt-8">
        <p class="text-white/30 text-sm font-medium">© 2026 YFlow. All rights reserved.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 极光波浪效果 */
.aurora-wave {
  position: absolute;
  width: 150%;
  height: 200%;
  border-radius: 40%;
  opacity: 0.1;
  animation: wave 15s ease-in-out infinite;
}

.aurora-wave-1 {
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.4) 0%, rgba(20, 184, 166, 0.4) 100%);
  top: -50%;
  left: -25%;
  animation-delay: 0s;
}

.aurora-wave-2 {
  background: linear-gradient(135deg, rgba(20, 184, 166, 0.3) 0%, rgba(34, 211, 238, 0.3) 100%);
  top: -60%;
  right: -25%;
  animation-delay: -5s;
  animation-duration: 18s;
}

.aurora-wave-3 {
  background: linear-gradient(135deg, rgba(34, 211, 238, 0.2) 0%, rgba(6, 182, 212, 0.2) 100%);
  bottom: -50%;
  left: 25%;
  animation-delay: -10s;
  animation-duration: 20s;
}

@keyframes wave {
  0%, 100% {
    transform: translate(0, 0) rotate(0deg) scale(1);
  }
  33% {
    transform: translate(30px, -50px) rotate(120deg) scale(1.1);
  }
  66% {
    transform: translate(-20px, 20px) rotate(240deg) scale(0.9);
  }
}

/* 漂浮粒子 */
.particle {
  position: absolute;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(6, 182, 212, 0.8) 0%, transparent 70%);
  animation: float-particle 20s ease-in-out infinite;
}

.particle-1 {
  width: 4px;
  height: 4px;
  top: 20%;
  left: 20%;
  animation-delay: 0s;
}

.particle-2 {
  width: 6px;
  height: 6px;
  top: 60%;
  left: 80%;
  animation-delay: -4s;
}

.particle-3 {
  width: 3px;
  height: 3px;
  top: 80%;
  left: 30%;
  animation-delay: -8s;
}

.particle-4 {
  width: 5px;
  height: 5px;
  top: 40%;
  left: 70%;
  animation-delay: -12s;
}

.particle-5 {
  width: 4px;
  height: 4px;
  top: 30%;
  right: 20%;
  animation-delay: -16s;
}

.particle-6 {
  width: 3px;
  height: 3px;
  bottom: 20%;
  right: 40%;
  animation-delay: -6s;
}

@keyframes float-particle {
  0%, 100% {
    transform: translate(0, 0) scale(1);
    opacity: 0.3;
  }
  25% {
    transform: translate(100px, -50px) scale(1.5);
    opacity: 0.8;
  }
  50% {
    transform: translate(-50px, 100px) scale(1);
    opacity: 0.5;
  }
  75% {
    transform: translate(50px, 50px) scale(1.2);
    opacity: 0.7;
  }
}

/* 光晕效果 */
.glow-circle {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
  animation: pulse-glow 8s ease-in-out infinite;
}

.glow-circle-1 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, rgba(6, 182, 212, 0.3) 0%, transparent 70%);
  top: -100px;
  left: -100px;
  animation-delay: 0s;
}

.glow-circle-2 {
  width: 300px;
  height: 300px;
  background: radial-gradient(circle, rgba(20, 184, 166, 0.25) 0%, transparent 70%);
  bottom: -50px;
  right: -50px;
  animation-delay: -4s;
}

.glow-circle-3 {
  width: 350px;
  height: 350px;
  background: radial-gradient(circle, rgba(34, 211, 238, 0.2) 0%, transparent 70%);
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  animation-delay: -2s;
}

@keyframes pulse-glow {
  0%, 100% {
    opacity: 0.5;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.2);
  }
}

/* 玻璃卡片 */
.glass-card {
  background: rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(40px);
  -webkit-backdrop-filter: blur(40px);
  border: 1px solid rgba(255, 255, 255, 0.15);
  box-shadow:
    0 8px 32px rgba(0, 0, 0, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

/* 自定义输入框样式 */
:deep(.custom-input .el-input__wrapper) {
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 16px;
  box-shadow: none;
  padding: 14px 18px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

:deep(.custom-input .el-input__wrapper:hover) {
  background: rgba(255, 255, 255, 0.12);
  border-color: rgba(6, 182, 212, 0.4);
  transform: translateY(-1px);
}

:deep(.custom-input .el-input__wrapper.is-focus) {
  background: rgba(255, 255, 255, 0.15);
  border-color: rgba(6, 182, 212, 0.6);
  box-shadow:
    0 0 0 4px rgba(6, 182, 212, 0.1),
    0 4px 20px rgba(6, 182, 212, 0.15),
    0 0 0 1px rgba(6, 182, 212, 0.2);
  transform: translateY(-1px);
}

:deep(.custom-input .el-input__inner) {
  color: #ffffff;
  font-size: 15px;
  font-weight: 500;
}

:deep(.custom-input .el-input__inner::placeholder) {
  color: rgba(255, 255, 255, 0.4);
}

:deep(.custom-input .el-input__prefix) {
  color: rgba(6, 182, 212, 0.7);
}

:deep(.custom-input .el-input__suffix) {
  color: rgba(6, 182, 212, 0.7);
}

:deep(.custom-input .el-input__clear) {
  color: rgba(6, 182, 212, 0.7);
}

:deep(.custom-input .el-input__password) {
  color: rgba(6, 182, 212, 0.7);
}

/* 表单项样式 */
:deep(.el-form-item) {
  margin-bottom: 20px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

:deep(.el-form-item__content) {
  position: relative;
  z-index: 10;
}

:deep(.el-form-item__error) {
  position: relative;
  z-index: 100;
  color: #fca5a5;
  font-size: 12px;
  margin-top: 8px;
  font-weight: 500;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  background: rgba(239, 68, 68, 0.15);
  padding: 4px 10px;
  border-radius: 8px;
  backdrop-filter: blur(8px);
  display: inline-block;
  border: 1px solid rgba(239, 68, 68, 0.3);
  animation: errorSlideIn 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.2);
}

@keyframes errorSlideIn {
  from {
    opacity: 0;
    transform: translateY(-8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 登录按钮样式 - 移除黑色外边框，添加优雅的焦点状态 */
.login-button {
  outline: none;
  position: relative;
  overflow: hidden;
  border: none;
}

.login-button::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(255, 255, 255, 0.3) 50%,
    transparent 100%
  );
  transition: left 0.6s ease;
}

.login-button:hover::before {
  left: 100%;
}

.login-button::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.15) 0%, transparent 50%, rgba(255, 255, 255, 0.1) 100%);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.login-button:hover::after {
  opacity: 1;
}

.login-button:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 3px rgba(6, 182, 212, 0.5),
    0 0 0 6px rgba(6, 182, 212, 0.2),
    0 8px 30px rgba(6, 182, 212, 0.3);
}

.login-button:active:not(:disabled) {
  transform: scale(0.98) translateY(0);
  box-shadow:
    0 0 0 3px rgba(6, 182, 212, 0.4),
    0 4px 15px rgba(6, 182, 212, 0.2);
}

/* 输入框错误状态 */
:deep(.el-form-item.is-error .el-input__wrapper) {
  border-color: rgba(239, 68, 68, 0.6);
  background: rgba(239, 68, 68, 0.08);
  animation: inputShake 0.4s cubic-bezier(0.36, 0.07, 0.19, 0.97);
}

@keyframes inputShake {
  0%, 100% { transform: translateX(0); }
  20%, 60% { transform: translateX(-4px); }
  40%, 80% { transform: translateX(4px); }
}

:deep(.el-form-item.is-error .el-input__wrapper:hover),
:deep(.el-form-item.is-error .el-input__wrapper.is-focus) {
  border-color: rgba(239, 68, 68, 0.8);
  box-shadow:
    0 0 0 4px rgba(239, 68, 68, 0.1),
    0 4px 20px rgba(239, 68, 68, 0.15);
}

/* 响应式设计 */
@media (max-width: 640px) {
  .glass-card {
    margin: 16px;
    padding: 32px 24px;
  }

  .glow-circle-1,
  .glow-circle-2,
  .glow-circle-3 {
    width: 200px;
    height: 200px;
  }
}
</style>
