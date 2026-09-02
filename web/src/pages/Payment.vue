<template>
  <div v-if="!isInitializing" class="container">
    <h2>Checkout</h2>
    <OrderSummary :tax-amount="taxAmount" />

    <h3>Payment Method</h3>
    <PaymentMethodSelector v-model="paymentMethod" :disabled="['delivery']" />

    <form @submit.prevent="submitPayment">
      <PaymentForm
        v-if="paymentMethod === 'online'"
        ref="paymentFormRef"
        :address="checkoutStore.shippingAddress"
        :client-secret="clientSecret"
        @ready="onPaymentReady"
        @error="onPaymentError"
      />

      <p v-if="paymentMethod === 'delivery'" class="delivery-note">
        Payment will be collected upon delivery
      </p>

      <button
        type="submit"
        class="btn-full-width mt-30"
        :disabled="isSubmitting || (paymentMethod === 'online' && !isPaymentReady)"
        :tabindex="0"
      >
        Place Order
      </button>
    </form>
    <p v-if="checkoutStore.paymentError" class="error">{{ checkoutStore.paymentError }}</p>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { Payment as PaymentForm, PaymentMethodSelector } from '@/components/forms'
import OrderSummary from '@/components/OrderSummary.vue'
import { useCountdown } from '@/composables/useCountdown'
import { useCartStore } from '@/store/cart'
import { useCheckoutStore } from '@/store/checkout'
import type { PaymentMethod } from '@/types'

const checkoutStore = useCheckoutStore()
const cartStore = useCartStore()
const router = useRouter()

const paymentMethod = ref<PaymentMethod>('online')
const isSubmitting = ref(false)
const isInitializing = ref(true)
const isPaymentReady = ref(false)
const taxAmount = ref(0)
const clientSecret = ref('')
const orderId = ref('')
const paymentFormRef = ref()

const { start: startTimer } = useCountdown(14 * 60, () => {
  router.push('/cart')
})

onMounted(async () => {
  try {
    await initializePayment()
  } catch (error: any) {
    handleInitError(error)
  } finally {
    isInitializing.value = false
  }
})

async function initializePayment() {
  if (!checkoutStore.isShippingAddressComplete) {
    router.push('/checkout/shipping')
    return
  }

  // populate cart and get tax estimate
  await cartStore.fetchCart()
  const { tax_amount } = await checkoutStore.estimateTax()
  taxAmount.value = tax_amount

  // get client secret for payment
  const res = await checkoutStore.preparePayment()
  if (!res) {
    await cartStore.fetchCart()
    router.push('/cart')
    return
  }
  clientSecret.value = res.client_secret
  orderId.value = res.order_id
  startTimer()
}

function handleInitError(error: any) {
  if (error.response?.status === 400) {
    checkoutStore.shippingError = 'Invalid shipping address'
    router.push('/checkout/shipping')
  }
}

function onPaymentReady() {
  isPaymentReady.value = true
}

function onPaymentError(error: string) {
  checkoutStore.paymentError = error
  console.error('Payment form error:', error)
}

async function submitPayment() {
  if (isSubmitting.value || !orderId.value) return

  isSubmitting.value = true
  checkoutStore.paymentError = null

  try {
    if (paymentMethod.value === 'delivery') {
      // TODO: call API to mark order as pay-on-delivery
      router.push('/checkout/confirmation')
      return
    }

    if (!paymentFormRef.value) return
    await paymentFormRef.value.confirmPayment(orderId.value)
    router.push('/checkout/confirmation')
  } catch (error) {
    const message =
      error instanceof Error
        ? error.message
        : 'Payment failed. Try again or use a different payment method.'
    checkoutStore.paymentError = message
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

.delivery-note {
  text-align: center;
  color: #555;
  padding: 20px 0;
}
</style>
