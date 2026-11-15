<template>
  <div class="container checkout-page">
    <div class="checkout-panel checkout-confirmation">
      <CheckCircleIcon class="confirmation-icon" />
      <h1>Order confirmed</h1>
      <p>A notification has been sent to your account.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { CheckCircleIcon } from '@heroicons/vue/24/outline'
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const router = useRouter()

import { getOrderOwner } from '@/services/api'
import { useCartStore } from '@/store/cart'
import { useCheckoutStore } from '@/store/checkout'
import { useInboxStore } from '@/store/inbox'

const route = useRoute()

const checkoutStore = useCheckoutStore()
const cartStore = useCartStore()
const inboxStore = useInboxStore()

// Sample redirect URL
// https://selfco.io/checkout/payment/checkout/confirmation?
// order_id=115...&payment_intent=pi_3SXv...&payment_intent_client_secret=pi_3SX...&redirect_status=succeeded

// Sample redirect URL
// https://selfco.io/checkout/confirmation?
// order_id=115...&payment_intent=pi_3SXv...&payment_intent_client_secret=pi_3SX...&redirect_status=failed

onMounted(async () => {
  clearCart()
  checkoutStore.resetCheckout()

  // Handle redirects from Stripe Payment
  const { redirect_status, order_id, payment_intent_client_secret } = route.query
  if (!order_id) return

  const order = await getOrderOwner(order_id as string)

  if (redirect_status === 'failed') {
    // Restore state so the retry reuses this order and its PaymentIntent
    // rather than creating a second one.
    checkoutStore.shippingAddress = order.address
    checkoutStore.order_id = order.id
    checkoutStore.stripe_client_secret = payment_intent_client_secret as string
    checkoutStore.paymentError = 'Payment failed.'
    router.push('/checkout/payment')
  }
})

// Poll until the webhook has been processed: the cart is cleared and the
// order notification has landed in the inbox.
const clearCart = () => {
  const poll = window.setInterval(async () => {
    await Promise.all([cartStore.fetchCart(), inboxStore.fetchConversations()])
    if (!cartStore.hasItems && inboxStore.conversations.length > 0) {
      window.clearInterval(poll)
    }
  }, 1500)

  // Stop polling after 10s regardless
  window.setTimeout(() => window.clearInterval(poll), 10000)
}
</script>
