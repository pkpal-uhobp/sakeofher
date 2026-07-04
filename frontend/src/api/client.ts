import axios from 'axios'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 15000,
  withCredentials: true,
})

export function bytesToGB(bytes: number): number {
  if (!bytes || bytes <= 0) return 0
  return Math.round(bytes / 1024 / 1024 / 1024)
}

export function formatBytesGB(bytes: number): string {
  const gb = bytesToGB(bytes)
  return `${gb} ГБ`
}

export function formatDate(value?: string | null): string {
  if (!value) return '—'

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'

  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

export function daysLeft(value?: string | null): number {
  if (!value) return 0

  const expiresAt = new Date(value).getTime()
  if (Number.isNaN(expiresAt)) return 0

  return Math.max(0, Math.ceil((expiresAt - Date.now()) / 86_400_000))
}

export function formatRub(value?: number | null): string {
  if (!value || value <= 0) return '0 ₽'

  return new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    maximumFractionDigits: 0,
  }).format(value)
}
