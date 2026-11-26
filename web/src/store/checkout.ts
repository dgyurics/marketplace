import { defineStore } from 'pinia'

import {
  createAddress as apiCreateAddress,
  updateAddress as apiUpdateAddress,
  createOrder as apiCreateOrder,
  createStripeIntent as apiCreateStripeIntent,
  confirmPayOnDelivery as apiConfirmPayOnDelivery,
  getTaxEstimate as apiGetTaxEstimate,
} from '@/services/api'
import type {
  Address,
  InsufficientStockItem,
  PayOnDeliveryResponse,
  PendingOrderResponse,
  StripeIntentResponse,
} from '@/types'

export const useCheckoutStore = defineStore('checkout', {
  state: () => ({
    shippingAddress: {} as Address,
    stripe_client_secret: '',
    stripe_total_amount: 0,
    order_id: '',
    shippingError: null as string | null,
    paymentError: null as string | null,
    insufficientStock: [] as InsufficientStockItem[],
  }),

  getters: {
    isShippingAddressComplete: (state) => Boolean(state.shippingAddress.id),
  },

  actions: {
    async saveShippingAddress(addressData: Address): Promise<Address> {
      this.shippingError = null // Clear any previous error before saving

      // Normalize country and state to uppercase
      addressData.country = addressData.country.toUpperCase()
      if (addressData.state) {
        addressData.state = addressData.state.toUpperCase()
      }

      // Check if we're updating an existing address or creating a new one
      const savedAddress = this.shippingAddress.id
        ? await apiUpdateAddress({ id: this.shippingAddress.id, ...addressData })
        : await apiCreateAddress(addressData)

      this.shippingAddress = savedAddress
      return savedAddress
    },

    async estimateTax(): Promise<{ tax_amount: number }> {
      return apiGetTaxEstimate(this.shippingAddress.country, this.shippingAddress.state)
    },

    async createPendingOrder(): Promise<PendingOrderResponse | null> {
      if (!this.shippingAddress.id) {
        throw new Error('Shipping address not found')
      }

      const result = await apiCreateOrder(this.shippingAddress.id, window.crypto.randomUUID())

      if (!result.success) {
        this.insufficientStock = result.items
        return null
      }

      this.order_id = result.data.order_id
      this.insufficientStock = []
      this.stripe_client_secret = ''
      this.stripe_total_amount = 0

      return result.data
    },

    /**
     * Creates the PaymentIntent. Called only once the user submits the Stripe
     * form. The result is cached so a retry after a declined card reuses the
     * same intent
     */
    async fetchStripeIntent(): Promise<StripeIntentResponse> {
      if (!this.order_id) {
        throw new Error('Order not found')
      }

      if (this.stripe_client_secret) {
        return {
          client_secret: this.stripe_client_secret,
          order_id: this.order_id,
          total_amount: this.stripe_total_amount,
        }
      }

      const res = await apiCreateStripeIntent(this.order_id, this.order_id)
      this.stripe_client_secret = res.client_secret
      this.stripe_total_amount = res.total_amount

      return res
    },

    /** Set order payment type NULL -> ON_DELIVERY */
    async confirmPayOnDelivery(): Promise<PayOnDeliveryResponse> {
      if (!this.order_id) {
        throw new Error('Order not found')
      }

      return apiConfirmPayOnDelivery(this.order_id)
    },

    resetCheckout() {
      this.shippingAddress = {} as Address
      this.stripe_client_secret = ''
      this.stripe_total_amount = 0
      this.order_id = ''
      this.shippingError = null
      this.paymentError = null
      this.insufficientStock = []
    },
  },
})
