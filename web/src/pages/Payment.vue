<template>
  <div v-if="!isInitializing" class="container">
    <h2>Checkout</h2>
    <OrderSummary :tax-amount="taxAmount" />
    <h3>payment details</h3>
    <form @submit.prevent="submitPayment">
      <div class="form-group-flex">
        <label for="cardholder-name">Name on Card</label>
        <input id="cardholder-name" v-model="cardholderName" type="text" required />
      </div>

      <div class="form-group-flex">
        <label>Card Number</label>
        <div id="card-number" class="stripe-input"></div>
      </div>

      <div class="form-row">
        <div class="form-group-flex">
          <label>Expiration Date</label>
          <div id="card-expiry" class="stripe-input"></div>
        </div>

        <div class="form-group-flex">
          <label>CVC</label>
          <div id="card-cvc" class="stripe-input"></div>
        </div>
      </div>

      <!-- Email Field Below Payment Information -->
      <div v-show="false" class="form-group-flex">
        <label for="email">Email</label>
        <input
          id="email"
          v-model="checkoutStore.shippingAddress.email"
          type="email"
          autocomplete="email"
        />
      </div>

      <div class="form-group-flex checkbox-group">
        <label class="checkbox-label">
          <input id="useBilling" v-model="checkoutStore.useShippingAddress" type="checkbox" />
          Billing address same as shipping
        </label>
      </div>
      <div v-if="!checkoutStore.useShippingAddress" class="billing-address-fields">
        <h3>Billing Address</h3>
        <BillingAddressForm v-model="checkoutStore.billingAddress" />
      </div>
      <button type="submit" class="btn-full-width mt-15" :disabled="isSubmitting">
        Place Order
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import type {
  StripeCardCvcElement,
  StripeCardExpiryElement,
  StripeCardNumberElement,
} from '@stripe/stripe-js'
import { onMounted, ref, nextTick } from 'vue'
import { useRouter } from 'vue-router'

import { BillingAddressForm } from '@/components/forms'
import OrderSummary from '@/components/OrderSummary.vue'
import { getStripe, confirmCardPayment } from '@/services/stripe'
import { useCartStore } from '@/store/cart'
import { useCheckoutStore } from '@/store/checkout'
import { getLocale } from '@/utilities'

const checkoutStore = useCheckoutStore()
const cartStore = useCartStore()
const router = useRouter()
const isSubmitting = ref(false)
const isInitializing = ref(true)
const cardholderName = ref('')
const taxAmount = ref(0)

const country = getLocale().country_code

let cardElement: StripeCardNumberElement,
  expiryElement: StripeCardExpiryElement,
  cvcElement: StripeCardCvcElement

onMounted(async () => {
  try {
    // Check if shipping address is complete before proceeding
    if (!checkoutStore.isShippingAddressComplete) {
      router.push('/checkout/shipping')
      return
    }

    // Fetch cart items
    await cartStore.fetchCart()

    // Estimate tax
    const { tax_amount } = await checkoutStore.estimateTax()
    taxAmount.value = tax_amount
  } catch {
    // Handle errors silently
  }

  isInitializing.value = false

  // Wait for next tick to ensure DOM elements are rendered
  await nextTick()
  await initializeStripe()
})

const elementStyles = {
  style: {
    base: {
      color: '#333',
      fontWeight: 'normal',
      fontFamily: 'Open Sans, sans-serif',
      fontSize: '16px' /* Match address input font-size */,
      lineHeight: '1.5' /* Improve readability */,
      padding: '10px' /* Match padding */,
      '::placeholder': {
        color: '#aaa',
      },
    },
    invalid: {
      color: '#c00',
      iconColor: '#c00',
    },
  },
}

async function initializeStripe() {
  const stripe = await getStripe()
  if (!stripe) {
    return
  }

  const elements = stripe.elements()

  cardElement = elements.create('cardNumber', elementStyles)
  cardElement.mount('#card-number')

  expiryElement = elements.create('cardExpiry', elementStyles)
  expiryElement.mount('#card-expiry')

  cvcElement = elements.create('cardCvc', elementStyles)
  cvcElement.mount('#card-cvc')
}

const submitPayment = async () => {
  // FIXME how to check if stripe fields have been filled correctly
  // then validate before submitting
  const selectedAddress = checkoutStore.selectedBillingAddress

  const billingDetails = {
    name: cardholderName.value,
    address: {
      line1: selectedAddress.line1,
      line2: selectedAddress.line2 ?? null,
      city: selectedAddress.city,
      state: selectedAddress.state ?? null,
      postal_code: selectedAddress.postal_code,
      country,
    },
  }

  try {
    // Disable the submit button to prevent multiple submissions
    isSubmitting.value = true

    const clientSecret = await checkoutStore.preparePayment()

    const { error, paymentIntent: confirmedIntent } = await confirmCardPayment(
      clientSecret,
      cardElement,
      billingDetails
    )

    if (error) {
      alert(`Payment failed: ${error.message}`)
      return
    }

    if (confirmedIntent.status === 'succeeded') {
      router.push('/checkout/confirmation')
    } else {
      alert('Payment processing or additional verification required')
    }
  } catch (error) {
    alert('Payment submission failed. Try again')
    console.error('Payment submission error:', error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.stripe-input {
  width: 100%; /* Match address input width */
  padding: 10px; /* Consistent padding */
  border: 1px solid #ccc; /* Consistent border */
  border-radius: 4px; /* Consistent border-radius */
  font-size: 16px; /* Match address input font-size */
  background: white;
  box-sizing: border-box; /* Ensure padding doesn't affect width */
  height: 44px;
  line-height: 1.5;
}

/* Existing styles for inputs to ensure consistency */
input[type='text'],
input[type='email'],
input[type='password'],
input[type='tel'],
input[type='number'],
input[type='search'] {
  width: 100%;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 16px;
  box-sizing: border-box;
}

h2,
h3 {
  text-align: center;
  margin-bottom: 10px;
  text-transform: capitalize;
}

.section-subtitle {
  text-align: center;
  font-size: 16px;
  color: #555;
  margin-bottom: 20px;
  text-transform: capitalize;
}

label {
  font-weight: 500;
  margin-bottom: 5px;
}

.form-row {
  display: flex;
  gap: 10px;
}

.form-row .form-group-flex {
  flex: 1;
}

.receipt-note {
  font-size: 10px;
  color: #666;
  margin-top: 2px;
}

.confirmation-note {
  text-align: center;
}

.checkbox-group {
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
