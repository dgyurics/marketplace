<template>
  <div class="floating-panel">
    <div class="wizard-header">
      <h2>Create New Product</h2>
      <div class="wizard-steps">
        <div :class="['step', { active: currentStep === 1, completed: currentStep > 1 }]">
          <span class="step-number">1</span>
          <span class="step-label">Images</span>
        </div>
        <div :class="['step', { active: currentStep === 2, completed: currentStep > 2 }]">
          <span class="step-number">2</span>
          <span class="step-label">Info</span>
        </div>
        <div :class="['step', { active: currentStep === 3 }]">
          <span class="step-number">3</span>
          <span class="step-label">Details</span>
        </div>
      </div>
    </div>

    <div class="wizard-content">
      <div v-show="currentStep === 1" class="step-content">
        <LoadingSpinner v-if="creating" />
        <template v-if="productId">
          <ImageUploader
            :product-id="productId"
            :images="images"
            @upload-success="handleImageUploadSuccess"
          />
          <ImageGallery v-if="images.length" :images="images" @image-deleted="handleImageDeleted" />
        </template>
        <div class="form-actions">
          <button v-if="!productId && !creating" type="button" @click="createDraft">
            Try again
          </button>
          <button
            type="button"
            class="btn-primary"
            :disabled="!images.length"
            @click="currentStep = 2"
          >
            Next
          </button>
        </div>
      </div>

      <form v-show="currentStep === 2" class="step-content" @submit.prevent="goToDetails">
        <div v-if="assistantEnabled" class="form-group">
          <button
            type="button"
            class="assistant-button"
            :disabled="autofilling || !images.some((image) => image.type === 'hero')"
            :aria-expanded="showAssistant"
            aria-controls="product-assistant"
            @click="showAssistant = !showAssistant"
          >
            <SparklesIcon class="assistant-icon" aria-hidden="true" />
            <span>Auto-fill</span>
          </button>
          <div v-if="showAssistant" id="product-assistant" class="assistant-context">
            <textarea
              v-model="instructions"
              class="assistant-input"
              aria-label="Additional context"
              placeholder="Optional context or instructions…"
              :disabled="autofilling"
              rows="2"
              maxlength="4000"
            ></textarea>
            <button
              type="button"
              class="btn-xs btn-subtle assistant-generate"
              :disabled="autofilling"
              @click="autoFill"
            >
              <span>{{ autofilling ? 'Generating…' : 'Generate' }}</span>
            </button>
          </div>
        </div>
        <div class="form-group">
          <InputText v-model="newProduct.name" label="Name" required />
        </div>
        <div class="form-row-2col">
          <InputCurrency
            id="new-product-price"
            v-model="displayPrice"
            label="Price"
            label-position="inside"
            required
          />
          <InputNumber v-model="newProduct.inventory" label="Inventory" min="0" required />
        </div>
        <div class="form-group">
          <InputText v-model="newProduct.summary" label="Summary" />
        </div>
        <div class="form-group">
          <SelectInput v-model="newProduct.category" label="Category" :options="categoryOptions" />
        </div>

        <div class="checkbox-group">
          <label class="checkbox-label">
            <input v-model="newProduct.featured" type="checkbox" />
            <span>Featured</span>
          </label>
          <label class="checkbox-label">
            <input v-model="newProduct.negotiable" type="checkbox" />
            <span>Negotiable</span>
          </label>
        </div>
        <div class="form-actions">
          <button type="button" class="btn-primary btn-outline" @click="currentStep = 1">
            Back
          </button>
          <button type="submit" class="btn-primary">Next</button>
        </div>
      </form>
      <form v-show="currentStep === 3" class="step-content" @submit.prevent="handleSubmit">
        <div class="form-group">
          <TextArea v-model="newProduct.description" label="Description" />
        </div>
        <div class="form-row-2col">
          <InputNumber v-model="newProduct.cart_limit" label="Cart Limit" min="1" />
          <InputNumber v-model="newProduct.sort_order" label="Sort Order" min="0" />
        </div>
        <div class="form-group">
          <InputText v-model="newProduct.tax_code" label="Tax Code" />
        </div>
        <KeyValueEditor
          :key="detailsVersion"
          ref="detailsEditor"
          v-model="newProduct.details"
          key-placeholder="Key (e.g., color, size, material)"
          value-placeholder="Value"
          pair-name="Detail"
          @update:model-value="detailsEdited = true"
        />

        <div class="form-actions">
          <button
            type="button"
            class="btn-primary btn-outline"
            :disabled="saving"
            @click="currentStep = 2"
          >
            Back
          </button>
          <button
            type="submit"
            class="btn-primary"
            :disabled="saving || autofilling || !images.length"
          >
            {{ saving ? 'Publishing…' : 'Publish' }}
          </button>
        </div>
      </form>
    </div>
    <div v-if="errorMessage" class="error" role="alert">{{ errorMessage }}</div>
  </div>
