import { ref, type Ref } from 'vue'
import { useRouter } from 'vue-router'

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

  function onReady() {
    isReady.value = true
  }

  function onError(message: string) {
    checkoutStore.paymentError = message
  }

  function isOrderGone(error: unknown) {
    const status = (error as { response?: { status?: number } }).response?.status
    return status === 404 || status === 409
  }

  /** Creates the intent; if the order expired/cancelled, creates a new one and retries once. */
  async function fetchIntentWithRetry() {
    try {
      await checkoutStore.fetchStripeIntent()
      return
    } catch (error: unknown) {
      if (!isOrderGone(error)) throw error
      const order = await checkoutStore.createPendingOrder()
      if (!order) throw error
    }
    await checkoutStore.fetchStripeIntent()
  }

  async function submit() {
    if (isSubmitting.value || !checkoutStore.order_id) return

    isSubmitting.value = true
    checkoutStore.paymentError = null

    try {
      if (!paymentFormRef.value) return
      // Done first so the order id passed below (used in return_url) is current.
      await fetchIntentWithRetry()
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
