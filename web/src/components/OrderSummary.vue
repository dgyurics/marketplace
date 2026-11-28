<template>
  <div class="order-summary">
    <h3>Order summary</h3>
    <div v-for="item in cartItems" :key="item.product.id" class="order-item">
      <img
        :src="
          item.product.images.find((img) => img.type === 'thumbnail')?.url ||
          item.product.images[0]?.url
        "
        :alt="item.product.name"
      />
      <div class="details">
        <h4>{{ item.product.name }}</h4>
        <p>Quantity: {{ item.quantity }}</p>
        <p>Unit Price: {{ formatPrice(item.product.price) }}</p>
      </div>
    </div>
    <div class="totals">
      <p>
        <span>Subtotal</span><span>{{ formatPrice(subtotal) }}</span>
      </p>
      <p>
        <span>Tax <span class="italic">(estimate)</span></span
        ><span>{{ formatPrice(taxAmount) }}</span>
      </p>
      <p>
        <span>Total</span><span>{{ formatPrice(total) }}</span>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed } from 'vue'

import { useCartStore } from '@/store/cart'
import { formatPrice } from '@/utilities'

const props = defineProps<{
  taxAmount?: number
}>()

const cartStore = useCartStore()
const { items: cartItems, subtotal } = storeToRefs(cartStore)

const taxAmount = computed(() => props.taxAmount || 0)

const total = computed(() => subtotal.value + taxAmount.value)
</script>

<style scoped>
.order-summary {
  min-width: 0;
  padding: 28px;
  border: 1px solid var(--border-color);
  background: var(--highlight-background);
  border-radius: var(--panel-radius);
  font-size: 14px;
}
.order-summary h3 {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 24px;
  letter-spacing: 0;
}
.order-item {
  display: flex;
  gap: 14px;
  align-items: flex-start;
  margin-bottom: 20px;
}
.order-item img {
  width: 64px;
  height: 64px;
  flex-shrink: 0;
  object-fit: contain;
  padding: 6px;
  background: var(--surface-background);
  border-radius: var(--control-radius);
}
.details {
  min-width: 0;
  overflow-wrap: anywhere;
}
.details h4 {
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0;
  margin-bottom: 6px;
}
.details p {
  font-size: 12px;
  color: var(--muted-text-color);
  margin-top: 3px;
}
.italic {
  font-style: italic;
  font-size: 12px;
}
.totals {
  border-top: 1px solid var(--highlight-border-color);
  padding-top: 16px;
  margin-top: 24px;
}
.totals p {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  margin-top: 10px;
}
.totals p:last-child {
  border-top: 1px solid var(--highlight-border-color);
  padding-top: 16px;
  margin-top: 16px;
  font-size: 18px;
  font-weight: 700;
}
@media (max-width: 600px) {
  .order-summary {
    padding: 20px;
  }
}
</style>
