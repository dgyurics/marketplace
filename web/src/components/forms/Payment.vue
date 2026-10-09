<template>
  <div class="payment-form">
    <div id="payment-element"></div>
  </div>
</template>

<script setup lang="ts">
import type { StripePaymentElement, StripeElements, Stripe } from '@stripe/stripe-js'
import { onMounted, onBeforeUnmount } from 'vue'

import { getStripe } from '@/services/stripe'
import type { Address } from '@/types'
import { getCurrency } from '@/utilities/config'

const props = defineProps<{
  amount: number // estimate - charged/actual amount comes from payment intent
  address: Address
}>()

const emit = defineEmits<{
  ready: []
  error: [error: string]
}>()

let elements: StripeElements | null = null
let paymentElement: StripePaymentElement | null = null

onMounted(async () => {
  try {
    const stripe = await getStripe()
    if (!stripe) return

    await initializePaymentElement(stripe)
  } catch (error) {
    emit('error', error instanceof Error ? error.message : 'Failed to initialize payment form')
  }
})

onBeforeUnmount(() => {
  paymentElement?.unmount()
})

async function initializePaymentElement(stripe: Stripe) {
  elements = stripe.elements({
    mode: 'payment',
    amount: props.amount,
    currency: getCurrency().toLowerCase(),
    appearance: {
      theme: 'stripe',
      variables: {
        fontFamily: "'Open Sans', sans-serif",
        fontSizeBase: '16px',
        borderRadius: '1px',
      },
    },
  })

  // @ts-ignore
  paymentElement = elements.create('payment', {
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
    fields: { billingDetails: 'auto' },
  })

  paymentElement.mount('#payment-element')
  paymentElement.on('ready', () => emit('ready'))
}

/**
 * Backend calculates final total with a respective payment-intent
 * Order goes from DRAFT -> PENDING
 * Payment type goes from UNDEFINED -> STRIPE
 */
async function confirmPayment(
  orderId: string,
  fetchClientSecret: () => Promise<{ client_secret: string; total_amount: number }>
) {
  if (!elements || !orderId) {
    throw new Error('Payment form not initialized or missing order ID')
  }

  const stripe = await getStripe()
  if (!stripe) throw new Error('Stripe not available')

  const { error: submitError } = await elements.submit()
  if (submitError) throw new Error(submitError.message)

  const { client_secret, total_amount } = await fetchClientSecret()

  // Keep wallets (Apple/Google Pay) in sync with the authoritative taxed total.
  if (total_amount && total_amount !== props.amount) {
    elements.update({ amount: total_amount })
  }

  const { error, paymentIntent } = await stripe.confirmPayment({
    elements,
    clientSecret: client_secret,
    confirmParams: {
      return_url: `${window.location.origin}/checkout/confirmation?order_id=${orderId}`,
    },
    redirect: 'if_required',
  })

  if (error) throw new Error(error.message)

  if (paymentIntent?.status !== 'succeeded' && paymentIntent?.status !== 'processing') {
    throw new Error('Payment was not completed. Please try again.')
  }
}

defineExpose({ confirmPayment })
</script>

<style scoped>
.payment-form {
  width: 100%;
}

#payment-element {
  min-height: 40px;
}
</style>
