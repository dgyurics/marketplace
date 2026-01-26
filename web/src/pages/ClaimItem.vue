<template>
  <div class="claim-container">
    <template v-if="claimed">
      <h2>Claim Successful</h2>
      <h3>
        <span class="capitalize">{{ product?.name }}</span> successfully claimed
      </h3>
    </template>
    <template v-else>
      <h2>Claim Item</h2>
      <h3 v-if="product" class="capitalize">{{ product.name }}</h3>

      <form @submit.prevent="handleSubmit">
        <div class="form-group-flex">
          <TextArea v-model="pickupNotes" label="Pickup Details" :resizable="false" required />
        </div>

        <div class="button-group">
          <button type="submit" class="btn-full-width mt-15" :tabindex="0" :disabled="isSubmitting">
            <span v-if="!isSubmitting">Claim Item</span>
            <span v-else>Claiming...</span>
          </button>
        </div>

        <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
      </form>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'

import { TextArea } from '@/components/forms'
import { claimItem, getProductById } from '@/services/api'
import type { Product } from '@/types'

const route = useRoute()

const product = ref<Product | null>(null)
const pickupNotes = ref('')
const claimed = ref(false)
const isSubmitting = ref(false)
const errorMessage = ref<string | null>(null)

onMounted(async () => {
  const productId = route.params['id'] as string
  try {
    product.value = await getProductById(productId)
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

  try {
    isSubmitting.value = true
    const productId = route.params['id'] as string
    await claimItem(productId, pickupNotes.value.trim())
    claimed.value = true
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
.claim-container {
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
