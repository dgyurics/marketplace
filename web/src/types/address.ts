export type Address = {
  id?: string
  country: string
  addressee?: string
  line1: string
  line2?: string
  city: string
  state: string
  postal_code: string
}

export type UpdateAddress = Required<Pick<Address, 'id'>> & Omit<Address, 'id'>
