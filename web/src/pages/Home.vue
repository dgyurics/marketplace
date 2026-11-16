<template>
  <div class="home">
    <section class="hero-wrapper">
      <div class="hero">
        <div class="hero-inner">
          <p class="hero-eyebrow">A little discovery. A great find.</p>
          <h1 class="hero-title">Your next favorite<br />is waiting here.</h1>
          <div class="hero-actions">
            <router-link class="hero-btn" to="/new"
              >Discover what's new <ArrowRightIcon
            /></router-link>
            <router-link class="hero-secondary" to="/popular">Explore popular finds</router-link>
          </div>
        </div>
        <div class="hero-art" aria-hidden="true">
          <div class="hero-orbit"></div>
          <div class="find-card find-card--back">
            <SparklesIcon />
            <span>Unexpected finds.<br />Everyday favorites.</span>
          </div>
          <div class="find-card find-card--front">
            <ShoppingBagIcon />
            <span>Good things,<br /><strong>closer to home.</strong></span>
          </div>
        </div>
      </div>
    </section>

    <section class="products" aria-labelledby="featured-heading">
      <div class="section-heading">
        <div>
          <p class="section-eyebrow">Some inspiration for your day…</p>
          <h2 id="featured-heading">Find your next favorite</h2>
        </div>
        <router-link to="/new">Shop all <ArrowRightIcon /></router-link>
      </div>
      <div class="product-grid">
        <ProductTile v-for="product in products" :key="product.id" :product="product" />
      </div>
      <IntersectionTrigger @intersect="fetchProducts" />
    </section>
  </div>
</template>

<script setup lang="ts">
import { ArrowRightIcon, ShoppingBagIcon, SparklesIcon } from '@heroicons/vue/24/outline'
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
      limit: 12,
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
.home {
  width: 100%;
}
.hero-wrapper,
.products {
  max-width: 1280px;
  margin: 0 auto;
  padding: 0 32px;
}
.hero-wrapper {
  padding-top: 28px;
}
.hero {
  display: flex;
  align-items: center;
  min-height: 370px;
  padding: 48px 56px;
  background: var(--highlight-background);
  border-radius: 16px;
  overflow: hidden;
  position: relative;
}
.hero-inner {
  position: relative;
  z-index: 1;
  flex: 1;
}
.hero-eyebrow,
.section-eyebrow {
  font-size: 13px;
  font-weight: 600;
  color: var(--accent-text-color);
  margin-bottom: 14px;
}
.hero-title {
  font-size: clamp(32px, 4vw, 50px);
  font-weight: 700;
  letter-spacing: -2px;
  line-height: 1.14;
  color: var(--text-color);
}
.hero-secondary {
  font-size: 13px;
  text-decoration: underline;
  text-underline-offset: 4px;
}
.hero-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 20px;
  margin-top: 26px;
}
.hero-btn {
  display: inline-flex;
  align-items: center;
  gap: 14px;
  background: var(--accent-color);
  color: var(--on-accent-color);
  border-radius: var(--control-radius);
  padding: 13px 20px;
  font-size: 14px;
  font-weight: 600;
}
.hero-btn:hover {
  background: var(--accent-hover-color);
}
.hero-art {
  position: relative;
  width: 340px;
  height: 280px;
  margin-left: 24px;
}
.hero-orbit {
  position: absolute;
  width: 290px;
  height: 290px;
  border-radius: 50%;
  background: var(--hero-orbit-background);
  top: 0;
  left: 20px;
}
.find-card {
  position: absolute;
  width: 180px;
  height: 210px;
  border-radius: var(--panel-radius);
  padding: 24px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.08);
  font-size: 14px;
}
.find-card svg {
  width: 64px;
  height: 64px;
  stroke-width: 1.2;
}
.find-card--back {
  background: var(--decoration-background);
  color: var(--decoration-text-color);
  left: 0;
  top: 16px;
  transform: rotate(-12deg);
}
.find-card--front {
  background: var(--surface-background);
  color: var(--accent-text-color);
  right: 0;
  top: 60px;
  transform: rotate(10deg);
}
.section-heading > a {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  font-weight: 600;
}
.hero-btn svg,
.section-heading svg {
  width: 17px;
  height: 17px;
}
.products {
  padding-top: 24px;
  padding-bottom: 56px;
}
.section-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 24px;
}
.section-eyebrow {
  color: var(--muted-text-color);
  font-weight: 400;
  margin-bottom: 6px;
}
.section-heading h2 {
  text-align: left;
  font-size: 26px;
  font-weight: 700;
  letter-spacing: -0.7px;
  margin: 0;
}
.section-heading > a {
  font-size: 13px;
  white-space: nowrap;
}
.product-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 28px 20px;
}
@media (max-width: 900px) {
  .hero {
    padding: 36px;
  }
  .hero-art {
    width: 260px;
    transform: scale(0.85);
    margin-left: 0;
  }
  .product-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (max-width: 600px) {
  .hero-wrapper,
  .products {
    padding-left: 16px;
    padding-right: 16px;
  }
  .hero-wrapper {
    padding-top: 16px;
  }
  .hero {
    padding: 30px 24px;
    min-height: auto;
  }
  .hero-title {
    letter-spacing: -1px;
  }
  .hero-art {
    display: none;
  }
  .hero-actions {
    gap: 16px;
  }
  .products {
    padding-top: 14px;
  }
  .section-heading h2 {
    font-size: 21px;
  }
  .product-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 24px 12px;
  }
}
</style>
