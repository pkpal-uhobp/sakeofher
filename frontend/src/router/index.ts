import { createRouter, createWebHistory } from 'vue-router'
import { api } from '../api/client'

import HomePage from '../pages/public/HomePage.vue'
import SubscriptionPage from '../pages/public/SubscriptionPage.vue'
import AdminDashboardPage from '../pages/admin/DashboardPage.vue'
import UsersPage from '../pages/admin/UsersPage.vue'
import UserDetailsPage from '../pages/admin/UserDetailsPage.vue'
import SubscriptionsPage from '../pages/admin/SubscriptionsPage.vue'
import TariffsPage from '../pages/admin/TariffsPage.vue'
import LoginPage from '../pages/auth/LoginPage.vue'

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

    // Pretty HTML subscription cabinet.
    // Direct URL /{secret}/sub/{telegramId} is now reserved for Base64 subscription content.
    {
      path: '/profile/:secret/sub/:telegramId',
      component: SubscriptionPage,
      meta: { public: true },
    },

    {
      path: '/login',
      component: LoginPage,
      meta: { public: true },
    },
    {
      path: '/panel',
      component: AdminDashboardPage,
      meta: { requiresAdmin: true },
    },
    {
      path: '/panel/users',
      component: UsersPage,
      meta: { requiresAdmin: true },
    },
    {
      path: '/panel/users/:id',
      component: UserDetailsPage,
      meta: { requiresAdmin: true },
    },
    {
      path: '/panel/subscriptions',
      component: SubscriptionsPage,
      meta: { requiresAdmin: true },
    },
    {
      path: '/panel/tariffs',
      component: TariffsPage,
      meta: { requiresAdmin: true },
    },
    {
      path: '/admin',
      redirect: '/panel',
    },
  ],
})

router.beforeEach(async (to) => {
  if (to.meta.public) {
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
