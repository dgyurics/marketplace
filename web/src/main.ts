import { createPinia } from 'pinia'
import { createApp } from 'vue'

import App from './App.vue'
import { createAppRouter } from './router'

import authDirective from '@/directives/auth'
import { useCartStore } from '@/store/cart'
import { initializeLocale } from '@/utilities/locale'

import './assets/style.css'

async function initApp() {
  const app = createApp(App)
  const pinia = createPinia()

  app.use(pinia)

  // Register auth directive
  app.use(authDirective)

  // Initialize cart store
  // Pre-flight http interceptor, apiClient.interceptors, will initialize auth store as well
  // Feels hacky but works for now...
  await useCartStore().fetchCart()

  const router = await createAppRouter()
  app.use(router)

  // Initialize locale before creating the app
  await initializeLocale()

  app.mount('#app')
}

initApp().catch(console.error)
