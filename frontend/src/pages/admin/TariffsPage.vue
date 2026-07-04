<template>
  <AdminLayout>
    <header class="admin-topbar">
      <div class="admin-title">
        <h1>Тарифы</h1>
        <p>Цена, срок, лимит и включение способов оплаты отдельно для Stars и CryptoBot.</p>
      </div>

      <button class="admin-button" type="button" @click="openCreate">
        + Добавить тариф
      </button>
    </header>

    <section class="admin-card admin-card-body">
      <div class="admin-toolbar">
        <div class="admin-filters">
          <label class="form-field">
            <span>Поиск</span>
            <input
              v-model.trim="search"
              class="admin-input"
              placeholder="Название, код или описание"
            />
          </label>

          <label class="form-field">
            <span>Статус</span>
            <select v-model="statusFilter" class="admin-select">
              <option value="">Все</option>
              <option value="active">Активные</option>
              <option value="disabled">Отключённые</option>
            </select>
          </label>
        </div>

        <button class="admin-ghost" type="button" @click="load">
          Обновить
        </button>
      </div>

      <p v-if="loading" class="admin-loading">Загружаем тарифы…</p>
      <p v-else-if="error" class="admin-error">{{ error }}</p>

      <div v-else-if="filteredTariffs.length" class="admin-table-wrap">
        <table class="admin-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Код</th>
              <th>Название</th>
              <th>Цена</th>
              <th>Оплата</th>
              <th>Срок</th>
              <th>Лимит</th>
              <th>Статус</th>
              <th></th>
            </tr>
          </thead>

          <tbody>
            <tr v-for="tariff in filteredTariffs" :key="tariff.id">
              <td>{{ tariff.id }}</td>
              <td><span class="admin-kbd">{{ tariff.code }}</span></td>
              <td>
                <strong>{{ tariff.title }}</strong>
                <br />
                <span>{{ tariff.description || '—' }}</span>
              </td>
              <td>
                <strong>{{ formatRub(tariff.price_rub) }}</strong>
              </td>
              <td>
                <div class="payment-badges">
                  <span v-for="label in paymentLabels(tariff)" :key="label" class="status-pill active">
                    {{ label }}
                  </span>
                  <span v-if="!paymentLabels(tariff).length" class="status-pill disabled">нет</span>
                </div>
              </td>
              <td>{{ tariff.duration_days }} дн.</td>
              <td>{{ formatBytesGB(tariff.traffic_limit_bytes) }}</td>
              <td>
                <span class="status-pill" :class="tariff.is_active ? 'active' : 'disabled'">
                  {{ tariff.is_active ? 'active' : 'disabled' }}
                </span>
              </td>
              <td>
                <div class="row-actions">
                  <button class="admin-link-button" type="button" @click="openEdit(tariff)">
                    Изменить
                  </button>

                  <button
                    v-if="tariff.is_active"
                    class="admin-ghost"
                    type="button"
                    @click="toggle(tariff.id, false)"
                  >
                    Отключить
                  </button>

                  <button
                    v-else
                    class="admin-ghost"
                    type="button"
                    @click="toggle(tariff.id, true)"
                  >
                    Включить
                  </button>

                  <button class="admin-danger" type="button" @click="remove(tariff)">
                    Удалить
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <p v-else class="empty-state">Тарифы не найдены.</p>
    </section>

    <div v-if="modalOpen" class="modal-backdrop" @click.self="closeModal">
      <section class="modal-card tariff-modal">
        <header class="modal-header">
          <h2>{{ editingId ? 'Изменить тариф' : 'Добавить тариф' }}</h2>
          <button class="admin-ghost" type="button" @click="closeModal">×</button>
        </header>

        <form class="modal-body" @submit.prevent="save">
          <div class="form-grid">
            <label class="form-field">
              <span>Код тарифа</span>
              <input v-model.trim="form.code" class="admin-input" required placeholder="Например: month_500" />
            </label>

            <label class="form-field">
              <span>Название</span>
              <input v-model.trim="form.title" class="admin-input" required placeholder="Например: 1 месяц" />
            </label>

            <label class="form-field">
              <span>Базовая цена, ₽</span>
              <input
                v-model.number="form.price_rub"
                class="admin-input"
                type="number"
                min="0"
                required
                placeholder="299"
              />
            </label>

            <label class="form-field">
              <span>Длительность, дней</span>
              <input v-model.number="form.duration_days" class="admin-input" type="number" min="1" required placeholder="30" />
            </label>

            <label class="form-field">
              <span>Период трафика, дней</span>
              <input v-model.number="form.period_days" class="admin-input" type="number" min="1" required placeholder="30" />
            </label>

            <label class="form-field">
              <span>Лимит, ГБ</span>
              <input v-model.number="form.traffic_limit_gb" class="admin-input" type="number" min="1" required placeholder="500" />
            </label>

            <label class="form-field">
              <span>Порядок сортировки</span>
              <input v-model.number="form.sort_order" class="admin-input" type="number" placeholder="100" />
            </label>

            <label class="form-field">
              <span>Статус</span>
              <select v-model="form.is_active" class="admin-select">
                <option :value="true">Активен</option>
                <option :value="false">Отключён</option>
              </select>
            </label>

            <label class="form-field wide">
              <span>Описание</span>
              <textarea
                v-model.trim="form.description"
                class="admin-textarea"
                placeholder="Короткое описание тарифа"
              />
            </label>
          </div>

          <section class="payment-settings">
            <h3>Способы оплаты</h3>

            <article class="payment-setting-card">
              <label class="switch-row">
                <input v-model="form.enable_stars" type="checkbox" />
                <span>
                  <strong>Telegram Stars</strong>
                  <small>Отдельная цена в звёздах</small>
                </span>
              </label>

              <label class="form-field">
                <span>Цена в Stars</span>
                <input
                  v-model.number="form.stars_amount"
                  class="admin-input"
                  type="number"
                  min="1"
                  placeholder="299"
                  :disabled="!form.enable_stars"
                />
              </label>
            </article>

            <article class="payment-setting-card">
              <label class="switch-row">
                <input v-model="form.enable_cryptobot_crypto" type="checkbox" />
                <span>
                  <strong>CryptoBot — крипта</strong>
                  <small>Инвойс CryptoBot в рублёвой цене с выбором активов</small>
                </span>
              </label>

              <label class="form-field">
                <span>Цена, ₽</span>
                <input
                  v-model.number="form.cryptobot_crypto_price_rub"
                  class="admin-input"
                  type="number"
                  min="1"
                  placeholder="299"
                  :disabled="!form.enable_cryptobot_crypto"
                />
              </label>

              <label class="form-field wide">
                <span>Активы через запятую</span>
                <input
                  v-model.trim="form.cryptobot_crypto_assets"
                  class="admin-input"
                  placeholder="USDT, TON, BTC, ETH, LTC, BNB, TRX, USDC"
                  :disabled="!form.enable_cryptobot_crypto"
                />
              </label>
            </article>

            <article class="payment-setting-card">
              <label class="switch-row">
                <input v-model="form.enable_cryptobot_rub" type="checkbox" />
                <span>
                  <strong>CryptoBot — рубли</strong>
                  <small>Отдельный рублёвый способ оплаты</small>
                </span>
              </label>

              <label class="form-field">
                <span>Цена, ₽</span>
                <input
                  v-model.number="form.cryptobot_rub_price_rub"
                  class="admin-input"
                  type="number"
                  min="1"
                  placeholder="299"
                  :disabled="!form.enable_cryptobot_rub"
                />
              </label>
            </article>
          </section>

          <div class="modal-actions">
            <button class="admin-ghost" type="button" @click="closeModal">Отмена</button>
            <button class="admin-button" type="submit" :disabled="saving">
              {{ saving ? 'Сохраняем…' : editingId ? 'Сохранить изменения' : 'Добавить тариф' }}
            </button>
          </div>
        </form>
      </section>
    </div>
  </AdminLayout>
