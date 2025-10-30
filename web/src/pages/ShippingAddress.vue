<template>
  <div v-if="!isInitializing" class="container">
    <h2>Checkout</h2>
    <div>
      <h3>shipping address</h3>
      <ShippingAddressForm
        :initial-address="checkoutStore.shippingAddress"
        :initial-email="checkoutStore.email"
        @submit="handleShippingSubmit"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { ShippingAddressForm } from '@/components/forms'
import { useCheckoutStore } from '@/store/checkout'
import type { Address } from '@/types'

const checkoutStore = useCheckoutStore()
const router = useRouter()
const isInitializing = ref(true)

onMounted(async () => {
  try {
    await checkoutStore.initializeOrder()
  } catch (error) {
    let status = 500
    if (error && typeof error === 'object' && 'response' in error) {
      status = (error as { response?: { status?: number } }).response?.status || 500
    }

    if (status === 400) {
      router.push('/')
      return
    }
    if (status === 401) {
      router.push('/auth')
      return
    }
    router.push(`/error?status=${status}`)
  } finally {
    isInitializing.value = false
  }
})

async function handleShippingSubmit(address: Address, email: string) {
  try {
    await checkoutStore.saveShippingAddress(address, email)
    // FIXME refactor shouldn't need this if above throws no error
    if (checkoutStore.canProceedToPayment) {
      router.push('/checkout/payment')
    }
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
