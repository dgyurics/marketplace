import type { Address } from './address'
import type { Product } from './product'

export interface OrderItem {
  product: Product
  thumbnail: string
  alt_text: string
  quantity: number
  unit_price: number
}

export type OrderStatus =
  'draft' | 'requested' | 'pending' | 'paid' | 'refunded' | 'shipped' | 'delivered' | 'canceled'

export type PaymentMethod = 'stripe' | 'on_delivery'

export type PaymentStatus = 'unpaid' | 'paid' | 'refunded'

export interface Order {
  id: string
  user_id: string
  address: Address
  items: OrderItem[]
  status?: OrderStatus
  amount?: number
  tax_amount?: number
  shipping_amount?: number
  total_amount?: number
  payment_method?: PaymentMethod
  payment_status?: PaymentStatus
  created_at: string
  updated_at: string
}

export interface InsufficientStockItem {
  product: Product
  quantity: number
  inventory: number
}

export interface CreateOrderConflict {
  success: false
  items: InsufficientStockItem[]
}

export interface PayOnDeliveryResponse {
  order_id: string
}

export interface PendingOrderResponse {
  order_id: string
}

export interface PendingOrderResult {
  success: true
  data: PendingOrderResponse
}

export interface StripeIntentResponse {
  client_secret: string
  order_id: string
  total_amount: number
}
