<template>
  <main class="subscription-page">
    <div class="page-grid"></div>
    <div class="glow glow-a"></div>
    <div class="glow glow-b"></div>

    <section class="subscription-shell">
      <header class="topbar">
        <RouterLink class="brand" to="/">
          <span></span>
          SakeOfHer
        </RouterLink>

        <button class="ghost-button" type="button" @click="load">
          Обновить
        </button>
      </header>

      <section v-if="loading" class="state-card">
        <span class="pulse"></span>
        <h1>Загружаем подписку</h1>
        <p>Проверяем статус, срок действия и лимиты трафика.</p>
      </section>

      <section v-else-if="error" class="state-card error">
        <span class="error-icon">!</span>
        <h1>Подписка не найдена</h1>
        <p>{{ error }}</p>
        <RouterLink class="primary-button" to="/">На главную</RouterLink>
      </section>

      <template v-else-if="data">
        <section class="hero-card">
          <div class="hero-content">
            <p class="eyebrow">Личный доступ</p>
            <h1>Ваша подписка активна и готова к использованию</h1>
            <p>
              Один URL работает в двух режимах: в браузере открывает эту страницу,
              а в приложениях отдаёт полную Remnawave Base64-подписку.
            </p>

            <div class="hero-actions">
              <button class="primary-button" type="button" @click="copyLink">
                Скопировать ссылку
              </button>

              <button class="secondary-button" type="button" @click="load">
                Обновить статус
              </button>
            </div>
          </div>

          <div class="status-orb" :class="subscription.status">
            <div>
              <strong>{{ statusLabel(subscription.status) }}</strong>
              <span>{{ daysLeft(subscription.expires_at) }} дней осталось</span>
            </div>
          </div>
        </section>

        <section class="cards-grid">
          <article class="glass-card">
            <span class="card-label">Пользователь</span>
            <strong>{{ userLabel }}</strong>
            <p>{{ data.user.telegram_id }}</p>
          </article>

          <article class="glass-card">
            <span class="card-label">Тариф</span>
            <strong>{{ data.tariff.title }}</strong>
            <p>{{ formatRub(data.tariff.price_rub) }} · {{ formatBytesGB(data.tariff.traffic_limit_bytes) }}</p>
          </article>

          <article class="glass-card">
            <span class="card-label">Истекает</span>
            <strong>{{ formatDate(subscription.expires_at) }}</strong>
            <p>{{ subscription.period_status }}</p>
          </article>
        </section>

        <section class="main-grid">
          <article class="panel-card">
            <div class="panel-heading">
              <div>
                <span class="card-label">Трафик</span>
                <h2>{{ usedGB }} / {{ limitGB }} ГБ</h2>
              </div>

              <strong>{{ trafficPercent }}%</strong>
            </div>

            <div class="progress">
              <span :style="{ width: `${trafficPercent}%` }"></span>
            </div>

            <div class="meta-row">
              <span>Использовано {{ formatBytesGB(subscription.traffic_used_bytes) }}</span>
              <span>Лимит {{ formatBytesGB(subscription.traffic_limit_bytes) }}</span>
            </div>
          </article>

          <article class="panel-card">
            <div class="panel-heading">
              <div>
                <span class="card-label">Срок действия</span>
                <h2>{{ daysLeft(subscription.expires_at) }} дней</h2>
              </div>

              <strong>{{ termPercent }}%</strong>
            </div>

            <div class="progress">
              <span :style="{ width: `${termPercent}%` }"></span>
            </div>

            <div class="meta-row">
              <span>Начало {{ formatDate(subscription.started_at || subscription.created_at) }}</span>
              <span>Конец {{ formatDate(subscription.expires_at) }}</span>
            </div>
          </article>
        </section>

        <section class="link-card">
          <div>
            <span class="card-label">Ссылка подписки</span>
            <h2>Один URL для страницы и подключения</h2>
          </div>

          <div class="copy-box">
            <input :value="subscriptionLink" readonly />
            <button class="primary-button" type="button" @click="copyLink">
              Copy
            </button>
          </div>

          <p class="copy-note">
            В браузере эта ссылка открывает страницу. Для приложений она отдаёт полную
            Base64-подписку из Remnawave. Принудительно:
            <code>?format=base64</code>.
          </p>

          <p v-if="copied" class="copy-success">Ссылка скопирована.</p>
        </section>

        <section class="apps-card">
          <div class="apps-heading">
            <span class="card-label">Приложения</span>
            <h2>Выберите устройство</h2>
            <p>Пул приложений ограничен списком как на Remnawave subscription page.</p>
          </div>

          <div class="device-tabs">
            <button
              v-for="device in devices"
              :key="device.key"
              type="button"
              :class="{ active: currentDevice === device.key }"
              @click="currentDevice = device.key"
            >
              {{ device.label }}
            </button>
          </div>

          <div class="apps-grid">
            <article v-for="app in filteredApps" :key="app.name + app.device" class="app-card">
              <div>
                <div class="app-title">
                  <strong>{{ app.name }}</strong>
                  <span>{{ coreLabel(app.core) }}</span>
                  <em v-if="app.hwid">HWID</em>
                </div>
                <p>{{ app.description }}</p>
              </div>

              <div class="app-actions">
                <a class="secondary-button" :href="app.download" target="_blank" rel="noreferrer">
                  Скачать
                </a>

                <a class="primary-button" :href="app.open">
                  Открыть
                </a>
              </div>
            </article>
          </div>
        </section>
      </template>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { api, bytesToGB, formatBytesGB, formatDate, formatRub, daysLeft } from '../../api/client'
