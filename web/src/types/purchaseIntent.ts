import type { Product } from './product'

export type PurchaseIntentStatus = 'pending' | 'accepted' | 'rejected' | 'canceled' | 'completed'
export interface PurchaseIntent {
  id: string
  user_id: string
  product: Product
  offer_price: number
  pickup_notes: string
  status: PurchaseIntentStatus
  created_at: string
  updated_at: string
}
