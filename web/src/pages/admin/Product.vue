<template>
  <div class="product-container">
    <div class="new-product-form">
      <form @submit.prevent="handleSubmit">
        <div class="form-row">
          <InputText v-model="newProduct.name" label="name" required />
          <InputNumber v-model="displayPrice" label="price" step="0.01" required />
          <InputNumber v-model="newProduct.inventory" label="inventory" required />
          <InputText v-model="newProduct.summary" label="summary" required />
          <SelectInput v-model="newProduct.category" label="category" :options="categoryOptions" />
          <InputText v-model="newProduct.tax_code" label="tax code" />
          <InputNumber v-model="newProduct.cart_limit" label="cart limit" />
        </div>

        <div class="textarea-row">
          <TextArea v-model="newProduct.description" label="description"></TextArea>
        </div>

        <!-- Details Section -->
        <KeyValueEditor
          ref="detailsEditor"
          v-model="newProduct.details"
          key-placeholder="Key (e.g., color, size, material)"
          value-placeholder="Value"
          pair-name="Detail"
        />

        <button type="submit" class="btn-full-width mt-15" :tabindex="0">Add Product</button>
      </form>
    </div>
    <div class="product-grid">
      <AdminProductTile v-for="product in products" :key="product.id" :product="product" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'

import AdminProductTile from '@/components/AdminProductTile.vue'
import { InputNumber, InputText, KeyValueEditor, SelectInput, TextArea } from '@/components/forms'
import { getProducts, createProduct, getCategories } from '@/services/api'
import { toMinorUnits, toMajorUnits } from '@/utilities'

const displayPrice = computed({
  get: () => (newProduct.value.price === '' ? '' : toMajorUnits(newProduct.value.price)),
  set: (value) => (newProduct.value.price = toMinorUnits(value)),
})

const categoryOptions = computed(() =>
  categories.value.map((category) => ({ value: category.slug, label: category.name }))
)

const products = ref([])
const categories = ref([])
const detailsEditor = ref(null)
const newProduct = ref({
  name: '',
  price: '',
  summary: '',
  description: '',
  tax_code: '',
  details: {},
  inventory: '',
  cart_limit: '',
  category: '',
})

const fetchProducts = async () => {
  try {
    const response = await getProducts({ page: 1, limit: 100 }) // Get max 100 products for now
    products.value = response
  } catch {
    // Handle error silently
  }
}

const fetchCategories = async () => {
  try {
    const response = await getCategories()
    categories.value = response
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
    const categoryId = categories.value.find((cat) => cat.slug === newProduct.value.category)?.id

    const productData = {
      name: newProduct.value.name,
      price: newProduct.value.price,
      summary: newProduct.value.summary,
      description: newProduct.value.description || undefined,
      tax_code: newProduct.value.tax_code || undefined,
      details: newProduct.value.details,
      inventory: newProduct.value.inventory,
      cart_limit: newProduct.value.cart_limit || undefined,
      // Include category if one is selected
      ...(categoryId && { category: { id: categoryId } }),
    }

    await createProduct(productData)

    // Reset form
    newProduct.value = {
      name: '',
      price: '',
      summary: '',
      description: '',
      tax_code: '',
      details: {},
      inventory: '',
      cart_limit: '',
    }
    detailsEditor.value?.reset()

    // Refresh products
    await fetchProducts()
  } catch {
    // Handle error silently
  }
}

onMounted(() => {
  fetchProducts()
  fetchCategories()
})
</script>

<style scoped>
.product-container {
  max-width: 1200px;
  margin: auto;
  padding: 20px;
  text-align: center;
}

.new-product-form {
  margin-bottom: 30px;
}

.form-row {
  display: flex;
  gap: 10px;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  margin-bottom: 15px;
}

.form-row :deep(.input-container) {
  flex: 1 1 calc(25% - 10px);
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
  font-family: 'Open Sans', sans-serif;
  margin-top: 20px;
}
</style>
