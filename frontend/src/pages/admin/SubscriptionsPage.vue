<template>
  <AdminLayout>
    <header class="admin-topbar">
      <div class="admin-title">
        <h1>Подписки</h1>
        <p>Ручное управление подписками, сроками и лимитами трафика.</p>
      </div>
    </header>

    <section class="admin-card admin-card-body">
      <div class="admin-toolbar">
        <div class="admin-filters">
          <label class="form-field">
            <span>Telegram ID</span>
            <input
              v-model.trim="filters.telegram_id"
              class="admin-input"
              placeholder="970706613"
              @keyup.enter="load"
            />
          </label>

          <label class="form-field">
            <span>Статус</span>
            <select v-model="filters.status" class="admin-select" @change="load">
              <option value="">Все</option>
              <option value="active">active</option>
              <option value="expired">expired</option>
              <option value="cancelled">cancelled</option>
            </select>
          </label>
        </div>

        <button class="admin-button" type="button" @click="load">Обновить</button>
      </div>

      <p v-if="loading" class="admin-loading">Загружаем подписки…</p>
      <p v-else-if="error" class="admin-error">{{ error }}</p>

      <div v-else-if="subscriptions.length" class="admin-table-wrap">
        <table class="admin-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Пользователь</th>
              <th>Тариф</th>
              <th>Статус</th>
              <th>Осталось</th>
              <th>Трафик</th>
              <th>Истекает</th>
              <th></th>
            </tr>
          </thead>

          <tbody>
            <tr v-for="item in subscriptions" :key="item.subscription.id">
              <td>{{ item.subscription.id }}</td>
              <td>
                <strong>{{ item.user.telegram_id }}</strong>
                <br />
                <RouterLink class="admin-kbd" :to="`/panel/users/${item.user.id}`">
                  {{ item.user.telegram_username ? `@${item.user.telegram_username}` : `user #${item.user.id}` }}
                </RouterLink>
              </td>
              <td>{{ item.tariff.title }}</td>
              <td>
                <span class="status-pill" :class="item.subscription.status">
                  {{ item.subscription.status }}
                </span>
                <br />
                <span class="status-pill" :class="item.subscription.period_status || ''" style="margin-top:6px">
                  {{ item.subscription.period_status || '—' }}
                </span>
              </td>
              <td>{{ daysLeft(item.subscription.expires_at) }} дн.</td>
              <td>
                {{ formatBytesGB(item.subscription.traffic_used_bytes) }}
                /
                {{ formatBytesGB(item.subscription.traffic_limit_bytes) }}
              </td>
              <td>{{ formatDate(item.subscription.expires_at) }}</td>
              <td>
                <div class="row-actions">
                  <button class="admin-link-button" type="button" @click="openExtend(item)">
                    Продлить
                  </button>
                  <button class="admin-ghost" type="button" @click="openTraffic(item)">
                    Лимит
                  </button>
                  <button class="admin-ghost" type="button" @click="setAction(item.subscription.id, 'enable')">
                    Enable
                  </button>
                  <button class="admin-ghost" type="button" @click="setAction(item.subscription.id, 'disable')">
                    Disable
                  </button>
                  <button class="admin-danger" type="button" @click="setAction(item.subscription.id, 'cancel')">
                    Cancel
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <p v-else class="empty-state">Подписки не найдены.</p>
    </section>

    <div v-if="extendModal" class="modal-backdrop" @click.self="extendModal = null">
      <section class="modal-card">
        <header class="modal-header">
          <h2>Продлить подписку #{{ extendModal.subscription.id }}</h2>
          <button class="admin-ghost" type="button" @click="extendModal = null">×</button>
        </header>

        <form class="modal-body" @submit.prevent="submitExtend">
          <div class="form-grid">
            <label class="form-field">
              <span>Тариф</span>
              <select v-model.number="extendForm.tariff_id" class="admin-select">
                <option :value="0">Оставить текущий</option>
                <option v-for="tariff in tariffs" :key="tariff.id" :value="tariff.id">
                  {{ tariff.title }} — {{ tariff.duration_days }} дней
                </option>
              </select>
            </label>

            <label class="form-field">
              <span>Дней</span>
              <input v-model.number="extendForm.days" class="admin-input" type="number" min="0" placeholder="по тарифу" />
            </label>
          </div>

          <div class="modal-actions">
            <button class="admin-ghost" type="button" @click="extendModal = null">Отмена</button>
            <button class="admin-button" type="submit">Продлить</button>
          </div>
        </form>
      </section>
    </div>

    <div v-if="trafficModal" class="modal-backdrop" @click.self="trafficModal = null">
      <section class="modal-card">
        <header class="modal-header">
          <h2>Изменить лимит #{{ trafficModal.subscription.id }}</h2>
          <button class="admin-ghost" type="button" @click="trafficModal = null">×</button>
        </header>

        <form class="modal-body" @submit.prevent="submitTraffic">
          <label class="form-field">
            <span>Новый лимит, ГБ</span>
            <input v-model.number="trafficForm.traffic_limit_gb" class="admin-input" type="number" min="1" required />
          </label>

          <div class="modal-actions">
            <button class="admin-ghost" type="button" @click="trafficModal = null">Отмена</button>
            <button class="admin-button" type="submit">Сохранить</button>
          </div>
        </form>
      </section>
    </div>
  </AdminLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import AdminLayout from './AdminLayout.vue'
