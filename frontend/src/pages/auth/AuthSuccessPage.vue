<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { api } from '../../api/client'

const loading = ref(true)
const error = ref('')
const me = ref<any>(null)

onMounted(async () => {
  try {
    const { data } = await api.get('/auth/me')
    me.value = data
  } catch (e: any) {
    error.value = e?.response?.data?.error || e.message || 'Не удалось проверить авторизацию'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <main class="page">
    <section class="card">
      <p class="badge">Auth</p>
      <h1>Авторизация</h1>
      <p v-if="loading">Проверяем сессию...</p>
      <p v-else-if="error" class="error">{{ error }}</p>
      <div v-else>
        <p class="ok">Вы вошли как Telegram ID: {{ me.user.telegram_id }}</p>
        <p>Админ: {{ me.is_admin ? 'да' : 'нет' }}</p>
        <RouterLink class="telegram" to="/admin">Перейти в админку</RouterLink>
      </div>
    </section>
  </main>
</template>

<style scoped>
.page { max-width: 760px; margin: 0 auto; padding: 64px 20px; }
.card { background: #111827; border: 1px solid #263244; border-radius: 24px; padding: 32px; box-shadow: 0 24px 80px rgba(0,0,0,.25); }
.badge { display: inline-block; padding: 8px 12px; border: 1px solid #334155; border-radius: 999px; color: #93c5fd; }
h1 { font-size: 42px; margin: 18px 0; }
.error { color: #fca5a5; }
.ok { color: #bbf7d0; }
.telegram { display: inline-block; margin-top: 18px; border: 0; border-radius: 16px; background: #60a5fa; color: #020617; padding: 15px 18px; font-weight: 800; cursor: pointer; text-decoration: none; }
</style>
