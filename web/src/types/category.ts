export interface Category {
  id: string
  name: string
  slug: string
  description?: string
  parent_id?: string
  created_at?: string
  updated_at?: string
}

// FIXME move this to some shared types file
// exists in product.ts too
type RequireOnly<T, K extends keyof T> = Partial<T> & Pick<T, K>

export type UpdateCategoryRequest = RequireOnly<Category, 'id'>
