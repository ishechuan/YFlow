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
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '240px'" class="layout-aside">
      <div class="logo-container">
        <div v-if="!isCollapse" class="logo-text">
          <h2>yflow</h2>
        </div>
        <div v-else class="logo-icon">
          <span>i18n</span>
        </div>
      </div>

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
          <template #title>{{ item.title }}</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 右侧容器 -->
    <el-container>
      <!-- 头部 -->
      <el-header class="layout-header">
        <div class="header-left">
          <el-button
            :icon="isCollapse ? Expand : Fold"
            text
            @click="toggleSidebar"
          />
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ pageTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <div class="header-right">
          <el-dropdown @command="handleLogout">
            <div class="user-info">
              <el-avatar :size="32" :src="''">
                <el-icon><User /></el-icon>
              </el-avatar>
              <span class="username">{{ authStore.user?.username }}</span>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item disabled>
                  <div class="user-detail">
                    <p><strong>用户 ID:</strong> {{ authStore.user?.id }}</p>
                    <p><strong>用户名:</strong> {{ authStore.user?.username }}</p>
                  </div>
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 主内容区 -->
      <el-main class="layout-main">
        <router-view v-slot="{ Component }">
          <transition name="fade-slide" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.layout-container {
  height: 100vh;
}

.layout-aside {
  background-color: #0f172a;
  transition: width 0.3s cubic-bezier(0.2, 0, 0, 1) 0s;
  overflow: hidden;
  box-shadow: 2px 0 8px 0 rgba(29, 35, 41, 0.05);
  z-index: 10;
}

.logo-container {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #0f172a;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.logo-text h2 {
  margin: 0;
  color: #ffffff;
  font-size: 20px;
  font-weight: 600;
  letter-spacing: 0.5px;
  background: linear-gradient(to right, #ffffff, #e2e8f0);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.logo-icon {
  color: #3b82f6;
  font-size: 20px;
  font-weight: 700;
}

.layout-menu {
  border-right: none;
  height: calc(100vh - 64px);
  background-color: #0f172a;
  overflow-y: auto;
}

/* Menu Item Styles */
:deep(.el-menu) {
  background-color: #0f172a;
  border-right: none;
}

:deep(.el-menu-item) {
  height: 50px;
  line-height: 50px;
  margin: 4px 8px;
  border-radius: 6px;
  color: #94a3b8;
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

/* Collapsed menu item centering */
:deep(.el-menu--collapse .el-menu-item) {
  margin: 4px 0;
  justify-content: center;
  padding: 0 !important;
}

:deep(.el-menu-item:hover) {
  background-color: rgba(255, 255, 255, 0.05);
  color: #ffffff;
}

:deep(.el-menu-item.is-active) {
  background-color: #3b82f6;
  color: #ffffff;
  font-weight: 500;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

:deep(.el-menu-item .el-icon) {
  font-size: 18px;
  margin-right: 12px;
}

:deep(.el-menu--collapse .el-menu-item .el-icon) {
  margin-right: 0;
}

.layout-menu:not(.el-menu--collapse) {
  width: 240px;
}

.layout-header {
  height: 64px;
  background: #ffffff;
  border-bottom: 1px solid #f1f5f9;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.02);
  z-index: 9;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: 6px;
  transition: all 0.3s;
}

.user-info:hover {
  background-color: #f8fafc;
}

.username {
  font-size: 14px;
  font-weight: 500;
  color: #334155;
}

.user-detail {
  padding: 8px 0;
  min-width: 180px;
}

.user-detail p {
  margin: 6px 0;
  font-size: 13px;
  color: #64748b;
  padding: 0 16px;
}

.layout-main {
  background-color: #f8fafc;
  padding: 24px;
  overflow-y: auto;
  overflow-x: hidden;
}

/* Page Transitions */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.3s ease;
}

.fade-slide-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.fade-slide-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

/* Responsive Design */
@media (max-width: 768px) {
  .layout-main {
    padding: 16px;
  }

  .username {
    display: none;
  }
  
  .layout-menu:not(.el-menu--collapse) {
    width: 200px;
    position: absolute;
    z-index: 100;
    height: 100vh;
  }
}
</style>
