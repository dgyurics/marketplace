import {
  type Stripe,
  type Address,
  type PaymentIntentResult,
  type StripeCardElement,
  type StripeCardNumberElement,
} from '@stripe/stripe-js'
import { loadStripe } from '@stripe/stripe-js/pure'

import { STRIPE_PUBLISHABLE_KEY } from '@/config'

let stripePromise: Promise<Stripe | null> | undefined

export function getStripe() {
  if (!stripePromise) {
    loadStripe.setLoadParameters({ advancedFraudSignals: false })
    stripePromise = loadStripe(STRIPE_PUBLISHABLE_KEY)
  }
  return stripePromise
}

export async function confirmCardPayment(
  clientSecret: string,
  cardElement: StripeCardElement | StripeCardNumberElement | { token: string },
  billingDetails: {
    name: string
    email: string
    address: Address
  }
): Promise<PaymentIntentResult> {
  const stripe = await getStripe()
  if (!stripe) {
    throw new Error('Stripe failed to initialize')
  }

  const result = await stripe.confirmCardPayment(clientSecret, {
    payment_method: {
      card: cardElement,
      billing_details: billingDetails,
    },
    // TODO
    // payment_method_options: {
    //   card: {
    //     cvc : cardElement.cvc, // Assuming cardElement has a cvc property
    //   },
    // },
  })

  return result
}
