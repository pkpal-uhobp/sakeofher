<template>
  <main class="login-page">
    <div class="login-glow glow-a"></div>
    <div class="login-glow glow-b"></div>

    <section class="login-card">
      <RouterLink class="login-brand" to="/">
        <span></span>
        SakeOfHer
      </RouterLink>

      <div class="badge">Панель управления</div>

      <h1>Вход в аккаунт</h1>

      <p>
        Введите логин и пароль из файла <b>.env</b>. После входа вы попадёте
        в закрытую панель управления.
      </p>

      <form class="login-form" @submit.prevent="submit">
        <label>
          <span>Логин</span>
          <input
            v-model.trim="form.login"
            autocomplete="username"
            placeholder="Например: admin"
            required
          />
          <small>Берётся из переменной ADMIN_LOGIN</small>
        </label>

        <label>
          <span>Пароль</span>
          <input
            v-model="form.password"
            autocomplete="current-password"
            placeholder="Введите пароль администратора"
            type="password"
            required
          />
          <small>Берётся из переменной ADMIN_PASSWORD</small>
        </label>

        <p v-if="error" class="reason">
          {{ error }}
        </p>

        <button type="submit" :disabled="loading">
          {{ loading ? 'Проверяем…' : 'Войти в панель' }}
        </button>
      </form>

      <RouterLink class="back-link" to="/">
        Вернуться на лендинг
      </RouterLink>
    </section>
  </main>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { api } from '../../api/client'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const error = ref('')

const form = reactive({
  login: '',
  password: '',
})

async function submit() {
  loading.value = true
  error.value = ''

  try {
    await api.post('/auth/login', {
      login: form.login,
      password: form.password,
    })

    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/panel'
    router.replace(redirect)
  } catch {
    error.value = 'Неверный логин или пароль. Проверьте ADMIN_LOGIN и ADMIN_PASSWORD в .env.'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  position: relative;
  min-height: 100vh;
  display: grid;
  place-items: center;
  overflow: hidden;
  padding: 32px;
  background:
    linear-gradient(rgba(255, 255, 255, 0.025) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.025) 1px, transparent 1px),
    radial-gradient(circle at top left, rgba(14, 165, 233, 0.24), transparent 32%),
    radial-gradient(circle at bottom right, rgba(37, 99, 235, 0.16), transparent 34%),
    linear-gradient(135deg, #020617, #080b12);
  background-size: 54px 54px, 54px 54px, 100% 100%, 100% 100%, 100% 100%;
  color: #f8fafc;
}

.login-glow {
  position: fixed;
  border-radius: 999px;
  filter: blur(95px);
  pointer-events: none;
}

.glow-a {
  top: -220px;
  right: 18%;
  width: 430px;
  height: 430px;
  background: rgba(14, 165, 233, 0.22);
}

.glow-b {
  bottom: -250px;
  left: 12%;
  width: 440px;
  height: 440px;
  background: rgba(37, 99, 235, 0.18);
}

.login-card {
  position: relative;
  z-index: 1;
  width: min(500px, 100%);
  padding: 34px;
  border-radius: 30px;
  background:
    linear-gradient(180deg, rgba(15, 23, 42, 0.9), rgba(2, 6, 23, 0.78)),
    radial-gradient(circle at top left, rgba(14, 165, 233, 0.16), transparent 44%);
  border: 1px solid rgba(148, 163, 184, 0.24);
  box-shadow:
    0 26px 90px rgba(0, 0, 0, 0.42),
    inset 0 1px 0 rgba(255, 255, 255, 0.06);
}

.login-brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 22px;
  color: #ffffff;
  font-size: 23px;
  font-weight: 900;
  letter-spacing: -0.04em;
  text-decoration: none;
}

.login-brand span {
  width: 11px;
  height: 11px;
  border-radius: 50%;
  background: #0ea5e9;
  box-shadow: 0 0 18px rgba(14, 165, 233, 0.9);
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
  font-size: 38px;
  line-height: 1.05;
  letter-spacing: -0.065em;
}

p {
  margin: 0 0 22px;
  color: #cbd5e1;
  line-height: 1.6;
}

.login-form {
  display: grid;
  gap: 16px;
}

label {
  display: grid;
  gap: 8px;
}

label span {
  color: #e2e8f0;
  font-size: 14px;
  font-weight: 800;
}

label small {
  color: #64748b;
  font-size: 12px;
}

input {
  width: 100%;
  min-height: 52px;
  padding: 13px 15px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 18px;
  outline: none;
  color: #f8fafc;
  background: rgba(2, 6, 23, 0.62);
}

input::placeholder {
  color: rgba(148, 163, 184, 0.62);
}

input:focus {
  border-color: rgba(56, 189, 248, 0.72);
  box-shadow: 0 0 0 4px rgba(14, 165, 233, 0.08);
}

.reason {
  margin: 0;
  padding: 12px 14px;
  border-radius: 16px;
  background: rgba(248, 113, 113, 0.12);
  color: #fecaca;
}

button {
  width: 100%;
  border: 0;
  border-radius: 18px;
  padding: 15px 18px;
  font-size: 16px;
  font-weight: 900;
  color: #ffffff;
  background: linear-gradient(135deg, #0786ff, #0057e5);
  cursor: pointer;
  box-shadow: 0 18px 46px rgba(0, 102, 255, 0.3);
}

button:hover {
  filter: brightness(1.08);
}

button:disabled {
  cursor: not-allowed;
  opacity: 0.7;
}

.back-link {
  display: block;
  margin-top: 20px;
  text-align: center;
  color: #93c5fd;
  text-decoration: none;
}

.back-link:hover {
  color: #ffffff;
}
</style>
