<template>
  <AdminLayout>
    <header class="admin-topbar">
      <div class="admin-title">
        <h1>Клиент #{{ route.params.id }}</h1>
        <p>Данные клиента, Remnawave squads и ручная выдача подписки.</p>
      </div>

      <RouterLink class="admin-link-button" to="/panel/users">
        ← Назад
      </RouterLink>
    </header>

    <p v-if="loading" class="admin-loading">Загружаем клиента…</p>
    <p v-else-if="error" class="admin-error">{{ error }}</p>
    <p v-if="success" class="empty-state" style="margin-bottom: 18px; color: #86efac">
      {{ success }}
    </p>

    <div v-if="!loading && user" class="admin-grid cols-2">
      <section class="admin-card admin-card-body">
        <h2>Данные клиента</h2>

        <form class="form-grid" style="margin-top: 16px" @submit.prevent="saveUser">
          <label class="form-field">
            <span>Telegram ID</span>
            <input class="admin-input" :value="user.telegram_id" readonly />
          </label>

          <label class="form-field">
            <span>Username</span>
            <input v-model.trim="form.telegram_username" class="admin-input" placeholder="username без @" />
          </label>

          <label class="form-field">
            <span>Статус</span>
            <select v-model="form.status" class="admin-select">
              <option value="active">active</option>
              <option value="blocked">blocked</option>
              <option value="deleted">deleted</option>
            </select>
          </label>

          <div>
            <button class="admin-button" type="submit" :disabled="savingUser">
              {{ savingUser ? 'Сохраняем…' : 'Сохранить клиента' }}
            </button>
          </div>
        </form>

        <dl class="detail-list" style="margin-top: 20px">
          <div class="detail-item">
            <dt>Remnawave</dt>
            <dd>
              <span class="status-pill" :class="user.remna_status || ''">
                {{ user.remna_status || '—' }}
              </span>
            </dd>
          </div>

          <div class="detail-item">
            <dt>Remna UUID</dt>
            <dd>{{ user.remna_uuid || '—' }}</dd>
          </div>

          <div class="detail-item">
            <dt>Ссылка подписки</dt>
            <dd>
              <div class="copy-box">
                <input class="admin-input" :value="publicSubscriptionURL" readonly />
                <button class="admin-ghost" type="button" @click="copy(publicSubscriptionURL)">
                  Copy
                </button>
                <a class="admin-link-button" :href="publicSubscriptionURL" target="_blank" rel="noreferrer">
                  Открыть
                </a>
              </div>
              <p class="hint-text">
                Это публичная ссылка сайта. В браузере открывает красивую страницу,
                для приложений отдаёт полную Remnawave Base64-подписку.
                Принудительно Base64: <code>?format=base64</code>.
              </p>
            </dd>
          </div>
        </dl>

        <div class="row-actions" style="margin-top: 18px">
          <button class="admin-ghost" type="button" @click="statusAction('block')">
            Заблокировать
          </button>

          <button class="admin-ghost" type="button" @click="statusAction('unblock')">
            Разблокировать
          </button>

          <button class="admin-danger" type="button" @click="statusAction('delete')">
            Удалить
          </button>
        </div>
      </section>

      <section class="admin-card admin-card-body">
        <h2>Выдать подписку вручную</h2>

        <p style="color:#94a3b8">
          Выбери тариф, лимит и squads Remnawave, в которые нужно добавить пользователя.
          HWID/device limit при создании и обновлении пользователя сбрасывается в 0.
        </p>

        <form class="form-grid" @submit.prevent="createSubscription">
          <label class="form-field">
            <span>Тариф</span>
            <select v-model.number="manual.tariff_id" class="admin-select" required>
              <option :value="0">Выберите тариф</option>
              <option v-for="tariff in activeTariffs" :key="tariff.id" :value="tariff.id">
                {{ tariff.title }} — {{ tariff.duration_days }} дней
              </option>
            </select>
          </label>

          <label class="form-field">
            <span>Лимит, ГБ</span>
            <input
              v-model.number="manual.traffic_limit_gb"
              class="admin-input"
              type="number"
              min="1"
              required
              placeholder="500"
            />
          </label>

          <section class="form-field wide">
            <span>Remnawave squads</span>

            <div class="squad-toolbar">
              <button class="admin-ghost" type="button" @click="selectAllSquads">
                Выбрать все
              </button>
              <button class="admin-ghost" type="button" @click="clearSquads">
                Снять все
              </button>
              <button class="admin-ghost" type="button" @click="loadSquads">
                Обновить squads
              </button>
            </div>

            <p v-if="squadsLoading" class="admin-loading">Загружаем squads…</p>

            <div v-else-if="squads.length" class="squad-list">
              <label v-for="squad in squads" :key="squad.uuid" class="squad-item">
                <input
                  v-model="manual.active_internal_squads"
                  type="checkbox"
                  :value="squad.uuid"
                />
                <span>
                  <strong>{{ squad.name }}</strong>
                  <small>{{ squad.uuid }}</small>
                </span>
              </label>
            </div>

            <p v-else class="empty-state">
              Squads не загружены. Проверь REMNAWAVE_BASE_URL и REMNAWAVE_API_TOKEN.
              Если список оставить пустым, backend попробует включить все squads автоматически.
            </p>
          </section>

          <div class="wide">
            <button class="admin-button" type="submit" :disabled="manualLoading || !manual.tariff_id">
              {{ manualLoading ? 'Создаём в Remnawave…' : 'Создать в Remnawave и выдать подписку' }}
            </button>
          </div>
        </form>
      </section>
    </div>
  </AdminLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import AdminLayout from './AdminLayout.vue'
