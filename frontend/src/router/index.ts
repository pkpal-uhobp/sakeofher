import { createRouter, createWebHistory } from 'vue-router'
import HomePage from '../pages/public/HomePage.vue'
import SubscriptionPage from '../pages/public/SubscriptionPage.vue'
import AdminDashboardPage from '../pages/admin/DashboardPage.vue'
import LoginPage from '../pages/auth/LoginPage.vue'
import AuthSuccessPage from '../pages/auth/AuthSuccessPage.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: HomePage },
    { path: '/s/:token', component: SubscriptionPage },
    { path: '/:secret/sub/:telegramId', component: SubscriptionPage },
    { path: '/login', component: LoginPage },
    { path: '/auth/success', component: AuthSuccessPage },
    { path: '/admin', component: AdminDashboardPage },
  ],
})
