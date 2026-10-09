<template>
  <form @submit.prevent="submit">
    <PaymentForm
      ref="paymentFormRef"
      :address="checkoutStore.shippingAddress"
      :amount="amount"
      @ready="onReady"
      @error="onError"
    />

    <button
      type="submit"
      class="btn-full-width mt-30"
      :disabled="isSubmitting || !isReady"
      :tabindex="0"
    >
      Place Order
    </button>
  </form>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import { Payment as PaymentForm } from '@/components/forms'
import { useStripeCheckout, type PaymentFormRef } from '@/composables/useStripeCheckout'
import { useCheckoutStore } from '@/store/checkout'

const checkoutStore = useCheckoutStore()
const paymentFormRef = ref<PaymentFormRef | null>(null)

defineProps<{ amount: number }>()

const { isReady, isSubmitting, onReady, onError, submit } = useStripeCheckout(paymentFormRef)
</script>
