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

  async function submit() {
    if (isSubmitting.value || !acknowledged.value) return

    isSubmitting.value = true
    checkoutStore.paymentError = null

    try {
      await checkoutStore.confirmPayOnDelivery()
      router.push('/checkout/confirmation')
    } catch (error: unknown) {
      const status = (error as { response?: { status?: number } }).response?.status
      if (status === 400) {
        checkoutStore.shippingError = 'Invalid shipping address'
        router.push('/checkout/shipping')
        return
      }
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