import {
  blockUser,
  createManualSubscription,
  deleteUser,
  getUser,
  listRemnawaveInternalSquads,
  listTariffs,
  unblockUser,
  updateUser,
  type RemnaInternalSquad,
  type Tariff,
  type User,
  type UserStatus,
} from '../../api/admin'
import { bytesToGB } from '../../api/client'

const route = useRoute()
const router = useRouter()

const user = ref<User | null>(null)
const tariffs = ref<Tariff[]>([])
const squads = ref<RemnaInternalSquad[]>([])
const loading = ref(false)
const squadsLoading = ref(false)
const manualLoading = ref(false)
const savingUser = ref(false)
const error = ref('')
const success = ref('')

const subscriptionSecret = import.meta.env.VITE_SUBSCRIPTION_PATH_SECRET || 'L0mENeiofHjdxC57'

const form = reactive({
  telegram_username: '',
  status: 'active' as UserStatus,
})

const manual = reactive({
  tariff_id: 0,
  traffic_limit_gb: 500,
  active_internal_squads: [] as string[],
})

const activeTariffs = computed(() => tariffs.value.filter((tariff) => tariff.is_active))

const publicSubscriptionURL = computed(() => {
  if (!user.value) return ''

  return `${window.location.origin}/${encodeURIComponent(subscriptionSecret)}/sub/${encodeURIComponent(user.value.telegram_id)}`
})

onMounted(async () => {
  await Promise.all([loadUser(), loadTariffs(), loadSquads()])
})

async function loadUser() {
  loading.value = true
  error.value = ''

  try {
    user.value = await getUser(Number(route.params.id))
    fillForm()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось загрузить клиента'
  } finally {
    loading.value = false
  }
}

function fillForm() {
  if (!user.value) return

  form.telegram_username = user.value.telegram_username || ''
  form.status = (user.value.status || 'active') as UserStatus
}

async function loadTariffs() {
  try {
    tariffs.value = await listTariffs()
    const firstActive = tariffs.value.find((item) => item.is_active)
    if (firstActive) {
      manual.tariff_id = firstActive.id
      manual.traffic_limit_gb = bytesToGB(firstActive.traffic_limit_bytes) || 500
    }
  } catch {
    tariffs.value = []
  }
}

async function loadSquads() {
  squadsLoading.value = true

  try {
    squads.value = await listRemnawaveInternalSquads()
    selectAllSquads()
  } catch {
    squads.value = []
    manual.active_internal_squads = []
  } finally {
    squadsLoading.value = false
  }
}

function selectAllSquads() {
  manual.active_internal_squads = squads.value.map((squad) => squad.uuid)
}

function clearSquads() {
  manual.active_internal_squads = []
}

async function saveUser() {
  if (!user.value) return

  savingUser.value = true
  error.value = ''
  success.value = ''

  try {
    user.value = await updateUser(user.value.id, {
      telegram_username: form.telegram_username || null,
      status: form.status,
    })
    fillForm()
    success.value = 'Клиент сохранён.'
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось сохранить клиента'
  } finally {
    savingUser.value = false
  }
}

async function statusAction(action: 'block' | 'unblock' | 'delete') {
  if (!user.value) return

  try {
    if (action === 'block') await blockUser(user.value.id)
    if (action === 'unblock') await unblockUser(user.value.id)
    if (action === 'delete') {
      const ok = window.confirm('Удалить клиента? Если он есть в Remnawave, он будет удалён и там.')
      if (!ok) return

      await deleteUser(user.value.id)
      router.replace('/panel/users')
      return
    }

    await loadUser()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось изменить статус'
  }
}

async function createSubscription() {
  if (!user.value) return

  manualLoading.value = true
  error.value = ''
  success.value = ''

  try {
    await createManualSubscription({
      user_id: user.value.id,
      tariff_id: manual.tariff_id,
      traffic_limit_gb: manual.traffic_limit_gb,
      active_internal_squads: manual.active_internal_squads,
    })

    await loadUser()
    success.value = 'Пользователь создан/синхронизирован в Remnawave, подписка выдана. HWID limit сброшен в 0.'
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось создать пользователя в Remnawave или выдать подписку'
  } finally {
    manualLoading.value = false
  }
}

async function copy(value: string | null | undefined) {
  if (!value) return
  await navigator.clipboard.writeText(value)
}
</script>

<style src="./admin.css"></style>
<style scoped>
.hint-text {
  margin: 8px 0 0;
  color: #94a3b8;
  font-size: 13px;
}

.hint-text code {
  color: #7dd3fc;
}
</style>
