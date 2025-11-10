export type Address = {
  id?: string
  addressee?: string
  line1: string
  line2?: string
  city: string
  state?: string
  postal_code: string
  country: string
  email: string
}

export type UpdateAddress = Required<Pick<Address, 'id'>> & Omit<Address, 'id'>

export type BillingAddress = Omit<Address, 'email'>