</template>

<script setup lang="ts">
import { SparklesIcon } from '@heroicons/vue/24/outline'
import { ref, computed, onMounted } from 'vue'

import {
  InputCurrency,
  InputNumber,
  InputText,
  KeyValueEditor,
  SelectInput,
  TextArea,
} from '@/components/forms'
import ImageGallery from '@/components/ImageGallery.vue'
import ImageUploader from '@/components/ImageUploader.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { createProduct, updateProduct, autofillProduct } from '@/services/api'
import type { Category, Image } from '@/types'
import { toMinorUnits, toMajorUnits } from '@/utilities'
import { getAssistantOptions } from '@/utilities/config'

const props = defineProps<{
  categories: Category[]
  onSuccess?: () => void
}>()

const assistantEnabled = getAssistantOptions().enabled
const showAssistant = ref(false)
const instructions = ref('')
const generatedText = { name: '', summary: '', description: '' }
const autofilling = ref(false)
const detailsVersion = ref(0)
const detailsEdited = ref(false)
const currentStep = ref(1)
const productId = ref('')
const images = ref<Image[]>([])
const creating = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const detailsEditor = ref<{ hasErrors: () => boolean } | null>(null)
const newProduct = ref({
  name: '',
  price: '' as number | '',
  inventory: '' as number | '',
  summary: '',
  category: '',
  description: '',
  cart_limit: '' as number | '',
  sort_order: '' as number | '',
  tax_code: '',
  featured: true,
  negotiable: false,
  details: {} as Record<string, string>,
})

const displayPrice = computed({
  get: () => (newProduct.value.price === '' ? '' : toMajorUnits(newProduct.value.price)),
  set: (value) => (newProduct.value.price = value === '' ? '' : toMinorUnits(value)),
})
const categoryOptions = computed(() =>
  props.categories.map((category) => ({ value: category.slug, label: category.name }))
)

const createDraft = async () => {
  if (creating.value || productId.value) return
  creating.value = true
  errorMessage.value = ''
  try {
    const product = await createProduct({
      name: '',
      price: 0,
      summary: '',
      description: '',
      details: {},
      status: 'draft',
    })
    productId.value = product.id
  } catch {
    errorMessage.value = 'Could not start a new product. Please try again.'
  } finally {
    creating.value = false
  }
}

const handleImageUploadSuccess = (uploadedImages: Image[]) => images.value.push(...uploadedImages)
const handleImageDeleted = (imageId: string) => {
  images.value = images.value.filter((image) => image.id !== imageId)
}

const autoFill = async () => {
  if (autofilling.value || !productId.value || !images.value.some((image) => image.type === 'hero'))
    return
  autofilling.value = true
  errorMessage.value = ''
  try {
    const fields = await autofillProduct(productId.value, instructions.value)
    for (const key of ['name', 'summary', 'description'] as const) {
      if (!newProduct.value[key].trim() || newProduct.value[key] === generatedText[key]) {
        newProduct.value[key] = fields[key]
        generatedText[key] = fields[key]
      }
    }
    if (!detailsEdited.value && Object.keys(fields.details).length) {
      newProduct.value.details = fields.details
      detailsVersion.value++
    }
  } catch {
    errorMessage.value = 'Could not auto-fill product fields. Please try again.'
  } finally {
    autofilling.value = false
  }
}

