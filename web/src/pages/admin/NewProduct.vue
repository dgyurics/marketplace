<template>
  <div class="container">
    <h2>Create Product</h2>
    <div>
      <NewProductForm :error-message="errorMessage" @submit="handleSubmit" />
      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import NewProductForm from '@/components/admin/NewProductForm.vue'
import { createProduct } from '@/services/api'
import type { CreateProductRequest } from '@/types'

const router = useRouter()
const errorMessage = ref('')

const handleSubmit = async (productData: CreateProductRequest, categorySlug: string) => {
  errorMessage.value = ''

  try {
    const result = await createProduct(productData, categorySlug)

    // Redirect to image upload page on success
    router.push(`/admin/products/${result.id}/images`)
  } catch (error) {
    errorMessage.value = `Failed to create product: ${error instanceof Error ? error.message : 'Unknown error'}`
  }
}
</script>

<style scoped>
.container {
  text-align: center;
}
</style>
