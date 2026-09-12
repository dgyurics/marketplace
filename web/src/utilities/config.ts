import { getConfig as fetchConfig } from '@/services/api'
import type { AppConfig, Locale, PaymentOptions } from '@/types'

let appConfig: AppConfig | null = null

/**
 * Initialize app config - call this once at app startup
 */
export async function initializeConfig(): Promise<void> {
  appConfig ??= await fetchConfig()
}

/**
 * Get cached app config (synchronous)
 * @returns The app config
 * @throws Error if config not initialized
 */
export function getAppConfig(): AppConfig {
  if (!appConfig) {
    throw new Error('Config not initialized. Call initializeConfig() first.')
  }
  return appConfig
}

/**
 * Get cached locale data (synchronous)
 * @returns The locale data
 * @throws Error if locale not initialized
 */
export function getLocale(): Locale {
  return getAppConfig().locale
}

/**
 * Get the payment options enabled for this deployment
 */
export function getPaymentOptions(): PaymentOptions {
  return getAppConfig().payment_options
}

/**
 * Get currency for current locale
 */
export function getCurrency(): string {
  return getLocale().currency
}

/**
 * Get minor units for current locale
 */
export function getMinorUnits(): number {
  return getLocale().minor_units
}

/**
 * Validate postal code using locale pattern
 */
export function validatePostalCode(postalCode: string): boolean {
  const locale = getLocale()
  const regex = new RegExp(locale.postal_code_pattern)
  return regex.test(postalCode)
}
