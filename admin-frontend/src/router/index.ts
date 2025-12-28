import { createRouter, createWebHistory } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import LoginView from '@/views/LoginView.vue'
import DashboardView from '@/views/DashboardView.vue'
import MainLayout from '@/layouts/MainLayout.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginView,
      meta: { requiresAuth: false },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/RegisterView.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/',
      component: MainLayout,
      meta: { requiresAuth: true },
      redirect: '/dashboard',
      children: [
        {
          path: '/dashboard',
          name: 'dashboard',
          component: DashboardView,
          meta: { title: '仪表板' },
        },
        {
          path: '/projects',
          name: 'projects',
          component: () => import('@/views/ProjectsView.vue'),
          meta: { title: '项目管理' },
        },
        {
          path: '/languages',
          name: 'languages',
          component: () => import('@/views/LanguagesView.vue'),
          meta: { title: '语言管理' },
        },
        {
          path: '/translations',
          name: 'translations',
          component: () => import('@/views/TranslationsView.vue'),
          meta: { title: '翻译管理' },
        },
        {
          path: '/users',
          name: 'users',
          component: () => import('@/views/UsersView.vue'),
          meta: { title: '用户管理', roles: ['admin'] },
        },
        {
          path: '/invitations',
          name: 'invitations',
          component: () => import('@/views/InvitationsView.vue'),
          meta: { title: '邀请管理', roles: ['admin'] },
        },
        // 未来可以在这里添加更多路由，如项目管理、语言管理等
      ],
    },
  ],
})

// 路由守卫
router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'login', query: { redirect: to.fullPath } })
  } else if (to.name === 'login' && authStore.isAuthenticated) {
    next({ name: 'dashboard' })
  } else if (to.meta.roles && Array.isArray(to.meta.roles)) {
    // 检查用户角色权限
    const userRole = authStore.user?.role
    if (userRole && !to.meta.roles.includes(userRole)) {
      ElMessage.warning('您没有权限访问此页面')
      next({ name: 'dashboard' })
    } else {
      next()
    }
  } else {
    next()
  }
})

export default router
