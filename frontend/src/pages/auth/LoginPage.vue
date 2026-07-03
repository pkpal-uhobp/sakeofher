<script setup lang="ts">
const authStartURL = '/api/v1/auth/telegram/oauth/start'
</script>

<template>
  <main class="page">
    <section class="card">
      <p class="badge">SakeOfHer Admin</p>
      <h1>Вход через Telegram</h1>
      <p class="subtitle">
        Авторизация проходит через Telegram OAuth 2.0 / OpenID Connect. После входа backend проверит Telegram ID в таблице admins.
      </p>
      <a class="telegram" :href="authStartURL">Войти через Telegram</a>
      <p class="note">
        Перед использованием добавь домен и callback URL в BotFather → Bot Settings → Web Login.
      </p>
    </section>
  </main>
</template>

<style scoped>
.page { max-width: 760px; margin: 0 auto; padding: 64px 20px; }
.card { background: #111827; border: 1px solid #263244; border-radius: 24px; padding: 32px; box-shadow: 0 24px 80px rgba(0,0,0,.25); }
.badge { display: inline-block; padding: 8px 12px; border: 1px solid #334155; border-radius: 999px; color: #93c5fd; }
h1 { font-size: 42px; margin: 18px 0; }
.subtitle, .note { color: #94a3b8; line-height: 1.6; }
.telegram { display: inline-block; margin-top: 18px; border: 0; border-radius: 16px; background: #60a5fa; color: #020617; padding: 15px 18px; font-weight: 800; cursor: pointer; text-decoration: none; }
.note { margin-top: 20px; font-size: 14px; }
</style>
