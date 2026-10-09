import { ref, type Ref } from 'vue'
import { useRouter } from 'vue-router'

import { useCountdown } from '@/composables/useCountdown'
import { useCheckoutStore } from '@/store/checkout'

export interface PaymentFormRef {
  confirmPayment: (
    orderId: string,
    fetchClientSecret: () => Promise<{ client_secret: string; total_amount: number }>
  ) => Promise<void>
}

/**
 * Owns the Stripe side of checkout: tracking readiness of the Payment Element,
 * and creating + confirming the PaymentIntent only once the user submits.
 */
export function useStripeCheckout(paymentFormRef: Ref<PaymentFormRef | null>) {
  const router = useRouter()
  const checkoutStore = useCheckoutStore()

  const isReady = ref(false)
  const isSubmitting = ref(false)

  // Orders hold inventory, so send the user back to the cart if they stall.
  const { start: startTimer } = useCountdown(14 * 60, () => {
    router.push('/cart')
  })

  function onReady() {
    isReady.value = true
    startTimer()
  }

  function onError(message: string) {
    checkoutStore.paymentError = message
  }

  async function submit() {
    if (isSubmitting.value || !checkoutStore.order_id) return

    isSubmitting.value = true
    checkoutStore.paymentError = null

    try {
      if (!paymentFormRef.value) return
      await paymentFormRef.value.confirmPayment(checkoutStore.order_id, () =>
        checkoutStore.fetchStripeIntent()
      )
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
    isReady,
    isSubmitting,
    onReady,
    onError,
    submit,
  }
}
