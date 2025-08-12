import type { Product } from './product'

export interface StripePaymentIntent {
  client_secret: string
}

export interface OrderItem {
  product: Product
  quantity: number
  unit_price: number
  description?: string
  thumbnail: string
}

export interface Order {
  id: string
  user_id?: string
  address_id?: string
  currency: string
  amount: number
  tax_amount: number
  total_amount: number
  status: 'pending' | 'paid' | 'refunded' | 'fulfilled' | 'shipped' | 'delivered' | 'canceled'
  items?: OrderItem[]
}
