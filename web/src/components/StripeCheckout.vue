<template>
  <form @submit.prevent="submit">
    <PaymentForm
      v-if="clientSecret"
      ref="paymentFormRef"
      :address="checkoutStore.shippingAddress"
      :client-secret="clientSecret"
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
import { onMounted, ref } from 'vue'

import { Payment as PaymentForm } from '@/components/forms'
import { useStripeCheckout, type PaymentFormRef } from '@/composables/useStripeCheckout'
import { useCheckoutStore } from '@/store/checkout'

const checkoutStore = useCheckoutStore()
const paymentFormRef = ref<PaymentFormRef | null>(null)

const { clientSecret, isReady, isSubmitting, prepare, onReady, onError, submit } =
  useStripeCheckout(paymentFormRef)

onMounted(prepare)
</script>
