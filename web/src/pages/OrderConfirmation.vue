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

// Sample redirect URL
// https://selfco.io/checkout/payment/checkout/confirmation?
// order_id=115...&payment_intent=pi_3SXv...&payment_intent_client_secret=pi_3SX...&redirect_status=succeeded

// Sample redirect URL
// https://selfco.io/checkout/confirmation?
// order_id=115...&payment_intent=pi_3SXv...&payment_intent_client_secret=pi_3SX...&redirect_status=failed

onMounted(async () => {
  if (checkoutStore.shippingAddress.email) {
    email.value = checkoutStore.shippingAddress.email
    checkoutStore.resetCheckout()
    return
  }

  // Handle redirects from Stripe Payment
  const { redirect_status, order_id, payment_intent_client_secret } = route.query
  if (!order_id) return

  const order = await getOrderOwner(order_id as string)
  email.value = order.address.email

  if (redirect_status === 'failed') {
    // Restore state for retry
    checkoutStore.shippingAddress = order.address
    checkoutStore.order_id = order.id
    checkoutStore.stripe_client_secret = payment_intent_client_secret as string
    checkoutStore.paymentError = 'Payment failed. Try again or use a different payment method.'
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
