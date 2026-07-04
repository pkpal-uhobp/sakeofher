<template>
  <main class="login-page">
    <section class="login-card">
      <div class="badge">SakeOfHer Admin</div>

      <h1>Вход в админ-панель</h1>

      <p>
        Лендинг и страница подписки доступны всем. Для управления проектом войдите
        с логином и паролем администратора.
      </p>

      <p v-if="reasonText" class="reason">
        {{ reasonText }}
      </p>

      <p v-if="errorText" class="reason">
        {{ errorText }}
      </p>

      <form @submit.prevent="login">
        <label>
          Логин
          <input v-model="username" type="text" autocomplete="username" required />
        </label>

        <label>
          Пароль
          <input v-model="password" type="password" autocomplete="current-password" required />
        </label>

        <button type="submit" :disabled="loading">
          {{ loading ? 'Входим…' : 'Войти' }}
        </button>
      </form>

      <RouterLink class="back-link" to="/">
        Вернуться на лендинг
      </RouterLink>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { api } from '../../api/client'

const route = useRoute()
const router = useRouter()

const username = ref('')
const password = ref('')
const loading = ref(false)
const errorText = ref('')

const reasonText = computed(() => {
  if (route.query.reason === 'admin_required') {
    return 'У вашей учётной записи нет прав администратора.'
  }

  if (route.query.reason === 'auth_required') {
    return 'Сначала нужно авторизоваться.'
  }

  return ''
})

async function login() {
  errorText.value = ''
  loading.value = true

  try {
    await api.post('/auth/login', {
      username: username.value.trim(),
      password: password.value,
    })

    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/admin'
    await router.replace(redirect)
  } catch (e: any) {
    errorText.value = e?.response?.data?.error || 'Неверный логин или пароль.'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 32px;
  background:
    radial-gradient(circle at top left, rgba(14, 165, 233, 0.24), transparent 32%),
    radial-gradient(circle at bottom right, rgba(37, 99, 235, 0.16), transparent 34%),
    linear-gradient(135deg, #020617, #080b12);
  color: #f8fafc;
}

.login-card {
  width: min(460px, 100%);
  padding: 32px;
  border-radius: 28px;
  background: rgba(15, 23, 42, 0.86);
  border: 1px solid rgba(148, 163, 184, 0.24);
  box-shadow:
    0 24px 80px rgba(0, 0, 0, 0.35),
    inset 0 1px 0 rgba(255, 255, 255, 0.06);
}

.badge {
  display: inline-flex;
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(59, 130, 246, 0.18);
  color: #bfdbfe;
  font-size: 13px;
  margin-bottom: 18px;
}

h1 {
  margin: 0 0 12px;
  font-size: 32px;
  line-height: 1.1;
  letter-spacing: -0.05em;
}

p {
  margin: 0 0 20px;
  color: #cbd5e1;
  line-height: 1.6;
}

.reason {
  padding: 12px 14px;
  border-radius: 16px;
  background: rgba(248, 113, 113, 0.12);
  color: #fecaca;
}

form {
  display: grid;
  gap: 16px;
}

label {
  display: grid;
  gap: 8px;
  color: #cbd5e1;
  font-size: 14px;
}

input {
  width: 100%;
  border: 1px solid rgba(148, 163, 184, 0.28);
  border-radius: 16px;
  padding: 14px 16px;
  font-size: 16px;
  color: #f8fafc;
  background: rgba(2, 6, 23, 0.72);
}

input:focus {
  outline: none;
  border-color: rgba(96, 165, 250, 0.8);
}

button {
  width: 100%;
  border: 0;
  border-radius: 18px;
  padding: 14px 18px;
  font-size: 16px;
  font-weight: 800;
  color: #ffffff;
  background: linear-gradient(135deg, #0786ff, #0057e5);
  cursor: pointer;
  box-shadow: 0 18px 46px rgba(0, 102, 255, 0.3);
}

button:disabled {
  opacity: 0.7;
  cursor: wait;
}

button:hover:not(:disabled) {
  filter: brightness(1.08);
}

.back-link {
  display: block;
  margin-top: 18px;
  text-align: center;
  color: #93c5fd;
  text-decoration: none;
}

.back-link:hover {
  color: #ffffff;
}
</style>
