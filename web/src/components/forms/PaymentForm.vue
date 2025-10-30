<template>
  <div class="payment-form">
    <div class="form-group-flex">
      <div id="payment-element" ref="paymentElementRef"></div>
    </div>

    <div class="form-group billing-checkbox">
      <label class="checkbox-label">
        <input
          v-model="useSameAddress"
          type="checkbox"
          @change="$emit('update:useSameAddress', useSameAddress)"
        />
        Billing address same as shipping
      </label>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { StripePaymentElement, StripeElements } from '@stripe/stripe-js'
import { onMounted, ref, onBeforeUnmount } from 'vue'

import { getStripe } from '@/services/stripe'

interface Props {
  useSameAddress: boolean
  clientSecret: string
}

interface Emits {
  (e: 'update:useSameAddress', value: boolean): void
  (e: 'payment-ready', elements: StripeElements): void
  (e: 'payment-error', error: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const useSameAddress = ref(props.useSameAddress)
const paymentElementRef = ref<HTMLElement>()

let paymentElement: StripePaymentElement | null = null
let elements: StripeElements | null = null

onMounted(async () => {
  await initializeStripeElements()
})

onBeforeUnmount(() => {
  if (paymentElement) {
    paymentElement.destroy()
  }
})

async function initializeStripeElements() {
  try {
    const stripe = await getStripe()
    if (!stripe || !props.clientSecret) {
      emit('payment-error', 'Failed to initialize payment system')
      return
    }

    elements = stripe.elements({
      clientSecret: props.clientSecret,
      appearance: {
        theme: 'stripe',
        variables: {
          colorPrimary: '#000000',
          colorText: '#333333',
          colorBackground: '#ffffff',
          colorDanger: '#df1b41',
          fontFamily: 'system-ui, sans-serif',
          spacingUnit: '2px',
          borderRadius: '4px',
          fontSizeBase: '16px',
        },
        rules: {
          '.Input': {
            padding: '10px',
            border: '1px solid #ccc',
            borderRadius: '4px',
            fontSize: '16px',
            backgroundColor: '#ffffff',
            boxSizing: 'border-box',
          },
          '.Input:focus': {
            border: '1px solid #ccc',
            outline: 'none',
            boxShadow: 'none',
          },
          '.Label': {
            fontWeight: 'normal',
            marginBottom: '5px',
            fontSize: '14px',
            color: '#333333',
          },
          '.Tab': {
            display: 'none',
          },
          '.TabIcon': {
            display: 'none',
          },
          '.Icon': {
            display: 'none',
          },
          '.CardBrandIcon': {
            display: 'none',
          },
        },
      },
    })

    paymentElement = elements.create('payment', {
      layout: {
        type: 'accordion',
        defaultCollapsed: false,
        radios: false,
        spacedAccordionItems: false,
      },
      paymentMethodOrder: ['card'],
      fields: {
        billingDetails: 'never',
      },
      wallets: {
        applePay: 'never',
        googlePay: 'never',
      },
    })

    if (paymentElementRef.value) {
      paymentElement.mount(paymentElementRef.value)
      emit('payment-ready', elements)
    }
  } catch {
    emit('payment-error', 'Failed to initialize payment form')
  }
}
</script>

<style scoped>
.payment-form {
  width: 100%;
}

.form-group-flex {
  margin-bottom: 15px;
}

#payment-element {
  width: 100%;
  min-height: 40px;
}

.billing-checkbox {
  margin-bottom: 15px;
}

.checkbox-label {
  display: inline-flex;
  align-items: center;
  font-weight: 500;
  gap: 6px;
  font-size: 13px;
  font-style: italic;
}

.checkbox-label input[type='checkbox'] {
  margin: 0;
}
</style>
