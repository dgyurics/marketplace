<template>
  <div class="container">
    <h2>Order Confirmed</h2>
    <div class="confirmation-message">
      <h3>Your order has been placed!</h3>
      <p class="confirmation-note">
        A confirmation email has been sent to
        <strong>{{ checkoutStore.shippingAddress.email }}</strong
        >.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'

import { useCheckoutStore } from '@/store/checkout'

const checkoutStore = useCheckoutStore()
const router = useRouter()

onUnmounted(() => {
  checkoutStore.orderConfirmed = false
  checkoutStore.resetCheckout()
})

onMounted(() => {
  if (!checkoutStore.orderConfirmed) {
    router.push('/')
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
