<template>
  <div class="product-tile" :tabindex="0" @click="handleClick" @keydown.enter="handleClick">
    <div class="image-container">
      <img v-if="imageUrl" :src="imageUrl" alt="" class="product-image" loading="lazy" />
      <span v-else class="no-image">No image</span>
    </div>
    <div class="product-info">
      <h3 class="product-title">{{ product.name }}</h3>
      <p class="product-id">ID: {{ product.id }}</p>
      <p class="product-price">{{ formatPrice(product.price) }}</p>
      <p v-if="product.summary" class="product-summary">{{ product.summary }}</p>
      <p v-if="product.tax_code" class="product-tax-code">Tax Code: {{ product.tax_code }}</p>
      <div class="product-details">
        <div class="details-list">
          <span
            v-for="(value, key) in product.details"
            :key="key"
            class="detail-item"
            :title="`${key}: ${value}`"
          >
            {{ key }}: {{ value }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'

import type { Product } from '@/types'
import { formatPrice } from '@/utilities/currency'

const props = defineProps<{ product: Product }>()
const router = useRouter()
const imageUrl = computed(
  () =>
    props.product.images?.find((image) => image.type === 'hero')?.url ||
    props.product.images?.[0]?.url
)

const handleClick = () => {
  router.push(`/admin/products/${props.product.id}`)
}
</script>

<style scoped>
.product-tile {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 20px;
  border: var(--tile-border);
  box-shadow: var(--tile-shadow);
  border-radius: var(--tile-radius);
  background-color: var(--surface-background);
  cursor: pointer;
  transition:
    transform 0.2s ease-in-out,
    box-shadow 0.2s ease-in-out;
}

.product-tile:hover {
  opacity: 1;
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
  height: 150px;
  margin-bottom: 10px;
}

.product-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.no-image {
  color: var(--subtle-text-color);
  font-size: 12px;
}

.product-info {
  width: 100%;
}

.product-title {
  font-size: 16px;
  font-weight: 500;
  letter-spacing: 2px;
  text-transform: uppercase;
  margin-bottom: 5px;
}

.product-id {
  font-size: 10px;
  color: var(--subtle-text-color);
  margin-bottom: 5px;
  font-family: 'Roboto Mono', monospace;
}

.product-price {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-color);
  margin: 5px 0;
  letter-spacing: 0.5px;
}

.product-summary {
  font-size: 12px;
  color: var(--muted-text-color);
  margin-top: 5px;
  padding: 0 10px;
  line-height: 1.4;
}

.product-tax-code {
  font-size: 10px;
  color: var(--subtle-text-color);
  margin-top: 5px;
  font-family: 'Roboto Mono', monospace;
  text-transform: uppercase;
}

.product-details {
  margin-top: 8px;
  text-align: left;
}

.details-title {
  font-size: 10px;
  font-weight: 600;
  color: var(--muted-text-color);
  margin-bottom: 4px;
  text-align: center;
}

.details-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.detail-item {
  font-size: 9px;
  color: var(--muted-text-color);
  background-color: var(--subtle-background);
  padding: 2px 4px;
  border-radius: 2px;
  font-family: 'Roboto Mono', monospace;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
