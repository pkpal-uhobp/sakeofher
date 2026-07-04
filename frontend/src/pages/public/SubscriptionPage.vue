<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, formatBytesGB, formatDate, daysLeft } from '../../api/client'

type PublicSubscription = {
  subscription?: {
    id?: number
    status?: string
    period_status?: string
    expires_at?: string
    current_period_start?: string
    current_period_end?: string
    traffic_limit_bytes?: number
    traffic_used_bytes?: number
  }
  user?: {
    telegram_id?: number
    telegram_username?: string | null
    alias?: string | null
  }
  tariff?: {
    title?: string
    price_rub?: number
    traffic_limit_bytes?: number
  }
  subscription_url?: string | null
  telegram_bot_url?: string | null
  bot_url?: string | null
}

const route = useRoute()

const loading = ref(true)
const error = ref('')
const item = ref<PublicSubscription | null>(null)
const copied = ref(false)

const secret = computed(() => String(route.params.secret || '').trim())
const token = computed(() => String(route.params.token || '').trim())
const telegramId = computed(() => Number(route.params.telegramId || 0))

const subscription = computed(() => item.value?.subscription || {})
const user = computed(() => item.value?.user || {})

const isExpired = computed(() => {
  const status = String(subscription.value.status || '').toLowerCase()
  const periodStatus = String(subscription.value.period_status || '').toLowerCase()
  const expiresAt = subscription.value.expires_at ? new Date(subscription.value.expires_at).getTime() : 0

  return (
    status === 'expired' ||
    status === 'cancelled' ||
    periodStatus === 'finished' ||
    periodStatus === 'traffic_exhausted' ||
    Boolean(expiresAt && expiresAt < Date.now())
  )
})

const title = computed(() => {
  if (isExpired.value) return 'Подписка истекла'
  return 'Ваша подписка активна и готова к использованию'
})

const subtitle = computed(() => {
  if (isExpired.value) {
    return 'Доступ временно отключён. Подключения будут удалены из профиля при обновлении подписки. Продлите доступ через Telegram-бота.'
  }

  return 'Один URL работает в двух режимах: в браузере открывает эту страницу, а в приложениях отдаёт полную Remnawave Base64-подписку.'
})

const statusText = computed(() => (isExpired.value ? 'Истекла' : 'Активна'))
const statusSubtext = computed(() => {
  if (isExpired.value) return 'Нужно продлить'

  const left = daysLeft(subscription.value.expires_at || null)
  if (left <= 0) return 'Активна'

  return `${left} дней осталось`
})

const subscriptionUrl = computed(() => {
  if (secret.value && telegramId.value) {
    return `${window.location.origin}/${secret.value}/sub/${telegramId.value}`
  }

  if (token.value) {
    return `${window.location.origin}/s/${token.value}`
  }

  return window.location.href
})

const base64Url = computed(() => {
  const separator = subscriptionUrl.value.includes('?') ? '&' : '?'
  return `${subscriptionUrl.value}${separator}format=base64`
})

const botUrl = computed(() => {
  const candidates = [
    item.value?.telegram_bot_url,
    item.value?.bot_url,
    import.meta.env.VITE_TELEGRAM_BOT_URL,
    import.meta.env.VITE_BOT_URL,
  ]

  for (const value of candidates) {
    const prepared = String(value || '').trim()
    if (prepared) return prepared
  }

  return 'https://t.me/'
})

const renewBotUrl = computed(() => {
  const raw = botUrl.value.trim()
  if (!raw || raw === 'https://t.me/') return raw

  try {
    const url = new URL(raw)

    // If admin already specified a concrete deep-link, do not break it.
    if (!url.searchParams.has('start')) {
      const id = user.value.telegram_id || telegramId.value || ''
      if (id) url.searchParams.set('start', `renew_${id}`)
    }

    return url.toString()
  } catch {
    return raw
  }
})

const trafficUsed = computed(() => Number(subscription.value.traffic_used_bytes || 0))
const trafficLimit = computed(() => Number(subscription.value.traffic_limit_bytes || item.value?.tariff?.traffic_limit_bytes || 0))
const trafficPercent = computed(() => {
  if (!trafficLimit.value) return 0
  return Math.min(100, Math.round((trafficUsed.value / trafficLimit.value) * 100))
})

const username = computed(() => {
  const alias = String(user.value.telegram_username || user.value.alias || '').trim()
  if (alias) return alias.startsWith('@') ? alias : `@${alias}`

  if (user.value.telegram_id) return String(user.value.telegram_id)

  return 'Пользователь'
})

