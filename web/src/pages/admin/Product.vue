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
            placeholder="Price"
            required
            @input="handlePriceInput"
          />
          <input v-model="newProduct.summary" type="text" placeholder="Summary" required />
          <input v-model="newProduct.tax_code" type="text" placeholder="Tax Code (optional)" />
          <input v-model="newProduct.inventory" type="number" placeholder="Quantity" required />
          <input v-model="newProduct.cart_limit" type="number" placeholder="Cart Limit" required />
          <select v-model="selectedCategorySlug" required>
            <option value="">Select Category</option>
            <option v-for="category in categories" :key="category.id" :value="category.slug">
              {{ category.slug }}
            </option>
          </select>
        </div>

        <div class="textarea-row">
          <textarea v-model="newProduct.description" placeholder="Description" rows="3"></textarea>
        </div>

        <!-- Details Section -->
        <KeyValueEditor
          ref="detailsEditor"
          v-model="newProduct.details"
          key-placeholder="Key (e.g., color, size, material)"
          value-placeholder="Value"
          pair-name="Detail"
        />

        <button type="submit" class="btn-full-width mt-15">Add Product</button>
      </form>
    </div>
    <div class="product-grid">
      <AdminProductTile v-for="product in products" :key="product.id" :product="product" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

import AdminProductTile from '@/components/AdminProductTile.vue'
import KeyValueEditor from '@/components/forms/KeyValueEditor.vue'
import { getProducts, createProduct, getCategories } from '@/services/api'

const products = ref([])
const categories = ref([])
const selectedCategorySlug = ref('')
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
      summary: newProduct.value.summary,
      description: newProduct.value.description || undefined,
      tax_code: newProduct.value.tax_code || undefined,
      details: newProduct.value.details,
      inventory: newProduct.value.inventory,
      cart_limit: newProduct.value.cart_limit,
    }

    await createProduct(productData, selectedCategorySlug.value)

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
    selectedCategorySlug.value = ''
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

.textarea-row {
  display: flex;
  justify-content: center;
  margin-bottom: 15px;
}

.form-row textarea,
.form-row input,
.form-row select,
.textarea-row textarea {
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 16px;
  background-color: transparent;
  min-width: 200px;
  font-family: inherit;
}

.textarea-row textarea {
  width: 100%;
  max-width: 600px;
  resize: vertical;
  line-height: 1.4;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
  font-family: 'Open Sans', sans-serif;
  margin-top: 20px;
}
</style>
