import { LOCALE } from '@/config'

// Based on supported countries in utilities/locale.go
export const LOCALE_MAP = {
  // Major English-speaking markets
  'en-US': { currency: 'USD', minorUnits: 2, country: 'US' },
  'en-CA': { currency: 'CAD', minorUnits: 2, country: 'CA' },
  'en-AU': { currency: 'AUD', minorUnits: 2, country: 'AU' },
  'en-GB': { currency: 'GBP', minorUnits: 2, country: 'GB' },
  'en-NZ': { currency: 'NZD', minorUnits: 2, country: 'NZ' },
  'en-ZA': { currency: 'ZAR', minorUnits: 2, country: 'ZA' },
  'en-IE': { currency: 'EUR', minorUnits: 2, country: 'IE' },
  'en-SG': { currency: 'SGD', minorUnits: 2, country: 'SG' },
  'en-HK': { currency: 'HKD', minorUnits: 2, country: 'HK' },

  // European markets
  'fr-FR': { currency: 'EUR', minorUnits: 2, country: 'FR' },
  'de-DE': { currency: 'EUR', minorUnits: 2, country: 'DE' },
  'de-AT': { currency: 'EUR', minorUnits: 2, country: 'AT' },
  'de-CH': { currency: 'CHF', minorUnits: 2, country: 'CH' },
  'es-ES': { currency: 'EUR', minorUnits: 2, country: 'ES' },
  'it-IT': { currency: 'EUR', minorUnits: 2, country: 'IT' },
  'nl-NL': { currency: 'EUR', minorUnits: 2, country: 'NL' },
  'pt-PT': { currency: 'EUR', minorUnits: 2, country: 'PT' },
  'fr-BE': { currency: 'EUR', minorUnits: 2, country: 'BE' },
  'pl-PL': { currency: 'PLN', minorUnits: 2, country: 'PL' },
  'da-DK': { currency: 'DKK', minorUnits: 2, country: 'DK' },
  'sv-SE': { currency: 'SEK', minorUnits: 2, country: 'SE' },
  'no-NO': { currency: 'NOK', minorUnits: 2, country: 'NO' },
  'fi-FI': { currency: 'EUR', minorUnits: 2, country: 'FI' },
  'cs-CZ': { currency: 'CZK', minorUnits: 2, country: 'CZ' },
  'hu-HU': { currency: 'HUF', minorUnits: 0, country: 'HU' },
  'ro-RO': { currency: 'RON', minorUnits: 2, country: 'RO' },
  'el-GR': { currency: 'EUR', minorUnits: 2, country: 'GR' },
  'hr-HR': { currency: 'EUR', minorUnits: 2, country: 'HR' },
  'sl-SI': { currency: 'EUR', minorUnits: 2, country: 'SI' },
  'sk-SK': { currency: 'EUR', minorUnits: 2, country: 'SK' },
  'et-EE': { currency: 'EUR', minorUnits: 2, country: 'EE' },
  'lv-LV': { currency: 'EUR', minorUnits: 2, country: 'LV' },
  'lt-LT': { currency: 'EUR', minorUnits: 2, country: 'LT' },
  'mt-MT': { currency: 'EUR', minorUnits: 2, country: 'MT' },
  'el-CY': { currency: 'EUR', minorUnits: 2, country: 'CY' },
  'is-IS': { currency: 'ISK', minorUnits: 0, country: 'IS' },
  'tr-TR': { currency: 'TRY', minorUnits: 2, country: 'TR' },

  // Additional European markets (small states)
  'de-LI': { currency: 'CHF', minorUnits: 2, country: 'LI' }, // Liechtenstein uses CHF (customs union with Switzerland)
  'fr-LU': { currency: 'EUR', minorUnits: 2, country: 'LU' }, // Luxembourg

  // Asian markets
  'ja-JP': { currency: 'JPY', minorUnits: 0, country: 'JP' },
  'ko-KR': { currency: 'KRW', minorUnits: 0, country: 'KR' },
  'zh-CN': { currency: 'CNY', minorUnits: 2, country: 'CN' },
  'zh-HK': { currency: 'HKD', minorUnits: 2, country: 'HK' },
  'zh-TW': { currency: 'TWD', minorUnits: 2, country: 'TW' },
  'th-TH': { currency: 'THB', minorUnits: 2, country: 'TH' },
  'vi-VN': { currency: 'VND', minorUnits: 0, country: 'VN' },
  'id-ID': { currency: 'IDR', minorUnits: 0, country: 'ID' },
  'ms-MY': { currency: 'MYR', minorUnits: 2, country: 'MY' },
  'tl-PH': { currency: 'PHP', minorUnits: 2, country: 'PH' },
  'hi-IN': { currency: 'INR', minorUnits: 2, country: 'IN' },
  'si-LK': { currency: 'LKR', minorUnits: 2, country: 'LK' },

  // Latin American markets
  'es-MX': { currency: 'MXN', minorUnits: 2, country: 'MX' },
  'pt-BR': { currency: 'BRL', minorUnits: 2, country: 'BR' },
  'es-AR': { currency: 'ARS', minorUnits: 2, country: 'AR' },
  'es-CL': { currency: 'CLP', minorUnits: 0, country: 'CL' },
  'es-CO': { currency: 'COP', minorUnits: 0, country: 'CO' },
  'es-PE': { currency: 'PEN', minorUnits: 2, country: 'PE' },
  'es-UY': { currency: 'UYU', minorUnits: 2, country: 'UY' },
  'es-PA': { currency: 'PAB', minorUnits: 2, country: 'PA' },

  // Middle East & Africa
  'ar-AE': { currency: 'AED', minorUnits: 2, country: 'AE' },
  'ar-SA': { currency: 'SAR', minorUnits: 2, country: 'SA' },
  'ar-EG': { currency: 'EGP', minorUnits: 2, country: 'EG' },
  'he-IL': { currency: 'ILS', minorUnits: 2, country: 'IL' },
  'ar-MA': { currency: 'MAD', minorUnits: 2, country: 'MA' },
  'sw-KE': { currency: 'KES', minorUnits: 2, country: 'KE' },
  'en-NG': { currency: 'NGN', minorUnits: 2, country: 'NG' },
  'en-GH': { currency: 'USD', minorUnits: 2, country: 'GH' }, // Ghana: GHS not in supported currencies, using USD
  'fr-CI': { currency: 'XOF', minorUnits: 0, country: 'CI' },
  'en-GI': { currency: 'GIP', minorUnits: 2, country: 'GI' },
} as const

