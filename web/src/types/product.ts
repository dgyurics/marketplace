export interface Product {
  id: string
  name: string
  price: number
  details: Record<string, unknown>
  description: string
  images: Image[]
  tax_code?: string
  created_at?: string
  updated_at?: string
}

export interface ProductWithInventory extends Product {
  quantity: number
}

// For creating new products (no ID, images, timestamps)
export interface CreateProductRequest {
  name: string
  price: number
  details: Record<string, unknown>
  description: string
  tax_code?: string
}

export interface UpdateProductRequest {
  name?: string
  price?: number
  details?: Record<string, unknown>
  description?: string
  tax_code?: string
}

export type ImageType = 'hero' | 'thumbnail' | 'gallery'

export interface Image {
  id: string
  product_id?: string
  url: string
  type: ImageType
  display_order: number
  alt_text?: string | null
}

export interface CartItem {
  quantity: number
  product: Product
  unit_price: number
}
