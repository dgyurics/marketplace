<template>
  <div class="product-container">
    <div class="new-product-form">
      <form @submit.prevent="handleSubmit">
        <div class="form-row">
          <input v-model="newProduct.name" type="text" placeholder="Product Name" required />
          <input
            v-model="newProduct.price"
            type="number"
            step="0.01"
            placeholder="Price (in dollars)"
            required
            @input="handlePriceInput"
          />
          <input v-model="newProduct.description" type="text" placeholder="Description" required />
          <input v-model="newProduct.tax_code" type="text" placeholder="Tax Code (optional)" />
          <select v-model="selectedCategorySlug" required>
            <option value="">Select Category</option>
            <option v-for="category in categories" :key="category.id" :value="category.slug">
              {{ category.name }}
            </option>
          </select>
        </div>

        <!-- Details Section -->
        <KeyValueEditor
          ref="detailsEditor"
          v-model="newProduct.details"
          key-placeholder="Key (e.g., color, size, material)"
          value-placeholder="Value"
          pair-name="Detail"
        />

        <button type="submit" class="submit-button">Add Product</button>
      </form>
    </div>
    <div class="product-grid">
      <AdminProductTile
        v-for="product in products"
        :key="product.id"
        :product="product"
        @delete="handleDelete"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

import AdminProductTile from '@/components/AdminProductTile.vue'
import KeyValueEditor from '@/components/KeyValueEditor.vue'
import { getProducts, createProduct, removeProduct, getCategories } from '@/services/api'

const products = ref([])
const categories = ref([])
const selectedCategorySlug = ref('')
const detailsEditor = ref(null)
const newProduct = ref({
  name: '',
  price: '',
  description: '',
  tax_code: '',
  details: {},
})

const fetchProducts = async () => {
  try {
    const response = await getProducts([], 1, 100) // Get max 100 products for now
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

const handlePriceInput = (event) => {
  // Keep the displayed value as-is for user input
  newProduct.value.price = event.target.value
}

const handleSubmit = async () => {
  // Check for validation errors in details
  if (detailsEditor.value?.hasErrors()) {
    return
  }

  try {
    const productData = {
      name: newProduct.value.name,
      price: Math.round(parseFloat(newProduct.value.price) * 100), // Convert to cents
      description: newProduct.value.description,
      tax_code: newProduct.value.tax_code || undefined,
      details: newProduct.value.details,
    }

    await createProduct(productData, selectedCategorySlug.value)

    // Reset form
    newProduct.value = { name: '', price: '', description: '', tax_code: '', details: {} }
    selectedCategorySlug.value = ''
    detailsEditor.value?.reset()

    // Refresh products
    await fetchProducts()
  } catch {
    // Handle error silently
  }
}

const handleDelete = async (productId) => {
  try {
    await removeProduct(productId)
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
}

.form-row input,
.form-row select {
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 16px;
  background-color: transparent;
  min-width: 200px;
}

.submit-button {
  padding: 10px 20px;
  background-color: #000;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  cursor: pointer;
  transition: background-color 0.2s ease-in-out;
}

.submit-button:hover {
  background-color: #333;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  font-family: 'Inter', sans-serif;
  margin-top: 20px;
}
</style>