const apps = [
  {
    title: 'Happ',
    platform: 'Android / iOS / Windows / macOS / Linux',
    downloadUrl: 'https://happ.su/',
    openScheme: 'happ://add/',
  },
  {
    title: 'Hiddify',
    platform: 'Android / iOS / Windows / macOS / Linux',
    downloadUrl: 'https://hiddify.com/',
    openScheme: 'hiddify://import/',
  },
  {
    title: 'v2rayNG',
    platform: 'Android',
    downloadUrl: 'https://github.com/2dust/v2rayNG/releases',
    openScheme: 'v2rayng://install-config?url=',
  },
  {
    title: 'Shadowrocket',
    platform: 'iOS',
    downloadUrl: 'https://apps.apple.com/app/shadowrocket/id932747118',
    openScheme: 'shadowrocket://add/sub://',
  },
]

function appOpenUrl(openScheme: string) {
  return `${openScheme}${encodeURIComponent(subscriptionUrl.value)}`
}

async function copySubscriptionUrl() {
  await navigator.clipboard.writeText(subscriptionUrl.value)
  copied.value = true
  window.setTimeout(() => {
    copied.value = false
  }, 1500)
}

async function loadSubscription() {
  loading.value = true
  error.value = ''

  try {
    if (secret.value && telegramId.value) {
      const response = await api.get(`/subscriptions/path/${encodeURIComponent(secret.value)}/telegram/${telegramId.value}`)
      item.value = response.data
      return
    }

    if (token.value) {
      const response = await api.get(`/subscriptions/public/${encodeURIComponent(token.value)}`)
      item.value = response.data
      return
    }

    throw new Error('Неверная ссылка подписки')
  } catch (err: any) {
    error.value = err?.response?.data?.error || err?.message || 'Не удалось загрузить подписку'
  } finally {
    loading.value = false
  }
}

onMounted(loadSubscription)
</script>

<template>
  <main class="subscription-page">
    <header class="topbar">
      <div class="brand">
        <span class="brand-dot"></span>
        <span>SakeOfHer</span>
      </div>

      <a
        class="top-renew"
        :href="renewBotUrl"
        target="_blank"
        rel="noopener noreferrer"
      >
        Продлить в Telegram
      </a>
    </header>

    <section v-if="loading" class="panel state-panel">
      <p class="eyebrow">Загрузка</p>
      <h1>Проверяем подписку</h1>
      <p>Сейчас загрузим статус доступа и ссылку для подключения.</p>
    </section>

    <section v-else-if="error" class="panel state-panel state-panel-error">
      <p class="eyebrow">Ошибка</p>
      <h1>Подписка не найдена</h1>
      <p>{{ error }}</p>

      <a
        class="btn btn-green"
        :href="renewBotUrl"
        target="_blank"
        rel="noopener noreferrer"
      >
        Перейти в Telegram-бота
      </a>
    </section>

    <template v-else>
      <section class="hero panel" :class="{ expired: isExpired }">
        <div class="hero-copy">
          <p class="eyebrow">Личный доступ</p>
          <h1>{{ title }}</h1>
          <p>{{ subtitle }}</p>

          <div class="hero-actions">
            <button class="btn btn-blue" type="button" @click="copySubscriptionUrl">
              {{ copied ? 'Скопировано' : 'Скопировать ссылку' }}
            </button>

            <a
              class="btn btn-green"
              :href="renewBotUrl"
              target="_blank"
              rel="noopener noreferrer"
            >
              Продлить в Telegram
            </a>

            <button class="btn btn-dark" type="button" @click="loadSubscription">
              Обновить статус
            </button>
          </div>
        </div>

        <div class="status-orb" :class="{ expired: isExpired }">
          <strong>{{ statusText }}</strong>
          <span>{{ statusSubtext }}</span>
        </div>
      </section>

      <section v-if="isExpired" class="panel expired-panel">
        <p class="eyebrow">Доступ отключён</p>
        <h2>Подключения удалены из подписки</h2>
        <p>
          При обновлении профиля клиент получит пустую подписку. Чтобы вернуть подключения,
          продлите доступ через Telegram-бота.
        </p>

        <a
          class="btn btn-green"
          :href="renewBotUrl"
          target="_blank"
          rel="noopener noreferrer"
        >
          Перейти к продлению
        </a>
      </section>

      <section class="cards">
        <article class="info-card">
          <p class="eyebrow">Пользователь</p>
          <h2>{{ username }}</h2>
          <span>{{ user.telegram_id || '—' }}</span>
        </article>

        <article class="info-card">
          <p class="eyebrow">Оплаченный доступ</p>
          <h2>{{ isExpired ? 'Отключён' : 'Активен' }}</h2>
          <span>{{ subscription.status || '—' }}</span>
        </article>

        <article class="info-card">
          <p class="eyebrow">Истекает</p>
          <h2>{{ formatDate(subscription.expires_at || null) }}</h2>
          <span>{{ statusSubtext }}</span>
        </article>
      </section>

      <section class="panel traffic-panel">
        <div class="section-head">
          <div>
            <p class="eyebrow">Трафик текущего периода</p>
            <h2>{{ formatBytesGB(trafficUsed) }} / {{ trafficLimit ? formatBytesGB(trafficLimit) : '∞' }}</h2>
          </div>
          <span>{{ trafficPercent }}%</span>
        </div>

        <div class="traffic-line">
          <div :style="{ width: `${trafficPercent}%` }"></div>
        </div>

        <p class="muted">
          Месячный лимит считается отдельно от общего срока доступа. При продлении срок суммируется,
          а текущий период трафика обновляется worker'ом.
        </p>
      </section>

      <section class="panel link-panel">
        <p class="eyebrow">Один URL для страницы и приложений</p>
        <div class="url-row">
          <input :value="subscriptionUrl" readonly />
          <button class="btn btn-blue" type="button" @click="copySubscriptionUrl">Copy</button>
        </div>

        <p class="muted">
          В браузере ссылка открывает красивую страницу. В приложении эта же ссылка отдаёт подписку.
          Для принудительного Base64: <code>{{ base64Url }}</code>
        </p>
      </section>

      <section class="panel apps-panel">
        <p class="eyebrow">Приложения</p>
        <h2>Выберите клиент</h2>
        <p class="muted">
          Кнопка «Открыть» передаёт текущую ссылку подписки прямо в приложение.
        </p>

        <div class="apps-grid">
          <article v-for="app in apps" :key="app.title" class="app-card">
            <h3>{{ app.title }}</h3>
            <p>{{ app.platform }}</p>

            <div class="app-actions">
              <a class="btn btn-dark" :href="app.downloadUrl" target="_blank" rel="noopener noreferrer">
                Скачать
              </a>
              <a class="btn btn-blue" :href="appOpenUrl(app.openScheme)">
                Открыть
              </a>
            </div>
          </article>
        </div>
      </section>
    </template>
  </main>
