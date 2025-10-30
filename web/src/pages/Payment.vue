<template>
  <div v-if="!isInitializing" class="container">
    <h2>Checkout</h2>
    <OrderSummary :order="checkoutStore.order" />
    <div>
      <h3>Payment Details</h3>
      <PaymentForm
        v-model:use-same-address="checkoutStore.useShippingAddress"
        :client-secret="clientSecret"
        @payment-ready="handlePaymentReady"
        @payment-error="handlePaymentError"
      />

      <div v-if="!checkoutStore.useShippingAddress" class="billing-address-section">
        <h3>Billing Address</h3>
        <BillingAddressForm v-model="checkoutStore.billingAddress" />
      </div>

      <div class="form-actions">
        <button type="button" class="btn-primary" @click="submitPayment">Place Order</button>
      </div>

      <div v-if="errorMessage" class="error">
        {{ errorMessage }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { StripeElements } from '@stripe/stripe-js'
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { BillingAddressForm, PaymentForm } from '@/components/forms'
import OrderSummary from '@/components/OrderSummary.vue'
import { getStripe } from '@/services/stripe'
import { useCheckoutStore } from '@/store/checkout'
import { getCountryForLocale, getAppLocale } from '@/utilities'

const checkoutStore = useCheckoutStore()
const router = useRouter()
const isSubmitting = ref(false)
const isInitializing = ref(true)
const errorMessage = ref('')
const clientSecret = ref('')

const country = getCountryForLocale(getAppLocale())
let stripeElements: StripeElements | null = null

onMounted(async () => {
  if (!checkoutStore.canProceedToPayment) {
    router.push('/checkout/shipping')
    return
  }

  try {
    await checkoutStore.estimateTax()
    clientSecret.value = await checkoutStore.preparePayment()
  } catch {
    errorMessage.value = 'Failed to initialize payment. Please try again.'
  } finally {
    isInitializing.value = false
  }
})

function handlePaymentReady(elements: StripeElements) {
  stripeElements = elements
  errorMessage.value = ''
}

function handlePaymentError(error: string) {
  errorMessage.value = error
}

async function submitPayment() {
  if (!stripeElements) {
    return
  }

  const selectedAddress = checkoutStore.selectedBillingAddress

  const billingDetails = {
    email: checkoutStore.email,
    address: {
      line1: selectedAddress.line1,
      line2: selectedAddress.line2 ?? null,
      city: selectedAddress.city,
      state: selectedAddress.state,
      postal_code: selectedAddress.postal_code,
      country,
    },
  }

  try {
    isSubmitting.value = true
    errorMessage.value = ''

    const stripe = await getStripe()
    if (!stripe) {
      throw new Error('Payment system unavailable')
    }

    const { error, paymentIntent } = await stripe.confirmPayment({
      elements: stripeElements,
      confirmParams: {
        return_url: `${window.location.origin}/checkout/confirmation`,
        payment_method_data: {
          billing_details: billingDetails,
        },
      },
      redirect: 'if_required',
    })

    if (error) {
      errorMessage.value = `Payment failed: ${error.message}`
      return
    }

    if (paymentIntent.status === 'succeeded') {
      checkoutStore.confirmOrder()
      router.push('/checkout/confirmation')
    } else {
      errorMessage.value = 'Payment processing or additional verification required'
    }
  } catch (error) {
    console.error('Payment error:', error)
    errorMessage.value = 'Payment submission failed.'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.container {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
}

h2,
h3 {
  text-align: center;
  margin-bottom: 10px;
  text-transform: capitalize;
}

.billing-address-section {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #e5e7eb;
}

h4 {
  margin-bottom: 16px;
  font-size: 16px;
  font-weight: 500;
  color: #333;
}

.form-actions {
  margin-top: 32px;
}

.btn-primary {
  width: 100%;
  padding: 12px 24px;
  background-color: #000000;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn-primary:hover:not(:disabled) {
  background-color: #333333;
}

.btn-primary:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}
</style>
