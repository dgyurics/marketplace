<template>
  <div class="product-detail-container">
    <div v-if="loading" class="loading">Loading product...</div>
    <div v-else-if="product" class="edit-form">
      <form @submit.prevent="handleSubmit">
        <div class="form-row">
          <input v-model="editProduct.name" type="text" placeholder="Product Name" required />
          <input
            v-model="editProduct.price"
            type="number"
            step="1.00"
            placeholder="Price"
            required
            @input="handlePriceInput"
          />
          <input v-model="editProduct.inventory" type="number" placeholder="Quantity" required />
          <input v-model="editProduct.summary" type="text" placeholder="Summary" required />
          <input v-model="editProduct.tax_code" type="text" placeholder="Tax Code (optional)" />
          <select v-model="selectedCategorySlug" required>
            <option value="">Select Category</option>
            <option v-for="category in categories" :key="category.id" :value="category.slug">
              {{ category.slug }}
            </option>
          </select>
        </div>

        <div class="textarea-row">
          <textarea v-model="editProduct.description" placeholder="Description" rows="4"></textarea>
        </div>

        <!-- Details Section -->
        <KeyValueEditor
          ref="detailsEditor"
          v-model="editProduct.details"
          title="Product Details"
          key-placeholder="Key (e.g., color, size, material)"
          value-placeholder="Value"
          pair-name="Detail"
        />

        <!-- Image Gallery Section -->
        <ImageGallery
          :images="product.images || []"
          @image-deleted="handleImageDeleted"
          @image-promoted="handleImagePromoted"
        />

        <!-- Image Upload Section -->
        <ImageUploader
          :product-id="product.id"
          :images="product.images || []"
          @upload-success="handleImageUploadSuccess"
          @upload-error="handleImageUploadError"
        />

        <div class="form-actions">
          <button type="submit" class="btn-full-width mt-15" :disabled="saving">
            Save Changes
          </button>
          <button type="button" class="btn-full-width btn-outline" @click="goBack">Cancel</button>
        </div>
      </form>
    </div>

    <div v-else class="error">Product not found or failed to load.</div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import KeyValueEditor from '@/components/forms/KeyValueEditor.vue'
import ImageGallery from '@/components/ImageGallery.vue'
import ImageUploader from '@/components/ImageUploader.vue'
import { getProductById, getCategories, updateProduct } from '@/services/api'

const route = useRoute()
const router = useRouter()

const product = ref(null)
const categories = ref([])
const loading = ref(true)
const saving = ref(false)
const selectedCategorySlug = ref('')
const detailsEditor = ref(null)

const editProduct = ref({
  name: '',
  price: '',
  summary: '',
  description: '',
  tax_code: '',
  details: {},
  category: '',
  inventory: '',
})

const fetchProduct = async () => {
  try {
    loading.value = true
    const productId = route.params.id
    const data = await getProductById(productId)
    product.value = data

    // Populate form with existing data
    editProduct.value = {
      name: data.name,
      price: (data.price / 100).toFixed(2),
      summary: data.summary,
      description: data.description,
      tax_code: data.tax_code ?? '',
      details: data.details,
      category: data.category,
      inventory: data.inventory,
    }

    // Set the category slug for the dropdown
    selectedCategorySlug.value = data.category?.slug ?? ''
  } catch {
    // Handle error silently
    product.value = null
  } finally {
    loading.value = false
  }
}

const fetchCategories = async () => {
  try {
    const data = await getCategories()
    categories.value = data
  } catch {
    // Handle error silently
  }
}

const handlePriceInput = (event) => {
  editProduct.value.price = event.target.value
}

const handleSubmit = async () => {
  // Check for validation errors in details
  if (detailsEditor.value?.hasErrors()) {
    return
  }

  try {
    saving.value = true

    // Find the selected category by slug
    const selectedCategory = categories.value.find((cat) => cat.slug === selectedCategorySlug.value)

    const _updateData = {
      id: product.value.id,
      name: editProduct.value.name,
      price: Math.round(parseFloat(editProduct.value.price) * 100), // Convert to cents
      summary: editProduct.value.summary,
      description: editProduct.value.description || undefined,
      tax_code: editProduct.value.tax_code || undefined,
      details: editProduct.value.details,
      inventory: editProduct.value.inventory,
      // Include category if one is selected
      ...(selectedCategory && { category: { id: selectedCategory.id } }),
    }

    await updateProduct(_updateData)

    // For now, just go back
    goBack()
  } catch {
    // Handle error silently
  } finally {
    saving.value = false
  }
}

const goBack = () => {
  router.back()
}

const handleImageUploadSuccess = async () => {
  // Refresh product data to show new images
  await fetchProduct()
}

const handleImageDeleted = async () => {
  // Refresh product data to remove deleted image
  await fetchProduct()
}

const handleImagePromoted = async () => {
  // Refresh product data to reflect promoted image
  await fetchProduct()
}

const handleImageUploadError = (_error) => {
  // Handle upload errors silently, consistent with other error handling in this component
}

onMounted(() => {
  fetchProduct()
  fetchCategories()
})
</script>

<style scoped>
.product-detail-container {
  max-width: 800px;
  margin: auto;
  padding: 20px;
}

.header {
  margin-bottom: 30px;
}

.header h1 {
  margin: 0;
  font-size: 24px;
  color: #333;
}

.loading,
.error {
  text-align: center;
  padding: 40px;
  font-size: 16px;
  color: #666;
}

.error {
  color: #e74c3c;
}

.edit-form {
  padding: 30px;
  border-radius: 8px;
}

.form-row {
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.textarea-row {
  margin-bottom: 20px;
}

.form-row input,
.form-row select {
  flex: 1;
  min-width: 200px;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 16px;
  background-color: white;
}

.textarea-row textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 16px;
  background-color: white;
  resize: vertical;
  font-family: inherit;
  line-height: 1.4;
}

.form-row input:focus,
.form-row select:focus,
.textarea-row textarea:focus {
  outline: none;
  border-color: #007bff;
}

.form-actions {
  display: flex;
  flex-direction: column;
  gap: 15px;
  margin-top: 30px;
}
</style>
