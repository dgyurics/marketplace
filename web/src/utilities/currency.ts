import { getLocale } from './locale'

/**
 * Get currency configuration (lazy-loaded)
 */
function getCurrencyConfig() {
  const locale = getLocale()
  return {
    code: locale.currency,
    minorUnits: locale.minor_units,
  }
}

const DEFAULT_CURRENCY_OPTIONS: Intl.NumberFormatOptions = {
  style: 'currency',
  get currency() {
    return getCurrencyConfig().code
  },
} as const

/**
 * Get currency code for the application's configured locale
 * @returns The currency code (e.g., 'USD')
 */
export function getCurrencyCode(): string {
  return getCurrencyConfig().code
}

/**
 * Cache for NumberFormat instances to avoid recreating them
 */
const formatters = new Map<string, Intl.NumberFormat>()

/**
 * Get or create a cached NumberFormat instance
 * @param locale - The locale code (e.g., 'en-US')
 * @param options - Intl.NumberFormat options
 * @returns A cached Intl.NumberFormat instance
 */
function getFormatter(locale: string, options: Intl.NumberFormatOptions): Intl.NumberFormat {
  const key = `${locale}-${JSON.stringify(options)}`
  if (!formatters.has(key)) {
    formatters.set(key, new Intl.NumberFormat(locale, options))
  }
  return formatters.get(key)!
}

// notation: 'compact': Order totals in admin dashboard ("$1.2M in sales")
// signDisplay: showing price changes or discounts ("+$5 increase")
// unit / unitDisplay: product specifications (weight: "2.5 kg", dimensions: "15 cm")

/**
 * Convert minor units (cents) to major units (dollars) and format as currency
 * @param amount - Amount in minor units (e.g., cents)
 * @param options - Intl.NumberFormat options to customize formatting
 * @param locale - Locale code (e.g., 'en-US'); defaults to app locale if not provided
 * @returns Formatted currency string (e.g., "$1,234.56")
 */
export function formatCurrency(
  amount: number,
  options: Intl.NumberFormatOptions = {},
  locale?: string
): string {
  // 300 cents -> 3.00 dollars
  const majorAmount = amount / Math.pow(10, getCurrencyConfig().minorUnits)
  const targetLocale = locale ?? getLocale().language

  const formatOptions = {
    ...DEFAULT_CURRENCY_OPTIONS,
    ...options,
  }

  const formatter = getFormatter(targetLocale, formatOptions)
  return formatter.format(majorAmount)
}

/**
 * Same as formatPrice, but with special handling for zero amounts
 * @param amount - Amount in minor units (e.g., cents)
 * @returns "free" if amount is zero, otherwise formatted price (e.g., "$1,234.56")
 */
export function displayPrice(amount: number): string {
  return amount === 0 ? 'free' : formatPrice(amount)
}

/**
 * Simple currency formatting with symbol (most common use case)
 * @param amount - Amount in minor units (e.g., cents)
 * @returns Formatted price with currency symbol (e.g., "$1,234.56")
 */
export function formatPrice(amount: number): string {
  return formatCurrency(amount)
}

/**
 * Format currency without symbol
 * @param amount - Amount in minor units (e.g., cents)
 * @returns Formatted amount without currency symbol (e.g., "1,234.56")
 */
export function formatAmount(amount: number): string {
  return formatCurrency(amount, { style: 'decimal' })
}

/**
 * Convert major units (dollars) to minor units (cents)
 * @param amount - Amount in major units (e.g., dollars)
 * @returns Amount in minor units (e.g., cents)
 */
export function toMinorUnits(amount: number): number {
  return Math.round(amount * Math.pow(10, getCurrencyConfig().minorUnits))
}

/**
 * Convert minor units (cents) to major units (dollars)
 * @param amount - Amount in minor units (e.g., cents)
 * @returns Amount in major units (e.g., dollars)
 */
export function toMajorUnits(amount: number): number {
  return amount / Math.pow(10, getCurrencyConfig().minorUnits)
}
