import { api } from './client'

export type UserStatus = 'active' | 'blocked' | 'deleted'
export type SubscriptionStatus = 'active' | 'expired' | 'cancelled'
export type PeriodStatus = 'active' | 'traffic_exhausted' | 'finished'

export interface User {
  id: number
  telegram_id: number
  telegram_username?: string | null
  telegram_first_name?: string | null
  telegram_last_name?: string | null
  language_code?: string | null
  alias?: string | null
  remna_uuid?: string | null
  remna_username?: string | null
  subscription_url?: string | null
  status?: UserStatus | string
  remna_status?: string | null
  created_at?: string
  updated_at?: string
}

export interface TariffPrice {
  id: number
  tariff_id: number
  provider: string
  payment_method: string
  currency: string
  amount_minor?: number | null
  stars_amount?: number | null
  accepted_assets: string[]
  is_active: boolean
  sort_order: number
}

export interface Tariff {
  id: number
  code: string
  title: string
  description?: string | null
  duration_days: number
  period_days: number
  traffic_limit_bytes: number
  price_rub: number
  is_active: boolean
  sort_order: number
  prices?: TariffPrice[]
  created_at?: string
  updated_at?: string
}

export interface RemnaInternalSquad {
  uuid: string
  name: string
  is_active: boolean
}

export interface Subscription {
  id: number
  user_id: number
  tariff_id?: number | null
  last_payment_id?: number | null
  status: SubscriptionStatus | string
  started_at?: string
  expires_at: string
  current_period_start?: string
  current_period_end?: string
  traffic_limit_bytes: number
  traffic_used_bytes: number
  period_status?: PeriodStatus | string
  public_token: string
  created_at?: string
  updated_at?: string
}

export interface PublicSubscription {
  subscription: Subscription
  user: User
  tariff: Tariff
  subscription_url?: string | null
}

export interface ListResponse<T> {
  items: T[]
  total: number
  limit: number
  offset: number
}

export interface AuthMe {
  user: User
  username?: string
  is_admin: boolean
}

export interface UserFilters {
  query?: string
  status?: string
  limit?: number
  offset?: number
}

export interface CreateUserInput {
  telegram_id: number
  telegram_username?: string | null
}

export interface SubscriptionFilters {
  user_id?: number | string
  telegram_id?: number | string
  status?: string
  limit?: number
  offset?: number
}

export interface TariffPaymentSettingsInput {
  telegram_stars: {
    enabled: boolean
    stars_amount: number
  }
  cryptobot_crypto: {
    enabled: boolean
    price_rub: number
    accepted_assets: string[]
  }
  tribute_rub: {
    enabled: boolean
    price_rub: number
  }
}

export interface CreateTariffInput {
  code: string
  title: string
  description?: string | null
  duration_days: number
  period_days: number
  traffic_limit_gb: number
  price_rub: number
  is_active?: boolean
  sort_order?: number
  payment_settings?: TariffPaymentSettingsInput
}

export interface UpdateTariffInput {
  code?: string
  title?: string
  description?: string | null
  duration_days?: number
  period_days?: number
  traffic_limit_gb?: number
  price_rub?: number
  is_active?: boolean
  sort_order?: number
  payment_settings?: TariffPaymentSettingsInput
}

export interface CreateManualSubscriptionInput {
  user_id: number
  tariff_id: number
  traffic_limit_gb: number
  active_internal_squads?: string[]
}

export interface ExtendSubscriptionInput {
  tariff_id?: number | null
  days?: number | null
}

export interface UpdateTrafficLimitInput {
  traffic_limit_gb: number
}

export interface UpdateUserInput {
  telegram_username?: string | null
  status?: UserStatus
}

export async function getMe(): Promise<AuthMe> {
  const { data } = await api.get('/auth/me')
  return data
}

export async function logout(): Promise<void> {
  await api.post('/auth/logout')
}

export async function listUsers(filters: UserFilters = {}): Promise<ListResponse<User>> {
  const { data } = await api.get<ListResponse<User>>('/users', { params: cleanParams(filters) })
  return data
}

