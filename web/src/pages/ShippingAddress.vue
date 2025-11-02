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
import { useRouter } from 'vue-router'

import { ShippingAddressForm } from '@/components/forms'
import { useCheckoutStore } from '@/store/checkout'
import type { Address } from '@/types'

const checkoutStore = useCheckoutStore()
const router = useRouter()

async function handleShippingSubmit(address: Address) {
  try {
    // upsert shipping address and get saved result
    const savedAddress = await checkoutStore.saveShippingAddress(address)

    // init order if not exists
    if (!checkoutStore.order.id) {
      await checkoutStore.initializeOrder(savedAddress.id!)
    }

    // complete order by inputting payment information
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
