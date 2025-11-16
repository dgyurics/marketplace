import axios from 'axios'

import { API_URL as baseURL, REQUEST_TIMEOUT as timeout } from '@/config'
import { useAuthStore } from '@/store/auth'
import type {
  CartItem,
  Address,
  Order,
  Category,
  StripePaymentIntent,
  AuthTokens,
  ImageType,
  Product,
  CreateProductRequest,
  UpdateProductRequest,
  ProductFilters,
  UserRecord,
  UpdateCategoryRequest,
  UpdateAddress,
  Locale,
} from '@/types'

const apiClient = axios.create({
  baseURL,
  timeout,
  headers: {
    'Content-Type': 'application/json',
  },
})

/**
 * Ensures the access token is valid and refreshes it if necessary.
 * @returns {Promise<string | null>} The access token if valid, otherwise null.
 */
async function ensureAccessToken(): Promise<string | null> {
  const authStore = useAuthStore()
  const { accessToken, setTokens, refreshToken, clearTokens, isTokenExpired } = authStore

  if (!refreshToken) return null

  if (accessToken && !isTokenExpired()) {
    return accessToken // Token is still valid
  }

  try {
    const authTokens: AuthTokens = await getNewAccessToken(refreshToken)
    setTokens(authTokens)
    return authTokens.token
  } catch (error) {
    console.debug('Token refresh failed:', error)
    clearTokens()
    return null
  }
}

