<template>
  <div class="new-image-form">
    <form @submit.prevent="handleSubmit">
      <div class="form-group">
        <label for="productId">Product ID:</label>
        <input
          id="productId"
          v-model="productId"
          type="text"
          placeholder="Enter product ID"
          required
        />
      </div>

      <div class="form-group">
        <label for="imageType">Image Type:</label>
        <select id="imageType" v-model="imageType" required>
          <option value="gallery" default>Gallery</option>
          <option value="hero">Hero</option>
          <option value="thumbnail">Thumbnail</option>
        </select>
      </div>

      <div class="form-group">
        <label for="imageFile">Select Image:</label>
        <input
          id="imageFile"
          ref="fileInput"
          type="file"
          accept="image/*"
          required
          @change="handleFileSelect"
        />
      </div>

      <div v-if="selectedFile" class="file-preview">
        <p>Selected: {{ selectedFile.name }}</p>
        <p>Size: {{ (selectedFile.size / 1024 / 1024).toFixed(2) }} MB</p>
      </div>

      <div class="form-actions">
        <button type="submit" :disabled="!productId || !selectedFile || !imageType || isUploading">
          {{ isUploading ? 'Uploading...' : 'Upload Image' }}
        </button>
        <button type="button" @click="resetForm">Reset</button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import type { ImageType } from '@/types'

const emit = defineEmits<{
  upload: [productId: string, file: File, imageType: string]
}>()

const productId = ref('')
const imageType = ref<ImageType>('gallery') // Default to 'gallery'
const selectedFile = ref<File | null>(null)
const isUploading = ref(false)
const fileInput = ref<HTMLInputElement>()

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    selectedFile.value = target.files[0]
  }
}

const handleSubmit = () => {
  if (productId.value && selectedFile.value) {
    isUploading.value = true
    emit('upload', productId.value, selectedFile.value, imageType.value)
  }
}

const resetForm = () => {
  productId.value = ''
  imageType.value = 'gallery'
  selectedFile.value = null
  isUploading.value = false
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

// Expose methods to reset uploading state and form
defineExpose({
  resetForm,
})
</script>

<style scoped>
.new-image-form {
  max-width: 400px;
  margin: 0 auto;
  padding: 20px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background: white;
}

label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
  color: #333;
}

input,
select {
  width: 100%;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 14px;
}

input:focus,
select:focus {
  outline: none;
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
}

.file-preview {
  background: #f8f9fa;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 20px;
}

.file-preview p {
  margin: 5px 0;
  font-size: 14px;
  color: #666;
}

.form-actions {
  display: flex;
  gap: 10px;
}

.form-actions button {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
}

.form-actions button[type='submit'] {
  background: #007bff;
  color: white;
  flex: 1;
}

.form-actions button[type='submit']:hover:not(:disabled) {
  background: #0056b3;
}

.form-actions button[type='submit']:disabled {
  background: #6c757d;
  cursor: not-allowed;
}

.form-actions button[type='button'] {
  background: #6c757d;
  color: white;
}

.form-actions button[type='button']:hover {
  background: #545b62;
}
</style>
