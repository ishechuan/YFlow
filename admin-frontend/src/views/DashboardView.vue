<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { getDashboardStats } from '@/services/dashboardService'
import {
  Folder,
  Document,
  MessageBox,
  ChatDotRound,
} from '@element-plus/icons-vue'

// 使用 vue-query 获取仪表板统计数据
const {
  data: stats,
  isLoading,
  isError,
  error,
} = useQuery({
  queryKey: ['dashboardStats'],
  queryFn: getDashboardStats,
})

// 统计卡片配置
const statCards = computed(() => [
  {
    title: '项目总数',
    value: stats.value?.total_projects ?? 0,
    icon: Folder,
    gradient: 'from-blue-500 to-blue-600',
    bgGradient: 'from-blue-50 to-blue-100',
    iconBg: 'bg-blue-500',
  },
  {
    title: '语言总数',
    value: stats.value?.total_languages ?? 0,
    icon: ChatDotRound,
    gradient: 'from-emerald-500 to-emerald-600',
    bgGradient: 'from-emerald-50 to-emerald-100',
    iconBg: 'bg-emerald-500',
  },
  {
    title: '翻译键总数',
    value: stats.value?.total_keys ?? 0,
    icon: Document,
    gradient: 'from-amber-500 to-amber-600',
    bgGradient: 'from-amber-50 to-amber-100',
    iconBg: 'bg-amber-500',
  },
  {
    title: '翻译总数',
    value: stats.value?.total_translations ?? 0,
    icon: MessageBox,
    gradient: 'from-rose-500 to-rose-600',
    bgGradient: 'from-rose-50 to-rose-100',
    iconBg: 'bg-rose-500',
  },
])
</script>

<template>
  <div class="dashboard-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div>
        <h1 class="page-title">仪表板</h1>
        <p class="page-subtitle">系统概览和统计信息</p>
      </div>
    </div>

    <!-- 统计卡片 -->
    <el-row v-if="isLoading" :gutter="20" class="stats-grid">
      <el-col v-for="i in 4" :key="i" :xs="24" :sm="12" :lg="6">
        <div class="stat-card stat-card-skeleton">
          <el-skeleton animated>
            <template #template>
              <div class="skeleton-icon">
                <el-skeleton-item variant="circle" style="width: 48px; height: 48px" />
              </div>
              <div class="skeleton-content">
                <el-skeleton-item variant="text" style="width: 60%; margin-bottom: 12px" />
                <el-skeleton-item variant="h1" style="width: 40%" />
              </div>
            </template>
          </el-skeleton>
        </div>
      </el-col>
    </el-row>

    <!-- 错误状态 -->
    <div v-else-if="isError" class="error-container">
      <div class="error-content">
        <svg class="error-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <h3 class="error-title">加载统计数据失败</h3>
        <p class="error-message">{{ error?.message || '请检查网络连接或稍后重试' }}</p>
      </div>
    </div>

    <!-- 统计卡片 -->
    <el-row v-else :gutter="20" class="stats-grid">
      <el-col
        v-for="card in statCards"
        :key="card.title"
        :xs="24"
        :sm="12"
        :lg="6"
      >
        <div :class="['stat-card', 'stat-card-hover']">
          <div :class="['stat-icon-wrapper', card.bgGradient]">
            <div :class="['stat-icon', card.iconBg]">
              <el-icon :size="24" color="#ffffff">
                <component :is="card.icon" />
              </el-icon>
            </div>
          </div>
          <div class="stat-content">
            <p class="stat-title">{{ card.title }}</p>
            <h3 :class="['stat-value', `text-gradient-${card.gradient}`]">
              {{ card.value.toLocaleString() }}
            </h3>
          </div>
          <div class="stat-decoration">
            <svg :class="['decoration-bg', card.iconBg]" viewBox="0 0 200 200" xmlns="http://www.w3.org/2000/svg">
              <path fill="currentColor" d="M44.7,-76.4C58.9,-69.2,71.8,-59.1,79.6,-46.9C87.4,-34.7,90.1,-20.4,85.8,-7.2C81.5,6,70.2,18.1,60.8,29.4C51.4,40.7,43.8,51.2,34.1,58.8C24.4,66.4,12.6,71.1,-0.4,71.8C-13.4,72.5,-27,68.9,-38.8,61.7C-50.6,54.5,-60.6,43.7,-68.9,31.3C-77.2,18.9,-83.8,4.9,-82.1,-8.5C-80.4,-21.9,-70.4,-34.7,-59.2,-43.1C-48,-51.5,-35.6,-55.5,-23.5,-63.6C-11.4,-71.7,0.4,-83.9,13.6,-86.3C26.8,-88.7,40.8,-81.3,44.7,-76.4Z" transform="translate(100 100)" />
            </svg>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 快捷操作 -->
    <el-row :gutter="20" class="actions-grid">
      <el-col :xs="24" :sm="12">
        <div class="action-card">
          <div class="action-icon-wrapper">
            <div class="action-icon bg-gradient-primary">
              <el-icon :size="28" color="#ffffff">
                <Folder />
              </el-icon>
            </div>
          </div>
          <div class="action-content">
            <h3 class="action-title">创建新项目</h3>
            <p class="action-description">开始一个新的翻译项目</p>
          </div>
          <el-button type="primary" round class="action-button">
            立即创建
          </el-button>
        </div>
      </el-col>
      <el-col :xs="24" :sm="12">
        <div class="action-card">
          <div class="action-icon-wrapper">
            <div class="action-icon bg-gradient-modern">
              <el-icon :size="28" color="#ffffff">
                <ChatDotRound />
              </el-icon>
            </div>
          </div>
          <div class="action-content">
            <h3 class="action-title">添加语言</h3>
            <p class="action-description">为项目添加新的语言支持</p>
          </div>
          <el-button type="primary" round class="action-button">
            立即添加
          </el-button>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.dashboard-page {
  max-width: 1400px;
  margin: 0 auto;
}