</template>

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
  type Tariff,
  type TariffPrice,
} from '../../api/admin'
import { bytesToGB, formatBytesGB, formatRub } from '../../api/client'

const tariffs = ref<Tariff[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const modalOpen = ref(false)
const editingId = ref<number | null>(null)
const search = ref('')
const statusFilter = ref('')

const form = reactive({
  code: '',
  title: '',
  description: '',
  duration_days: 30,
  period_days: 30,
  traffic_limit_gb: 500,
  price_rub: 299,
  is_active: true,
  sort_order: 100,

  enable_stars: false,
  stars_amount: 299,

  enable_cryptobot_crypto: false,
  cryptobot_crypto_price_rub: 299,
  cryptobot_crypto_assets: 'USDT, TON, BTC, ETH, LTC, BNB, TRX, USDC',

  enable_cryptobot_rub: false,
  cryptobot_rub_price_rub: 299,
})

const filteredTariffs = computed(() => {
  const q = search.value.toLowerCase()

  return tariffs.value.filter((tariff) => {
    const statusOk =
      !statusFilter.value ||
      (statusFilter.value === 'active' && tariff.is_active) ||
      (statusFilter.value === 'disabled' && !tariff.is_active)

    const queryOk =
      !q ||
      tariff.code.toLowerCase().includes(q) ||
      tariff.title.toLowerCase().includes(q) ||
      (tariff.description || '').toLowerCase().includes(q)

    return statusOk && queryOk
  })
})

onMounted(load)

async function load() {
  loading.value = true
  error.value = ''

  try {
    tariffs.value = await listTariffs()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось загрузить тарифы'
  } finally {
    loading.value = false
  }
}

function paymentLabels(tariff: Tariff): string[] {
  const prices = tariff.prices || []
  const labels: string[] = []

  if (prices.some((price) => price.provider === 'telegram' && price.payment_method === 'stars')) {
    labels.push('Stars')
  }
  if (prices.some((price) => price.provider === 'cryptobot' && price.payment_method === 'crypto')) {
    labels.push('CryptoBot crypto')
  }
  if (prices.some((price) => price.provider === 'cryptobot' && price.payment_method === 'rub')) {
    labels.push('CryptoBot RUB')
  }

  return labels
}

function openCreate() {
  editingId.value = null

  Object.assign(form, {
    code: '',
    title: '',
    description: '',
    duration_days: 30,
    period_days: 30,
    traffic_limit_gb: 500,
    price_rub: 299,
    is_active: true,
    sort_order: 100,

    enable_stars: false,
    stars_amount: 299,

    enable_cryptobot_crypto: false,
    cryptobot_crypto_price_rub: 299,
    cryptobot_crypto_assets: 'USDT, TON, BTC, ETH, LTC, BNB, TRX, USDC',

    enable_cryptobot_rub: false,
    cryptobot_rub_price_rub: 299,
  })

  modalOpen.value = true
}

function openEdit(tariff: Tariff) {
  editingId.value = tariff.id

  const stars = findPrice(tariff, 'telegram', 'stars')
  const crypto = findPrice(tariff, 'cryptobot', 'crypto')
  const rub = findPrice(tariff, 'cryptobot', 'rub')

  Object.assign(form, {
    code: tariff.code,
    title: tariff.title,
    description: tariff.description || '',
    duration_days: tariff.duration_days,
    period_days: tariff.period_days,
    traffic_limit_gb: bytesToGB(tariff.traffic_limit_bytes),
    price_rub: tariff.price_rub || 0,
    is_active: tariff.is_active,
    sort_order: tariff.sort_order,

    enable_stars: Boolean(stars),
    stars_amount: stars?.stars_amount || tariff.price_rub || 299,

    enable_cryptobot_crypto: Boolean(crypto),
    cryptobot_crypto_price_rub: crypto?.amount_minor ? Math.round(crypto.amount_minor / 100) : tariff.price_rub || 299,
    cryptobot_crypto_assets: crypto?.accepted_assets?.length
      ? crypto.accepted_assets.join(', ')
      : 'USDT, TON, BTC, ETH, LTC, BNB, TRX, USDC',

    enable_cryptobot_rub: Boolean(rub),
    cryptobot_rub_price_rub: rub?.amount_minor ? Math.round(rub.amount_minor / 100) : tariff.price_rub || 299,
  })

  modalOpen.value = true
}

function findPrice(tariff: Tariff, provider: string, method: string): TariffPrice | undefined {
  return (tariff.prices || []).find((price) => price.provider === provider && price.payment_method === method)
}

function closeModal() {
  modalOpen.value = false
}

function buildPaymentSettings() {
  return {
    telegram_stars: {
      enabled: form.enable_stars,
      stars_amount: Number(form.stars_amount || 0),
    },
    cryptobot_crypto: {
      enabled: form.enable_cryptobot_crypto,
      price_rub: Number(form.cryptobot_crypto_price_rub || 0),
      accepted_assets: form.cryptobot_crypto_assets
        .split(',')
        .map((item) => item.trim().toUpperCase())
        .filter(Boolean),
    },
    cryptobot_rub: {
      enabled: form.enable_cryptobot_rub,
      price_rub: Number(form.cryptobot_rub_price_rub || 0),
    },
  }
}

async function save() {
  saving.value = true
  error.value = ''

  try {
    const payload = {
      code: form.code,
      title: form.title,
      description: form.description || null,
      duration_days: form.duration_days,
      period_days: form.period_days,
      traffic_limit_gb: form.traffic_limit_gb,
      price_rub: form.price_rub,
      is_active: form.is_active,
      sort_order: form.sort_order,
      payment_settings: buildPaymentSettings(),
    }

    if (editingId.value) {
      await updateTariff(editingId.value, payload)
    } else {
      await createTariff(payload)
    }

    closeModal()
    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось сохранить тариф'
  } finally {
    saving.value = false
  }
}

async function toggle(id: number, active: boolean) {
  error.value = ''

  try {
    if (active) {
      await enableTariff(id)
    } else {
      await disableTariff(id)
    }

    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось изменить статус тарифа'
  }
}

async function remove(tariff: Tariff) {
  const ok = window.confirm(`Удалить тариф "${tariff.title}"? Он будет скрыт из списка, но история подписок сохранится.`)
  if (!ok) return

  error.value = ''

  try {
    await deleteTariff(tariff.id)
    await load()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Не удалось удалить тариф'
  }
}
</script>

<style src="./admin.css"></style>
<style scoped>
.payment-badges {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
}

.tariff-modal {
  width: min(980px, calc(100vw - 28px));
}

.payment-settings {
  display: grid;
  gap: 14px;
  margin-top: 24px;
}

.payment-settings h3 {
  margin: 0;
  color: #f8fafc;
  font-size: 22px;
}

.payment-setting-card {
  display: grid;
  grid-template-columns: minmax(260px, 1fr) minmax(220px, 0.75fr);
  gap: 14px;
  align-items: end;
  padding: 18px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 20px;
  background: rgba(15, 23, 42, 0.42);
}

.payment-setting-card .wide {
  grid-column: 1 / -1;
}

.switch-row {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 52px;
  color: #f8fafc;
}

.switch-row input {
  width: 20px;
  height: 20px;
  accent-color: #0ea5e9;
}

.switch-row span {
  display: grid;
  gap: 4px;
}

.switch-row small {
  color: #94a3b8;
}

@media (max-width: 760px) {
  .payment-setting-card {
    grid-template-columns: 1fr;
  }
}
</style>
