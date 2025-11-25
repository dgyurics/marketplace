<template>
  <div class="payment-form">
    <div class="form-group-flex">
      <div id="payment-element" ref="paymentElementRef"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { StripePaymentElement, StripeElements } from '@stripe/stripe-js'
import { onMounted, ref, onBeforeUnmount } from 'vue'

import { getStripe } from '@/services/stripe'
import type { Address } from '@/types'

const props = defineProps<{
  clientSecret: string
  address: Address
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
        variables: {
          fontFamily: "'Open Sans', sans-serif",
          fontSizeBase: '16px',
          borderRadius: '1px',
        },
      },
    })

    // @ts-ignore - Payment element type not properly defined in current Stripe types
    paymentElement = elements.create('payment', {
      paymentMethodOrder: ['card'], // Card appears first in the list
      defaultValues: {
        billingDetails: {
          name: props.address.name,
          email: props.address.email,
          address: {
            line1: props.address.line1,
            line2: props.address.line2,
            city: props.address.city,
            state: props.address.state,
            postal_code: props.address.postal_code,
            country: props.address.country,
          },
        },
      },
      fields: {
        billingDetails: 'auto',
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

async function confirmPayment() {
  if (!elements) {
    throw new Error('Payment form not initialized')
  }

  const stripe = await getStripe()
  if (!stripe) {
    throw new Error('Stripe not available')
  }

  const { error } = await stripe.confirmPayment({
    elements,
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
