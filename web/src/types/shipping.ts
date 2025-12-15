export interface ShippingZone {
  id?: string
  country: string
  state?: string | null
  postal_code?: string | null
}

export interface ExcludedShippingZone {
  id?: string
  country: string
  postal_code: string
}
