import { createRouter, createWebHistory } from 'vue-router'
import { api } from '../api/client'

import HomePage from '../pages/public/HomePage.vue'
import SubscriptionPage from '../pages/public/SubscriptionPage.vue'
import AdminDashboardPage from '../pages/admin/DashboardPage.vue'
import LoginPage from '../pages/auth/LoginPage.vue'

const publicRoutes = new Set([
  '/',
  '/login',
])

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: HomePage,
      meta: { public: true },
    },
    {
      path: '/s/:token',
      component: SubscriptionPage,
      meta: { public: true },
    },
    {
      path: '/:secret/sub/:telegramId',
      component: SubscriptionPage,
      meta: { public: true },
    },
    {
      path: '/login',
      component: LoginPage,
      meta: { public: true },
    },
    {
      path: '/admin',
      component: AdminDashboardPage,
      meta: { requiresAdmin: true },
    },
  ],
})

router.beforeEach(async (to) => {
  if (publicRoutes.has(to.path) || to.meta.public) {
    return true
  }

  if (!to.meta.requiresAdmin) {
    return true
  }

  try {
    const response = await api.get('/auth/me')
    const isAdmin = Boolean(response.data?.is_admin)

    if (!isAdmin) {
      return {
        path: '/login',
        query: {
          reason: 'admin_required',
          redirect: to.fullPath,
        },
      }
    }

    return true
  } catch {
    return {
      path: '/login',
      query: {
        reason: 'auth_required',
        redirect: to.fullPath,
      },
    }
  }
})
