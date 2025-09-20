<template>
  <div class="image-uploader">
    <div class="uploader-container">
      <div class="upload-section">
        <div class="form-row">
          <select v-model="imageType" class="image-type-select">
            <option
              v-for="option in availableImageTypes"
              :key="option.value"
              :value="option.value"
              :disabled="option.disabled"
            >
              {{ option.label }}
            </option>
          </select>
          <input
            ref="fileInput"
            type="file"
            accept="image/jpeg,image/png,image/webp,image/gif,image/bmp,image/tiff"
            class="file-input"
            @change="handleFileSelect"
          />
          <button
            type="button"
            class="upload-btn"
            :disabled="!selectedFile || uploading"
            @click="handleUpload"
          >
            {{ uploading ? 'Uploading...' : 'Upload Image' }}
          </button>
        </div>

        <div v-if="selectedFile" class="file-preview">
          <span class="file-info">
            {{ selectedFile.name }} ({{ (selectedFile.size / 1024 / 1024).toFixed(2) }} MB)
          </span>
        </div>

        <div class="checkbox-row">
          <label class="checkbox-label">
            <input v-model="removeBackground" type="checkbox" />
            Remove background automatically
          </label>
        </div>

        <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>
        <div v-if="successMessage" class="success-message">{{ successMessage }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

import { uploadImage } from '@/services/api'
import type { Image, ImageType } from '@/types'

const props = defineProps<{
  productId: string
  images: Image[]
}>()

// Provide default for images
const images = computed(() => props.images || [])

const emit = defineEmits(['upload-success', 'upload-error'])

const fileInput = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)
const imageType = ref<ImageType>('gallery')
const removeBackground = ref(false)
const uploading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

// Compute available image types based on existing images
const availableImageTypes = computed(() => {
  const existingTypes = images.value.map((img) => img.type)
  const hasHero = existingTypes.includes('hero')
  const hasThumbnail = existingTypes.includes('thumbnail')

  const options = [
    { value: 'hero', label: 'Hero', disabled: hasHero },
    { value: 'thumbnail', label: 'Thumbnail', disabled: hasThumbnail },
    { value: 'gallery', label: 'Gallery', disabled: false },
  ]

  // Logic for defaults and enabled options
  if (!hasHero) {
    // No hero image - only allow hero
    return options.map((opt) => ({ ...opt, disabled: opt.value !== 'hero' }))
  }
  if (!hasThumbnail) {
    // Has hero but no thumbnail - only allow thumbnail
    return options.map((opt) => ({ ...opt, disabled: opt.value !== 'thumbnail' }))
  }
  // Has both hero and thumbnail - only allow gallery
  return options.map((opt) => ({ ...opt, disabled: opt.value !== 'gallery' }))
})

// Set default imageType based on available options
const setDefaultImageType = () => {
  const availableOption = availableImageTypes.value.find((opt) => !opt.disabled)
  if (availableOption) {
    imageType.value = availableOption.value as ImageType
  }
}

// Watch for changes in images to update default
watch(
  images,
  () => {
    setDefaultImageType()
  },
  { immediate: true }
)

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    selectedFile.value = target.files[0]
    errorMessage.value = ''
    successMessage.value = ''
  }
}

const handleUpload = async () => {
  if (!selectedFile.value || !props.productId) return

  try {
    uploading.value = true
    errorMessage.value = ''
    successMessage.value = ''

    await uploadImage(props.productId, selectedFile.value, imageType.value, removeBackground.value)

    successMessage.value = 'Image uploaded successfully!'
    emit('upload-success')

    // Reset form
    resetForm()

    // Clear success message after 3 seconds
    setTimeout(() => {
      successMessage.value = ''
    }, 3000)
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
    emit('upload-error', error)
  } finally {
    uploading.value = false
  }
}

const resetForm = () => {
  selectedFile.value = null
  setDefaultImageType()
  removeBackground.value = false
  errorMessage.value = ''
  successMessage.value = ''
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

const getErrorMessage = (error: unknown) => {
  const status = (error as { response?: { status?: number } }).response?.status
  switch (status) {
    case 400:
      return 'Invalid request. Please check the file format.'
    case 401:
      return 'Unauthorized. Please log in again.'
    case 403:
      return 'Forbidden. You do not have permission to upload images.'
    case 404:
      return 'Product not found.'
    case 413:
      return 'File too large. Please choose a smaller image.'
    case 415:
      return 'Unsupported file format. Please use JPEG, PNG, WebP, GIF, BMP, or TIFF.'
    default:
      return 'Something went wrong'
  }
}

// Expose methods for parent component
defineExpose({
  resetForm,
})
</script>

<style scoped>
.image-uploader {
  margin: 20px 0;
}

.uploader-container {
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 15px;
  background-color: #f9f9f9;
}

.upload-section {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.form-row {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}

.image-type-select {
  padding: 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 14px;
  background-color: white;
  min-width: 120px;
}

.image-type-select option:disabled {
  color: #999;
  background-color: #f5f5f5;
}

.file-input {
  flex: 1;
  padding: 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 14px;
  background-color: white;
  min-width: 200px;
}

.upload-btn {
  padding: 8px 16px;
  background-color: #000;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
  white-space: nowrap;
}

.upload-btn:hover:not(:disabled) {
  background-color: #333;
}

.upload-btn:disabled {
  background-color: #6c757d;
  cursor: not-allowed;
}

.file-preview {
  padding: 8px;
  background-color: #e9ecef;
  border-radius: 4px;
  font-size: 14px;
}

.file-info {
  color: #495057;
}

.checkbox-row {
  display: flex;
  align-items: center;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  cursor: pointer;
}

.checkbox-label input[type='checkbox'] {
  margin: 0;
}

.error-message {
  color: #e74c3c;
  font-size: 14px;
  padding: 8px;
  background-color: #fdeaea;
  border: 1px solid #f5c6cb;
  border-radius: 4px;
}

.success-message {
  color: #155724;
  font-size: 14px;
  padding: 8px;
  background-color: #d4edda;
  border: 1px solid #c3e6cb;
  border-radius: 4px;
}

@media (max-width: 768px) {
  .form-row {
    flex-direction: column;
    align-items: stretch;
  }

  .file-input,
  .image-type-select,
  .upload-btn {
    min-width: 100%;
  }
}
</style>