import { SUBSCRIPTION_DEVICES, buildSubscriptionApps } from './subscriptionClients'
import type { ClientCore } from './subscriptionClients'

interface User {
  telegram_id: number
  telegram_username?: string | null
  subscription_url?: string | null
}

interface Tariff {
  title: string
  traffic_limit_bytes: number
  price_rub: number
}

interface Subscription {
  id: number
  status: string
  period_status: string
  expires_at: string
  started_at?: string
  created_at?: string
  traffic_used_bytes: number
  traffic_limit_bytes: number
}

interface PublicSubscription {
  user: User
  tariff: Tariff
  subscription: Subscription
  subscription_url?: string | null
}

const route = useRoute()
const loading = ref(false)
const error = ref('')
const data = ref<PublicSubscription | null>(null)
const copied = ref(false)
const currentDevice = ref('android')

const devices = SUBSCRIPTION_DEVICES

const subscription = computed(() => data.value!.subscription)

const userLabel = computed(() => {
  const username = data.value?.user.telegram_username
  return username ? `@${username}` : `ID ${data.value?.user.telegram_id || '—'}`
})

const directSubscriptionPath = computed(() => {
  const secret = String(route.params.secret || '')
  const telegramId = String(route.params.telegramId || data.value?.user.telegram_id || '')

  if (secret && telegramId) {
    return `${window.location.origin}/${encodeURIComponent(secret)}/sub/${encodeURIComponent(telegramId)}`
  }

  return ''
})

const subscriptionLink = computed(() => {
  return directSubscriptionPath.value || data.value?.subscription_url || data.value?.user.subscription_url || window.location.href
})

const appSubscriptionLink = computed(() => {
  const glue = subscriptionLink.value.includes('?') ? '&' : '?'
  return `${subscriptionLink.value}${glue}format=base64`
})

const usedGB = computed(() => bytesToGB(subscription.value.traffic_used_bytes))
const limitGB = computed(() => bytesToGB(subscription.value.traffic_limit_bytes))

const trafficPercent = computed(() => {
  const limit = subscription.value.traffic_limit_bytes
  if (!limit || limit <= 0) return 0

  return Math.min(100, Math.round((subscription.value.traffic_used_bytes / limit) * 100))
})

const termPercent = computed(() => {
  const started = new Date(subscription.value.started_at || subscription.value.created_at || Date.now()).getTime()
  const expires = new Date(subscription.value.expires_at).getTime()
  const now = Date.now()

  if (!Number.isFinite(started) || !Number.isFinite(expires) || expires <= started) return 0

  return Math.max(0, Math.min(100, Math.round(((now - started) / (expires - started)) * 100)))
})

