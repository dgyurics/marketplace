import { defineStore } from 'pinia'

import {
  createAddress as apiCreateAddress,
  updateAddress as apiUpdateAddress,
  createOrder as apiCreateOrder,
  getTaxEstimate as apiGetTaxEstimate,
} from '@/services/api'
import type { Address } from '@/types'

export const useCheckoutStore = defineStore('checkout', {
  state: () => ({
    shippingAddress: {} as Address,
    stripe_client_secret: '',
    shippingError: null as string | null,
  }),

  getters: {
    isShippingAddressComplete: (state) => Boolean(state.shippingAddress.id),
  },

  actions: {
    async saveShippingAddress(addressData: Address): Promise<Address> {
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

    async preparePayment(): Promise<string> {
      if (this.stripe_client_secret) {
        return this.stripe_client_secret
      }
      if (!this.shippingAddress.id) {
        throw new Error('Shipping address not found')
      }

      const { client_secret } = await apiCreateOrder(this.shippingAddress.id)
      this.stripe_client_secret = client_secret

      return client_secret
    },

    setShippingError(message: string) {
      this.shippingError = message
    },
    clearShippingError() {
      this.shippingError = null
    },

    resetCheckout() {
      this.shippingAddress = {} as Address
      this.stripe_client_secret = ''
    },
  },
})
