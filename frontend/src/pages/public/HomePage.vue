<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { api, bytesToGB } from '../../api/client'

type Tariff = {
  id: number
  code: string
  title: string
  description?: string
  duration_days: number
  period_days: number
  traffic_limit_bytes: number
}

type CheckoutLink = {
  action: 'purchase' | 'renew'
  start_payload: string
  telegram_bot_url: string
  telegram_bot_username: string
  tariff: Tariff
  traffic_limit_gb: number
  traffic_limit_bytes: number
  next_expires_at_preview: string
  note: string
}

const tariffs = ref<Tariff[]>([])
const selectedTariffId = ref<number | null>(null)
const trafficLimitGB = ref<number>(300)
const checkout = ref<CheckoutLink | null>(null)
const loading = ref(false)
const error = ref('')

const selectedTariff = computed(() => tariffs.value.find((item) => item.id === selectedTariffId.value))

watch(selectedTariff, (tariff) => {
  if (tariff) {
    trafficLimitGB.value = bytesToGB(tariff.traffic_limit_bytes)
    checkout.value = null
  }
})

onMounted(async () => {
  const { data } = await api.get('/tariffs')
  tariffs.value = data
  if (tariffs.value.length > 0) {
    selectedTariffId.value = tariffs.value[0].id
    trafficLimitGB.value = bytesToGB(tariffs.value[0].traffic_limit_bytes)
  }
})

async function createCheckoutLink() {
  error.value = ''
  checkout.value = null
  loading.value = true
  try {
    const { data } = await api.post('/checkout/purchase', {
      tariff_id: selectedTariffId.value,
      traffic_limit_gb: Number(trafficLimitGB.value),
    })
    checkout.value = data
  } catch (e: any) {
    error.value = e?.response?.data?.error || e.message || 'Не удалось сформировать ссылку в Telegram'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="page">
    <section class="hero">
      <p class="badge">SakeOfHer VPN</p>
      <h1>Выберите подписку на сайте, оплатите в Telegram-боте</h1>
      <p class="subtitle">
        Сайт только помогает выбрать тариф и лимит трафика. Оплата, создание пользователя и выдача VPN-ссылки будут выполняться на стороне Telegram-бота.
      </p>
    </section>

    <section class="card">
      <h2>1. Тариф</h2>
      <div class="tariffs">
        <button
          v-for="tariff in tariffs"
          :key="tariff.id"
          class="tariff"
          :class="{ active: selectedTariffId === tariff.id }"
          @click="selectedTariffId = tariff.id"
        >
          <b>{{ tariff.title }}</b>
          <span>{{ tariff.duration_days }} дней</span>
          <small>{{ bytesToGB(tariff.traffic_limit_bytes) }} GB по умолчанию</small>
        </button>
      </div>
    </section>

    <section class="card form">
      <h2>2. Лимит трафика</h2>
      <p class="muted">
        При первой покупке лимит задаётся вручную. При продлении он будет автоматически подтягиваться из текущей подписки.
      </p>
      <label>
        Лимит трафика, GB
        <input v-model.number="trafficLimitGB" type="number" min="1" />
      </label>

      <button class="primary" :disabled="loading || !selectedTariffId || !trafficLimitGB" @click="createCheckoutLink">
        {{ loading ? 'Готовим ссылку...' : 'Перейти к оплате в Telegram' }}
      </button>

      <p v-if="error" class="error">{{ error }}</p>

      <div v-if="checkout" class="success">
        <div>
          <b>Ссылка готова</b>
          <p>{{ checkout.note }}</p>
          <p class="payload">Payload для бота: <code>{{ checkout.start_payload }}</code></p>
        </div>
        <a class="telegram" :href="checkout.telegram_bot_url" target="_blank" rel="noreferrer">
          Открыть Telegram-бота
        </a>
      </div>
    </section>

    <section class="card muted-card">
      <h2>Как будет работать дальше</h2>
      <ol>
        <li>Пользователь выбирает тариф и лимит на сайте.</li>
        <li>Сайт формирует Telegram deep-link с параметрами покупки.</li>
        <li>Пользователь оплачивает в боте.</li>
        <li>Бот активирует подписку и выдаёт страницу подписки.</li>
      </ol>
    </section>
  </main>
</template>

<style scoped>
.page { max-width: 1080px; margin: 0 auto; padding: 48px 20px 80px; }
.hero { margin-bottom: 28px; }
.badge { display: inline-block; padding: 8px 12px; border: 1px solid #334155; border-radius: 999px; color: #93c5fd; }
h1 { max-width: 820px; font-size: 46px; line-height: 1.05; margin: 18px 0; }
.subtitle { max-width: 800px; color: #cbd5e1; font-size: 18px; }
.card { background: #111827; border: 1px solid #263244; border-radius: 24px; padding: 24px; margin: 18px 0; box-shadow: 0 24px 80px rgba(0,0,0,.25); }
.tariffs { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 14px; }
.tariff { text-align: left; border: 1px solid #334155; background: #0f172a; color: #e5e7eb; border-radius: 18px; padding: 18px; cursor: pointer; display: grid; gap: 8px; }
.tariff.active { border-color: #60a5fa; box-shadow: 0 0 0 3px rgba(96,165,250,.15); }
.tariff span, .tariff small, .muted, .muted-card { color: #94a3b8; }
.form { display: grid; gap: 14px; }
label { display: grid; gap: 8px; color: #cbd5e1; }
input { background: #0f172a; color: #e5e7eb; border: 1px solid #334155; border-radius: 14px; padding: 14px 16px; font-size: 16px; }
.primary, .telegram { margin-top: 8px; border: 0; border-radius: 16px; background: #60a5fa; color: #020617; padding: 15px 18px; font-weight: 800; cursor: pointer; text-decoration: none; text-align: center; }
.primary:disabled { opacity: .5; cursor: not-allowed; }
.error { color: #fca5a5; }
.success { display: grid; gap: 14px; background: #052e16; border: 1px solid #166534; border-radius: 18px; padding: 18px; color: #bbf7d0; }
.success p { margin: 8px 0 0; }
.payload { color: #d9f99d; }
code { background: #020617; border: 1px solid #334155; border-radius: 10px; padding: 3px 8px; color: #c4b5fd; }
ol { padding-left: 22px; }
li { margin: 8px 0; }
</style>