const apps = computed(() =>
  buildSubscriptionApps({
    subscriptionURL: appSubscriptionLink.value,
  }),
)

const filteredApps = computed(() => apps.value.filter((app) => app.device === currentDevice.value))

onMounted(load)

async function load() {
  loading.value = true
  error.value = ''
  copied.value = false

  try {
    const token = String(route.params.token || '')
    const secret = String(route.params.secret || '')
    const telegramId = String(route.params.telegramId || '')

    const url =
      secret && telegramId
        ? `/subscriptions/path/${encodeURIComponent(secret)}/telegram/${encodeURIComponent(telegramId)}`
        : `/subscriptions/public/${encodeURIComponent(token)}`

    const response = await api.get<PublicSubscription>(url)
    data.value = response.data
  } catch {
    data.value = null
    error.value = 'Проверьте ссылку или обратитесь к администратору.'
  } finally {
    loading.value = false
  }
}

async function copyLink() {
  await navigator.clipboard.writeText(subscriptionLink.value)
  copied.value = true
}

function statusLabel(status: string): string {
  switch (status) {
    case 'active':
      return 'Активна'
    case 'expired':
      return 'Истекла'
    case 'cancelled':
      return 'Отменена'
    default:
      return status
  }
}

function coreLabel(core: ClientCore): string {
  switch (core) {
    case 'xray':
      return 'Xray'
    case 'mihomo':
      return 'Mihomo'
    case 'singbox':
      return 'Sing-box'
    default:
      return core
  }
}
</script>

