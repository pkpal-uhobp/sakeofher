<template>
  <main class="auth-success">
    <section class="card">
      <h1>Вход через Telegram отключён</h1>
      <p>
        Используйте вход по логину и паролю из переменных окружения.
      </p>

      <RouterLink to="/login" class="home-link">
        Перейти ко входу
      </RouterLink>
    </section>
  </main>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
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
