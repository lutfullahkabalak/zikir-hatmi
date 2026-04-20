import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import ui from '@nuxt/ui/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), tailwindcss(), ui()],
  server: {
    host: true,
    allowedHosts: [
      'localhost',
      '127.0.0.1',
      'zikirhatmi.abapnews.tr',
      '.abapnews.info',
    ],
    proxy: {
      '/hatims': {
        target: process.env.VITE_PROXY_TARGET || 'http://localhost:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: (process.env.VITE_PROXY_TARGET || 'http://localhost:8080').replace('http', 'ws'),
        ws: true,
        changeOrigin: true,
      },
    },
  },
})
