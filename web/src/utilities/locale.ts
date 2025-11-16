import { getLocale as fetchLocale } from '@/services/api'
import type { Locale } from '@/types/locale'

let localeData: Locale | null = null

/**
 * Initialize locale data - call this once at app startup
 */
export async function initializeLocale(): Promise<void> {
  localeData ??= await fetchLocale()
}

/**
 * Get cached locale data (synchronous)
 * @returns The locale data
 * @throws Error if locale not initialized
 */
export function getLocale(): Locale {
  if (!localeData) {
    throw new Error('Locale not initialized. Call initializeLocale() first.')
  }
  return localeData
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
