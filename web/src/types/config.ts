import type { Locale } from './locale'

export type PaymentOptions = {
  stripe: boolean
  on_delivery: boolean
}

// Assistant features available in the UI, currently used for product autofill.
export type AssistantOptions = {
  enabled: boolean
}

export type ImageTransformOptions = {
  rembg: boolean
  openai: boolean
}

/**
 * Application metadata served by GET /config
 */
export type AppConfig = {
  assistant_options: AssistantOptions
  image_transform_options: ImageTransformOptions
  payment_options: PaymentOptions
  locale: Locale
}
