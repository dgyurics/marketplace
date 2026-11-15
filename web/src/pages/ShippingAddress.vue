<template>
  <div class="container checkout-page">
    <header class="checkout-heading">
      <p class="checkout-eyebrow">Step 1 of 2 · Shipping</p>
      <h1>Where should we send it?</h1>
      <p>Enter your shipping details to continue.</p>
    </header>
    <div class="checkout-panel">
      <form @submit.prevent="handleShippingSubmit">
        <AddressForm
          :model-value="checkoutStore.shippingAddress"
          @update:model-value="handleAddressUpdate"
        />
        <button type="submit" class="btn-full-width mt-15" :tabindex="0">
          Continue to payment
        </button>
      </form>
      <p v-if="checkoutStore.shippingError" class="error">{{ checkoutStore.shippingError }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'

import { Address as AddressForm } from '@/components/forms'
import { useCartStore } from '@/store/cart'
import { useCheckoutStore } from '@/store/checkout'
import type { Address } from '@/types'

const checkoutStore = useCheckoutStore()
const cartStore = useCartStore()
const router = useRouter()

function handleAddressUpdate(address: Address) {
  checkoutStore.shippingAddress = address
}

onMounted(async () => {
  // Ensure cart is loaded for checkout
  await cartStore.fetchCart()

  // Redirect to cart if no items
  if (cartStore.items.length === 0) {
    router.push('/cart')
  }
})

onUnmounted(() => {
  checkoutStore.shippingError = null
})

async function handleShippingSubmit() {
  try {
    // Save the shipping address
    await checkoutStore.saveShippingAddress(checkoutStore.shippingAddress)

    // Clear any previous errors
    checkoutStore.shippingError = null

    // Navigate to payment
    router.push('/checkout/payment')
  } catch (error: any) {
    const status = error.response?.status
    if (status === 400) {
      checkoutStore.shippingError = 'Invalid shipping address'
    } else if (status === 422) {
      checkoutStore.shippingError = 'Shipping not available to the provided address'
    } else {
      checkoutStore.shippingError = 'Something went wrong'
    }
  }
}
</script>
