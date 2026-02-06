<template>
  <div class="product-container">
    <!-- Create Product Wizard -->
    <CreateProductWizard :categories="categories" :on-success="fetchProducts" />

    <!-- Products Grid -->
    <div class="products-section">
      <div class="product-grid">
        <AdminProductTile v-for="product in products" :key="product.id" :product="product" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

import AdminProductTile from '@/components/AdminProductTile.vue'
import CreateProductWizard from '@/components/CreateProductWizard.vue'
import { getProducts, getCategories } from '@/services/api'

const products = ref([])
const categories = ref([])

const fetchProducts = async () => {
  try {
    const response = await getProducts({ page: 1, limit: 100 })
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

onMounted(() => {
  fetchProducts()
  fetchCategories()
})
</script>

<style scoped>
.product-container {
  max-width: 1000px;
  margin: auto;
  padding: 20px;
}

/* Products Section */
.products-section {
  margin-top: 50px;
}

.products-section h2 {
  margin: 0 0 20px 0;
  font-size: 1.5rem;
  font-weight: 300;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
  font-family: 'Open Sans', sans-serif;
}
</style>
