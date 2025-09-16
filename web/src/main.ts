import { createPinia } from 'pinia'
import { createApp } from 'vue'

import App from './App.vue'
import { createAppRouter } from './router'
import './assets/style.css'

async function initApp() {
  const app = createApp(App)
  const pinia = createPinia()

  app.use(pinia)

  const router = await createAppRouter()
  app.use(router)

  app.mount('#app')
}

initApp().catch(console.error)
