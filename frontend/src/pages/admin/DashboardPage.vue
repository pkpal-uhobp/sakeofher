<template>
  <AdminLayout>
    <header class="admin-topbar">
      <div class="admin-title">
        <h1>Панель управления</h1>
        <p>Управление пользователями, подписками и тарифами проекта.</p>
      </div>

      <RouterLink class="admin-button" to="/panel/users">
        Открыть пользователей
      </RouterLink>
    </header>

    <section class="admin-grid cols-3">
      <article class="admin-card metric-card">
        <p class="metric-label">Пользователи</p>
        <p class="metric-value">{{ stats.users }}</p>
        <p class="metric-note">Всего в базе</p>
      </article>

      <article class="admin-card metric-card">
        <p class="metric-label">Подписки</p>
        <p class="metric-value">{{ stats.subscriptions }}</p>
        <p class="metric-note">По текущему фильтру API</p>
      </article>

      <article class="admin-card metric-card">
        <p class="metric-label">Тарифы</p>
        <p class="metric-value">{{ stats.tariffs }}</p>
        <p class="metric-note">Активные и отключённые</p>
      </article>
    </section>

    <section class="admin-grid cols-3" style="margin-top: 18px">
      <RouterLink class="admin-card admin-card-body" to="/panel/users">
        <h2>Пользователи</h2>
        <p>Поиск, блокировка, разблокировка, просмотр Remnawave-данных.</p>
      </RouterLink>

      <RouterLink class="admin-card admin-card-body" to="/panel/subscriptions">
        <h2>Подписки</h2>
        <p>Ручная выдача, продление, изменение лимита, отключение и отмена.</p>
      </RouterLink>

      <RouterLink class="admin-card admin-card-body" to="/panel/tariffs">
        <h2>Тарифы</h2>
        <p>Создание, редактирование, включение и отключение тарифов.</p>
      </RouterLink>
    </section>

    <p v-if="error" class="admin-error" style="margin-top: 18px">
      {{ error }}
    </p>
  </AdminLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import AdminLayout from './AdminLayout.vue'
import { listSubscriptions, listTariffs, listUsers } from '../../api/admin'

const stats = reactive({
  users: '—',
  subscriptions: '—',
  tariffs: '—',
})

const error = ref('')

onMounted(loadStats)

async function loadStats() {
  error.value = ''

  try {
    const [users, subscriptions, tariffs] = await Promise.all([
      listUsers({ limit: 1 }),
      listSubscriptions({ limit: 1 }),
      listTariffs(),
    ])

    stats.users = String(users.total ?? users.items.length)
    stats.subscriptions = String(subscriptions.total ?? subscriptions.items.length)
    stats.tariffs = String(tariffs.length)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось загрузить статистику'
  }
}
</script>

<style scoped>
a.admin-card {
  color: inherit;
  text-decoration: none;
}

a.admin-card h2 {
  margin: 0 0 8px;
  color: #fff;
  letter-spacing: -0.04em;
}

a.admin-card p {
  margin: 0;
  color: #94a3b8;
  line-height: 1.5;
}
</style>
