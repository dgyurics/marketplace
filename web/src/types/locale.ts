export type Locale = {
  country_code: string
  country: string
  // TODO add custom label for line 2
  // line2_label: string // e.g. Apt # / Suite
  postal_code_label: string
  postal_code_pattern: string
  state_label: string
  state_required: boolean
  state_codes?: { [key: string]: string }
  currency: string
  minor_units: number
  language: string
}
