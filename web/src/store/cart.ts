import { defineStore } from 'pinia'

import {
  getCart as apiGetCart,
  addItemToCart as apiAddItemToCart,
  removeItemFromCart as apiRemoveItemFromCart,
} from '@/services/api'
import type { CartItem } from '@/types'

export const useCartStore = defineStore('cart', {
  state: () => ({
    items: [] as CartItem[],
  }),

  getters: {
    itemCountByProductId: (state) => {
      return (productId: string) =>
        state.items.find((item) => item.product.id === productId)?.quantity ?? 0
    },

    subtotal: (state) =>
      state.items.reduce((total, item) => total + item.unit_price * item.quantity, 0),
  },

  actions: {
    async fetchCart() {
      try {
        const cartItems = await apiGetCart()
        this.items = cartItems

        return this.items
      } catch (err) {
        console.error('Error fetching cart:', err)
        this.items = []
        return []
      }
    },

    async addToCart(productId: string, quantity: number) {
      try {
        // TODO: re-design cart endpoints to return the updated cart
        await apiAddItemToCart(productId, quantity)
        await this.fetchCart() // Refresh cart after adding item
      } catch (err) {
        console.error('Error adding item to cart:', err)
        throw err
      }
    },

    async removeFromCart(productId: string) {
      try {
        await apiRemoveItemFromCart(productId)
        await this.fetchCart() // Refresh cart after removal
      } catch (err) {
        console.error('Error removing item from cart:', err)
        throw err
      }
    },
  },
})
