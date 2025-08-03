import vue from '@vitejs/plugin-vue'
import path from 'path'
import { fileURLToPath } from 'url'
import { defineConfig, type UserConfig } from 'vite'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

export default defineConfig(({ command }) => {
  const config: UserConfig = {
    plugins: [vue()],
    server: {
      host: true,
      port: 5173,
    },
    resolve: {
      alias: {
        '@': path.resolve(__dirname, 'src'),
      },
    },
    logLevel: 'warn',
  }

  // Development-specific configuration
  if (command === 'serve') {
    config.envDir = '../deploy/local'
    config.logLevel = 'info'
  }

  return config
})
