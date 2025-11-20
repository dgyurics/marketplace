<template>
  <div class="product-tile" :tabindex="0" @click="handleClick" @keydown.enter="handleClick">
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
import { useRouter } from 'vue-router'

import type { Product } from '@/types'
import { formatPrice } from '@/utilities/currency'

const props = defineProps<{ product: Product }>()
const router = useRouter()

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
  border: 1px solid rgba(0, 0, 0, 0.1);
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.05);
  border-radius: 8px;
  background-color: #fff;
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
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.1);
}

.product-info {
  width: 100%;
}

.product-title {
  font-size: 16px;
  font-weight: 500;
  letter-spacing: 3px;
  text-transform: uppercase;
  margin-bottom: 5px;
}

.product-id {
  font-size: 10px;
  color: #888;
  margin-bottom: 5px;
  font-family: 'Roboto Mono', monospace;
}

.product-price {
  font-size: 14px;
  font-weight: 700;
  color: #222;
  margin: 5px 0;
  letter-spacing: 0.5px;
}

.product-summary {
  font-size: 12px;
  color: #666;
  margin-top: 5px;
  padding: 0 10px;
  line-height: 1.4;
}

.product-tax-code {
  font-size: 10px;
  color: #888;
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
  color: #555;
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
  color: #666;
  background-color: #f5f5f5;
  padding: 2px 4px;
  border-radius: 2px;
  font-family: 'Roboto Mono', monospace;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
