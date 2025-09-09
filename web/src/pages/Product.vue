<template>
  <div class="product-container">
    <div class="product-grid">
      <ProductTile v-for="product in products" :key="product.id" :product="product" />
    </div>
    <IntersectionTrigger @intersect="fetchProducts" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'

import IntersectionTrigger from '@/components/IntersectionTrigger.vue'
import ProductTile from '@/components/ProductTile.vue'
import { getProducts } from '@/services/api'
import type { Product, SortBy, ProductFilters } from '@/types'

interface Props {
  sortBy?: SortBy | null
  category?: string | null
}

const props = defineProps<Props>()
const products = ref<Product[]>([])
const page = ref(1)
const hasMore = ref(true)
const isLoading = ref(false)

const fetchProducts = async () => {
  if (isLoading.value || !hasMore.value) return // Prevent multiple calls
  isLoading.value = true

  try {
    const filters: ProductFilters = {
      page: page.value,
      limit: 9,
    }

    // Set category if we have one (category routes)
    if (props.category) {
      filters.categories = [props.category]
    }

    // Set sorting if we have it (new/popular routes)
    if (props.sortBy) {
      filters.sortBy = props.sortBy
    }

    const response = await getProducts(filters)
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

// Watch for category or sortBy changes
// When these change, it means user navigated to different page
watch(
  () => [props.category, props.sortBy],
  () => {
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