const goToDetails = () => {
  if (
    !newProduct.value.name.trim() ||
    newProduct.value.price === '' ||
    newProduct.value.inventory === ''
  )
    return
  currentStep.value = 3
}

const handleSubmit = async () => {
  if (
    saving.value ||
    autofilling.value ||
    !productId.value ||
    !images.value.length ||
    detailsEditor.value?.hasErrors()
  )
    return
  if (
    !newProduct.value.name.trim() ||
    newProduct.value.price === '' ||
    newProduct.value.inventory === ''
  )
    return
  saving.value = true
  errorMessage.value = ''
  try {
    const category = props.categories.find(
      (category) => category.slug === newProduct.value.category
    )
    await updateProduct({
      id: productId.value,
      name: newProduct.value.name.trim(),
      price: newProduct.value.price,
      inventory: Number(newProduct.value.inventory),
      summary: newProduct.value.summary,
      ...(newProduct.value.description && { description: newProduct.value.description }),
      ...(newProduct.value.tax_code && { tax_code: newProduct.value.tax_code }),
      details: newProduct.value.details,
      ...(newProduct.value.cart_limit !== '' && {
        cart_limit: Number(newProduct.value.cart_limit),
      }),
      sort_order: Number(newProduct.value.sort_order || 0),
      featured: newProduct.value.featured,
      negotiable: newProduct.value.negotiable,
      ...(category && { category }),
      status: 'published',
    })
    props.onSuccess?.()
  } catch {
    errorMessage.value = 'Could not publish the product. Please try again.'
  } finally {
    saving.value = false
  }
}

onMounted(createDraft)
</script>

<style scoped>
.assistant-button {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
  padding: 4px 0;
  border: none;
  background: transparent;
  color: #666;
  font-size: 12px;
  line-height: 16px;
}

.assistant-button > span {
  display: block;
}

.assistant-button:hover {
  background: transparent;
  color: #222;
}

.assistant-button:disabled {
  background: transparent;
  color: #aaa;
}

.assistant-icon {
  width: 14px;
  height: 14px;
  display: block;
  flex-shrink: 0;
}

.assistant-context {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 6px;
}

.assistant-input {
  width: 100%;
  box-sizing: border-box;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 1px;
  font-family: inherit;
  font-size: 12px;
  line-height: 1.5;
  resize: vertical;
}

.assistant-input::placeholder {
  color: #888;
}

.assistant-generate {
  align-self: flex-start;
}

.wizard-header {
  margin-bottom: 30px;
}

.wizard-steps {
  display: flex;
  gap: 20px;
  justify-content: center;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  opacity: 0.5;
  transition: opacity 0.3s ease;
}

.step.active {
  opacity: 1;
}

.step.completed {
  opacity: 0.7;
}

.step-number {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 1.1rem;
  transition: all 0.3s ease;
}

.step.active .step-number {
  background: #000;
  color: #fff;
}

.step-label {
  font-size: 0.9rem;
  font-weight: 500;
}

.wizard-content {
  min-height: 300px;
}

.step-content {
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.step-content h3 {
  margin: 0 0 10px 0;
  font-size: 1.3rem;
  font-weight: 400;
}

.step-description {
  margin: 0 0 20px 0;
  color: #666;
  font-size: 0.95rem;
}

.form-group {
  margin-bottom: 10px;
}

.checkbox-group {
  display: flex;
  gap: 20px;
  margin-bottom: 10px;
  margin-top: 8px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 13px;
  color: #666;
  margin-bottom: 0;
  margin-top: 0;
  margin-left: 0;
}

.checkbox-label input[type='checkbox'] {
  cursor: pointer;
  width: 18px;
  height: 18px;
  accent-color: #000;
}

.form-row-2col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 20px;
}

.details-editor-wrapper {
  margin: 30px 0;
}

.form-actions {
  display: flex;
  gap: 10px;
  margin-top: 30px;
  justify-content: center;
}

.btn-primary {
  flex: 1;
  max-width: 300px;
}

@media (max-width: 480px) {
  .form-row-2col {
    grid-template-columns: 1fr;
  }
}
</style>
