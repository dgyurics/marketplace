<template>
  <div class="product-container">
    <div class="product-grid">
      <ProductTile v-for="product in products" :key="product.id" :product="product" />
    </div>
    <IntersectionTrigger @intersect="fetchProducts" />
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'

import IntersectionTrigger from '@/components/IntersectionTrigger.vue'
import ProductTile from '@/components/ProductTile.vue'
import { getProducts } from '@/services/api'

const route = useRoute()
const products = ref([])
const page = ref(1)
const hasMore = ref(true)
const isLoading = ref(false)
const category = ref(route.params.category) // e.g. "furniture" or "wall-decor"

const fetchProducts = async () => {
  if (isLoading.value || !hasMore.value) return // Prevent multiple calls
  isLoading.value = true

  try {
    const response = await getProducts([category.value], page.value, 9)
    if (response.length === 0) {
      hasMore.value = false
    } else {
      products.value.push(...response)
      page.value += 1
    }
  } catch (error) {
    console.error('Error fetching products:', error)
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  fetchProducts()
})

// Watch for route param changes
watch(
  () => route.params.category,
  (newCategory) => {
    category.value = newCategory
    resetAndFetch()
  }
)

const resetAndFetch = () => {
  products.value = []
  page.value = 1
  hasMore.value = true
  fetchProducts()
}
</script>

<style scoped>
.product-container {
  max-width: 1200px;
  margin: auto;
  padding: 20px;
  text-align: center;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  font-family: 'Inter', sans-serif;
  margin-top: 20px;
}
</style>
