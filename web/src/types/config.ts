import type { Locale } from './locale'

export type PaymentOptions = {
  stripe: boolean
  pay_on_delivery: boolean
}

/**
 * Application metadata served by GET /config
 */
export type AppConfig = {
  payment_options: PaymentOptions
  locale: Locale
}
