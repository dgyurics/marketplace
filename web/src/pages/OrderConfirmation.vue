<template>
  <div class="container">
    <h2>Order Confirmed</h2>
    <div class="confirmation-message">
      <h3>Your order has been placed!</h3>
      <p class="confirmation-note">
        A confirmation email has been sent to <strong>{{ email }}</strong
        >.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const router = useRouter()

import { getOrderOwner } from '@/services/api'
import { useCheckoutStore } from '@/store/checkout'

const route = useRoute()
const checkoutStore = useCheckoutStore()
const email = ref('')

// https://selfco.io/checkout/payment/checkout/confirmation?
// order_id=115...&payment_intent=pi_3SXv...&payment_intent_client_secret=pi_3SX...&redirect_status=succeeded

// https://selfco.io/checkout/confirmation?
// order_id=115...&payment_intent=pi_3SXv...&payment_intent_client_secret=pi_3SX...&redirect_status=failed
onMounted(async () => {
  if (checkoutStore.shippingAddress.email) {
    email.value = checkoutStore.shippingAddress.email
    checkoutStore.resetCheckout()
    return
  }

  // Handle redirects from Stripe Payment
  const redirect = route.query['redirect_status'] // 'succeeded' | 'failed' | undefined

  if (redirect === 'succeeded') {
    const data = await getOrderOwner(route.query['order_id'] as string)
    email.value = data.address.email
    return
  }

  if (redirect === 'failed') {
    const orderID = route.query['order_id'] as string
    const data = await getOrderOwner(orderID)
    checkoutStore.shippingAddress = data.address
    checkoutStore.order_id = orderID
    checkoutStore.paymentError = 'Payment failed. Try again or use a different payment method.'
    checkoutStore.stripe_client_secret = route.query['payment_intent_client_secret'] as string
    router.push('/checkout/payment')
  }
})
</script>

<style scoped>
h2,
h3 {
  text-align: center;
  margin-bottom: 10px;
}

.confirmation-note {
  text-align: center;
}
</style>
