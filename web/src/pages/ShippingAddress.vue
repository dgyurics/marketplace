<template>
  <div class="container">
    <h2>Checkout</h2>
    <div>
      <h3>shipping address</h3>
      <ShippingAddressForm
        :model-value="checkoutStore.shippingAddress"
        @submit="handleShippingSubmit"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { ShippingAddressForm } from '@/components/forms'
import { useCartStore } from '@/store/cart'
import { useCheckoutStore } from '@/store/checkout'
import type { Address } from '@/types'

const checkoutStore = useCheckoutStore()
const cartStore = useCartStore()
const router = useRouter()

onMounted(async () => {
  // Ensure cart is loaded for checkout
  await cartStore.fetchCart()

  // Redirect to cart if no items
  if (cartStore.items.length === 0) {
    router.push('/cart')
  }
})

async function handleShippingSubmit(address: Address) {
  try {
    // Save the shipping address
    await checkoutStore.saveShippingAddress(address)

    // Navigate to payment
    router.push('/checkout/payment')
  } catch {
    router.push('/error')
  }
}
</script>

<style scoped>
h2,
h3 {
  text-align: center;
  margin-bottom: 10px;
  text-transform: capitalize;
}
</style>
