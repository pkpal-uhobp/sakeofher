import axios from 'axios'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 15000,
})

export function bytesToGB(bytes: number): number {
  if (!bytes || bytes <= 0) return 0
  return Math.round(bytes / 1024 / 1024 / 1024)
}
