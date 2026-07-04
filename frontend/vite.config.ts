import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

const proxyTarget = process.env.VITE_DEV_PROXY_TARGET || 'http://127.0.0.1:8080'
const subscriptionPathSecret = process.env.VITE_SUBSCRIPTION_PATH_SECRET || 'L0mENeiofHjdxC57'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': {
        target: proxyTarget,
        changeOrigin: true,
      },

      // Local dev behavior:
      // GET http://localhost:5173/L0mENeiofHjdxC57/sub/213
      // must return Base64 subscription text from backend, not Vite SPA HTML.
      [`^/${subscriptionPathSecret}/sub/.*`]: {
        target: proxyTarget,
        changeOrigin: true,
      },
    },
  },
})
