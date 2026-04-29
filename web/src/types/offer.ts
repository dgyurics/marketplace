import type { Product } from './product'

export type OfferStatus = 'pending' | 'accepted' | 'rejected' | 'canceled' | 'completed'
export interface Offer {
  id: string
  user_id: string
  product: Product
  amount: number
  comment: string
  status: OfferStatus
  created_at: string
  updated_at: string
}