<style scoped>
.subscription-page {
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  color: #f8fafc;
  background:
    radial-gradient(circle at 75% 10%, rgba(14, 165, 233, 0.24), transparent 32%),
    radial-gradient(circle at 10% 90%, rgba(37, 99, 235, 0.18), transparent 30%),
    linear-gradient(135deg, #020617 0%, #050816 45%, #02040a 100%);
  font-family:
    Inter,
    ui-sans-serif,
    system-ui,
    -apple-system,
    BlinkMacSystemFont,
    "Segoe UI",
    sans-serif;
}

.page-grid {
  position: fixed;
  inset: 0;
  opacity: 0.18;
  pointer-events: none;
  background-image:
    linear-gradient(rgba(148, 163, 184, 0.14) 1px, transparent 1px),
    linear-gradient(90deg, rgba(148, 163, 184, 0.1) 1px, transparent 1px);
  background-size: clamp(36px, 5vw, 56px) clamp(36px, 5vw, 56px);
  mask-image: radial-gradient(circle at 50% 20%, black, transparent 75%);
}

.glow {
  position: fixed;
  width: min(430px, 60vw);
  height: min(430px, 60vw);
  border-radius: 999px;
  filter: blur(95px);
  pointer-events: none;
}

.glow-a {
  right: 10%;
  top: -220px;
  background: rgba(14, 165, 233, 0.22);
}

.glow-b {
  left: 7%;
  bottom: -250px;
  background: rgba(37, 99, 235, 0.18);
}

.subscription-shell {
  position: relative;
  z-index: 1;
  width: min(1180px, calc(100% - clamp(24px, 5vw, 56px)));
  margin: 0 auto;
  padding: clamp(18px, 3vw, 32px) 0 clamp(32px, 6vw, 58px);
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 62px;
  margin-bottom: clamp(18px, 3vw, 28px);
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  color: #ffffff;
  font-size: clamp(22px, 4vw, 28px);
  font-weight: 900;
  letter-spacing: -0.05em;
  text-decoration: none;
}

.brand span {
  width: 12px;
  height: 12px;
  border-radius: 999px;
  background: #0ea5e9;
  box-shadow: 0 0 22px rgba(14, 165, 233, 0.9);
}

.hero-card,
.glass-card,
.panel-card,
.link-card,
.apps-card,
.state-card {
  border: 1px solid rgba(148, 163, 184, 0.18);
  background:
    linear-gradient(180deg, rgba(15, 23, 42, 0.84), rgba(2, 6, 23, 0.66)),
    radial-gradient(circle at top left, rgba(14, 165, 233, 0.12), transparent 42%);
  box-shadow:
    0 24px 80px rgba(0, 0, 0, 0.26),
    inset 0 1px 0 rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(18px);
}

.hero-card {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(180px, 260px);
  gap: clamp(20px, 4vw, 34px);
  align-items: center;
  min-height: clamp(280px, 42vw, 360px);
  padding: clamp(24px, 5vw, 42px);
  border-radius: clamp(24px, 4vw, 36px);
}

.eyebrow,
.card-label {
  display: block;
  margin-bottom: 10px;
  color: #7dd3fc;
  font-size: 12px;
  font-weight: 900;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.hero-card h1 {
  max-width: 760px;
  margin: 0;
  font-size: clamp(38px, 8vw, 76px);
  line-height: 0.95;
  letter-spacing: -0.075em;
}

.hero-card p {
  max-width: 650px;
  margin: 22px 0 0;
  color: rgba(203, 213, 225, 0.86);
  font-size: clamp(16px, 2.2vw, 19px);
  line-height: 1.55;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 28px;
}

.primary-button,
.secondary-button,
.ghost-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 46px;
  padding: 12px 18px;
  border: 0;
  border-radius: 15px;
  color: #ffffff;
  font: inherit;
  font-weight: 900;
  text-decoration: none;
  cursor: pointer;
  white-space: nowrap;
}

.primary-button {
  background: linear-gradient(135deg, #0786ff, #0057e5);
  box-shadow: 0 18px 42px rgba(0, 102, 255, 0.28);
}

.secondary-button,
.ghost-button {
  border: 1px solid rgba(148, 163, 184, 0.2);
  background: rgba(15, 23, 42, 0.56);
}

.status-orb {
  display: grid;
  place-items: center;
  width: clamp(165px, 24vw, 220px);
  height: clamp(165px, 24vw, 220px);
  margin: auto;
  border-radius: 50%;
  background:
    radial-gradient(circle at 40% 30%, rgba(34, 197, 94, 0.4), rgba(14, 165, 233, 0.14) 45%, rgba(15, 23, 42, 0.62) 70%),
    #020617;
  border: 1px solid rgba(34, 197, 94, 0.28);
  box-shadow:
    0 0 70px rgba(34, 197, 94, 0.2),
    inset 0 0 40px rgba(14, 165, 233, 0.12);
  text-align: center;
}

.status-orb.expired,
.status-orb.cancelled {
  border-color: rgba(248, 113, 113, 0.28);
  background:
    radial-gradient(circle at 40% 30%, rgba(248, 113, 113, 0.34), rgba(251, 146, 60, 0.12) 45%, rgba(15, 23, 42, 0.62) 70%),
    #020617;
}

.status-orb strong {
  display: block;
  font-size: clamp(26px, 4vw, 32px);
  letter-spacing: -0.05em;
}

.status-orb span {
  display: block;
  margin-top: 8px;
  color: #cbd5e1;
}

.cards-grid,
.main-grid {
  display: grid;
  gap: 18px;
  margin-top: 18px;
}

.cards-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.main-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.glass-card,
.panel-card,
.link-card,
.apps-card {
  border-radius: clamp(22px, 3vw, 30px);
  padding: clamp(20px, 3vw, 26px);
}

.glass-card strong {
  display: block;
  font-size: clamp(22px, 4vw, 27px);
  letter-spacing: -0.04em;
  overflow-wrap: anywhere;
}

.glass-card p {
  margin: 8px 0 0;
  color: #94a3b8;
}

.panel-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.panel-heading h2,
.link-card h2,
.apps-heading h2 {
  margin: 0;
  font-size: clamp(25px, 4vw, 31px);
  letter-spacing: -0.05em;
}

.panel-heading strong {
  color: #7dd3fc;
  font-size: clamp(24px, 4vw, 30px);
}

.progress {
  height: 12px;
  margin-top: 22px;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.18);
}

.progress span {
  display: block;
  height: 100%;
  min-width: 2%;
  border-radius: inherit;
  background: linear-gradient(90deg, #0ea5e9, #22c55e);
  box-shadow: 0 0 24px rgba(14, 165, 233, 0.55);
}

.meta-row {
  display: flex;
  justify-content: space-between;
  gap: 14px;
  margin-top: 12px;
  color: #94a3b8;
  font-size: 14px;
}

.link-card,
.apps-card {
  margin-top: 18px;
}

.copy-box {
  display: flex;
  gap: 12px;
  margin-top: 18px;
}

.copy-box input {
  flex: 1;
  min-width: 0;
  min-height: 50px;
  padding: 13px 15px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 16px;
  color: #e2e8f0;
  background: rgba(2, 6, 23, 0.64);
}

.copy-note {
  margin: 12px 0 0;
  color: #94a3b8;
  line-height: 1.5;
}

.copy-note code {
  color: #7dd3fc;
}

.copy-success {
  margin: 12px 0 0;
  color: #86efac;
}

.apps-heading p {
  margin: 10px 0 0;
  color: #94a3b8;
}

.device-tabs {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  margin-top: 18px;
}

.device-tabs button {
  min-height: 42px;
  padding: 10px 14px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 14px;
  color: #cbd5e1;
  background: rgba(15, 23, 42, 0.56);
  font-weight: 900;
  cursor: pointer;
}

.device-tabs button.active {
  color: #ffffff;
  border-color: rgba(14, 165, 233, 0.45);
  background: linear-gradient(135deg, rgba(14, 165, 233, 0.28), rgba(37, 99, 235, 0.18));
}

.apps-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
  margin-top: 18px;
}

.app-card {
  display: grid;
  gap: 18px;
  padding: 18px;
  border: 1px solid rgba(148, 163, 184, 0.14);
  border-radius: 22px;
  background: rgba(2, 6, 23, 0.46);
}

.app-title {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.app-card strong {
  font-size: 20px;
}

.app-title span,
.app-title em {
  display: inline-flex;
  align-items: center;
  min-height: 23px;
  padding: 4px 8px;
  border-radius: 999px;
  color: #bae6fd;
  background: rgba(14, 165, 233, 0.12);
  font-size: 12px;
  font-style: normal;
  font-weight: 900;
}

.app-title em {
  color: #bbf7d0;
  background: rgba(34, 197, 94, 0.12);
}

.app-card p {
  margin: 8px 0 0;
  color: #94a3b8;
  line-height: 1.45;
}

.app-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.state-card {
  display: grid;
  place-items: center;
  min-height: 420px;
  padding: 42px;
  border-radius: 34px;
  text-align: center;
}

.state-card h1 {
  margin: 18px 0 8px;
  font-size: clamp(34px, 7vw, 46px);
  letter-spacing: -0.06em;
}

.state-card p {
  max-width: 460px;
  margin: 0;
  color: #94a3b8;
}

.pulse {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  border: 4px solid rgba(14, 165, 233, 0.18);
  border-top-color: #0ea5e9;
  animation: spin 0.9s linear infinite;
}

.error-icon {
  display: grid;
  place-items: center;
  width: 58px;
  height: 58px;
  border-radius: 50%;
  color: #fecaca;
  background: rgba(248, 113, 113, 0.13);
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 980px) {
  .hero-card,
  .cards-grid,
  .main-grid,
  .apps-grid {
    grid-template-columns: 1fr;
  }

  .hero-card {
    text-align: left;
  }

  .status-orb {
    margin: 0;
  }
}

@media (max-width: 560px) {
  .subscription-shell {
    width: min(100% - 20px, 1180px);
  }

  .topbar {
    align-items: stretch;
    flex-direction: column;
  }

  .hero-actions,
  .copy-box,
  .app-actions {
    flex-direction: column;
  }

  .primary-button,
  .secondary-button,
  .ghost-button {
    width: 100%;
  }

  .meta-row {
    flex-direction: column;
  }
}
</style>
