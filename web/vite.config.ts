import vue from '@vitejs/plugin-vue'
import path from 'path'
import { fileURLToPath } from 'url'
import { defineConfig } from 'vite'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8000', // Go backend URL
        changeOrigin: true,
        secure: false, // Disable SSL verification if backend uses HTTPS with self-signed certs
        rewrite: (path) => path.replace(/^\/api/, ''), // Remove "/api" prefix
      },
    },
  },
  logLevel: 'info',
})
