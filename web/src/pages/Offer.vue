<template>
  <div class="offer-container">
    <template v-if="success">
      <h2>Offer Submitted Successfully</h2>
    </template>
    <template v-else-if="hasPendingOffer">
      <h2>Offer Pending</h2>
      <p>You already have a pending offer for this product.</p>
    </template>
    <template v-else>
      <h2>Submit Offer</h2>
      <h3 v-if="product" class="capitalize">{{ product.name }}</h3>

      <form @submit.prevent="handleSubmit">
        <div class="form-group-flex">
          <InputCurrency id="offer-price" v-model="displayOfferPrice" label="Your Offer" required />
          <p class="help-text">Asking price: {{ displayPrice(product?.price || 0) }}</p>
        </div>

        <!-- Optional comment -->
        <div class="form-group-flex">
          <TextArea v-model="comment" label="Comment" :resizable="false"></TextArea>
        </div>

        <div class="button-group">
          <button
            type="submit"
            class="btn-full-width mt-15"
            :tabindex="0"
            :disabled="isSubmitting || !offersLoaded"
          >
            <span v-if="!isSubmitting">Submit Offer</span>
          </button>
        </div>

        <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
      </form>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'

import { InputCurrency, TextArea } from '@/components/forms'
import { createOffer, getProductById, getOffersByProductId } from '@/services/api'
import type { Product } from '@/types'
import { displayPrice, toMajorUnits, toMinorUnits } from '@/utilities/currency'

const route = useRoute()

const product = ref<Product | null>(null)
const comment = ref('')
const offerPriceValue = ref<number | ''>('')
const success = ref(false)
const isSubmitting = ref(false)
const errorMessage = ref<string | null>(null)
const hasPendingOffer = ref(false)
const offersLoaded = ref(false)

const displayOfferPrice = computed({
  get: () => (offerPriceValue.value === '' ? '' : toMajorUnits(offerPriceValue.value)),
  set: (value) => (offerPriceValue.value = value === '' ? '' : toMinorUnits(value)),
})

onMounted(async () => {
  try {
    const productId = route.params['id'] as string
    product.value = await getProductById(productId)
    const offers = await getOffersByProductId(productId)
    hasPendingOffer.value = offers.some((offer) => offer.status === 'pending')
    offersLoaded.value = true
  } catch (error) {
    console.error('Error loading offer form:', error)
    errorMessage.value = 'Unable to load offer details.'
  }
})

const handleSubmit = async () => {
  if (!offersLoaded.value || hasPendingOffer.value || isSubmitting.value) return

  errorMessage.value = null

  if (offerPriceValue.value === '' || offerPriceValue.value < 0) {
    errorMessage.value = 'Invalid offer amount'
    return
  }

  try {
    isSubmitting.value = true
    const productId = route.params['id'] as string
    const finalOfferPrice = offerPriceValue.value

    await createOffer(productId, finalOfferPrice, comment.value.trim())
    success.value = true
  } catch (error: any) {
    const status = error.response?.status
    if (status === 404) {
      errorMessage.value = 'Product not found'
    } else if (status === 409) {
      errorMessage.value = 'Item no longer available'
    } else {
      errorMessage.value = 'Something went wrong'
    }
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.offer-container {
  max-width: 600px;
  margin: auto;
  padding: 20px;
  text-align: center;
  position: relative;
  top: -20px;
}

h3 {
  font-size: 16px;
  font-weight: 400;
  color: #666;
  margin-bottom: 40px;
  margin-top: 0;
}

.form-group-flex {
  display: flex;
  flex-direction: column;
  margin-bottom: 20px;
  text-align: left;
}

.help-text {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
  margin-bottom: 0;
}

.button-group {
  margin-top: 30px;
}

.error {
  text-align: center;
  font-size: 12px;
  color: #c00;
  margin-top: 8px;
}

.capitalize {
  text-transform: capitalize;
}
</style>
