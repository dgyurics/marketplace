import { defineStore } from 'pinia'

import {
  createAddress as apiCreateAddress,
  updateAddress as apiUpdateAddress,
  createOrder as apiCreateOrder,
  getTaxEstimate as apiGetTaxEstimate,
} from '@/services/api'
import type { Address, CreateOrderResponse } from '@/types'

export const useCheckoutStore = defineStore('checkout', {
  state: () => ({
    shippingAddress: {} as Address,
    stripe_client_secret: '',
    order_id: '',
    shippingError: null as string | null,
    paymentError: null as string | null,
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

    async preparePayment(): Promise<CreateOrderResponse> {
      if (this.stripe_client_secret) {
        return { client_secret: this.stripe_client_secret, order_id: this.order_id }
      }
      if (!this.shippingAddress.id) {
        throw new Error('Shipping address not found')
      }

      const resBody: CreateOrderResponse = await apiCreateOrder(this.shippingAddress.id)
      this.stripe_client_secret = resBody.client_secret
      this.order_id = resBody.order_id

      return resBody
    },

    resetCheckout() {
      this.shippingAddress = {} as Address
      this.stripe_client_secret = ''
      this.order_id = ''
      this.shippingError = null
      this.paymentError = null
    },
  },
})
