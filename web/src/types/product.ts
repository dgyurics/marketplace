import type { Category } from './category'

export interface Product {
  id: string
  name: string
  price: number
  summary: string
  description: string
  details: Record<string, unknown>
  images: Image[]
  category?: Category
  tax_code?: string
  inventory: number
  created_at?: string
  updated_at?: string
}

export type SortBy = 'price' | 'popularity' | 'newest'

export interface ProductFilters {
  categories?: string[]
  sortBy?: SortBy
  inStock?: boolean
  page?: number
  limit?: number
}

export interface CreateProductRequest {
  name: string
  price: number
  details: Record<string, unknown>
  description: string
  tax_code?: string
}

type RequireOnly<T, K extends keyof T> = Partial<T> & Pick<T, K>

export type UpdateProductRequest = RequireOnly<Product, 'id'>

export type ImageType = 'hero' | 'thumbnail' | 'gallery'

export interface Image {
  id: string
  product_id?: string // FIXME pretty sure this never is used/exists
  url: string
  type: ImageType
  alt_text?: string | null
}

export interface CartItem {
  quantity: number
  product: Product
  unit_price: number
}
