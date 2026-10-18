import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { useCheckoutStore } from '@/store/checkout'

/**
 * Owns the pay-on-delivery side of checkout: acknowledging the terms and
 * confirming the pending order. No payment is collected, so the order is final.
 */
export function useDeliveryCheckout() {
  const router = useRouter()
  const checkoutStore = useCheckoutStore()

  const acknowledged = ref(false)
  const isSubmitting = ref(false)

  function isOrderGone(error: unknown) {
    const status = (error as { response?: { status?: number } }).response?.status
    return status === 404 || status === 409
  }

  /** Confirms the order; if it expired/cancelled, creates a new one and retries once. */
  async function confirmWithRetry() {
    try {
      await checkoutStore.confirmPayOnDelivery()
      return
    } catch (error: unknown) {
      if (!isOrderGone(error)) throw error
      const order = await checkoutStore.createPendingOrder()
      if (!order) throw error
    }
    await checkoutStore.confirmPayOnDelivery()
  }

  async function submit() {
    if (isSubmitting.value || !acknowledged.value) return

    isSubmitting.value = true
    checkoutStore.paymentError = null

    try {
      await confirmWithRetry()
      router.push('/checkout/confirmation')
    } catch {
      checkoutStore.paymentError = 'Unable to place order. Please try again.'
    } finally {
      isSubmitting.value = false
    }
  }

  return {
    acknowledged,
    isSubmitting,
    submit,
  }
}
