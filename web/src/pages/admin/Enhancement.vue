<template>
  <div class="enhancement-container">
    <h1>Improve image</h1>

    <form v-if="image" @submit.prevent="generate">
      <div class="context-field">
        <label for="image-context">Describe the changes you'd like to make</label>
        <textarea
          id="image-context"
          v-model="productContext"
          rows="4"
          maxlength="1000"
          :disabled="busy"
        ></textarea>
      </div>
      <button :disabled="busy || !productContext.trim()">
        {{ busy ? 'Please wait…' : 'Create preview' }}
      </button>

      <div class="comparison">
        <section>
          <h2>Current</h2>
          <div class="image-frame">
            <img :src="image.url" alt="Current product image" />
          </div>
          <button type="button" class="btn-outline" :disabled="busy" @click="goBack">
            {{ preview ? 'Keep original' : 'Cancel' }}
          </button>
        </section>

        <section>
          <h2>Preview</h2>
          <div class="image-frame">
            <img v-if="preview" :src="preview.after_url" alt="Enhanced product image" />
            <div v-else-if="generating" class="preview-status" role="status">
              <LoadingSpinner class="preview-spinner" aria-hidden="true" />
              <span>Creating preview…</span>
            </div>
            <p v-else class="preview-placeholder">Your preview will appear here</p>
          </div>
          <button type="button" :disabled="busy || !preview" @click="accept">
            {{ busy ? 'Please wait…' : 'Use improved' }}
          </button>
        </section>
      </div>
    </form>

    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <button v-if="!image" type="button" class="btn-outline" @click="goBack">Cancel</button>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { getProductById, resolveEnhancement, generateEnhancement } from '@/services/api'
import type { Enhancement, Image } from '@/types'
import { getAppConfig } from '@/utilities/config'

const route = useRoute()
const router = useRouter()
const productId = String(route.params['id'])
const imageId = String(route.params['imageId'])
const productDraft = window.history.state.productDraft
const image = ref<Image | null>(null)
const preview = ref<Enhancement | null>(null)
const busy = ref(false)
const generating = ref(false)
const error = ref('')
const productContext = ref(
  'Keep the product unchanged. Add a subtle shadow if needed. Remove any background if one exists. Return a transparent PNG.'
)
let disposed = false
const generationController = new window.AbortController()

const getErrorMessage = (err: unknown) => {
  if (!axios.isAxiosError<{ message?: string }>(err))
    return 'Could not process the image. Your current image is unchanged.'
  const response = err.response
  if (response?.data?.message) return response.data.message
  if (response?.status === 404 || response?.status === 409)
    return 'This preview is unavailable or the image changed. Go back and try again.'
  return 'Could not process the image. Your current image is unchanged.'
}

const discardPreview = async () => {
  if (!preview.value) return
  await resolveEnhancement(imageId, preview.value.id, false)
  preview.value = null
}

const generate = async () => {
  busy.value = true
  generating.value = true
  error.value = ''
  try {
    await discardPreview()
    preview.value = await generateEnhancement(
      imageId,
      productContext.value,
      generationController.signal
    )
  } catch (err) {
    error.value = getErrorMessage(err)
  } finally {
    busy.value = false
    generating.value = false
  }

  if (disposed) void discardPreview().catch(() => {})
}

const accept = async () => {
  if (!preview.value) return
  busy.value = true
  try {
    await resolveEnhancement(imageId, preview.value.id, true)
    preview.value = null
    await goBack()
  } catch {
    error.value = 'Could not confirm the update. Go back and reload to check the current image.'
  } finally {
    busy.value = false
  }
}

const goBack = () =>
  router.replace({
    path: `/admin/products/${productId}`,
    state: { productDraft },
  })

onMounted(async () => {
  if (!getAppConfig().ai_options.image_editing) {
    await goBack()
    return
  }

  let product
  try {
    product = await getProductById(productId)
  } catch {
    error.value = 'Could not load the image.'
    return
  }

  if (disposed) return
  image.value = product.images?.find((item) => item.id === imageId) ?? null
  if (!image.value) error.value = 'Image not found.'
})

onUnmounted(() => {
  disposed = true
  generationController.abort()
  void discardPreview().catch(() => {})
})
</script>

<style scoped>
.enhancement-container {
  max-width: 960px;
  margin: 20px auto;
  padding: 20px;
}

h1 {
  margin-bottom: 24px;
}

h2 {
  margin-bottom: 12px;
  font-size: 18px;
}

.context-field {
  display: grid;
  gap: 8px;
  margin-bottom: 16px;
}

textarea {
  width: 100%;
  resize: vertical;
  padding: 12px;
  border: 1px solid #ccc;
  font: inherit;
}

.comparison {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 24px;
  margin-top: 32px;
}

.image-frame {
  display: grid;
  place-items: center;
  height: 320px;
  margin-bottom: 16px;
  border: 1px solid #ddd;
  background: repeating-conic-gradient(#f3f3f3 0% 25%, white 0% 50%) 0 / 20px 20px;
}

img {
  width: 100%;
  height: 100%;
  min-height: 0;
  object-fit: contain;
}

.preview-status {
  display: flex;
  align-items: center;
  gap: 8px;
}

.preview-status .preview-spinner {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  border-color: #489fad;
  border-top-color: transparent;
}

.preview-placeholder {
  color: #666;
}

button {
  padding: 10px 16px;
  cursor: pointer;
}

.comparison button {
  width: 100%;
}

@media (max-width: 600px) {
  .comparison {
    grid-template-columns: 1fr;
  }

  .image-frame {
    height: 240px;
  }
}
</style>