apiClient.interceptors.request.use(async (config) => {
  const token = await ensureAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

export const login = async (email: string, password: string): Promise<AuthTokens> => {
  const response = await apiClient.post('/users/login', { email, password })
  return response.data
}

export const register = async (email: string): Promise<void> => {
  const response = await apiClient.post('/register', { email })
  return response.data
}

export const registerConfirm = async (
  email: string,
  password: string,
  registrationCode: string
): Promise<AuthTokens> => {
  const response = await apiClient.post('/register/confirm', {
    email,
    password,
    registration_code: registrationCode,
  })
  return response.data
}

// Send password reset email
export const passwordReset = async (email: string): Promise<void> => {
  const response = await apiClient.post('/users/password-reset', { email })
  return response.data
}

// Update password using reset code (code in email link)
export const passwordUpdate = async (
  email: string,
  newPassword: string,
  resetCode: string
): Promise<void> => {
  const response = await apiClient.post('/users/password-reset/confirm', {
    email,
    password: newPassword,
    reset_code: resetCode,
  })
  return response.data
}

export const updateCredentials = async (email: string, password: string): Promise<AuthTokens> => {
  const response = await apiClient.put('/users/credentials', { email, password })
  return response.data
}

export const logout = async () => {
  const response = await apiClient.post('/users/logout')
  return response.data
}

export const getNewAccessToken = async (refreshToken: string): Promise<AuthTokens> => {
  const response = await axios.post(`${baseURL}/users/refresh-token`, {
    refresh_token: refreshToken,
  })
  return response.data
}

export const createAddress = async (address: Address): Promise<Address> => {
  const response = await apiClient.post('/addresses', address)
  return response.data
}

export const updateAddress = async (address: UpdateAddress): Promise<Address> => {
  const response = await apiClient.put('/addresses', address)
  return response.data
}

export const removeUserAddress = async (addressId: string): Promise<void> => {
  return apiClient.delete(`/addresses/${addressId}`)
}

export const getProducts = async (filters: ProductFilters = {}): Promise<Product[]> => {
  const { categories = [], sortBy, inStock, page = 1, limit = 10 } = filters

  const params = new URLSearchParams()

  // Add pagination
  params.append('page', page.toString())
  params.append('limit', limit.toString())

  // Add categories
  categories.forEach((category: string) => params.append('category', category))

  // Add sorting
  if (sortBy) {
    params.append('sort_by', sortBy)
  }

  // Add in-stock filter
  if (inStock !== undefined) {
    params.append('in_stock', inStock.toString())
  }

  const response = await apiClient.get(`/products?${params}`)
  return response.data
}

export const uploadImage = async (
  productId: string,
  file: File,
  type: ImageType,
  removeBg: Boolean
): Promise<{ path: string }> => {
  const formData = new FormData()
  formData.append('image', file)
  formData.append('type', type)
  const response = await apiClient.post(
    `/images/products/${productId}?remove_bg=${removeBg}`,
    formData,
    {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    }
  )
  return response.data
}

export const removeImage = async (id: string): Promise<void> => {
  await apiClient.delete(`/images/${id}`)
}

export const promoteImage = async (id: string): Promise<void> => {
  await apiClient.post(`/images/${id}`)
}

export const getProductById = async (id: string): Promise<Product> => {
  const response = await apiClient.get(`/products/${id}`)
  return response.data
}

export const createGuestUser = async (): Promise<AuthTokens> => {
  const response = await apiClient.post('/users/guest')
  return response.data
}

export const getCart = async (): Promise<CartItem[]> => {
  const response = await apiClient.get('/carts')
  return response.data // Updated to return the array directly
}

export const addItemToCart = async (productId: string, quantity: number) => {
  const response = await apiClient.post(`/carts/items/${productId}`, { quantity })
  return response.data
}

export const removeItemFromCart = async (productId: string) => {
  const response = await apiClient.delete(`/carts/items/${productId}`)
  return response.data
}

export const createOrder = async (shippingID: string): Promise<StripePaymentIntent> => {
  const params = new URLSearchParams()
  params.append('shipping_id', shippingID)

  const response = await apiClient.post(`/orders?${params}`)
  return response.data
}

export const getUsers = async (page: number = 1, limit: number = 50): Promise<UserRecord[]> => {
  const params = new URLSearchParams()

  // Add pagination
  params.append('page', page.toString())
  params.append('limit', limit.toString())

  const response = await apiClient.get(`/users?${params}`)
  return response.data
}

export const getOrders = async (page: number = 1, limit: number = 50): Promise<Order[]> => {
  const params = new URLSearchParams()

  // Add pagination
  params.append('page', page.toString())
  params.append('limit', limit.toString())

  const response = await apiClient.get(`/orders?${params}`)
  return response.data
}

export const getOrderPublic = async (orderId: string): Promise<Order> => {
  const response = await apiClient.post(`/orders/${orderId}/public`)
  return response.data
}

export const getOrderOwner = async (orderId: string): Promise<Order> => {
  const response = await apiClient.post(`/orders/${orderId}/owner`)
  return response.data
}

export const getOrderAdmin = async (orderId: string): Promise<Order> => {
  const response = await apiClient.post(`/orders/${orderId}/admin`)
  return response.data
}

export const getTaxEstimate = async (
  country: string,
  state?: string
): Promise<{ tax_amount: number }> => {
  const params = new URLSearchParams()
  params.append('country', country)

  if (state) {
    params.append('state', state)
  }

  const response = await apiClient.get(`/tax/estimate?${params}`)
  return response.data
}

export const createCategory = async (category: Partial<Category>): Promise<Category> => {
  const response = await apiClient.post('/categories', category)
  return response.data
}

export const updateCategory = async (category: UpdateCategoryRequest): Promise<Category> => {
  const response = await apiClient.put('/categories', category)
  return response.data
}

export const getCategories = async (): Promise<Category[]> => {
  const response = await apiClient.get('/categories')
  return response.data
}

export const getCategoryById = async (categoryId: string): Promise<Category> => {
  const response = await apiClient.get(`/categories/${categoryId}`)
  return response.data
}

export const removeCategory = async (categoryId: string): Promise<void> => {
  await apiClient.delete(`/categories/${categoryId}`)
}

export const createProduct = async (product: CreateProductRequest): Promise<Product> => {
  const reponse = await apiClient.post('/products', product)
  return reponse.data
}

export const removeProduct = async (productId: string): Promise<void> => {
  await apiClient.delete(`/products/${productId}`)
}

export const updateProduct = async (product: UpdateProductRequest): Promise<Product> => {
  const response = await apiClient.put('/products', product)
  return response.data
}

export const getLocale = async (): Promise<Locale> => {
  const response = await apiClient.get('/locale')
  return response.data
}
