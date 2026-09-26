import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { useCartStore } from '@/store/cart'
import { useCheckoutStore } from '@/store/checkout'

/**
 * Owns the pay-on-delivery side of checkout: acknowledging the terms and
 * placing the order. There is no payment step, so the order is final.
 */
export function useDeliveryCheckout() {
  const router = useRouter()
  const cartStore = useCartStore()
  const checkoutStore = useCheckoutStore()

  const acknowledged = ref(false)
  const isSubmitting = ref(false)

  async function submit() {
    if (isSubmitting.value || !acknowledged.value) return

    isSubmitting.value = true
    checkoutStore.paymentError = null

    try {
      const res = await checkoutStore.placePayOnDeliveryOrder()
      if (!res) {
        await cartStore.fetchCart()
        router.push('/cart')
        return
      }
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
