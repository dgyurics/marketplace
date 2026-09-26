<template>
  <div v-if="!isInitializing" class="container">
    <h2>Checkout</h2>
    <OrderSummary :tax-amount="taxAmount" />

    <h3>Payment Method</h3>
    <PaymentMethodSelector v-model="paymentMethod" :enabled="enabledMethods" />

    <StripeCheckout v-if="paymentMethod === 'stripe'" />

    <DeliveryCheckout v-else-if="paymentMethod === 'pay_on_delivery'" />

    <NoPaymentMethods v-else />

    <p v-if="checkoutStore.paymentError" class="error">{{ checkoutStore.paymentError }}</p>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import DeliveryCheckout from '@/components/DeliveryCheckout.vue'
import { PaymentMethodSelector } from '@/components/forms'
import NoPaymentMethods from '@/components/NoPaymentMethods.vue'
import OrderSummary from '@/components/OrderSummary.vue'
import StripeCheckout from '@/components/StripeCheckout.vue'
import { getPaymentMethods } from '@/services/api'
import { useCartStore } from '@/store/cart'
import { useCheckoutStore } from '@/store/checkout'
import type { PaymentMethod } from '@/types'

const checkoutStore = useCheckoutStore()
const cartStore = useCartStore()
const router = useRouter()

const paymentMethod = ref<PaymentMethod | null>(null)
const enabledMethods = ref<PaymentMethod[]>([])
const isInitializing = ref(true)
const taxAmount = ref(0)

onMounted(async () => {
  try {
    await initializePayment()
  } catch (error: unknown) {
    handleInitError(error)
  } finally {
    isInitializing.value = false
  }
})

// Errors are scoped to this page, so don't let them leak into the next visit.
onUnmounted(() => {
  checkoutStore.paymentError = null
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

  // server decides which payment methods are available
  const options = await getPaymentMethods()
  enabledMethods.value = (Object.keys(options) as PaymentMethod[]).filter((m) => options[m])
  paymentMethod.value = enabledMethods.value[0] ?? null
}

function handleInitError(error: unknown) {
  const status = (error as { response?: { status?: number } })?.response?.status
  if (status === 400) {
    checkoutStore.shippingError = 'Invalid shipping address'
    router.push('/checkout/shipping')
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
