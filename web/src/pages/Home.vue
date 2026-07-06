<template>
  <div class="home">
    <!-- Hero -->
    <section class="hero-wrapper">
      <div class="hero unselectable">
        <div class="hero-inner">
          <p class="hero-eyebrow">curated goods for modern life</p>
          <h1 class="hero-title">essential living</h1>
          <div class="hero-actions">
            <button class="hero-btn" @click="$router.push('/new')">shop new</button>
            <button class="hero-btn hero-btn--outline" @click="$router.push('/popular')">
              shop popular
            </button>
          </div>
        </div>
      </div>
    </section>

    <!-- Products -->
    <section class="products">
      <div class="product-grid">
        <ProductTile v-for="product in products" :key="product.id" :product="product" />
      </div>
      <IntersectionTrigger @intersect="fetchProducts" />
    </section>
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
      featured: true,
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
/* ── Layout ── */
.home {
  width: 100%;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* ── Hero ── */
.hero-wrapper {
  width: 100%;
  max-width: 1200px;
  margin: 10px auto 0 auto;
  padding: 0 20px;
  box-sizing: border-box;
}

.hero {
  width: 100%;
  height: calc(50vh - 10px);
  max-height: 480px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(160deg, #0a0a0a 0%, #1a1a2e 50%, #16213e 100%);
  border-radius: 8px;
  position: relative;
  overflow: hidden;
}

.hero::before {
  content: '';
  position: absolute;
  inset: 0;
  background: radial-gradient(ellipse at 30% 50%, rgba(255, 255, 255, 0.03) 0%, transparent 70%);
  pointer-events: none;
}

.hero-inner {
  position: relative;
  text-align: center;
  color: #fff;
  padding: 0 24px;
}

.hero-eyebrow {
  font-family: 'Open Sans', sans-serif;
  font-size: 0.75rem;
  font-weight: 400;
  letter-spacing: 4px;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.45);
  margin-bottom: 1.5rem;
}

.hero-title {
  font-family: 'Josefin Sans', sans-serif;
  font-size: clamp(2.5rem, 6vw, 4rem);
  font-weight: 100;
  letter-spacing: 10px;
  text-transform: uppercase;
  margin: 0;
  line-height: 1.1;
}

.hero-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  margin-top: 3rem;
}

.hero-btn {
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 2px;
  text-transform: uppercase;
  padding: 14px 32px;
  border: none;
  cursor: pointer;
  transition: all 0.3s ease;
  background: #fff;
  color: #0a0a0a;
  min-width: auto;
}

.hero-btn:hover {
  background: rgba(255, 255, 255, 0.85);
  transform: translateY(-1px);
}

.hero-btn--outline {
  background: transparent;
  color: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.25);
}

.hero-btn--outline:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.5);
  color: #fff;
}

/* ── Products ── */
.products {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px 20px 40px;
  box-sizing: border-box;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
  font-family: 'Open Sans', sans-serif;
  margin-top: 20px;
}

/* ── Responsive ── */
@media (max-width: 768px) {
  .hero-title {
    letter-spacing: 6px;
  }

  .hero-actions {
    flex-direction: column;
    align-items: center;
    gap: 0.75rem;
  }

  .hero-btn {
    width: 200px;
  }

  .product-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
  }
}

@media (max-width: 480px) {
  .hero {
    height: calc(45vh - 10px);
  }

  .product-grid {
    grid-template-columns: 1fr;
    gap: 20px;
  }

  .products {
    padding: 15px 15px 30px;
  }

  .hero-wrapper {
    padding: 0 15px;
  }
}

@media (orientation: landscape) and (max-height: 600px) {
  .hero {
    height: calc(60vh - 10px);
    max-height: 500px;
  }
}
</style>
