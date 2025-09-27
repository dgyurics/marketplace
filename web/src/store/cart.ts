import { defineStore } from 'pinia'

import {
  getCart as apiGetCart,
  addItemToCart as apiAddItemToCart,
  updateCartItem as apiUpdateCartItem,
  removeItemFromCart as apiRemoveItemFromCart,
} from '@/services/api'
import type { CartItem } from '@/types'

export const useCartStore = defineStore('cart', {
  state: () => ({
    items: [] as CartItem[],
    loading: false,
    error: '',
  }),

  getters: {
    itemCountByProductId: (state) => {
      return (productId: string) =>
        state.items.find((item) => item.product.id === productId)?.quantity ?? 0
    },

    subtotal: (state) =>
      state.items.reduce((total, item) => total + item.unit_price * item.quantity, 0),

    isEmpty: (state) => state.items.length === 0,
  },

  actions: {
    async fetchCart() {
      try {
        this.loading = true
        this.error = ''

        const cartItems = await apiGetCart()
        this.items = cartItems

        return this.items
      } catch (err) {
        this.error = 'Failed to fetch cart'
        console.error('Error fetching cart:', err)
        this.items = []
        return []
      } finally {
        this.loading = false
      }
    },

    async addToCart(productId: string, quantity: number) {
      try {
        this.loading = true
        this.error = ''

        // TODO: re-design cart endpoints to return the updated cart
        await apiAddItemToCart(productId, quantity)
        await this.fetchCart() // Refresh cart after adding item
      } catch (err) {
        this.error = 'Failed to add item to cart'
        console.error('Error adding item to cart:', err)
        throw err
      } finally {
        this.loading = false
      }
    },

    async updateItemQuantity(productId: string, quantity: number) {
      try {
        this.loading = true
        this.error = ''

        await apiUpdateCartItem(productId, quantity)
        await this.fetchCart() // Refresh cart after updating item
      } catch (err) {
        this.error = 'Failed to update cart item'
        console.error('Error updating cart item:', err)
        throw err
      } finally {
        this.loading = false
      }
    },

    async removeFromCart(productId: string) {
      try {
        this.loading = true
        this.error = ''

        await apiRemoveItemFromCart(productId)
        await this.fetchCart() // Refresh cart after removal
      } catch (err) {
        this.error = 'Failed to remove item from cart'
        console.error('Error removing item from cart:', err)
        throw err
      } finally {
        this.loading = false
      }
    },

    clearCart() {
      this.items = []
      this.error = ''
    },

    clearError() {
      this.error = ''
    },
  },
})