export type SupportedLocale = keyof typeof LOCALE_MAP

/**
 * Get the application's configured locale
 * @returns The locale code (e.g., 'en-US')
 */
export function getAppLocale(): SupportedLocale {
  return LOCALE as SupportedLocale
}

/**
 * Get country code for a given locale
 * @param locale - The locale code (e.g., 'en-US')
 * @returns The country code (e.g., 'US')
 */
export function getCountryForLocale(locale: SupportedLocale): string {
  return LOCALE_MAP[locale].country
}

/**
 * Retrieve locale data for a given locale
 * @param locale - The locale code (e.g., 'en-US')
 * @returns An object containing currency, minorUnits, and country
 */
function getLocaleData(locale: SupportedLocale) {
  return LOCALE_MAP[locale]
}

/** Get currency code for a given locale
 * @param locale - The locale code (e.g., 'en-US')
 * @returns The currency code (e.g., 'USD')
 */
export function getCurrencyForLocale(locale: SupportedLocale) {
  return getLocaleData(locale).currency
}

/** Get minor units for a given locale
 * @param locale - The locale code (e.g., 'en-US')
 * @returns The number of minor units (e.g., 2 for USD)
 */
export function getMinorUnitsForLocale(locale: SupportedLocale): number {
  return getLocaleData(locale).minorUnits
}