</template>

<style scoped>
.subscription-page {
  min-height: 100vh;
  padding: 28px;
  color: #f8fbff;
  background:
    radial-gradient(circle at 80% 0%, rgba(20, 150, 220, 0.35), transparent 32%),
    linear-gradient(180deg, #020713 0%, #040916 100%);
  font-family: Inter, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 1480px;
  margin: 0 auto 28px;
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  font-weight: 900;
  font-size: 28px;
}

.brand-dot {
  width: 14px;
  height: 14px;
  border-radius: 999px;
  background: #17a8ff;
  box-shadow: 0 0 28px rgba(23, 168, 255, 0.8);
}

.top-renew {
  color: #fff;
  text-decoration: none;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 999px;
  padding: 14px 20px;
  background: rgba(9, 18, 34, 0.75);
  font-weight: 800;
}

.panel {
  max-width: 1480px;
  margin: 0 auto 24px;
  border: 1px solid rgba(139, 171, 218, 0.18);
  border-radius: 34px;
  background:
    linear-gradient(135deg, rgba(15, 28, 52, 0.92), rgba(4, 9, 22, 0.88)),
    rgba(7, 14, 29, 0.9);
  box-shadow: 0 24px 80px rgba(0, 0, 0, 0.28);
}

.hero {
  min-height: 360px;
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 40px;
  align-items: center;
  padding: clamp(32px, 4vw, 54px);
}

.hero.expired {
  border-color: rgba(255, 74, 108, 0.25);
  background:
    radial-gradient(circle at 85% 25%, rgba(255, 46, 88, 0.18), transparent 30%),
    linear-gradient(135deg, rgba(15, 28, 52, 0.92), rgba(4, 9, 22, 0.88));
}

.eyebrow {
  margin: 0 0 20px;
  color: #7ddcff;
  text-transform: uppercase;
  letter-spacing: 0.22em;
  font-weight: 900;
  font-size: 14px;
}

.hero h1,
.state-panel h1 {
  max-width: 980px;
  margin: 0;
  font-size: clamp(44px, 6vw, 92px);
  line-height: 0.98;
  letter-spacing: -0.06em;
}

.hero-copy p:not(.eyebrow),
.state-panel p {
  max-width: 760px;
  margin: 28px 0 0;
  color: #bac6d7;
  font-size: clamp(18px, 2vw, 25px);
  line-height: 1.55;
}

.hero-actions,
.app-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  margin-top: 34px;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 58px;
  padding: 0 24px;
  border: 0;
  border-radius: 17px;
  color: #fff;
  text-decoration: none;
  font: inherit;
  font-weight: 900;
  cursor: pointer;
}

