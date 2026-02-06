<template>
  <div class="wizard-section">
    <div class="wizard-header">
      <h2>Create New Product</h2>
      <div class="wizard-steps">
        <div :class="['step', { active: currentStep === 1, completed: currentStep > 1 }]">
          <span class="step-number">1</span>
          <span class="step-label">Info</span>
        </div>
        <div :class="['step', { active: currentStep === 2, completed: currentStep > 2 }]">
          <span class="step-number">2</span>
          <span class="step-label">Images</span>
        </div>
        <div :class="['step', { active: currentStep === 3 }]">
          <span class="step-number">3</span>
          <span class="step-label">Details</span>
        </div>
      </div>
    </div>

    <div class="wizard-content">
      <!-- Step 1: Basic Info -->
      <div v-show="currentStep === 1" class="step-content">
        <form @submit.prevent="goToStep(2)">
          <div class="form-group">
            <InputText v-model="newProduct.name" label="Name" required />
          </div>

          <div class="form-row-2col">
            <div class="form-group">
              <InputNumber v-model="displayPrice" label="Price" step="0.01" required />
            </div>
            <div class="form-group">
              <InputNumber v-model="newProduct.inventory" label="Inventory" required />
            </div>
          </div>

          <div class="form-group">
            <InputText
              v-model="newProduct.summary"
              label="Summary"
              required
              placeholder="Brief 1-2 sentence description"
            />
          </div>

          <div class="form-row-2col">
            <div class="form-group">
              <SelectInput
                v-model="newProduct.category"
                label="Category"
                :options="categoryOptions"
              />
            </div>
            <div class="form-group">
              <InputNumber v-model="newProduct.cart_limit" label="Cart Limit" />
            </div>
          </div>

          <div class="form-group">
            <TextArea
              v-model="newProduct.description"
              label="Description"
              placeholder="Detailed description of the product"
            ></TextArea>
          </div>

          <div class="form-actions">
            <button type="submit" class="btn-primary" :tabindex="0">Next</button>
          </div>
        </form>
      </div>

      <!-- Step 2: Images -->
      <div v-show="currentStep === 2" class="step-content">
        <!-- Image Gallery Section -->
        <ImageGallery
          :images="tempProductImages"
          @image-deleted="handleImageDeleted"
          @image-promoted="handleImagePromoted"
        />

        <!-- Image Upload Section -->
        <div v-if="tempProductId" class="step-content">
          <ImageUploader
            :product-id="tempProductId"
            :images="tempProductImages"
            @upload-success="handleImageUploadSuccess"
            @upload-error="handleImageUploadError"
            @image-deleted="handleImageDeleted"
          />
        </div>

        <div class="form-actions">
          <button
            type="button"
            class="btn-primary"
            :disabled="tempProductImages.length === 0"
            :tabindex="0"
            @click="goToStep(3)"
          >
            Next
          </button>
        </div>
      </div>

      <!-- Step 3: Details & Review -->
      <div v-show="currentStep === 3" class="step-content">
        <div v-if="successMessage" class="success-message">
          <p>{{ successMessage }}</p>
          <button class="btn-primary" @click="resetAndStart">Create Another</button>
        </div>

        <div v-else>
          <div class="form-group">
            <InputText v-model="newProduct.tax_code" label="Tax Code" />
          </div>

          <div class="details-editor-wrapper">
            <KeyValueEditor
              ref="detailsEditor"
              v-model="newProduct.details"
              key-placeholder="Key (e.g., color, size, material)"
              value-placeholder="Value"
              pair-name="Detail"
            />
          </div>

          <div class="form-actions">
            <button
              v-auth="'admin'"
              type="button"
              class="btn-primary"
              :tabindex="0"
              @click="handleSubmit"
            >
              Complete
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

import { InputNumber, InputText, KeyValueEditor, SelectInput, TextArea } from '@/components/forms'
import ImageGallery from '@/components/ImageGallery.vue'
import ImageUploader from '@/components/ImageUploader.vue'
import * as authAPI from '@/services/api'
import { toMinorUnits, toMajorUnits } from '@/utilities'

const props = defineProps({
  categories: {
    type: Array,
    required: true,
  },
  onSuccess: {
    type: Function,
    default: null,
  },
})

const currentStep = ref(1)
const displayPrice = computed({
  get: () => toMajorUnits(newProduct.value.price),
  set: (value) => (newProduct.value.price = toMinorUnits(value)),
})

const categoryOptions = computed(() =>
  props.categories.map((category) => ({ value: category.slug, label: category.name }))
)

const detailsEditor = ref(null)
const tempProductId = ref(null)
const tempProductImages = ref([])
const successMessage = ref('')

const newProduct = ref({
  name: '',
  price: 0,
  summary: '',
  description: '',
  tax_code: '',
  details: {},
  inventory: '',
  cart_limit: '',
  category: '',
})

