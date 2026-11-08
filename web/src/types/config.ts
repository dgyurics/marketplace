import type { Locale } from './locale'

export type PaymentOptions = {
  stripe: boolean
  on_delivery: boolean
}

export type ImageTransformOptions = {
  rembg: boolean
  openai: boolean
}

/**
 * Application metadata served by GET /config
 */
export type AppConfig = {
  assistant_enabled: boolean
  image_transform_options: ImageTransformOptions
  payment_options: PaymentOptions
  locale: Locale
}