/* 页面头部 */
.page-header {
  margin-bottom: 32px;
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

/* 统计卡片网格 */
.stats-grid {
  margin-bottom: 24px;
}

.stat-card {
  position: relative;
  background: #ffffff;
  border-radius: 16px;
  padding: 24px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  align-items: center;
  gap: 16px;
  border: 1px solid #f1f5f9;
  height: 120px;
}

.stat-card-hover {
  cursor: default;
}

.stat-card-hover:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.1);
  border-color: transparent;
}

.stat-card-skeleton {
  pointer-events: none;
}

/* 骨架屏样式 */
.skeleton-icon {
  display: flex;
  align-items: center;
  justify-content: center;
}

.skeleton-content {
  flex: 1;
}

/* 统计图标 */
.stat-icon-wrapper {
  width: 64px;
  height: 64px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}

.stat-icon {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

/* 统计内容 */
.stat-content {
  flex: 1;
  z-index: 1;
}

.stat-title {
  margin: 0 0 8px;
  font-size: 14px;
  color: #64748b;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  margin: 0;
  font-size: 36px;
  font-weight: 800;
  line-height: 1;
  background: linear-gradient(135deg, #1e293b 0%, #475569 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* 装饰背景 */
.stat-decoration {
  position: absolute;
  right: -20px;
  bottom: -20px;
  width: 120px;
  height: 120px;
  opacity: 0.08;
  z-index: 0;
}

.decoration-bg {
  width: 100%;
  height: 100%;
}

/* 渐变类 */
.from-blue-50 {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
}

.from-blue-500 {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
}

.from-emerald-50 {
  background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
}

.from-emerald-500 {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
}

.from-amber-50 {
  background: linear-gradient(135deg, #fffbeb 0%, #fed7aa 100%);
}

.from-amber-500 {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
}

.from-rose-50 {
  background: linear-gradient(135deg, #fff1f2 0%, #ffe4e6 100%);
}

.from-rose-500 {
  background: linear-gradient(135deg, #f43f5e 0%, #e11d48 100%);
}

/* 错误状态 */
.error-container {
  background: #ffffff;
  border-radius: 16px;
  padding: 48px 24px;
  text-align: center;
  border: 1px solid #f1f5f9;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.error-content {
  max-width: 400px;
  margin: 0 auto;
}

.error-icon {
  width: 64px;
  height: 64px;
  color: #f43f5e;
  margin: 0 auto 20px;
}

.error-title {
  margin: 0 0 8px;
  font-size: 20px;
  font-weight: 700;
  color: #0f172a;
}

.error-message {
  margin: 0;
  font-size: 14px;
  color: #64748b;
}

/* 快捷操作 */
.actions-grid {
  margin-top: 24px;
}

.action-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  border: 1px solid #f1f5f9;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  align-items: center;
  gap: 20px;
  height: 100%;
}

.action-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.08);
  border-color: transparent;
}

.action-icon-wrapper {
  flex-shrink: 0;
}

.action-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(6, 182, 212, 0.3);
}

.action-content {
  flex: 1;
}

.action-title {
  margin: 0 0 4px;
  font-size: 18px;
  font-weight: 700;
  color: #0f172a;
}

.action-description {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}

.action-button {
  flex-shrink: 0;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .page-title {
    font-size: 24px;
  }

  .stat-card {
    padding: 20px;
    height: 100px;
  }

  .stat-icon-wrapper {
    width: 52px;
    height: 52px;
  }

  .stat-icon {
    width: 42px;
    height: 42px;
  }

  .stat-value {
    font-size: 28px;
  }

  .action-card {
    flex-direction: column;
    text-align: center;
    gap: 16px;
  }

  .action-button {
    width: 100%;
  }
}
</style>
