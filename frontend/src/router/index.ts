import { createRouter, createWebHistory } from 'vue-router'
import HomePage from '../pages/public/HomePage.vue'
import SubscriptionPage from '../pages/public/SubscriptionPage.vue'
import AdminDashboardPage from '../pages/admin/DashboardPage.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: HomePage },
    { path: '/s/:token', component: SubscriptionPage },
    { path: '/admin', component: AdminDashboardPage },
  ],
})
