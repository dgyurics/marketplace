import { createPinia } from 'pinia'
import { createApp } from 'vue'

import App from './App.vue'
import { createAppRouter } from './router'

import authDirective from '@/directives/auth'
import { useCartStore } from '@/store/cart'
import { initializeConfig } from '@/utilities/config'

import './assets/style.css'

async function initApp() {
  const app = createApp(App)
  const pinia = createPinia()

  app.use(pinia)

  // Register auth directive
  app.use(authDirective)

  // Cart loads in the background; it must not delay the first render.
  // The apiClient request interceptor refreshes an expired access token before each request
  // (concurrent requests share one refresh).
  useCartStore().fetchCart()

  // Config and categories (router) are required before render; load them concurrently.
  const [router] = await Promise.all([createAppRouter(), initializeConfig()])
  app.use(router)

  app.mount('#app')
}

initApp().catch(console.error)
