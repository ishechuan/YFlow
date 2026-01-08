<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  Fold,
  Expand,
  User,
  Odometer,
  FolderOpened,
  ChatDotRound,
  Document,
  Setting,
  UserFilled,
  Postcard,
  ArrowDown,
  SwitchButton,
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

// 侧边栏折叠状态
const isCollapse = ref(false)

// 判断是否为管理员
const isAdmin = computed(() => authStore.user?.role === 'admin')

// 菜单项配置
const menuItems = computed(() => {
  const items = [
    {
      index: '/dashboard',
      title: '仪表板',
      icon: Odometer,
    },
    {
      index: '/projects',
      title: '项目管理',
      icon: FolderOpened,
    },
    {
      index: '/languages',
      title: '语言管理',
      icon: ChatDotRound,
    },
    {
      index: '/translations',
      title: '翻译管理',
      icon: Document,
    },
  ]

  // 管理员才能看到用户管理和邀请管理
  if (isAdmin.value) {
    items.push({
      index: '/users',
      title: '用户管理',
      icon: UserFilled,
    })
    items.push({
      index: '/invitations',
      title: '邀请管理',
      icon: Postcard,
    })
  }

  // 所有用户都能看到系统设置
  items.push({
    index: '/settings',
    title: '系统设置',
    icon: Setting,
  })

  return items
})

// 当前激活的菜单项
const activeMenu = computed(() => {
  return route.path
})

// 页面标题
const pageTitle = computed(() => {
  return route.meta.title as string || '仪表板'
})

// 切换侧边栏
const toggleSidebar = () => {
  isCollapse.value = !isCollapse.value
}

// 处理菜单点击
const handleMenuSelect = (index: string) => {
  router.push(index)
}

// 退出登录
const handleLogout = () => {
  authStore.logout()
}
</script>

<template>
  <el-container class="layout-container">
    <!-- 侧边栏 - 深海风格 -->
    <el-aside :width="isCollapse ? '72px' : '260px'" class="layout-aside">
      <!-- Logo 区域 -->
      <div class="logo-section">
        <div v-if="!isCollapse" class="logo-expanded">
          <div class="logo-icon-wrapper">
            <div class="logo-icon">Y</div>
          </div>
          <div class="logo-text">
            <span class="logo-title">YFlow</span>
            <span class="logo-subtitle">Translation Platform</span>
          </div>
        </div>
        <div v-else class="logo-collapsed">
          <div class="logo-icon-small">Y</div>
        </div>
      </div>

      <!-- 导航菜单 -->
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        :collapse-transition="false"
        class="layout-menu"
        @select="handleMenuSelect"
      >
        <el-menu-item
          v-for="item in menuItems"
          :key="item.index"
          :index="item.index"
        >
          <el-icon><component :is="item.icon" /></el-icon>
          <template #title>
            <span class="menu-title">{{ item.title }}</span>
          </template>
        </el-menu-item>
      </el-menu>

      <!-- 侧边栏底部装饰 -->
      <div v-if="!isCollapse" class="sidebar-footer">
        <div class="footer-decoration">
          <div class="decoration-line"></div>
          <p class="footer-text">Pro Edition</p>
        </div>
      </div>
    </el-aside>

    <!-- 右侧容器 -->
    <el-container class="main-container">
      <!-- 头部 -->
      <el-header class="layout-header">
        <div class="header-left">
          <el-button
            :icon="isCollapse ? Expand : Fold"
            class="toggle-btn"
            @click="toggleSidebar"
          />
          <el-breadcrumb separator="/" class="breadcrumb">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ pageTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <div class="header-right">
          <!-- 用户信息下拉 -->
          <el-dropdown @command="handleLogout" trigger="click">
            <div class="user-dropdown">
              <el-avatar :size="36" class="user-avatar">
                <el-icon><User /></el-icon>
              </el-avatar>
              <div class="user-info">
                <span class="user-name">{{ authStore.user?.username }}</span>
                <span class="user-role">{{ isAdmin ? '管理员' : '普通用户' }}</span>
              </div>
              <el-icon class="dropdown-icon"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item disabled>
                  <div class="user-detail">
                    <div class="detail-row">
                      <span class="detail-label">用户 ID:</span>
                      <span class="detail-value">{{ authStore.user?.id }}</span>
                    </div>
                    <div class="detail-row">
                      <span class="detail-label">用户名:</span>
                      <span class="detail-value">{{ authStore.user?.username }}</span>
                    </div>
                  </div>
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  <span>退出登录</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 主内容区 -->
      <el-main class="layout-main">
        <router-view v-slot="{ Component }">
          <transition name="page-transition" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
