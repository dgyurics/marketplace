export type Address = {
  id?: string
  name?: string
  line1: string
  line2?: string
  city: string
  state?: string
  postal_code: string
  country: string
  email: string
}

export type UpdateAddress = Required<Pick<Address, 'id'>> & Omit<Address, 'id'>
