import type { Address } from './address'
import type { Product } from './product'

export interface OrderItem {
  product: Product
  thumbnail: string
  alt_text: string
  quantity: number
  unit_price: number
}

export type OrderStatus = 'pending' | 'paid' | 'refunded' | 'shipped' | 'delivered' | 'canceled'

export interface Order {
  id: string
  user_id: string
  address: Address
  items: OrderItem[]
  status: OrderStatus
  amount: number
  tax_amount: number
  shipping_amount: number
  total_amount: number
  created_at: string
  updated_at: string
}

export interface InsufficientStockItem {
  product: Product
  quantity: number
  inventory: number
}

export interface CreateOrderResult {
  success: true
  data: CreateOrderResponse
}

export interface CreateOrderConflict {
  success: false
  items: InsufficientStockItem[]
}

export interface CreateOrderResponse {
  client_secret: string
  order_id: string
}