.btn-blue {
  background: linear-gradient(135deg, #1599ff, #075ee8);
}

.btn-green {
  background: linear-gradient(135deg, #24c866, #109946);
}

.btn-dark {
  background: rgba(20, 31, 52, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.12);
}

.status-orb {
  width: min(280px, 30vw);
  aspect-ratio: 1;
  border-radius: 999px;
  display: grid;
  place-content: center;
  text-align: center;
  background:
    radial-gradient(circle, rgba(23, 205, 133, 0.24), rgba(9, 20, 44, 0.7));
  border: 1px solid rgba(35, 209, 122, 0.35);
}

.status-orb.expired {
  background:
    radial-gradient(circle, rgba(255, 69, 107, 0.25), rgba(9, 20, 44, 0.7));
  border-color: rgba(255, 80, 107, 0.36);
}

.status-orb strong {
  font-size: clamp(28px, 3vw, 42px);
}

.status-orb span {
  margin-top: 10px;
  color: #c4cedd;
  font-size: 18px;
}

.state-panel,
.expired-panel,
.traffic-panel,
.link-panel,
.apps-panel {
  padding: 32px;
}

.state-panel-error {
  border-color: rgba(255, 80, 107, 0.36);
}

.expired-panel {
  border-color: rgba(255, 80, 107, 0.25);
  background:
    radial-gradient(circle at 20% 0%, rgba(255, 55, 99, 0.2), transparent 30%),
    rgba(9, 14, 30, 0.92);
}

.expired-panel h2,
.apps-panel h2,
.traffic-panel h2 {
  margin: 0;
  font-size: clamp(30px, 3vw, 44px);
  letter-spacing: -0.04em;
}

.expired-panel p,
.muted {
  color: #b8c4d5;
  line-height: 1.55;
  font-size: 18px;
}

.cards {
  max-width: 1480px;
  margin: 0 auto 24px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 24px;
}

.info-card,
.app-card {
  border: 1px solid rgba(139, 171, 218, 0.18);
  border-radius: 28px;
  background: rgba(8, 15, 31, 0.88);
  padding: 28px;
}

.info-card h2 {
  margin: 0 0 8px;
  font-size: 30px;
}

.info-card span {
  color: #96a6bc;
  font-size: 18px;
}

.section-head,
.url-row {
  display: flex;
  gap: 18px;
  align-items: center;
  justify-content: space-between;
}

.traffic-line {
  height: 14px;
  margin: 24px 0;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.1);
}

.traffic-line div {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #16b7ff, #1fd67c);
}

.url-row input {
  flex: 1;
  min-width: 0;
  height: 58px;
  border: 1px solid rgba(139, 171, 218, 0.22);
  border-radius: 17px;
  padding: 0 18px;
  color: #eaf2ff;
  background: rgba(1, 7, 20, 0.72);
  font: inherit;
}

code {
  color: #7ddcff;
  overflow-wrap: anywhere;
}

.apps-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
  margin-top: 24px;
}

.app-card h3 {
  margin: 0 0 10px;
  font-size: 26px;
}

.app-card p {
  color: #aebbd0;
  font-size: 17px;
  line-height: 1.45;
}

@media (max-width: 980px) {
  .subscription-page {
    padding: 18px;
  }

  .topbar {
    gap: 16px;
    align-items: flex-start;
  }

  .hero {
    grid-template-columns: 1fr;
  }

  .status-orb {
    width: 220px;
  }

  .cards,
  .apps-grid {
    grid-template-columns: 1fr;
  }

  .section-head,
  .url-row {
    align-items: stretch;
    flex-direction: column;
  }
}

@media (max-width: 620px) {
  .topbar {
    flex-direction: column;
  }

  .hero h1,
  .state-panel h1 {
    font-size: 44px;
  }

  .btn {
    width: 100%;
  }
}
</style>
