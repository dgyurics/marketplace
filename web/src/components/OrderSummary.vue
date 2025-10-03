<template>
  <div v-if="order" class="order-summary">
    <h3>Summary</h3>
    <div v-for="item in order.items" :key="item.product.id" class="order-item">
      <img :src="item.thumbnail" />
      <div class="details">
        <h4>{{ item.product.name }}</h4>
        <p>Quantity: {{ item.quantity }}</p>
        <p>Unit Price: {{ formatPrice(item.unit_price) }}</p>
      </div>
    </div>
    <div class="totals">
      <p>
        <span>Subtotal</span><span>{{ formatPrice(order.amount) }}</span>
      </p>
      <p>
        <span>Tax <span class="italic">(estimate)</span></span
        ><span>{{ formatPrice(order.tax_amount) }}</span>
      </p>
      <p>
        <span>Total</span><span>{{ formatPrice(order.total_amount) }}</span>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Order } from '@/types'
import { formatPrice } from '@/utilities'

defineProps<{ order: Order | null }>()
</script>

<style scoped>
.order-summary {
  width: 320px;
  padding: 24px;
  left: 40px;
  position: absolute;
  margin-top: -70px;
  z-index: 9999;
  border: 1px solid #ddd;
  background: #fafafa;
  border-radius: 6px;
  font-size: 14px;
  font-family: 'Open Sans', sans-serif;
}

.order-summary h3 {
  margin-bottom: 20px;
  text-align: center;
}
.order-item {
  display: flex;
  margin-bottom: 20px;
  align-items: flex-start;
}
.order-item img {
  width: 80px;
  height: auto;
  margin-right: 15px;
  border-radius: 4px;
}
.details h4 {
  margin: 0 0 6px;
  font-weight: 500;
  color: #333;
}
.details p {
  font-size: 13px;
  color: #888;
  margin: 2px 0;
}
.italic {
  font-style: italic;
}
.totals {
  border-top: 1px solid #ddd;
  margin-top: 20px;
  padding-top: 15px;
  color: #555;
}
.totals p {
  display: flex;
  justify-content: space-between;
  margin: 5px 0;
}
</style>
