const DEFAULT_DATE_OPTIONS: Intl.DateTimeFormatOptions = {
  year: 'numeric',
  month: 'numeric',
  day: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
  hour12: false,
}

// TODO retrieve from environment/config
function getLocale(): string {
  return 'en-US'
}

/**
 * Format a date using the app's standard format
 */
export function formatDate(
  date: string | Date,
  options: Intl.DateTimeFormatOptions = DEFAULT_DATE_OPTIONS,
  locale?: string
): string {
  const dateObj = typeof date === 'string' ? new Date(date) : date
  const targetLocale = locale ?? getLocale()

  return dateObj.toLocaleString(targetLocale, options)
}

/**
 * Format date for display in tables/lists
 */
export function formatTableDate(date: string | Date): string {
  return formatDate(date)
}

/**
 * Format date with shorter format (no time)
 */
export function formatShortDate(date: string | Date): string {
  return formatDate(date, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}
