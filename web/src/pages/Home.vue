<template>
  <div class="home-container">
    <div class="feature-container">
      <div class="hero-content unselectable">
        <h1 class="hero-title">essential living</h1>
        <div class="hero-buttons">
          <button class="btn-sm" :tabindex="0" @click="$router.push('/new')">shop new</button>
          <button class="btn-sm" :tabindex="0" @click="$router.push('/popular')">
            shop popular
          </button>
        </div>
      </div>
    </div>
    <div class="product-section">
      <div class="product-grid">
        <ProductTile v-for="product in products" :key="product.id" :product="product" />
      </div>
      <IntersectionTrigger @intersect="fetchProducts" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

import IntersectionTrigger from '@/components/IntersectionTrigger.vue'
import ProductTile from '@/components/ProductTile.vue'
import { getProducts } from '@/services/api'
import type { Product, ProductFilters } from '@/types'

const products = ref<Product[]>([])
const page = ref(1)
const hasMore = ref(true)
const isLoading = ref(false)

const fetchProducts = async () => {
  if (isLoading.value || !hasMore.value) return
  isLoading.value = true

  try {
    const filters: ProductFilters = {
      page: page.value,
      limit: 9,
      sortBy: 'newest',
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
</script>

<style scoped>
.home-container {
  width: 100vw;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  padding: 10px;
  box-sizing: border-box;
}

.feature-container {
  width: 100vw;
  height: calc(40vh - 10px);
  overflow: hidden;
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  background-color: #000;
  border-radius: 12px;
}

.product-section {
  width: 100vw;
  min-height: 50vh;
  background-color: #fff;
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
  text-align: center;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  font-family: 'Open Sans', sans-serif;
  margin-top: 20px;
}

.feature-image {
  width: 100vw;
  height: 100vh;
  object-fit: cover;
  position: absolute;
  filter: grayscale(0.3);
  top: 0;
  left: 0;
  z-index: -1;
}

.hero-content {
  position: relative;
  z-index: 1;
  text-align: center;
  color: white;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.5);
}

.hero-title {
  font-size: 3rem;
  font-weight: 100;
  margin: 0;
  text-transform: uppercase;
  font-family: 'Josefin Sans', sans-serif;
  position: relative;
}

.hero-title::after {
  content: '';
  position: absolute;
  bottom: -12px;
  left: 50%;
  transform: translateX(-50%);
  width: 80px;
  height: 2px;
  background: white;
  opacity: 0.7;
}

.hero-buttons {
  display: flex;
  gap: 1rem;
  margin-top: 3rem;
  justify-content: center;
  align-items: center;
}

.hero-buttons button {
  color: #fff;
  background: #0000;
  border: 1px solid #fff;
  flex: none !important;
  display: inline-block !important;
}
button:hover {
  background: #ffffff;
  color: #000;
}
</style>
