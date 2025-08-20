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

export const register = async (
  email: string,
  password: string,
  inviteCode?: string
): Promise<AuthTokens> => {
  const response = await apiClient.post('/users/register', {
    email,
    password,
    invite_code: inviteCode,
  })
  return response.data
}

export const passwordReset = async (email: string): Promise<void> => {
  const response = await apiClient.post('/users/password-reset', { email })
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

export const getUserAddresses = async (): Promise<Address[]> => {
  const response = await apiClient.get('/addresses')
  return response.data
}

export const createAddress = async (address: Address): Promise<Address> => {
  const response = await apiClient.post('/addresses', address)
  return response.data
}

export const removeUserAddress = async (addressId: string): Promise<void> => {
  return apiClient.delete(`/addresses/${addressId}`)
}

export const getProducts = async (
  categories: string[],
  page: number = 1,
  limit: number = 10
): Promise<Product[]> => {
  let params = `?page=${page}&limit=${limit}`
  if (categories.length > 0) {
    params += `&${categories.map((category) => `category=${category}`).join('&')}`
  }
  const response = await apiClient.get(`/products${params}`)
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

export const updateCartItem = async (productId: string, quantity: number) => {
  const response = await apiClient.patch('/carts/items', { product_id: productId, quantity })
  return response.data
}

export const removeItemFromCart = async (productId: string) => {
  const response = await apiClient.delete(`/carts/items/${productId}`)
  return response.data
}

export const createOrder = async (): Promise<Order> => {
  const response = await apiClient.post('/orders', {})
  return response.data
}

export const getTaxEstimate = async (orderId: string): Promise<{ tax_amount: number }> => {
  const response = await apiClient.get(`/orders/${orderId}/tax-estimate`)
  return response.data
}

export const confirmOrder = async (orderId: string): Promise<StripePaymentIntent> => {
  const response = await apiClient.post(`/orders/${orderId}/confirm`)
  return response.data
}

export const updateOrder = async (
  orderId: string,
  addressId: string,
  email: string
): Promise<Order> => {
  const response = await apiClient.patch(`/orders/${orderId}`, {
    address_id: addressId,
    email,
  })
  return response.data
}

export const createCategory = async (category: Partial<Category>): Promise<Category> => {
  const response = await apiClient.post('/categories', category)
  return response.data
}

export const getCategories = async (): Promise<Category[]> => {
  const response = await apiClient.get('/categories')
  return response.data
}

export const removeCategory = async (categoryId: string): Promise<void> => {
  await apiClient.delete(`/categories/${categoryId}`)
}

export const createProduct = async (
  product: CreateProductRequest,
  categorySlug: string
): Promise<Product> => {
  const reponse = await apiClient.post(`/products/categories/${categorySlug}`, product)
  return reponse.data
}

export const removeProduct = async (productId: string): Promise<void> => {
  await apiClient.delete(`/products/${productId}`)
}

export const updateProduct = async (product: UpdateProductRequest): Promise<Product> => {
  const response = await apiClient.put('/products', product)
  return response.data
}
