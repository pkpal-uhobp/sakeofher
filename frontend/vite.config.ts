import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const proxyTarget = env.VITE_DEV_PROXY_TARGET || 'http://host.docker.internal:8080'

  return {
    plugins: [vue()],
    server: {
      host: '0.0.0.0',
      port: 5173,
      proxy: {
        '/api': {
          target: proxyTarget,
          changeOrigin: true,
          secure: false,
        },

        // Same URL works in two modes:
        //
        // Browser/Vue:
        //   /profile/L0mENeiofHjdxC57/sub/213
        //
        // Subscription clients:
        //   /L0mENeiofHjdxC57/sub/213
        //
        // In dev mode Happ/clients request localhost:5173. Without this proxy,
        // Vite returns 404 and Happ shows "server replied: Not Found".
        '^/[A-Za-z0-9_-]+/sub/[0-9]+$': {
          target: proxyTarget,
          changeOrigin: true,
          secure: false,
        },
      },
    },
  }
})
