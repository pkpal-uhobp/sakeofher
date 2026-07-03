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
    error.value = e?.response?.data?.error || e.message || 'Нужно войти через Telegram'
  } finally {
    loading.value = false
  }
})

async function logout() {
  await api.post('/auth/logout')
  window.location.href = '/login'
}
</script>

<template>
  <main class="page">
    <h1>Админка</h1>
    <section class="card" v-if="loading">Проверяем авторизацию...</section>
    <section class="card" v-else-if="error">
      <p class="error">{{ error }}</p>
      <RouterLink class="telegram" to="/login">Войти через Telegram</RouterLink>
    </section>
    <section class="card" v-else>
      <p>Telegram ID: <b>{{ me.user.telegram_id }}</b></p>
      <p>Username: <b>{{ me.user.telegram_username || '—' }}</b></p>
      <p>Админ: <b>{{ me.is_admin ? 'да' : 'нет' }}</b></p>
      <p v-if="!me.is_admin" class="error">Этот Telegram ID не найден в таблице admins. Доступ к админским действиям будет закрыт.</p>
      <button class="secondary" @click="logout">Выйти</button>
    </section>
  </main>
</template>

<style scoped>
.page { max-width: 920px; margin: 0 auto; padding: 48px 20px; }
.card { background: #111827; border: 1px solid #263244; border-radius: 24px; padding: 28px; margin-top: 20px; box-shadow: 0 24px 80px rgba(0,0,0,.25); }
.error { color: #fca5a5; }
.telegram, .secondary { display: inline-block; margin-top: 18px; border: 0; border-radius: 16px; padding: 14px 18px; font-weight: 800; cursor: pointer; text-decoration: none; }
.telegram { background: #60a5fa; color: #020617; }
.secondary { background: #1f2937; color: #e5e7eb; }
</style>
