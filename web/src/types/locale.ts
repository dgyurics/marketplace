export type Locale = {
  country_code: string
  country: string
  line2_label: string
  postal_code_label: string
  postal_code_pattern: string
  state_label: string
  state_required: boolean
  state_codes?: { [key: string]: string }
  currency: string
  minor_units: number
  language: string
}
