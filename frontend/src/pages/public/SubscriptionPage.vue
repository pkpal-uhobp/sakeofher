<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, bytesToGB } from '../../api/client'

const route = useRoute()
const loading = ref(true)
const error = ref('')
const item = ref<any>(null)
const checkout = ref<any>(null)
const creatingRenewLink = ref(false)

const token = computed(() => String(route.params.token || ''))
const secret = computed(() => String(route.params.secret || ''))
const telegramId = computed(() => String(route.params.telegramId || ''))
const isTelegramSubscriptionPath = computed(() => Boolean(secret.value && telegramId.value))
const trafficGB = computed(() => bytesToGB(item.value?.subscription?.traffic_limit_bytes || 0))
const usedGB = computed(() => bytesToGB(item.value?.subscription?.traffic_used_bytes || 0))
const daysLeft = computed(() => {
  const expires = item.value?.subscription?.expires_at
  if (!expires) return 0
  const ms = new Date(expires).getTime() - Date.now()
  return Math.max(0, Math.ceil(ms / 86400000))
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const endpoint = isTelegramSubscriptionPath.value
      ? `/subscriptions/path/${encodeURIComponent(secret.value)}/telegram/${encodeURIComponent(telegramId.value)}`
      : `/subscriptions/public/${encodeURIComponent(token.value)}`
    const { data } = await api.get(endpoint)
    item.value = data
  } catch (e: any) {
    error.value = e?.response?.data?.error || e.message || 'Подписка не найдена'
  } finally {
    loading.value = false
  }
}

async function createRenewLink() {
  creatingRenewLink.value = true
  error.value = ''
  checkout.value = null
  try {
    const { data } = await api.post('/checkout/renew', {
      public_token: item.value?.subscription?.public_token || token.value,
    })
    checkout.value = data
  } catch (e: any) {
    error.value = e?.response?.data?.error || e.message || 'Ошибка формирования ссылки продления'
  } finally {
    creatingRenewLink.value = false
  }
}

async function copyLink() {
  const link = item.value?.subscription_url
  if (link) await navigator.clipboard.writeText(link)
}

onMounted(load)
</script>

<template>
  <main class="page">
    <RouterLink class="back" to="/">← На главную</RouterLink>

    <section class="card" v-if="loading">Загрузка...</section>
    <section class="card error" v-else-if="error && !item">{{ error }}</section>

    <section v-else-if="item" class="card">
      <p class="badge">{{ item.subscription.status }}</p>
      <h1>Подписка {{ item.tariff.title }}</h1>
      <p v-if="isTelegramSubscriptionPath" class="note">Страница открыта по Telegram ID: {{ telegramId }}</p>

      <div class="grid">
        <div>
          <span>Осталось</span>
          <b>{{ daysLeft }} дней</b>
        </div>
        <div>
          <span>Лимит</span>
          <b>{{ trafficGB }} GB</b>
        </div>
        <div>
          <span>Использовано</span>
          <b>{{ usedGB }} GB</b>
        </div>
      </div>

      <div class="linkbox">
        <span>VPN-ссылка</span>
        <code>{{ item.subscription_url || 'Будет доступна после подключения Remnawave' }}</code>
      </div>

      <div class="actions">
        <button class="secondary" @click="copyLink">Скопировать ссылку</button>
        <button class="primary" :disabled="creatingRenewLink" @click="createRenewLink">
          {{ creatingRenewLink ? 'Готовим ссылку...' : 'Продлить в Telegram-боте' }}
        </button>
      </div>

      <p class="note">
        При продлении лимит трафика не вводится заново. Бот подтянет текущий лимит: {{ trafficGB }} GB.
      </p>
      <p v-if="error" class="error">{{ error }}</p>

      <div v-if="checkout" class="checkout">
        <b>Ссылка продления готова</b>
        <p>{{ checkout.note }}</p>
        <p>Payload для бота: <code>{{ checkout.start_payload }}</code></p>
        <a class="telegram" :href="checkout.telegram_bot_url" target="_blank" rel="noreferrer">
          Открыть Telegram-бота
        </a>
      </div>
    </section>
  </main>
</template>

<style scoped>
.page { max-width: 920px; margin: 0 auto; padding: 48px 20px 80px; }
.back { color: #93c5fd; text-decoration: none; }
.card { background: #111827; border: 1px solid #263244; border-radius: 24px; padding: 28px; margin-top: 20px; box-shadow: 0 24px 80px rgba(0,0,0,.25); }
.badge { display: inline-block; padding: 8px 12px; border: 1px solid #334155; border-radius: 999px; color: #bbf7d0; }
h1 { font-size: 40px; margin: 18px 0 24px; }
.grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(170px, 1fr)); gap: 14px; }
.grid div { background: #0f172a; border: 1px solid #334155; border-radius: 18px; padding: 18px; display: grid; gap: 8px; }
.grid span, .linkbox span, .note { color: #94a3b8; }
.grid b { font-size: 26px; }
.linkbox { display: grid; gap: 8px; margin-top: 18px; }
code { display: inline-block; overflow: auto; background: #020617; border: 1px solid #334155; border-radius: 10px; padding: 4px 8px; color: #c4b5fd; }
.linkbox code { display: block; border-radius: 14px; padding: 14px; }
.actions { display: flex; gap: 12px; flex-wrap: wrap; margin-top: 18px; }
button, .telegram { border: 0; border-radius: 16px; padding: 14px 18px; font-weight: 800; cursor: pointer; text-decoration: none; }
.primary, .telegram { background: #60a5fa; color: #020617; }
.secondary { background: #1f2937; color: #e5e7eb; }
.error { color: #fca5a5; }
.checkout { display: grid; gap: 10px; background: #052e16; border: 1px solid #166534; border-radius: 18px; padding: 18px; margin-top: 18px; color: #bbf7d0; }
.checkout p { margin: 0; }
</style>
