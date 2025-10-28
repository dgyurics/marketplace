import { defineStore } from 'pinia'

import {
  createAddress as apiCreateAddress,
  createOrder as apiCreateOrder,
  confirmOrder as apiConfirmOrder,
  getTaxEstimate as apiGetTaxEstimate,
  updateOrder as apiUpdateOrder,
} from '@/services/api'
import type { Address, Order } from '@/types'

export const useCheckoutStore = defineStore('checkout', {
  state: () => ({
    email: '',
    orderConfirmed: false,
    loading: false,
    error: '',
    shippingAddress: <Address>{
      id: '',
      addressee: '',
      line1: '',
      line2: '',
      city: '',
      state: '',
      postal_code: '',
    },
    billingAddress: <Address>{
      id: '',
      addressee: '',
      line1: '',
      line2: '',
      city: '',
      state: '',
      postal_code: '',
    },
    order: <Order>{
      id: '',
      email: '',
      items: [],
      status: 'pending',
      currency: '',
      amount: 0,
      tax_amount: 0,
      shipping_amount: 0,
      total_amount: 0,
      created_at: '',
      updated_at: '',
    },
    stripe_client_secret: '',
    useShippingAddress: true,
  }),

  getters: {
    canProceedToPayment: (state) => state.shippingAddress.id && state.email && state.order.id,

    selectedBillingAddress: (state) =>
      state.useShippingAddress ? state.shippingAddress : state.billingAddress,
  },

  actions: {
    async initializeOrder() {
      try {
        this.loading = true
        this.error = ''

        const newOrder = await apiCreateOrder()
        Object.assign(this.order, newOrder)

        return this.order
      } catch (err) {
        this.error = 'Failed to create order'
        console.error('Failed to create order:', err)
        throw err
      } finally {
        this.loading = false
      }
    },

    async saveShippingAddress(addressData: Address, emailData: string) {
      try {
        this.loading = true
        this.error = ''

        // Save email
        this.email = emailData

        // Create shipping address
        const createdAddress = await apiCreateAddress(addressData)
        Object.assign(this.shippingAddress, createdAddress)

        // Update order with shipping address and email
        if (this.order.id && this.shippingAddress.id) {
          const updatedOrder = await apiUpdateOrder(
            this.order.id,
            this.shippingAddress.id,
            emailData
          )
          Object.assign(this.order, updatedOrder)

          // Estimate tax
          await this.estimateTax()
        }

        return this.shippingAddress
      } catch (err) {
        this.error = 'Failed to save shipping address'
        console.error('Failed to save shipping address:', err)
        throw err
      } finally {
        this.loading = false
      }
    },

    async estimateTax(): Promise<void> {
      if (!this.order.id || !this.shippingAddress.id) {
        console.warn('Cannot estimate tax: no order ID')
        return
      }

      try {
        const estimate = await apiGetTaxEstimate(this.order.id)
        this.order.tax_amount = estimate.tax_amount
        this.order.total_amount = this.order.amount + estimate.tax_amount
      } catch (err) {
        console.error('Failed to estimate tax:', err)
      }
    },

    async preparePayment(): Promise<string> {
      if (!this.order.id) {
        throw new Error('Order not found')
      }

      if (this.stripe_client_secret) {
        return this.stripe_client_secret
      }

      try {
        this.loading = true
        this.error = ''

        const { client_secret } = await apiConfirmOrder(this.order.id)
        this.stripe_client_secret = client_secret

        return client_secret
      } catch (err) {
        this.error = 'Failed to prepare payment'
        console.error('Failed to prepare payment:', err)
        throw err
      } finally {
        this.loading = false
      }
    },

    confirmOrder() {
      this.orderConfirmed = true
    },

    clearError() {
      this.error = ''
    },

    resetCheckout() {
      // Reset state
      this.email = ''
      this.orderConfirmed = false
      this.loading = false
      this.error = ''
      this.useShippingAddress = true

      // Reset addresses
      Object.assign(this.shippingAddress, {
        id: '',
        addressee: '',
        line1: '',
        line2: '',
        city: '',
        state: '',
        postal_code: '',
      })
      Object.assign(this.billingAddress, {
        id: '',
        addressee: '',
        line1: '',
        line2: '',
        city: '',
        state: '',
        postal_code: '',
      })

      // Reset order
      Object.assign(this.order, {
        id: '',
        currency: '',
        amount: 0,
        tax_amount: 0,
        total_amount: 0,
        status: 'pending',
      })

      // Reset Stripe client secret
      this.stripe_client_secret = ''

      // Reset payment info
    },
  },
})
