import { defineStore } from 'pinia'

import {
  createAddress as apiCreateAddress,
  updateAddress as apiUpdateAddress,
  createOrder as apiCreateOrder,
  confirmOrder as apiConfirmOrder,
  getTaxEstimate as apiGetTaxEstimate,
  updateOrder as apiUpdateOrder,
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
    canProceedToPayment: (state) => state.shippingAddress.id && state.order.id,

    selectedBillingAddress: (state) =>
      state.useShippingAddress ? state.shippingAddress : state.billingAddress,
  },

  actions: {
    async initializeOrder() {
      const order = await apiCreateOrder()
      this.order = order

      // Populate existing address and email if they exist in the order
      if (order.address) {
        this.shippingAddress = order.address
      }

      return this.order
    },

    async saveShippingAddress(addressData: Address) {
      // Check if we're updating an existing address or creating a new one
      const savedAddress = this.shippingAddress.id
        ? await apiUpdateAddress({ id: this.shippingAddress.id, ...addressData })
        : await apiCreateAddress(addressData)

      this.shippingAddress = savedAddress

      // Update order with shipping address and email
      if (this.order.id && this.shippingAddress.id) {
        this.order = await apiUpdateOrder(this.order.id, this.shippingAddress.id)

        // Estimate tax with the new address
        await this.estimateTax()
      }

      return this.shippingAddress
    },

    async estimateTax(): Promise<void> {
      if (!this.order.id || !this.shippingAddress.id) {
        return
      }

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
