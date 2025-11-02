import { defineStore } from 'pinia'

import {
  createAddress as apiCreateAddress,
  updateAddress as apiUpdateAddress,
  createOrder as apiCreateOrder,
  confirmOrder as apiConfirmOrder,
  getTaxEstimate as apiGetTaxEstimate,
} from '@/services/api'
import type { Address, BillingAddress, Order } from '@/types'

export const useCheckoutStore = defineStore('checkout', {
  state: () => ({
    orderConfirmed: false,
    shippingAddress: {} as Address,
    billingAddress: {} as BillingAddress,
    order: {} as Order,
    stripe_client_secret: '',
    useShippingAddress: true,
  }),

  getters: {
    selectedBillingAddress: (state) =>
      state.useShippingAddress ? state.shippingAddress : state.billingAddress,
  },

  actions: {
    async initializeOrder(addressID: string) {
      this.order = await apiCreateOrder(addressID)
    },

    async saveShippingAddress(addressData: Address): Promise<Address> {
      // Check if we're updating an existing address or creating a new one
      const savedAddress = this.shippingAddress.id
        ? await apiUpdateAddress({ id: this.shippingAddress.id, ...addressData })
        : await apiCreateAddress(addressData)

      this.shippingAddress = savedAddress
      return savedAddress
    },

    async estimateTax(): Promise<void> {
      const estimate = await apiGetTaxEstimate(
        this.shippingAddress.state,
        this.shippingAddress.country
      )
      this.order.tax_amount = estimate.tax_amount
      this.order.total_amount = this.order.amount + estimate.tax_amount
    },

    async preparePayment(): Promise<string> {
      if (!this.order.id) {
        throw new Error('Order not found')
      }

      if (this.stripe_client_secret) {
        return this.stripe_client_secret
      }

      const { client_secret } = await apiConfirmOrder(this.order.id)
      this.stripe_client_secret = client_secret

      return client_secret
    },

    confirmOrder() {
      this.orderConfirmed = true
    },

    resetCheckout() {
      this.orderConfirmed = false
      this.useShippingAddress = true
      this.shippingAddress = {} as Address
      this.billingAddress = {} as Address
      this.order = {} as Order
      this.stripe_client_secret = ''
    },
  },
})