/* 布局容器 */
.layout-container {
  height: 100vh;
  overflow: hidden;
}

.main-container {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 侧边栏样式 - 深海极光风格 */
.layout-aside {
  background: linear-gradient(180deg, #0c4a6e 0%, #020617 100%);
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  box-shadow: 4px 0 24px rgba(0, 0, 0, 0.15);
  z-index: 100;
  display: flex;
  flex-direction: column;
  border-right: 1px solid rgba(6, 182, 212, 0.1);
}

/* Logo 区域 */
.logo-section {
  height: 72px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 16px;
  border-bottom: 1px solid rgba(6, 182, 212, 0.1);
  background: rgba(0, 0, 0, 0.2);
}

.logo-expanded {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-icon-wrapper {
  position: relative;
}

.logo-icon {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #06b6d4 0%, #14b8a6 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 700;
  color: white;
  box-shadow: 0 0 20px rgba(6, 182, 212, 0.4);
}

.logo-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.logo-title {
  font-size: 18px;
  font-weight: 700;
  color: #ffffff;
  letter-spacing: 0.5px;
}

.logo-subtitle {
  font-size: 10px;
  color: rgba(6, 182, 212, 0.6);
  font-weight: 500;
  letter-spacing: 0.3px;
}

.logo-collapsed {
  display: flex;
  justify-content: center;
}

.logo-icon-small {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, #06b6d4 0%, #14b8a6 100%);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 700;
  color: white;
  box-shadow: 0 0 20px rgba(6, 182, 212, 0.4);
}

/* 菜单样式 */
.layout-menu {
  flex: 1;
  border-right: none;
  background: transparent;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 12px 8px;
}

/* 自定义滚动条 */
.layout-menu::-webkit-scrollbar {
  width: 4px;
}

.layout-menu::-webkit-scrollbar-track {
  background: transparent;
}

.layout-menu::-webkit-scrollbar-thumb {
  background: rgba(6, 182, 212, 0.3);
  border-radius: 4px;
}

.layout-menu::-webkit-scrollbar-thumb:hover {
  background: rgba(6, 182, 212, 0.5);
}

:deep(.el-menu) {
  background-color: transparent;
  border-right: none;
}

:deep(.el-menu-item) {
  height: 48px;
  line-height: 48px;
  margin: 0 0 4px 0;
  border-radius: 12px;
  color: rgba(255, 255, 255, 0.6);
  display: flex;
  align-items: center;
  transition: all 0.2s ease;
  position: relative;
  overflow: hidden;
}

:deep(.el-menu-item::before) {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  height: 100%;
  width: 3px;
  background: linear-gradient(180deg, #06b6d4 0%, #14b8a6 100%);
  border-radius: 0 3px 3px 0;
  opacity: 0;
  transition: opacity 0.2s ease;
}

:deep(.el-menu-item:hover) {
  background-color: rgba(6, 182, 212, 0.1);
  color: rgba(255, 255, 255, 0.9);
}

:deep(.el-menu-item:hover::before) {
  opacity: 1;
}

:deep(.el-menu-item.is-active) {
  background: linear-gradient(90deg, rgba(6, 182, 212, 0.2) 0%, rgba(20, 184, 166, 0.2) 100%);
  color: #22d3ee;
  font-weight: 500;
  box-shadow: 0 0 20px rgba(6, 182, 212, 0.2);
}

:deep(.el-menu-item.is-active::before) {
  opacity: 1;
}

:deep(.el-menu-item .el-icon) {
  font-size: 18px;
  margin-right: 12px;
  width: 18px;
  text-align: center;
}

:deep(.el-menu--collapse .el-menu-item) {
  margin: 0 0 4px 0;
  justify-content: center;
  padding: 0 !important;
}

:deep(.el-menu--collapse .el-menu-item .el-icon) {
  margin-right: 0;
}

.menu-title {
  font-size: 14px;
  font-weight: 500;
  letter-spacing: 0.3px;
}

/* 侧边栏底部 */
.sidebar-footer {
  padding: 16px;
  border-top: 1px solid rgba(6, 182, 212, 0.1);
  background: rgba(0, 0, 0, 0.2);
}

.footer-decoration {
  text-align: center;
}

.decoration-line {
  width: 40px;
  height: 2px;
  background: linear-gradient(90deg, transparent, rgba(6, 182, 212, 0.5), transparent);
  margin: 0 auto 8px;
}

.footer-text {
  font-size: 11px;
  color: rgba(6, 182, 212, 0.5);
  font-weight: 500;
  letter-spacing: 1px;
  margin: 0;
}

/* 头部样式 */
.layout-header {
  height: 64px;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.05);
  z-index: 50;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.toggle-btn {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  transition: all 0.2s ease;
}

.toggle-btn:hover {
  background-color: #f1f5f9;
}

.breadcrumb {
  font-size: 14px;
}

:deep(.el-breadcrumb__item) {
  font-size: 14px;
}

:deep(.el-breadcrumb__inner) {
  color: #64748b;
  font-weight: 500;
}

:deep(.el-breadcrumb__inner.is-link) {
  color: #94a3b8;
}

:deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
  color: #0f172a;
  font-weight: 600;
}

.header-right {
  display: flex;
  align-items: center;
}

/* 用户下拉框 */
.user-dropdown {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: 12px;
  transition: all 0.2s ease;
}

.user-dropdown:hover {
  background-color: #f8fafc;
}

.user-avatar {
  background: linear-gradient(135deg, #06b6d4 0%, #14b8a6 100%);
  color: white;
}

.user-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.user-name {
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
  line-height: 1.2;
}

.user-role {
  font-size: 11px;
  color: #64748b;
  font-weight: 500;
  line-height: 1.2;
}

.dropdown-icon {
  color: #94a3b8;
  font-size: 14px;
  transition: transform 0.2s ease;
}

.user-dropdown:hover .dropdown-icon {
  transform: rotate(180deg);
}

/* 下拉菜单样式 */
:deep(.user-detail) {
  padding: 4px 0;
  min-width: 200px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 16px;
}

.detail-label {
  font-size: 13px;
  color: #64748b;
  font-weight: 500;
}

.detail-value {
  font-size: 13px;
  color: #0f172a;
  font-weight: 600;
}

:deep(.el-dropdown-menu__item) {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  padding: 8px 16px;
}

/* 主内容区 */
.layout-main {
  flex: 1;
  background: linear-gradient(180deg, #f8fafc 0%, #f1f5f9 100%);
  padding: 24px;
  overflow-y: auto;
  overflow-x: hidden;
}

/* 页面过渡动画 */
.page-transition-enter-active,
.page-transition-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.page-transition-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.page-transition-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .layout-main {
    padding: 16px;
  }

  .layout-header {
    padding: 0 16px;
  }

  .user-info {
    display: none;
  }

  .dropdown-icon {
    display: none;
  }

  .layout-aside {
    position: fixed !important;
    height: 100vh;
    z-index: 1000;
  }

  .logo-expanded {
    display: none;
  }

  .logo-collapsed {
    display: flex;
  }
}
</style>
