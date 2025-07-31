interface ImportMetaEnv {
  readonly VITE_API_URL: string
  readonly VITE_STRIPE_PUBLISHABLE_KEY: string
  readonly VITE_COUNTRY: string
  readonly VITE_REQUEST_TIMEOUT?: string
  // Vite built-in properties
  readonly MODE: string
  readonly BASE_URL: string
  readonly PROD: boolean
  readonly DEV: boolean
  readonly SSR: boolean
}

const env = import.meta.env as ImportMetaEnv

export const API_URL = env.VITE_API_URL
export const STRIPE_PUBLISHABLE_KEY = env.VITE_STRIPE_PUBLISHABLE_KEY
export const COUNTRY = env.VITE_COUNTRY
export const REQUEST_TIMEOUT = env.VITE_REQUEST_TIMEOUT
  ? parseInt(env.VITE_REQUEST_TIMEOUT, 10)
  : 30000
