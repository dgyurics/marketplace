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
  min-width: 0;
  cursor: pointer;
  text-align: left;
}
.image-container {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  aspect-ratio: 1;
  padding: 16px;
  background: #f5f5f3;
  border-radius: 10px;
  overflow: hidden;
}
.product-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
  transition: transform 0.2s ease;
}
.product-tile:hover .product-image {
  transform: scale(1.04);
}
.product-info {
  padding: 12px 2px 0;
}
.product-title {
  font-size: 14px;
  font-weight: 400;
  letter-spacing: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.product-price {
  font-size: 16px;
  font-weight: 700;
  margin-top: 6px;
}
.product-summary {
  font-size: 12px;
  color: var(--muted-text-color);
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
