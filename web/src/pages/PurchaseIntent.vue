<template>
  <div class="purchase-intent-container">
    <template v-if="success">
      <h2>{{ successMessage }}</h2>
      <h3>
        <span class="capitalize">{{ product?.name }}</span>
      </h3>
    </template>
    <template v-else>
      <h2>{{ isFreeItem ? 'Claim Item' : 'Submit Offer' }}</h2>
      <h3 v-if="product" class="capitalize">{{ product.name }}</h3>

      <form @submit.prevent="handleSubmit">
        <!-- Offer field for non-free items -->
        <!-- TODO: convert to component -->
        <div v-if="!isFreeItem" class="form-group-flex">
          <label for="offer-price">Your Offer</label>
          <div class="price-input-wrapper">
            <span class="price-symbol">$</span>
            <input
              id="offer-price"
              v-model.number="displayOfferPrice"
              type="number"
              step="0.01"
              min="0"
              placeholder="0.00"
              required
            />
          </div>
          <p class="help-text">Product price: {{ displayPrice(product?.price || 0) }}</p>
        </div>

        <!-- Pickup details for all items -->
        <div class="form-group-flex">
          <TextArea v-model="pickupNotes" label="Pickup Details" :resizable="false" required />
        </div>

        <div class="button-group">
          <button type="submit" class="btn-full-width mt-15" :tabindex="0" :disabled="isSubmitting">
            <span v-if="!isSubmitting">{{ isFreeItem ? 'Claim Item' : 'Submit Offer' }}</span>
            <span v-else>{{ isFreeItem ? 'Claiming...' : 'Submitting...' }}</span>
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

import { TextArea } from '@/components/forms'
import { createPurchaseIntent, getProductById } from '@/services/api'
import type { Product } from '@/types'
import { displayPrice, toMajorUnits, toMinorUnits } from '@/utilities/currency'

const route = useRoute()

const product = ref<Product | null>(null)
const pickupNotes = ref('')
const offerPriceValue = ref<number>(0)
const success = ref(false)
const isSubmitting = ref(false)
const errorMessage = ref<string | null>(null)

const isFreeItem = computed(() => product.value?.price === 0)

const displayOfferPrice = computed({
  get: () => toMajorUnits(offerPriceValue.value),
  set: (value) => (offerPriceValue.value = toMinorUnits(value)),
})

const successMessage = computed(() => {
  return isFreeItem.value ? 'Item Claimed Successfully' : 'Offer Submitted Successfully'
})

onMounted(async () => {
  const productId = route.params['id'] as string
  try {
    product.value = await getProductById(productId)
    if (!isFreeItem.value) {
      offerPriceValue.value = product.value?.price || 0
    }
  } catch (error) {
    console.error('Error fetching product:', error)
  }
})

const handleSubmit = async () => {
  errorMessage.value = null

  if (!pickupNotes.value.trim()) {
    errorMessage.value = 'Pickup details must be provided'
    return
  }

  if (!isFreeItem.value && (offerPriceValue.value === null || offerPriceValue.value < 0)) {
    errorMessage.value = 'Please enter a valid offer amount'
    return
  }

  try {
    isSubmitting.value = true
    const productId = route.params['id'] as string
    const finalOfferPrice = isFreeItem.value ? 0 : offerPriceValue.value

    await createPurchaseIntent(productId, finalOfferPrice, pickupNotes.value.trim())
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
.purchase-intent-container {
  max-width: 600px;
  margin: auto;
  padding: 20px;
  text-align: center;
  position: relative;
  top: -20px;
}

h2 {
  font-size: 22px;
  font-weight: 300;
  margin-bottom: 10px;
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

.form-group-flex label {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 8px;
}

.form-group-flex input {
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.price-input-wrapper {
  display: flex;
  align-items: center;
  border: 1px solid #ddd;
  border-radius: 4px;
  overflow: hidden;
}

.price-symbol {
  padding: 10px 8px;
  background-color: #f5f5f5;
  border-right: 1px solid #ddd;
  font-weight: 500;
  color: #333;
}

.price-input-wrapper input {
  flex: 1;
  border: none;
  padding: 10px;
  font-size: 14px;
  margin: 0;
  border-radius: 0;
}

.price-input-wrapper input::-webkit-outer-spin-button,
.price-input-wrapper input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.price-input-wrapper input[type='number'] {
  -moz-appearance: textfield;
}

.price-input-wrapper input:focus {
  outline: none;
  background-color: #fafafa;
}

.help-text {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
  margin-bottom: 0;
}

.button-group {
  display: flex;
  justify-content: space-between;
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
