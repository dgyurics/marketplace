<template>
  <div class="image-gallery">
    <h3>Product Images</h3>
    <div v-if="!images.length" class="no-images">
      <p>No images uploaded yet</p>
    </div>
    <div v-else class="table-container">
      <table class="images-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>URL</th>
            <th>Type</th>
            <th></th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="image in images" :key="image.id" class="image-row">
            <td class="id-cell">{{ image.id }}</td>
            <td class="url-cell">
              <span
                class="url-link"
                @mouseenter="showPreview(image, $event)"
                @mouseleave="hidePreview"
                @click="togglePreview(image, $event)"
              >
                {{ truncateUrl(image.url) }}
              </span>
            </td>
            <td class="type-cell">{{ image.type }}</td>
            <td class="actions-cell">
              <button
                type="button"
                class="promote-btn"
                title="Promote image"
                @click="handlePromote(image.id)"
              >
                ↑
              </button>
            </td>
            <td class="actions-cell">
              <button
                type="button"
                class="remove-btn"
                title="Delete image"
                @click="handleDelete(image.id)"
              >
                ×
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Image Preview Popup -->
    <div v-if="previewImage" class="image-preview" :style="previewStyle" @click="hidePreview">
      <img
        :src="previewImage.url"
        :alt="previewImage.alt_text || 'Preview'"
        class="preview-img"
        @load="adjustPreviewPosition"
      />
      <div class="preview-info">
        <span class="preview-type">{{ previewImage.type }}</span>
        <span class="preview-id">{{ previewImage.id }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

import { removeImage, promoteImage } from '@/services/api'

defineProps({
  images: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits<{
  'image-deleted': [imageId: string]
  'image-promoted': [imageId: string]
}>()

const previewImage = ref(null)
const previewStyle = ref({})
const isClickPreview = ref(false)

const truncateUrl = (url) => {
  if (url.length <= 50) return url
  return `${url.substring(0, 47)}...`
}

const handleDelete = async (imageId, productId) => {
  try {
    await removeImage(imageId, productId)
    emit('image-deleted', imageId)
  } catch {
    // Handle error silently, consistent with other error handling in the app
  }
}

const handlePromote = async (imageId) => {
  try {
    await promoteImage(imageId)
    emit('image-promoted', imageId)
  } catch {
    // Handle error silently, consistent with other error handling in the app
  }
}

const showPreview = (image, event) => {
  if (isClickPreview.value) return // Don't show hover preview if click preview is active

  previewImage.value = image
  updatePreviewPosition(event)
}

const hidePreview = () => {
  if (isClickPreview.value) return // Don't hide if it's a click preview
  previewImage.value = null
}

const togglePreview = (image, event) => {
  if (previewImage.value && previewImage.value.id === image.id && isClickPreview.value) {
    // Clicking on the same image again - hide preview
    previewImage.value = null
    isClickPreview.value = false
  } else {
    // Show click preview
    previewImage.value = image
    isClickPreview.value = true
    updatePreviewPosition(event)
  }
}

const updatePreviewPosition = (event) => {
  const rect = event.target.getBoundingClientRect()
  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight

  let left = rect.right + 10
  let top = rect.top

  // Adjust if preview would go off-screen
  if (left + 300 > viewportWidth) {
    left = rect.left - 310
  }

  if (top + 200 > viewportHeight) {
    top = viewportHeight - 220
  }

  previewStyle.value = {
    position: 'fixed',
    left: `${left}px`,
    top: `${top}px`,
    zIndex: 1000,
  }
}

const adjustPreviewPosition = () => {
  // Called when image loads to readjust position if needed
  if (previewImage.value) {
    const preview = document.querySelector('.image-preview')
    if (preview) {
      const rect = preview.getBoundingClientRect()
      const viewportWidth = window.innerWidth
      const viewportHeight = window.innerHeight

      let { left, top } = previewStyle.value
      left = parseInt(left)
      top = parseInt(top)

      if (rect.right > viewportWidth) {
        left = viewportWidth - rect.width - 10
      }
      if (rect.bottom > viewportHeight) {
        top = viewportHeight - rect.height - 10
      }

      previewStyle.value = {
        ...previewStyle.value,
        left: `${left}px`,
        top: `${top}px`,
      }
    }
  }
}

// Hide click preview when clicking outside
const handleClickOutside = (event) => {
  if (
    isClickPreview.value &&
    !event.target.closest('.image-preview') &&
    !event.target.closest('.url-link')
  ) {
    previewImage.value = null
    isClickPreview.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.image-gallery {
  margin: 20px 0;
}

.image-gallery h3 {
  margin: 0 0 15px 0;
  font-size: 16px;
  color: #333;
}

.no-images {
  text-align: center;
  padding: 40px 20px;
  color: #666;
  font-style: italic;
  border: 1px solid #ddd;
  border-radius: 4px;
  background-color: #f9f9f9;
}

.table-container {
  border: 1px solid #ddd;
  border-radius: 4px;
  overflow-x: auto;
  background-color: white;
}

.images-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.images-table th {
  background-color: #f8f9fa;
  padding: 12px 8px;
  text-align: left;
  font-weight: 600;
  color: #495057;
  border-bottom: 2px solid #dee2e6;
}

.images-table td {
  padding: 10px 8px;
  border-bottom: 1px solid #dee2e6;
  vertical-align: middle;
}

.image-row:hover {
  background-color: #f8f9fa;
}

.id-cell {
  font-family: 'Roboto Mono', monospace;
  font-size: 12px;
  color: #6c757d;
  width: 120px;
}

.url-cell {
  min-width: 200px;
  max-width: 300px;
}

.url-link {
  color: #007bff;
  cursor: pointer;
  text-decoration: underline;
  font-family: 'Roboto Mono', monospace;
  font-size: 12px;
  display: block;
  word-break: break-all;
}

.url-link:hover {
  color: #0056b3;
  background-color: #e3f2fd;
  padding: 2px 4px;
  border-radius: 3px;
}

.type-cell {
  width: 100px;
  text-transform: capitalize;
  color: #495057;
}

.actions-cell {
  width: 40px;
  text-align: center;
  padding: 10px 4px;
}

.remove-btn {
  width: 24px;
  height: 24px;
  border: none;
  background: none;
  color: #333;
  cursor: pointer;
  font-size: 16px;
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
  flex-shrink: 0;
}

.remove-btn:hover {
  color: #000;
}

.promote-btn {
  width: 24px;
  height: 24px;
  border: none;
  background: none;
  color: #333;
  cursor: pointer;
  font-size: 16px;
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
  flex-shrink: 0;
}

.promote-btn:hover {
  color: #000;
}

/* Image Preview Popup */
.image-preview {
  position: fixed;
  background: white;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  overflow: hidden;
  max-width: 300px;
  cursor: pointer;
  animation: fadeIn 0.2s ease-out;
}

.preview-img {
  width: 100%;
  max-width: 300px;
  max-height: 200px;
  object-fit: cover;
  display: block;
}

.preview-info {
  padding: 8px 12px;
  background-color: #f8f9fa;
  border-top: 1px solid #dee2e6;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.preview-type {
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  color: #495057;
}

.preview-id {
  font-family: 'Roboto Mono', monospace;
  font-size: 11px;
  color: #6c757d;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: scale(0.9);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .table-container {
    font-size: 12px;
  }

  .images-table th,
  .images-table td {
    padding: 8px 4px;
  }

  .id-cell {
    width: 80px;
  }

  .url-cell {
    min-width: 150px;
    max-width: 200px;
  }

  .image-preview {
    max-width: 250px;
  }
}
</style>
