<template>
  <div v-if="!isInitializing" class="container">
    <h2>Checkout</h2>
    <OrderSummary :tax-amount="taxAmount" />

    <h3>Payment Details</h3>
    <form @submit.prevent="submitPayment">
      <PaymentForm
        ref="paymentFormRef"
        :client-secret="clientSecret"
        @ready="onPaymentReady"
        @error="onPaymentError"
      />

      <div class="form-group checkbox-group">
        <label class="checkbox-label">
          <input v-model="checkoutStore.useShippingAddress" type="checkbox" :tabindex="0" />
          Billing information same as shipping
        </label>
      </div>

      <div v-if="!checkoutStore.useShippingAddress" class="billing-section">
        <BillingAddressForm v-model="checkoutStore.billingAddress" />
      </div>

      <button
        type="submit"
        class="btn-full-width mt-15"
        :disabled="isSubmitting || !isPaymentReady"
        :tabindex="0"
      >
        Place Order
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import type { BillingDetails } from '@stripe/stripe-js'
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { BillingAddressForm, PaymentForm } from '@/components/forms'
import OrderSummary from '@/components/OrderSummary.vue'
import { useCartStore } from '@/store/cart'
import { useCheckoutStore } from '@/store/checkout'
import { getLocale } from '@/utilities'

const checkoutStore = useCheckoutStore()
const cartStore = useCartStore()
const router = useRouter()

const isSubmitting = ref(false)
const isInitializing = ref(true)
const isPaymentReady = ref(false)
const taxAmount = ref(0)
const clientSecret = ref('')
const paymentFormRef = ref()

const country = getLocale().country_code

onMounted(async () => {
  try {
    // Check if shipping address is complete
    if (!checkoutStore.isShippingAddressComplete) {
      router.push('/checkout/shipping')
      return
    }

    // Fetch cart and estimate tax
    await cartStore.fetchCart()
    const { tax_amount } = await checkoutStore.estimateTax()
    taxAmount.value = tax_amount

    // Get client secret for payment
    clientSecret.value = await checkoutStore.preparePayment()
  } catch (error: any) {
    const status = error.response?.status
    if (status === 400) {
      checkoutStore.setShippingError('Invalid shipping address')
      router.push('/checkout/shipping')
    }
  } finally {
    isInitializing.value = false
  }
})

function onPaymentReady() {
  isPaymentReady.value = true
}

function onPaymentError(error: string) {
  alert(`Payment form error: ${error}`)
}

async function submitPayment() {
  if (!paymentFormRef.value || isSubmitting.value) return

  isSubmitting.value = true

  try {
    const selectedAddress = checkoutStore.selectedBillingAddress

    const billingDetails: BillingDetails = {
      name: selectedAddress.name ?? '', // FIXME if empty/invalid throw error
      email: selectedAddress.email,
      address: {
        line1: selectedAddress.line1,
        line2: selectedAddress.line2 ?? null,
        city: selectedAddress.city,
        state: selectedAddress.state ?? '', // FIXME if empty/invalid throw error
        postal_code: selectedAddress.postal_code,
        country,
      },
    }

    await paymentFormRef.value.confirmPayment(billingDetails)
    router.push('/checkout/confirmation')
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Please try again'
    alert(`Payment failed: ${message}`)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
h2,
h3 {
  text-align: center;
  margin-bottom: 20px;
  text-transform: capitalize;
}

.form-group {
  margin-bottom: 20px;
}

label {
  display: block;
  font-weight: 500;
  margin-bottom: 5px;
}

input[type='text'] {
  width: 100%;
  padding: 12px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 16px;
  box-sizing: border-box;
}

.checkbox-group {
  margin: 30px 0;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-style: italic;
}

.billing-section {
  margin-top: 20px;
}
</style>