export async function createUser(input: CreateUserInput): Promise<User> {
  const payload = cleanParams({
    telegram_id: input.telegram_id,
    telegram_username: normalizeUsername(input.telegram_username),
  })

  const { data } = await api.post('/users/telegram', payload)
  return data
}

export async function listRemnawaveInternalSquads(): Promise<RemnaInternalSquad[]> {
  const { data } = await api.get('/remnawave/internal-squads')
  return data
}

export async function getUser(id: number): Promise<User> {
  const { data } = await api.get(`/users/${id}`)
  return data
}

export async function updateUser(id: number, input: UpdateUserInput): Promise<User> {
  const { data } = await api.patch(`/users/${id}`, cleanParams(input))
  return data
}

export async function blockUser(id: number): Promise<User> {
  const { data } = await api.post(`/users/${id}/block`)
  return data
}

export async function unblockUser(id: number): Promise<User> {
  const { data } = await api.post(`/users/${id}/unblock`)
  return data
}

export async function deleteUser(id: number): Promise<User | null> {
  try {
    const { data } = await api.post(`/users/${id}/delete`)
    return data
  } catch (err: any) {
    if (err?.response?.status === 404) return null
    throw err
  }
}

export async function listSubscriptions(filters: SubscriptionFilters = {}): Promise<ListResponse<PublicSubscription>> {
  const { data } = await api.get<ListResponse<PublicSubscription>>('/subscriptions', { params: cleanParams(filters) })
  return data
}

export async function getSubscription(id: number): Promise<PublicSubscription> {
  const { data } = await api.get(`/subscriptions/${id}`)
  return data
}

export async function createManualSubscription(input: CreateManualSubscriptionInput): Promise<PublicSubscription> {
  const { data } = await api.post('/subscriptions', input)
  return data
}

export async function extendSubscription(id: number, input: ExtendSubscriptionInput): Promise<PublicSubscription> {
  const { data } = await api.post(`/subscriptions/${id}/extend`, cleanParams(input))
  return data
}

export async function updateTrafficLimit(id: number, input: UpdateTrafficLimitInput): Promise<PublicSubscription> {
  const { data } = await api.patch(`/subscriptions/${id}/traffic-limit`, input)
  return data
}

export async function disableSubscription(id: number): Promise<PublicSubscription> {
  const { data } = await api.post(`/subscriptions/${id}/disable`)
  return data
}

export async function enableSubscription(id: number): Promise<PublicSubscription> {
  const { data } = await api.post(`/subscriptions/${id}/enable`)
  return data
}

export async function cancelSubscription(id: number): Promise<PublicSubscription> {
  const { data } = await api.post(`/subscriptions/${id}/cancel`)
  return data
}

export async function deleteSubscription(id: number): Promise<void> {
  await api.delete(`/subscriptions/${id}`)
}

export async function listTariffs(): Promise<Tariff[]> {
  const { data } = await api.get('/tariffs/all')
  return data
}

export async function createTariff(input: CreateTariffInput): Promise<Tariff> {
  const { data } = await api.post('/tariffs', input)
  return data
}

export async function updateTariff(id: number, input: UpdateTariffInput): Promise<Tariff> {
  const { data } = await api.patch(`/tariffs/${id}`, cleanParams(input))
  return data
}

export async function enableTariff(id: number): Promise<Tariff> {
  const { data } = await api.post(`/tariffs/${id}/enable`)
  return data
}

export async function disableTariff(id: number): Promise<Tariff> {
  const { data } = await api.post(`/tariffs/${id}/disable`)
  return data
}

export async function deleteTariff(id: number): Promise<void> {
  await api.delete(`/tariffs/${id}`)
}

function normalizeUsername(value: string | null | undefined): string | null {
  const username = String(value || '').trim().replace(/^@+/, '')
  return username || null
}

function cleanParams<T extends Record<string, any>>(params: T): T {
  const next = { ...params }

  for (const key of Object.keys(next)) {
    const value = next[key]
    if (value === '' || value === null || typeof value === 'undefined') {
      delete next[key]
    }
  }

  return next
}
