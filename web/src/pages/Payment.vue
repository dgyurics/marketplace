<template>
  <div v-if="!isInitializing" class="container">
    <h2>Checkout</h2>
    <OrderSummary :tax-amount="taxAmount" />

    <h3>Payment Details</h3>
    <form @submit.prevent="submitPayment">
      <PaymentForm
        ref="paymentFormRef"
        :address="checkoutStore.shippingAddress"
        :client-secret="clientSecret"
        @ready="onPaymentReady"
        @error="onPaymentError"
      />

      <button
        type="submit"
        class="btn-full-width mt-30"
        :disabled="isSubmitting || !isPaymentReady"
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

import { Payment as PaymentForm } from '@/components/forms'
import OrderSummary from '@/components/OrderSummary.vue'
import { useCartStore } from '@/store/cart'
import { useCheckoutStore } from '@/store/checkout'

const checkoutStore = useCheckoutStore()
const cartStore = useCartStore()
const router = useRouter()

const isSubmitting = ref(false)
const isInitializing = ref(true)
const isPaymentReady = ref(false)
const taxAmount = ref(0)
const clientSecret = ref('')
const orderId = ref('')
const paymentFormRef = ref()

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
  clientSecret.value = res.client_secret
  orderId.value = res.order_id
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
  if (!paymentFormRef.value || isSubmitting.value || !orderId.value) return

  isSubmitting.value = true
  checkoutStore.paymentError = null

  try {
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
</style>
