<template>
  <div class="payment-form">
    <div class="form-group-flex">
      <div id="payment-element" ref="paymentElementRef"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { StripePaymentElement, StripeElements, BillingDetails } from '@stripe/stripe-js'
import { onMounted, ref, onBeforeUnmount } from 'vue'

import { getStripe } from '@/services/stripe'

const props = defineProps<{
  clientSecret: string
}>()

const emit = defineEmits<{
  ready: []
  error: [error: string]
}>()

const paymentElementRef = ref<HTMLElement>()
let elements: StripeElements | null = null
let paymentElement: StripePaymentElement | null = null

onMounted(async () => {
  try {
    const stripe = await getStripe()
    if (!stripe || !props.clientSecret) return

    elements = stripe.elements({
      clientSecret: props.clientSecret,
      appearance: {
        theme: 'stripe',
      },
    })

    paymentElement = elements.create('payment', {
      fields: {
        billingDetails: {
          name: 'never',
          email: 'never',
          phone: 'auto',
          address: 'never',
        },
      },
    })
    paymentElement.mount('#payment-element')

    paymentElement.on('ready', () => {
      emit('ready')
    })
  } catch (error) {
    emit('error', error instanceof Error ? error.message : 'Failed to initialize payment form')
  }
})

onBeforeUnmount(() => {
  if (paymentElement) {
    paymentElement.unmount()
  }
})

async function confirmPayment(billingDetails: BillingDetails) {
  if (!elements) {
    throw new Error('Payment form not initialized')
  }

  const stripe = await getStripe()
  if (!stripe) {
    throw new Error('Stripe not available')
  }

  const { error } = await stripe.confirmPayment({
    elements,
    confirmParams: {
      payment_method_data: {
        billing_details: billingDetails,
      },
    },
    redirect: 'if_required',
  })

  if (error) {
    throw new Error(error.message)
  }
}

defineExpose({
  confirmPayment,
})
</script>

<style scoped>
.payment-form {
  width: 100%;
}

#payment-element {
  min-height: 40px;
}
</style>
