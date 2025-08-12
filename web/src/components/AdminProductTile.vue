<template>
  <div class="product-tile" @click="handleClick">
    <button class="delete-button" title="Delete product" @click="handleDelete">×</button>
    <div class="product-info">
      <h3 class="product-title">{{ product.name }}</h3>
      <p class="product-id">ID: {{ product.id }}</p>
      <p class="product-price">${{ (product.price / 100).toFixed(2) }}</p>
      <p v-if="product.description" class="product-description">{{ product.description }}</p>
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

const props = defineProps<{ product: Product }>()
const emit = defineEmits<{
  delete: [productId: string]
}>()

const router = useRouter()

const handleDelete = (event: Event) => {
  event.stopPropagation()
  emit('delete', props.product.id)
}

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

.delete-button {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 24px;
  height: 24px;
  border: none;
  background: none;
  color: #333;
  font-size: 16px;
  font-weight: bold;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

.delete-button:hover {
  color: #000;
}

.product-tile:hover .delete-button {
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
  font-weight: 400;
  color: #888;
  margin-bottom: 5px;
  font-family: monospace;
}

.product-price {
  font-size: 14px;
  font-weight: 700;
  color: #222;
  margin: 5px 0;
  letter-spacing: 0.5px;
}

.product-description {
  font-size: 12px;
  color: #666;
  margin-top: 5px;
  padding: 0 10px;
  line-height: 1.4;
}

.product-tax-code {
  font-size: 10px;
  font-weight: 600;
  color: #555;
  margin-top: 5px;
  font-family: monospace;
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
  font-family: monospace;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
