<template>
  <main class="admin-shell">
    <div class="admin-layout">
      <aside class="admin-sidebar">
        <RouterLink class="admin-brand" to="/">
          <span></span>
          <strong>SakeOfHer</strong>
        </RouterLink>

        <section class="admin-user">
          <p class="admin-user-label">Вы вошли</p>
          <p class="admin-user-name">{{ displayName }}</p>
          <p class="admin-user-note">Панель управления</p>
        </section>

        <nav class="admin-nav">
          <RouterLink to="/panel">Обзор</RouterLink>
          <RouterLink to="/panel/users">Пользователи</RouterLink>
          <RouterLink to="/panel/subscriptions">Подписки</RouterLink>
          <RouterLink to="/panel/tariffs">Тарифы</RouterLink>
          <RouterLink to="/">Лендинг</RouterLink>
          <button type="button" @click="handleLogout">Выйти</button>
        </nav>
      </aside>

      <section class="admin-main">
        <div class="admin-mobile-bar">
          <RouterLink class="admin-brand mobile" to="/">
            <span></span>
            <strong>SakeOfHer</strong>
          </RouterLink>
          <RouterLink class="admin-link-button" to="/panel">Панель</RouterLink>
        </div>

        <slot />
      </section>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { getMe, logout, type User } from '../../api/admin'

const router = useRouter()
const user = ref<User | null>(null)
const fallbackUsername = ref('')

const displayName = computed(() => {
  if (fallbackUsername.value) return fallbackUsername.value
  if (!user.value) return 'Загрузка…'

  return user.value.telegram_username
    ? String(user.value.telegram_username)
    : user.value.telegram_first_name || `ID ${user.value.telegram_id}`
})

onMounted(async () => {
  try {
    const me = await getMe()
    user.value = me.user
    fallbackUsername.value = (me as any).username || ''
  } catch {
    user.value = null
  }
})

async function handleLogout() {
  try {
    await logout()
  } finally {
    router.replace('/login')
  }
}
</script>

<style src="./admin.css"></style>
