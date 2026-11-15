<template>
  <div
    v-if="imgURL"
    class="product-tile"
    :tabindex="0"
    @click="goToProductPage"
    @keydown.enter="goToProductPage"
  >
    <div class="image-container">
      <img :src="imgURL" :alt="product.name" class="product-image" />
    </div>
    <div class="product-info">
      <h3 class="product-title">{{ product.name }}</h3>
      <p v-if="product.summary" class="product-summary">{{ product.summary }}</p>
      <p class="product-price">{{ displayPrice(product.price) }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'

import type { Product } from '@/types'
import { displayPrice } from '@/utilities/currency'

const props = defineProps<{ product: Product }>()

const router = useRouter()

const goToProductPage = () => {
  if (!props.product.id) {
    console.error('Product is missing or has no ID:', props.product)
    return
  }
  router.push(`/products/${props.product.id}`)
}

const imgURL = computed(() => props.product.images.find((img) => img.type === 'hero')?.url ?? '')
</script>

<style scoped>
.product-tile {
  display: flex;
  cursor: pointer;
  flex-direction: column;
  align-items: center;
  justify-content: space-between;
  text-align: center;
  padding: 20px;
  border: var(--tile-border);
  box-shadow: var(--tile-shadow);
  border-radius: var(--tile-radius);
  background-color: var(--surface-background);
  transition:
    transform 0.2s ease-in-out,
    box-shadow 0.2s ease-in-out;
}

.product-tile:hover {
  transform: scale(1.02);
  box-shadow: var(--tile-hover-shadow);
}

.image-container {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 250px; /* Ensures consistent height */
}

.product-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain; /* Ensures entire image is visible without cropping */
}

.product-info {
  width: 100%;
  margin-top: 10px;
}

.product-title {
  font-size: 16px;
  font-weight: 500;
  letter-spacing: 2px;
  text-transform: uppercase;
  margin-bottom: 5px;
}

.product-price {
  font-size: 11px;
  font-weight: 700;
  color: #222;
  margin: 5px 0;
  letter-spacing: 0.5px;
  text-transform: uppercase;
}

.product-summary {
  font-size: 14px;
  color: var(--muted-text-color);
  margin-top: 5px;
  padding: 0 10px;
}

/* Mobile adjustments */
@media (max-width: 480px) {
  .product-tile {
    padding: 15px;
  }

  .image-container {
    height: 200px;
  }

  .product-title {
    font-size: 14px;
    letter-spacing: 1px;
  }

  .product-summary {
    font-size: 13px;
    padding: 0 5px;
  }
}
</style>
