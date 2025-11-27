<template>
  <div class="container">
    <h2>Checkout</h2>
    <div>
      <h3>shipping address</h3>
      <form @submit.prevent="handleShippingSubmit">
        <AddressForm
          :model-value="checkoutStore.shippingAddress"
          @update:model-value="handleAddressUpdate"
        />
        <button type="submit" class="btn-full-width mt-15" :tabindex="0">Continue</button>
      </form>
      <p v-if="checkoutStore.shippingError" class="error">{{ checkoutStore.shippingError }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
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

async function handleShippingSubmit() {
  try {
    // Save the shipping address
    await checkoutStore.saveShippingAddress(checkoutStore.shippingAddress)

    // Clear any previous errors
    checkoutStore.shippingError = null

    // Navigate to payment
    router.push('/checkout/payment')
  } catch (error: unknown) {
    const status = (error as { response?: { status?: number } })?.response?.status
    checkoutStore.shippingError =
      status === 400 ? 'Invalid shipping address' : 'Something went wrong'
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
</style>
