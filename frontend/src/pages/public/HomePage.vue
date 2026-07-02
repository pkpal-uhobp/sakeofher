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

const tariffs = ref<Tariff[]>([])
const selectedTariffId = ref<number | null>(null)
const telegramId = ref('')
const telegramUsername = ref('')
const telegramFirstName = ref('')
const trafficLimitGB = ref<number>(300)
const publicToken = ref('')
const subscriptionUrl = ref('')
const loading = ref(false)
const error = ref('')

const selectedTariff = computed(() => tariffs.value.find((item) => item.id === selectedTariffId.value))

watch(selectedTariff, (tariff) => {
  if (tariff && !trafficLimitGB.value) {
    trafficLimitGB.value = bytesToGB(tariff.traffic_limit_bytes)
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

async function purchase() {
  error.value = ''
  publicToken.value = ''
  subscriptionUrl.value = ''
  loading.value = true
  try {
    const { data } = await api.post('/site/subscriptions/purchase', {
      telegram_id: Number(telegramId.value),
      telegram_username: telegramUsername.value || undefined,
      telegram_first_name: telegramFirstName.value || undefined,
      tariff_id: selectedTariffId.value,
      traffic_limit_gb: Number(trafficLimitGB.value),
    })
    publicToken.value = data.subscription.public_token
    subscriptionUrl.value = `/s/${publicToken.value}`
  } catch (e: any) {
    error.value = e?.response?.data?.error || e.message || 'Ошибка покупки'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="page">
    <section class="hero">
      <p class="badge">SakeOfHer VPN</p>
      <h1>Покупка подписки без платежной механики</h1>
      <p class="subtitle">
        Сейчас сайт создаёт подписку напрямую. Лимит трафика задаётся при покупке, а при продлении будет подтягиваться автоматически из текущей подписки.
      </p>
    </section>

    <section class="card">
      <h2>1. Выберите тариф</h2>
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
      <h2>2. Данные пользователя</h2>
      <label>
        Telegram ID
        <input v-model="telegramId" placeholder="123456789" />
      </label>
      <label>
        Username
        <input v-model="telegramUsername" placeholder="username без @" />
      </label>
      <label>
        Имя
        <input v-model="telegramFirstName" placeholder="Имя пользователя" />
      </label>
      <label>
        Лимит трафика при покупке, GB
        <input v-model.number="trafficLimitGB" type="number" min="1" />
      </label>

      <button class="primary" :disabled="loading || !selectedTariffId || !telegramId || !trafficLimitGB" @click="purchase">
        {{ loading ? 'Создаём...' : 'Создать подписку' }}
      </button>

      <p v-if="error" class="error">{{ error }}</p>
      <div v-if="publicToken" class="success">
        <b>Подписка создана.</b>
        <RouterLink :to="subscriptionUrl">Открыть страницу подписки</RouterLink>
      </div>
    </section>

    <section class="card muted">
      <h2>Что дальше</h2>
      <p>
        Следующим шагом добавим отдельную страницу продления. В ней лимит трафика не будет вводиться заново — backend возьмёт его из последней подписки пользователя.
      </p>
    </section>
  </main>
</template>

<style scoped>
.page { max-width: 1080px; margin: 0 auto; padding: 48px 20px 80px; }
.hero { margin-bottom: 28px; }
.badge { display: inline-block; padding: 8px 12px; border: 1px solid #334155; border-radius: 999px; color: #93c5fd; }
h1 { max-width: 760px; font-size: 46px; line-height: 1.05; margin: 18px 0; }
.subtitle { max-width: 760px; color: #cbd5e1; font-size: 18px; }
.card { background: #111827; border: 1px solid #263244; border-radius: 24px; padding: 24px; margin: 18px 0; box-shadow: 0 24px 80px rgba(0,0,0,.25); }
.tariffs { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 14px; }
.tariff { text-align: left; border: 1px solid #334155; background: #0f172a; color: #e5e7eb; border-radius: 18px; padding: 18px; cursor: pointer; display: grid; gap: 8px; }
.tariff.active { border-color: #60a5fa; box-shadow: 0 0 0 3px rgba(96,165,250,.15); }
.tariff span, .tariff small, .muted { color: #94a3b8; }
.form { display: grid; gap: 14px; }
label { display: grid; gap: 8px; color: #cbd5e1; }
input { background: #0f172a; color: #e5e7eb; border: 1px solid #334155; border-radius: 14px; padding: 14px 16px; font-size: 16px; }
.primary { margin-top: 8px; border: 0; border-radius: 16px; background: #60a5fa; color: #020617; padding: 15px 18px; font-weight: 800; cursor: pointer; }
.primary:disabled { opacity: .5; cursor: not-allowed; }
.error { color: #fca5a5; }
.success { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; color: #bbf7d0; }
.success a { color: #93c5fd; }
</style>
