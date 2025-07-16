<template>
  <form @submit.prevent="handleSubmit">
    <div class="form-group">
      <label for="imageType">Image Type</label>
      <select id="imageType" v-model="imageType" required>
        <option value="gallery">Gallery</option>
        <option value="hero">Hero</option>
        <option value="thumbnail">Thumbnail</option>
      </select>
    </div>

    <div class="form-group">
      <label for="imageFile">Select Image</label>
      <input
        id="imageFile"
        ref="fileInput"
        type="file"
        accept="image/jpeg,image/png,image/webp,image/gif,image/bmp,image/tiff"
        required
        @change="handleFileSelect"
      />
      <small v-if="selectedFile" class="file-info">
        Selected: {{ selectedFile.name }} ({{ (selectedFile.size / 1024 / 1024).toFixed(2) }} MB)
      </small>

      <div class="checkbox-group">
        <input id="removeBackground" v-model="removeBackground" type="checkbox" />
        <label for="removeBackground" class="checkbox-label">Remove Background</label>
      </div>
    </div>

    <button type="submit" class="submit-button" :disabled="props.isUploading">
      <LoadingSpinner v-if="props.isUploading" size="md" />
      <SuccessCheck v-else-if="props.isSuccess" size="md" />
      <span v-else>Upload</span>
    </button>

    <p v-if="props.errorMessage" class="error">{{ props.errorMessage }}</p>
  </form>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import LoadingSpinner from '@/components/LoadingSpinner.vue'
import SuccessCheck from '@/components/SuccessCheck.vue'
import type { ImageType } from '@/types'

interface Props {
  isUploading?: boolean
  isSuccess?: boolean
  errorMessage?: string
}

const props = withDefaults(defineProps<Props>(), {
  isUploading: false,
  isSuccess: false,
  errorMessage: '',
})

const emit = defineEmits<{
  submit: [file: File, imageType: ImageType, removeBackground: boolean]
}>()

const imageType = ref<ImageType>('gallery') // Default to 'gallery'
const selectedFile = ref<File | null>(null)
const fileInput = ref<HTMLInputElement>()
const removeBackground = ref(true)

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    selectedFile.value = target.files[0]
  }
}

const handleSubmit = () => {
  if (selectedFile.value) {
    emit('submit', selectedFile.value, imageType.value, removeBackground.value)
  }
}

const resetForm = () => {
  imageType.value = 'gallery'
  selectedFile.value = null
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
.checkbox-group {
  display: flex;
  align-items: center;
  margin-top: 10px;
}

input[type='checkbox'] {
  width: auto;
  margin-right: 8px;
  cursor: pointer;
}

.checkbox-label {
  display: inline;
  margin-bottom: 0;
  cursor: pointer;
}

label {
  font-weight: 500;
  font-size: 14px;
  display: block;
  margin-bottom: 5px;
}

input {
  width: 100%;
  max-width: 100%;
  font-size: 18px;
}

input[type='text'],
input[type='email'],
input[type='password'],
input[type='tel'],
input[type='number'],
input[type='search'],
input[type='file'],
select {
  width: 100%;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 18px;
  box-sizing: border-box;
  background-color: transparent;
}

.file-info {
  font-size: 10px;
  color: #666;
  margin-top: 2px;
}
</style>
