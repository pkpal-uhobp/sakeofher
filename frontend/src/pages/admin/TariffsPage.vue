<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AdminLayout from './AdminLayout.vue'
import {
  createTariff,
  deleteTariff,
  disableTariff,
  enableTariff,
  listTariffs,
  updateTariff,
  type CreateTariffInput,
  type Tariff,
  type TariffPaymentSettingsInput,
} from '../../api/admin'
import { bytesToGB, formatBytesGB } from '../../api/client'

type TariffForm = {
  id: number | null
  code: string
  title: string
  description: string
  duration_days: number
  period_days: number
  traffic_limit_gb: number
  price_rub: number
  is_active: boolean
  sort_order: number
  payment_settings: TariffPaymentSettingsInput
}

const tariffs = ref<Tariff[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const success = ref('')

const defaultAssets = 'USDT, TON, BTC, ETH, LTC, BNB, TRX, USDC'

const form = reactive<TariffForm>({
  id: null,
  code: 'vpn_1m_300gb',
  title: '1 месяц',
  description: 'Доступ на 30 дней, 300 ГБ на период',
  duration_days: 30,
  period_days: 30,
  traffic_limit_gb: 300,
  price_rub: 65,
  is_active: true,
  sort_order: 10,
  payment_settings: {
    telegram_stars: {
      enabled: true,
      stars_amount: 50,
    },
    cryptobot_crypto: {
      enabled: true,
      price_rub: 65,
      accepted_assets: assetsFromString(defaultAssets),
    },
    tribute_rub: {
      enabled: false,
      price_rub: 65,
    },
  },
})

const isEditing = computed(() => Boolean(form.id))

const activeTariffs = computed(() => tariffs.value.filter((item) => item.is_active))
const inactiveTariffs = computed(() => tariffs.value.filter((item) => !item.is_active))

async function loadTariffs() {
  loading.value = true
  error.value = ''

  try {
    tariffs.value = await listTariffs()
  } catch (err: any) {
    error.value = err?.response?.data?.error || err?.message || 'Не удалось загрузить тарифы'
  } finally {
    loading.value = false
  }
}

function preset(months: 1 | 2 | 3) {
  if (months === 1) {
    Object.assign(form, {
      id: null,
      code: 'vpn_1m_300gb',
      title: '1 месяц',
      description: 'Доступ на 30 дней, 300 ГБ на период',
      duration_days: 30,
      period_days: 30,
      traffic_limit_gb: 300,
      price_rub: 65,
      is_active: true,
      sort_order: 10,
      payment_settings: {
        telegram_stars: { enabled: true, stars_amount: 50 },
        cryptobot_crypto: { enabled: true, price_rub: 65, accepted_assets: assetsFromString(defaultAssets) },
        tribute_rub: { enabled: false, price_rub: 65 },
      },
    })
  }

  if (months === 2) {
    Object.assign(form, {
      id: null,
      code: 'vpn_2m_300gb',
      title: '2 месяца',
      description: 'Доступ на 60 дней, 300 ГБ на период',
      duration_days: 60,
      period_days: 30,
      traffic_limit_gb: 300,
      price_rub: 140,
      is_active: true,
      sort_order: 20,
      payment_settings: {
        telegram_stars: { enabled: true, stars_amount: 100 },
        cryptobot_crypto: { enabled: true, price_rub: 130, accepted_assets: assetsFromString(defaultAssets) },
        tribute_rub: { enabled: true, price_rub: 140 },
      },
    })
  }

  if (months === 3) {
    Object.assign(form, {
      id: null,
      code: 'vpn_3m_300gb',
      title: '3 месяца',
      description: 'Доступ на 90 дней, 300 ГБ на период',
      duration_days: 90,
      period_days: 30,
      traffic_limit_gb: 300,
      price_rub: 210,
      is_active: true,
      sort_order: 30,
      payment_settings: {
        telegram_stars: { enabled: true, stars_amount: 150 },
        cryptobot_crypto: { enabled: true, price_rub: 195, accepted_assets: assetsFromString(defaultAssets) },
        tribute_rub: { enabled: true, price_rub: 210 },
      },
    })
  }

  success.value = ''
  error.value = ''
}

function editTariff(tariff: Tariff) {
  const stars = findPrice(tariff, 'telegram_stars', 'stars')
  const crypto = findPrice(tariff, 'crypto_bot', 'crypto')
  const tribute = findPrice(tariff, 'tribute', 'rub')

  Object.assign(form, {
    id: tariff.id,
    code: tariff.code,
    title: tariff.title,
    description: tariff.description || '',
    duration_days: tariff.duration_days,
    period_days: tariff.period_days,
    traffic_limit_gb: bytesToGB(tariff.traffic_limit_bytes),
    price_rub: rubFromMinor(tariff.price_rub),
    is_active: tariff.is_active,
    sort_order: tariff.sort_order,
    payment_settings: {
      telegram_stars: {
        enabled: Boolean(stars?.is_active),
        stars_amount: Number(stars?.stars_amount || 0),
      },
      cryptobot_crypto: {
        enabled: Boolean(crypto?.is_active),
        price_rub: rubFromMinor(crypto?.amount_minor || 0),
        accepted_assets: crypto?.accepted_assets?.length ? crypto.accepted_assets : assetsFromString(defaultAssets),
      },
      tribute_rub: {
        enabled: Boolean(tribute?.is_active),
        price_rub: rubFromMinor(tribute?.amount_minor || 0),
      },
    },
  })

  success.value = ''
  error.value = ''
}

function resetForm() {
  preset(1)
}

async function saveTariff() {
  saving.value = true
  error.value = ''
  success.value = ''

  try {
    const payload = buildPayload()

    if (form.id) {
      await updateTariff(form.id, payload)
      success.value = 'Тариф обновлён'
    } else {
      await createTariff(payload as CreateTariffInput)
      success.value = 'Тариф создан'
    }

    await loadTariffs()
  } catch (err: any) {
    error.value = err?.response?.data?.error || err?.message || 'Не удалось сохранить тариф'
  } finally {
    saving.value = false
  }
}

async function toggleTariff(tariff: Tariff) {
  error.value = ''
  success.value = ''

  try {
    if (tariff.is_active) {
      await disableTariff(tariff.id)
      success.value = 'Тариф выключен'
    } else {
      await enableTariff(tariff.id)
      success.value = 'Тариф включён'
    }

    await loadTariffs()
  } catch (err: any) {
    error.value = err?.response?.data?.error || err?.message || 'Не удалось изменить статус тарифа'
  }
}

async function removeTariff(tariff: Tariff) {
  if (!confirm(`Удалить тариф "${tariff.title}"?`)) return

  error.value = ''
  success.value = ''

  try {
    await deleteTariff(tariff.id)
    success.value = 'Тариф удалён'
    await loadTariffs()

    if (form.id === tariff.id) resetForm()
  } catch (err: any) {
    error.value = err?.response?.data?.error || err?.message || 'Не удалось удалить тариф'
  }
}

function buildPayload(): CreateTariffInput {
  return {
    code: form.code.trim(),
    title: form.title.trim(),
    description: form.description.trim() || null,
    duration_days: Number(form.duration_days),
    period_days: Number(form.period_days),
    traffic_limit_gb: Number(form.traffic_limit_gb),
    price_rub: Number(form.price_rub),
    is_active: Boolean(form.is_active),
    sort_order: Number(form.sort_order),
    payment_settings: {
      telegram_stars: {
        enabled: Boolean(form.payment_settings.telegram_stars.enabled),
        stars_amount: Number(form.payment_settings.telegram_stars.stars_amount),
      },
      cryptobot_crypto: {
        enabled: Boolean(form.payment_settings.cryptobot_crypto.enabled),
        price_rub: Number(form.payment_settings.cryptobot_crypto.price_rub),
        accepted_assets: assetsFromString(form.payment_settings.cryptobot_crypto.accepted_assets.join(', ')),
      },
      tribute_rub: {
        enabled: Boolean(form.payment_settings.tribute_rub.enabled),
        price_rub: Number(form.payment_settings.tribute_rub.price_rub),
      },
    },
  }
}

function findPrice(tariff: Tariff, provider: string, method: string) {
  return tariff.prices?.find((item) => item.provider === provider && item.payment_method === method)
}

function rubFromMinor(value: number | null | undefined) {
  return Math.round(Number(value || 0) / 100)
}

function assetsFromString(value: string): string[] {
  return String(value || '')
    .split(',')
    .map((item) => item.trim().toUpperCase())
    .filter(Boolean)
}

function priceSummary(tariff: Tariff) {
  const stars = findPrice(tariff, 'telegram_stars', 'stars')
  const crypto = findPrice(tariff, 'crypto_bot', 'crypto')
  const tribute = findPrice(tariff, 'tribute', 'rub')

  const items: string[] = []

  if (stars?.is_active) items.push(`${stars.stars_amount} ⭐`)
  if (crypto?.is_active) items.push(`${rubFromMinor(crypto.amount_minor)} ₽ CryptoBot`)
  if (tribute?.is_active) items.push(`${rubFromMinor(tribute.amount_minor)} ₽ Tribute`)

  return items.length ? items.join(' · ') : 'Нет активных способов оплаты'
}

onMounted(loadTariffs)
</script>

<template>
  <AdminLayout>
    <section class="tariffs-page">
      <div class="page-head">
        <div>
          <h1>Тарифы</h1>
          <p>
            Управление сроком, лимитом и способами оплаты: Telegram Stars, CryptoBot crypto и Tribute RUB.
          </p>
        </div>

        <button class="btn btn-primary" type="button" @click="resetForm">
          + Новый тариф
        </button>
      </div>

      <div class="preset-row">
        <button class="pill" type="button" @click="preset(1)">
          1 месяц · 50⭐ · 65₽ crypto
        </button>
        <button class="pill" type="button" @click="preset(2)">
          2 месяца · 100⭐ · 130₽ crypto · 140₽ Tribute
        </button>
        <button class="pill" type="button" @click="preset(3)">
          3 месяца · 150⭐ · 195₽ crypto · 210₽ Tribute
        </button>
      </div>

      <p v-if="error" class="alert alert-error">{{ error }}</p>
      <p v-if="success" class="alert alert-success">{{ success }}</p>

      <div class="grid">
        <form class="panel form-panel" @submit.prevent="saveTariff">
          <div class="section-title">
            <h2>{{ isEditing ? 'Редактировать тариф' : 'Создать тариф' }}</h2>
            <span>{{ isEditing ? `ID ${form.id}` : 'новый' }}</span>
          </div>

          <div class="form-grid">
            <label>
              Код
              <input v-model="form.code" placeholder="vpn_1m_300gb" />
            </label>

            <label>
              Название
              <input v-model="form.title" placeholder="1 месяц" />
            </label>

            <label class="full">
              Описание
              <textarea v-model="form.description" placeholder="Доступ на 30 дней, 300 ГБ на период"></textarea>
            </label>

            <label>
              Длительность, дней
              <input v-model.number="form.duration_days" type="number" min="1" />
            </label>

            <label>
              Период трафика, дней
              <input v-model.number="form.period_days" type="number" min="1" />
            </label>

            <label>
              Лимит, ГБ
              <input v-model.number="form.traffic_limit_gb" type="number" min="1" />
            </label>

            <label>
              Базовая цена, ₽
              <input v-model.number="form.price_rub" type="number" min="0" />
            </label>

            <label>
              Сортировка
              <input v-model.number="form.sort_order" type="number" />
            </label>

            <label class="check-line">
              <input v-model="form.is_active" type="checkbox" />
              Тариф активен
            </label>
          </div>

          <h3>Способы оплаты</h3>

          <div class="payment-card">
            <label class="payment-check">
              <input v-model="form.payment_settings.telegram_stars.enabled" type="checkbox" />
              <span>
                <strong>Telegram Stars</strong>
                <small>Отдельная цена в звёздах</small>
              </span>
            </label>

            <label>
              Цена в Stars
              <input v-model.number="form.payment_settings.telegram_stars.stars_amount" type="number" min="1" />
            </label>
          </div>

          <div class="payment-card">
            <label class="payment-check">
              <input v-model="form.payment_settings.cryptobot_crypto.enabled" type="checkbox" />
              <span>
                <strong>CryptoBot — крипта</strong>
                <small>Рублёвая цена инвойса, оплата активами</small>
              </span>
            </label>

            <label>
              Цена, ₽
              <input v-model.number="form.payment_settings.cryptobot_crypto.price_rub" type="number" min="1" />
            </label>

            <label class="full">
              Активы через запятую
              <input
                :value="form.payment_settings.cryptobot_crypto.accepted_assets.join(', ')"
                @input="form.payment_settings.cryptobot_crypto.accepted_assets = assetsFromString(($event.target as HTMLInputElement).value)"
                placeholder="USDT, TON, BTC, ETH, LTC, BNB, TRX, USDC"
              />
            </label>
          </div>

          <div class="payment-card">
            <label class="payment-check">
              <input v-model="form.payment_settings.tribute_rub.enabled" type="checkbox" />
              <span>
                <strong>Tribute — рубли</strong>
                <small>Отдельный рублёвый способ оплаты</small>
              </span>
            </label>

            <label>
              Цена, ₽
              <input v-model.number="form.payment_settings.tribute_rub.price_rub" type="number" min="1" />
            </label>
          </div>

          <div class="actions">
            <button class="btn btn-primary" :disabled="saving" type="submit">
              {{ saving ? 'Сохраняю...' : 'Сохранить тариф' }}
            </button>

            <button class="btn btn-muted" type="button" @click="resetForm">
              Сбросить
            </button>
          </div>
        </form>

        <section class="panel list-panel">
          <div class="section-title">
            <h2>Активные тарифы</h2>
            <button class="btn btn-muted" type="button" @click="loadTariffs">
              Обновить
            </button>
          </div>

          <p v-if="loading" class="muted">Загрузка...</p>

          <article v-for="tariff in activeTariffs" :key="tariff.id" class="tariff-row">
            <div>
              <div class="row-title">
                <strong>{{ tariff.title }}</strong>
                <span>{{ tariff.code }}</span>
              </div>
              <p>{{ tariff.duration_days }} дн. · {{ formatBytesGB(tariff.traffic_limit_bytes) }} · {{ priceSummary(tariff) }}</p>
            </div>

            <div class="row-actions">
              <button class="btn btn-muted" type="button" @click="editTariff(tariff)">
                Изменить
              </button>
              <button class="btn btn-muted" type="button" @click="toggleTariff(tariff)">
                Выкл.
              </button>
              <button class="btn btn-danger" type="button" @click="removeTariff(tariff)">
                Удалить
              </button>
            </div>
          </article>

          <div v-if="inactiveTariffs.length" class="inactive">
            <h3>Выключенные</h3>

            <article v-for="tariff in inactiveTariffs" :key="tariff.id" class="tariff-row">
              <div>
                <div class="row-title">
                  <strong>{{ tariff.title }}</strong>
                  <span>{{ tariff.code }}</span>
                </div>
                <p>{{ priceSummary(tariff) }}</p>
              </div>

              <div class="row-actions">
                <button class="btn btn-muted" type="button" @click="editTariff(tariff)">
                  Изменить
                </button>
                <button class="btn btn-primary" type="button" @click="toggleTariff(tariff)">
                  Вкл.
                </button>
              </div>
            </article>
          </div>
        </section>
      </div>
    </section>
  </AdminLayout>
</template>

<style scoped>
.tariffs-page {
  padding: 34px;
  color: #f7fbff;
}

.page-head,
.section-title,
.actions,
.row-actions,
.row-title,
.preset-row {
  display: flex;
  align-items: center;
  gap: 14px;
}

.page-head,
.section-title {
  justify-content: space-between;
}

.page-head h1 {
  margin: 0;
  font-size: 52px;
  line-height: 1;
  letter-spacing: -0.05em;
}

.page-head p,
.muted,
.tariff-row p,
.payment-check small {
  color: #9fb1c7;
}

.preset-row {
  flex-wrap: wrap;
  margin: 24px 0;
}

.pill,
.btn {
  border: 0;
  border-radius: 16px;
  padding: 14px 18px;
  color: #fff;
  background: rgba(20, 31, 52, 0.92);
  font: inherit;
  font-weight: 900;
  cursor: pointer;
}

.pill {
  border: 1px solid rgba(36, 170, 255, 0.28);
}

.grid {
  display: grid;
  grid-template-columns: minmax(420px, 1fr) minmax(420px, 1fr);
  gap: 24px;
}

.panel {
  border: 1px solid rgba(139, 171, 218, 0.18);
  border-radius: 28px;
  background: rgba(8, 15, 31, 0.88);
  box-shadow: 0 20px 80px rgba(0, 0, 0, 0.24);
}

.form-panel,
.list-panel {
  padding: 28px;
}

.section-title h2 {
  margin: 0;
  font-size: 30px;
}

.section-title span {
  color: #7ddcff;
  font-weight: 900;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
  margin-top: 24px;
}

label {
  display: grid;
  gap: 8px;
  color: #aebbd0;
  font-weight: 900;
}

.full {
  grid-column: 1 / -1;
}

input,
textarea {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid rgba(139, 171, 218, 0.2);
  border-radius: 16px;
  padding: 16px;
  color: #fff;
  background: #040a19;
  font: inherit;
}

textarea {
  min-height: 90px;
  resize: vertical;
}

.check-line,
.payment-check {
  display: flex;
  align-items: center;
  gap: 14px;
}

.check-line input,
.payment-check input {
  width: 22px;
  height: 22px;
}

h3 {
  margin: 26px 0 14px;
}

.payment-card {
  display: grid;
  grid-template-columns: minmax(240px, 1fr) minmax(200px, 0.8fr);
  gap: 18px;
  margin-top: 16px;
  padding: 22px;
  border: 1px solid rgba(139, 171, 218, 0.14);
  border-radius: 24px;
  background: rgba(11, 22, 42, 0.78);
}

.payment-check span {
  display: grid;
  gap: 6px;
}

.btn-primary {
  background: linear-gradient(135deg, #1599ff, #075ee8);
}

.btn-muted {
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.btn-danger {
  background: rgba(105, 29, 47, 0.9);
}

.alert {
  padding: 18px 20px;
  border-radius: 18px;
  font-weight: 800;
}

.alert-error {
  color: #ffd4d4;
  background: rgba(105, 29, 47, 0.45);
  border: 1px solid rgba(255, 80, 107, 0.28);
}

.alert-success {
  color: #c9ffe4;
  background: rgba(22, 132, 72, 0.36);
  border: 1px solid rgba(45, 220, 130, 0.28);
}

.tariff-row {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  padding: 20px 0;
  border-bottom: 1px solid rgba(139, 171, 218, 0.1);
}

.tariff-row:last-child {
  border-bottom: 0;
}

.row-title strong {
  font-size: 22px;
}

.row-title span {
  color: #7ddcff;
}

.inactive {
  margin-top: 26px;
  opacity: 0.75;
}

@media (max-width: 1180px) {
  .grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .tariffs-page {
    padding: 18px;
  }

  .page-head,
  .section-title,
  .tariff-row {
    align-items: flex-start;
    flex-direction: column;
  }

  .form-grid,
  .payment-card {
    grid-template-columns: 1fr;
  }

  .row-actions {
    flex-wrap: wrap;
  }
}
</style>
