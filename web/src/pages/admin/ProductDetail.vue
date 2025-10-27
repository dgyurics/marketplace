<template>
  <div class="product-detail-container">
    <div v-if="loading" class="loading">Loading product...</div>
    <div v-else-if="product" class="edit-form">
      <form @submit.prevent="handleSubmit">
        <div class="form-row">
          <InputText v-model="editProduct.name" label="name" required />
          <InputNumber v-model="displayPrice" label="Price" step="0.01" required />
          <InputNumber v-model="editProduct.inventory" label="inventory" required />
          <InputNumber v-model="editProduct.cart_limit" label="cart limit" />
          <InputText v-model="editProduct.tax_code" label="tax code" />
          <SelectInput v-model="editProduct.category" label="category" :options="categoryOptions" />
          <InputText v-model="editProduct.summary" label="summary" required />
        </div>

        <div class="textarea-row">
          <TextArea v-model="editProduct.description" label="description"></TextArea>
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
          <button type="submit" class="btn-full-width mt-15">Save</button>
          <button type="button" class="btn-full-width btn-outline" @click="goBack">Cancel</button>
          <button type="button" class="btn-full-width btn-subtle" @click="handleDelete">
            Remove
          </button>
        </div>
      </form>
    </div>

    <div v-else class="error">Product not found or failed to load.</div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { InputNumber, InputText, KeyValueEditor, SelectInput, TextArea } from '@/components/forms'
import ImageGallery from '@/components/ImageGallery.vue'
import ImageUploader from '@/components/ImageUploader.vue'
import { getProductById, getCategories, updateProduct, removeProduct } from '@/services/api'
import { toMinorUnits, toMajorUnits } from '@/utilities'

const route = useRoute()
const router = useRouter()

const product = ref(null)
const categories = ref([])
const loading = ref(true)
const detailsEditor = ref(null)

const categoryOptions = computed(() =>
  categories.value.map((category) => ({ value: category.slug, label: category.name }))
)

const displayPrice = computed({
  get: () => toMajorUnits(editProduct.value.price),
  set: (value) => (editProduct.value.price = toMinorUnits(value)),
})

const editProduct = ref({
  name: '',
  price: '',
  summary: '',
  description: '',
  tax_code: '',
  details: {},
  category: '',
  inventory: '',
  cart_limit: '',
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
      price: data.price,
      summary: data.summary,
      description: data.description,
      tax_code: data.tax_code ?? '',
      details: data.details,
      category: data.category?.slug ?? '',
      inventory: data.inventory,
      cart_limit: data.cart_limit,
    }
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

const handleSubmit = async () => {
  // Check for validation errors in details
  if (detailsEditor.value?.hasErrors()) {
    return
  }

  try {
    const categoryId = categories.value.find((cat) => cat.slug === editProduct.value.category)?.id

    const _updateData = {
      id: product.value.id,
      name: editProduct.value.name,
      price: editProduct.value.price,
      summary: editProduct.value.summary,
      description: editProduct.value.description || undefined,
      tax_code: editProduct.value.tax_code || undefined,
      details: editProduct.value.details,
      inventory: editProduct.value.inventory,
      cart_limit: editProduct.value.cart_limit || undefined,
      // Include category if one is selected
      ...(categoryId && { category: { id: categoryId } }),
    }

    await updateProduct(_updateData)

    // For now, just go back
    goBack()
  } catch {
    // Handle error silently
  }
}

const goBack = () => {
  router.back()
}

const handleDelete = async () => {
  try {
    await removeProduct(product.value.id)
    router.push('/admin/products')
  } catch {
    // Handle error silently
  }
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

.form-row :deep(.input-container) {
  flex: 1 1 calc(33.333% - 10px);
}

.form-actions {
  display: flex;
  flex-direction: column;
  gap: 15px;
  margin-top: 30px;
}
</style>
