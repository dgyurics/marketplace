import type { Locale } from './locale'

export type PaymentOptions = {
  stripe: boolean
  on_delivery: boolean
}

export type AIOptions = {
  image_editing: boolean
  field_suggestions: boolean
}

/**
 * Application metadata served by GET /config
 */
export type AppConfig = {
  ai_options: AIOptions
  payment_options: PaymentOptions
  locale: Locale
}