import {
  cancelSubscription,
  disableSubscription,
  enableSubscription,
  extendSubscription,
  listSubscriptions,
  listTariffs,
  updateTrafficLimit,
  type PublicSubscription,
  type Tariff,
} from '../../api/admin'
import { bytesToGB, daysLeft, formatBytesGB, formatDate } from '../../api/client'

const subscriptions = ref<PublicSubscription[]>([])
const tariffs = ref<Tariff[]>([])
const loading = ref(false)
const error = ref('')

const filters = reactive({
  telegram_id: '',
  status: '',
})

const extendModal = ref<PublicSubscription | null>(null)
const trafficModal = ref<PublicSubscription | null>(null)

const extendForm = reactive({
  tariff_id: 0,
  days: 0,
})

const trafficForm = reactive({
  traffic_limit_gb: 500,
})

onMounted(async () => {
  await Promise.all([load(), loadTariffs()])
})

async function load() {
  loading.value = true
  error.value = ''

  try {
    const response = await listSubscriptions({
      telegram_id: filters.telegram_id,
      status: filters.status,
      limit: 100,
    })

    subscriptions.value = response.items
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось загрузить подписки'
  } finally {
    loading.value = false
  }
}

async function loadTariffs() {
  try {
    tariffs.value = await listTariffs()
  } catch {
    tariffs.value = []
  }
}

function openExtend(item: PublicSubscription) {
  extendModal.value = item
  extendForm.tariff_id = 0
  extendForm.days = 0
}

function openTraffic(item: PublicSubscription) {
  trafficModal.value = item
  trafficForm.traffic_limit_gb = bytesToGB(item.subscription.traffic_limit_bytes)
}

async function submitExtend() {
  if (!extendModal.value) return

  try {
    await extendSubscription(extendModal.value.subscription.id, {
      tariff_id: extendForm.tariff_id || undefined,
      days: extendForm.days || undefined,
    })

    extendModal.value = null
    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось продлить подписку'
  }
}

async function submitTraffic() {
  if (!trafficModal.value) return

  try {
    await updateTrafficLimit(trafficModal.value.subscription.id, {
      traffic_limit_gb: trafficForm.traffic_limit_gb,
    })

    trafficModal.value = null
    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось изменить лимит'
  }
}

async function setAction(id: number, action: 'enable' | 'disable' | 'cancel') {
  try {
    if (action === 'enable') await enableSubscription(id)
    if (action === 'disable') await disableSubscription(id)
    if (action === 'cancel') {
      const ok = window.confirm('Отменить подписку?')
      if (!ok) return
      await cancelSubscription(id)
    }

    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось изменить подписку'
  }
}
</script>

<style src="./admin.css"></style>
