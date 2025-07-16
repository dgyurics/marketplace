<template>
  <div class="container">
    <h2>Checkout</h2>
    <div>
      <h3>shipping address</h3>
      <ShippingAddressForm @submit="handleShippingSubmit" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

import ShippingAddressForm from '@/components/ShippingAddressForm.vue'
import { useCheckoutStore } from '@/store/checkout'
import type { Address } from '@/types'

const checkoutStore = useCheckoutStore()
const router = useRouter()

onMounted(async () => {
  await checkoutStore.initializeOrder()
})

async function handleShippingSubmit(address: Address, email: string) {
  await checkoutStore.saveShippingAddress(address, email)
  if (checkoutStore.canProceedToPayment) {
    router.push('/checkout/payment')
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
