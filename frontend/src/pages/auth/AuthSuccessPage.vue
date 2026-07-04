<template>
  <main class="auth-success">
    <section class="card">
      <h1>{{ title }}</h1>
      <p>{{ message }}</p>

      <RouterLink v-if="showHomeLink" to="/" class="home-link">
        На главную
      </RouterLink>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { api } from '../../api/client'

const router = useRouter()
const title = ref('Проверяем доступ…')
const message = ref('Пожалуйста, подождите.')
const showHomeLink = ref(false)

onMounted(async () => {
  try {
    const response = await api.get('/auth/me')
    const isAdmin = Boolean(response.data?.is_admin)

    if (!isAdmin) {
      title.value = 'Вход выполнен, но прав администратора нет'
      message.value = 'Лендинг остаётся доступным, но админ-панель открыта только администраторам.'
      showHomeLink.value = true

      setTimeout(() => {
        router.replace('/')
      }, 1600)
      return
    }

    title.value = 'Вход выполнен'
    message.value = 'Перенаправляем в админ-панель.'

    setTimeout(() => {
      router.replace('/admin')
    }, 500)
  } catch {
    title.value = 'Авторизация не удалась'
    message.value = 'Попробуйте войти ещё раз.'
    setTimeout(() => {
      router.replace({ path: '/login', query: { reason: 'auth_required' } })
    }, 900)
  }
})
</script>

<style scoped>
.auth-success {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 32px;
  background:
    radial-gradient(circle at top left, rgba(14, 165, 233, 0.2), transparent 32%),
    #0f172a;
  color: #f8fafc;
}

.card {
  width: min(440px, 100%);
  padding: 28px;
  border-radius: 24px;
  background: rgba(15, 23, 42, 0.9);
  border: 1px solid rgba(148, 163, 184, 0.24);
}

h1 {
  margin: 0 0 10px;
  letter-spacing: -0.04em;
}

p {
  margin: 0;
  color: #cbd5e1;
  line-height: 1.55;
}

.home-link {
  display: inline-flex;
  margin-top: 18px;
  color: #93c5fd;
  text-decoration: none;
}

.home-link:hover {
  color: #ffffff;
}
</style>
