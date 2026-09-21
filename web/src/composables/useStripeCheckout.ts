import { ref, type Ref } from 'vue'
import { useRouter } from 'vue-router'

import { useCountdown } from '@/composables/useCountdown'
import { useCartStore } from '@/store/cart'
import { useCheckoutStore } from '@/store/checkout'

export interface PaymentFormRef {
  confirmPayment: (orderId: string) => Promise<void>
}

/**
 * Owns the Stripe side of checkout: creating the order + PaymentIntent,
 * tracking readiness of the Payment Element, and confirming the payment.
 */
export function useStripeCheckout(paymentFormRef: Ref<PaymentFormRef | null>) {
  const router = useRouter()
  const cartStore = useCartStore()
  const checkoutStore = useCheckoutStore()

  const clientSecret = ref('')
  const orderId = ref('')
  const isReady = ref(false)
  const isSubmitting = ref(false)

  // Orders hold inventory, so send the user back to the cart if they stall.
  const { start: startTimer } = useCountdown(14 * 60, () => {
    router.push('/cart')
  })

  /** Creates the order and fetches the client secret for the Payment Element. */
  async function prepare() {
    try {
      const res = await checkoutStore.preparePayment()
      if (!res) {
        await cartStore.fetchCart()
        router.push('/cart')
        return
      }
      clientSecret.value = res.client_secret
      orderId.value = res.order_id
      startTimer()
    } catch (error: unknown) {
      const status = (error as { response?: { status?: number } }).response?.status
      if (status === 400) {
        checkoutStore.shippingError = 'Invalid shipping address'
        router.push('/checkout/shipping')
        return
      }
      checkoutStore.paymentError = 'Unable to start payment. Please try again.'
    }
  }

  function onReady() {
    isReady.value = true
  }

  function onError(message: string) {
    checkoutStore.paymentError = message
  }

  async function submit() {
    if (isSubmitting.value || !orderId.value) return

    isSubmitting.value = true
    checkoutStore.paymentError = null

    try {
      if (!paymentFormRef.value) return
      await paymentFormRef.value.confirmPayment(orderId.value)
      router.push('/checkout/confirmation')
    } catch (error) {
      const message =
        error instanceof Error
          ? error.message
          : 'Payment failed. Try again or use a different payment method.'
      checkoutStore.paymentError = message
    } finally {
      isSubmitting.value = false
    }
  }

  return {
    clientSecret,
    isReady,
    isSubmitting,
    prepare,
    onReady,
    onError,
    submit,
  }
}
