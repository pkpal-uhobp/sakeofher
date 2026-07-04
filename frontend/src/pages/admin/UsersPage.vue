<template>
  <AdminLayout>
    <header class="admin-topbar">
      <div class="admin-title">
        <h1>Клиенты</h1>
        <p>Добавление, поиск, редактирование, блокировка, разблокировка и удаление клиентов.</p>
      </div>

      <button class="admin-button" type="button" @click="openCreate">
        + Добавить клиента
      </button>
    </header>

    <section class="admin-card admin-card-body">
      <div class="admin-toolbar">
        <div class="admin-filters">
          <label class="form-field">
            <span>Поиск</span>
            <input
              v-model.trim="filters.query"
              class="admin-input"
              placeholder="telegram_id или username"
              @keyup.enter="loadUsers"
            />
          </label>

          <label class="form-field">
            <span>Статус</span>
            <select v-model="filters.status" class="admin-select" @change="loadUsers">
              <option value="">Все, кроме удалённых</option>
              <option value="active">active</option>
              <option value="blocked">blocked</option>
              <option value="deleted">deleted</option>
            </select>
          </label>
        </div>

        <button class="admin-button" type="button" @click="loadUsers">
          Обновить
        </button>
      </div>

      <p v-if="loading" class="admin-loading">Загружаем клиентов…</p>
      <p v-else-if="error" class="admin-error">{{ error }}</p>

      <div v-else-if="visibleUsers.length" class="admin-table-wrap">
        <table class="admin-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Telegram ID</th>
              <th>Username</th>
              <th>Статус</th>
              <th>Remnawave</th>
              <th></th>
            </tr>
          </thead>

          <tbody>
            <tr v-for="user in visibleUsers" :key="user.id">
              <td>{{ user.id }}</td>
              <td>
                <strong>{{ user.telegram_id }}</strong>
              </td>
              <td>
                <span v-if="user.telegram_username">@{{ user.telegram_username }}</span>
                <span v-else>—</span>
              </td>
              <td>
                <span class="status-pill" :class="user.status || ''">
                  {{ user.status || '—' }}
                </span>
              </td>
              <td>
                <span class="status-pill" :class="user.remna_status || ''">
                  {{ user.remna_status || '—' }}
                </span>
                <br />
                <span v-if="user.remna_username" style="color:#94a3b8">@{{ user.remna_username }}</span>
              </td>
              <td>
                <div class="row-actions">
                  <RouterLink class="admin-link-button" :to="`/panel/users/${user.id}`">
                    Карточка
                  </RouterLink>

                  <button
                    v-if="user.status !== 'blocked'"
                    class="admin-ghost"
                    type="button"
                    @click="changeStatus(user.id, 'block')"
                  >
                    Block
                  </button>

                  <button
                    v-else
                    class="admin-ghost"
                    type="button"
                    @click="changeStatus(user.id, 'unblock')"
                  >
                    Unblock
                  </button>

                  <button class="admin-danger" type="button" @click="changeStatus(user.id, 'delete')">
                    Delete
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-else class="empty-state">
        <strong>Клиенты не найдены.</strong>
        <p style="margin: 8px 0 0; color: #94a3b8">
          Нажми «Добавить клиента», чтобы создать клиента вручную.
        </p>
      </div>
    </section>

    <div v-if="createModal" class="modal-backdrop" @click.self="createModal = false">
      <section class="modal-card">
        <header class="modal-header">
          <h2>Добавить клиента</h2>
          <button class="admin-ghost" type="button" @click="createModal = false">×</button>
        </header>

        <form class="modal-body" @submit.prevent="submitCreate">
          <div class="form-grid">
            <label class="form-field">
              <span>Telegram ID *</span>
              <input
                v-model.number="createForm.telegram_id"
                class="admin-input"
                type="number"
                min="1"
                required
                placeholder="Например: 970706613"
              />
            </label>

            <label class="form-field">
              <span>Username</span>
              <input v-model.trim="createForm.telegram_username" class="admin-input" placeholder="username без @" />
            </label>
          </div>

          <div class="modal-actions">
            <button class="admin-ghost" type="button" @click="createModal = false">Отмена</button>
            <button class="admin-button" type="submit" :disabled="saving">
              {{ saving ? 'Добавляем…' : 'Добавить клиента' }}
            </button>
          </div>
        </form>
      </section>
    </div>
  </AdminLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import AdminLayout from './AdminLayout.vue'
import {
  blockUser,
  createUser,
  deleteUser,
  listUsers,
  unblockUser,
  type User,
} from '../../api/admin'

const router = useRouter()

const users = ref<User[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const filters = reactive({
  query: '',
  status: '',
})
const createModal = ref(false)

const createForm = reactive({
  telegram_id: null as number | null,
  telegram_username: '',
})

const visibleUsers = computed(() => {
  if (filters.status === 'deleted') {
    return users.value
  }

  return users.value.filter((user) => user.status !== 'deleted')
})

onMounted(loadUsers)

async function loadUsers() {
  loading.value = true
  error.value = ''

  try {
    const response = await listUsers({
      query: filters.query,
      status: filters.status,
      limit: 200,
      offset: 0,
    })

    users.value = response.items
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось загрузить клиентов'
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(createForm, {
    telegram_id: null,
    telegram_username: '',
  })

  createModal.value = true
}

async function submitCreate() {
  if (!createForm.telegram_id || createForm.telegram_id <= 0) {
    error.value = 'Укажи корректный Telegram ID'
    return
  }

  saving.value = true
  error.value = ''

  try {
    const user = await createUser({
      telegram_id: createForm.telegram_id,
      telegram_username: createForm.telegram_username || null,
    })

    createModal.value = false
    await loadUsers()

    if (user?.id) {
      router.push(`/panel/users/${user.id}`)
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось добавить клиента'
  } finally {
    saving.value = false
  }
}

async function changeStatus(id: number, action: 'block' | 'unblock' | 'delete') {
  try {
    if (action === 'block') await blockUser(id)
    if (action === 'unblock') await unblockUser(id)
    if (action === 'delete') {
      const ok = window.confirm('Удалить клиента из списка? Если он есть в Remnawave, он будет удалён и там.')
      if (!ok) return

      await deleteUser(id)
      users.value = users.value.filter((user) => user.id !== id)
      return
    }

    await loadUsers()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось изменить статус клиента'
  }
}
</script>

<style src="./admin.css"></style>
