import { defineStore } from 'pinia'

import {
  createAddress as apiCreateAddress,
  updateAddress as apiUpdateAddress,
  createOrder as apiCreateOrder,
  getTaxEstimate as apiGetTaxEstimate,
} from '@/services/api'
import type { Address, BillingAddress } from '@/types'

export const useCheckoutStore = defineStore('checkout', {
  state: () => ({
    shippingAddress: {} as Address,
    billingAddress: {} as BillingAddress,
    stripe_client_secret: '',
    useShippingAddress: true,
  }),

  getters: {
    selectedBillingAddress: (state) =>
      state.useShippingAddress ? state.shippingAddress : state.billingAddress,

    isShippingAddressComplete: (state) => {
      const addr = state.shippingAddress
      return Boolean(
        addr?.line1?.trim() &&
          addr?.city?.trim() &&
          addr?.state?.trim() &&
          addr?.postal_code?.trim() &&
          addr?.country?.trim()
      )
    },
  },

  actions: {
    async saveShippingAddress(addressData: Address): Promise<Address> {
      // Check if we're updating an existing address or creating a new one
      const savedAddress = this.shippingAddress.id
        ? await apiUpdateAddress({ id: this.shippingAddress.id, ...addressData })
        : await apiCreateAddress(addressData)

      this.shippingAddress = savedAddress
      return savedAddress
    },

    async estimateTax(): Promise<{ tax_amount: number }> {
      return apiGetTaxEstimate(this.shippingAddress.state, this.shippingAddress.country)
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

    resetCheckout() {
      this.useShippingAddress = true
      this.shippingAddress = {} as Address
      this.billingAddress = {} as Address
      this.stripe_client_secret = ''
    },
  },
})