const goToStep = async (step) => {
  if (step === 2 && currentStep.value === 1) {
    // Create product draft and save ID for image uploads
    await createProductDraft()
    currentStep.value = step
  } else if (step === 2 && currentStep.value === 3) {
    // Coming back from step 3, update product with any changes
    if (tempProductId.value) {
      await updateProductDraft()
    }
    currentStep.value = step
  } else if (step === 3 && currentStep.value === 2) {
    // Validate step 2 (must have at least one image)
    if (tempProductImages.value.length === 0) {
      return
    }
    currentStep.value = step
  } else if (step < currentStep.value) {
    currentStep.value = step
  }
}

const handleImageUploadSuccess = (images) => tempProductImages.value.push(...images)

const handleImageUploadError = (error) => {
  console.error('Image upload error:', error)
}

const handleImageDeleted = (imageId) => {
  tempProductImages.value = tempProductImages.value.filter((img) => img.id !== imageId)
}

const handleImagePromoted = (imageId) => {
  const idx = tempProductImages.value.findIndex((img) => img.id === imageId)
  if (idx > 0) {
    const [promoted] = tempProductImages.value.splice(idx, 1)
    tempProductImages.value.unshift(promoted)
  }
}

const createProductDraft = async () => {
  try {
    const categoryId = props.categories.find((cat) => cat.slug === newProduct.value.category)?.id

    const productData = {
      name: newProduct.value.name,
      price: newProduct.value.price,
      summary: newProduct.value.summary,
      description: newProduct.value.description || undefined,
      tax_code: newProduct.value.tax_code || undefined,
      inventory: newProduct.value.inventory,
      cart_limit: newProduct.value.cart_limit || undefined,
      ...(categoryId && { category: { id: categoryId } }),
    }

    const createdProduct = await authAPI.createProduct(productData)
    tempProductId.value = createdProduct.id
  } catch (error) {
    console.error('Error creating product draft:', error)
  }
}

const updateProductDraft = async () => {
  try {
    const categoryId = props.categories.find((cat) => cat.slug === newProduct.value.category)?.id

    const updateData = {
      id: tempProductId.value,
      name: newProduct.value.name,
      price: newProduct.value.price,
      summary: newProduct.value.summary,
      description: newProduct.value.description || undefined,
      tax_code: newProduct.value.tax_code || undefined,
      details: newProduct.value.details,
      inventory: newProduct.value.inventory,
      cart_limit: newProduct.value.cart_limit || undefined,
      ...(categoryId && { category: { id: categoryId } }),
    }

    await authAPI.updateProduct(updateData)
  } catch (error) {
    console.error('Error updating product draft:', error)
  }
}

const handleSubmit = async () => {
  // Check for validation errors in details
  if (detailsEditor.value?.hasErrors()) {
    return
  }

  try {
    // Update product with all fields from all steps
    const categoryId = props.categories.find((cat) => cat.slug === newProduct.value.category)?.id

    const updateData = {
      id: tempProductId.value,
      name: newProduct.value.name,
      price: newProduct.value.price,
      summary: newProduct.value.summary,
      description: newProduct.value.description || undefined,
      tax_code: newProduct.value.tax_code || undefined,
      details: newProduct.value.details,
      inventory: newProduct.value.inventory,
      cart_limit: newProduct.value.cart_limit || undefined,
      ...(categoryId && { category: { id: categoryId } }),
    }

    await authAPI.updateProduct(updateData)

    // Show success message
    successMessage.value = `"${newProduct.value.name}" created successfully!`

    // Call success callback if provided
    if (props.onSuccess) {
      props.onSuccess()
    }
  } catch (error) {
    console.error('Error finalizing product:', error)
  }
}

const resetAndStart = () => {
  successMessage.value = ''
  currentStep.value = 1
  newProduct.value = {
    name: '',
    price: 0,
    summary: '',
    description: '',
    tax_code: '',
    details: {},
    inventory: '',
    cart_limit: '',
    category: '',
  }
  tempProductId.value = null
  tempProductImages.value = []
  detailsEditor.value?.reset()
}
</script>

<style scoped>
.wizard-section {
  background: #fff;
  border-radius: 12px;
  padding: 30px;
  margin-bottom: 40px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.wizard-header {
  margin-bottom: 30px;
}

.wizard-header h2 {
  font-size: 20px;
  font-weight: 300;
  margin-bottom: 20px;
  color: #333;
  text-align: center;
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

.form-row-2col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 20px;
}

.image-uploader-wrapper {
  margin: 30px 0;
  padding: 20px;
  background: #f9f9f9;
  border-radius: 8px;
  border: 2px dashed #ddd;
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

.success-message {
  text-align: center;
  padding: 40px;
  background: #f0f8f0;
  border-radius: 8px;
}

.success-message p {
  font-size: 18px;
  color: #333;
  margin-bottom: 20px;
}

.success-message .btn-primary {
  max-width: 250px;
}
</style>
