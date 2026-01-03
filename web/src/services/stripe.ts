import { type Stripe } from '@stripe/stripe-js'
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
