<template>
  <div class="container">
    <h2>Image Upload</h2>
    <div>
      <NewImageForm
        ref="imageFormRef"
        :is-uploading="isUploading"
        :is-success="isSuccess"
        :error-message="errorMessage"
        @submit="handleImageUpload"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'

import NewImageForm from '@/components/admin/NewImageForm.vue'
import { uploadProductImage } from '@/services/api'
import type { ApiError, ImageType } from '@/types'

const route = useRoute()
const productId = route.params['id'] as string

const isUploading = ref(false)
const isSuccess = ref(false)
const imageFormRef = ref()

const errorMessage = ref<string>('')

const handleImageUpload = async (file: File, imageType: ImageType, removeBg: Boolean) => {
  isUploading.value = true
  errorMessage.value = ''

  try {
    await uploadProductImage(productId, file, imageType, removeBg)

    // Immediately stop loading and show success
    isUploading.value = false
    isSuccess.value = true

    // Reset form and clear success state after 5 seconds
    setTimeout(() => {
      isSuccess.value = false
      imageFormRef.value?.resetForm()
    }, 5000)
  } catch (error) {
    isUploading.value = false
    errorMessage.value = handleApiError(error)
  }
}

const handleApiError = (error: unknown): string => {
  if (error && typeof error === 'object' && 'response' in error) {
    const apiError = error as ApiError
    switch (apiError.response?.status) {
      case 400:
        return 'Invalid request. Please check the input.'
      case 401:
        return 'Unauthorized. Please log in again.'
      case 403:
        return 'Forbidden. You do not have permission to perform this action.'
      case 404:
        return 'Invalid Product ID.'
      case 500:
        return 'Server error. Please try again later.'
      default:
        return apiError.message || 'An error occurred uploading the image.'
    }
  }

  if (error instanceof Error) {
    return error.message
  }

  return 'An unexpected error occurred. Please try again.'
}
</script>

<style scoped>
.container {
  text-align: center;
}
</style>
