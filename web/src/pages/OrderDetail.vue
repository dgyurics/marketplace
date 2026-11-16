<template>
  <div class="order-detail-container">
    <div v-if="order" class="order-detail">
      <div class="order-header">
        <h1>Order Details</h1>
        <div class="order-meta">
          <span class="order-id">ID: {{ order.id }}</span>
          <span class="order-status">
            {{ (order.status ?? 'pending').toUpperCase() }}
          </span>
        </div>
      </div>

      <div class="order-info">
        <div v-if="order.address.email" class="info-row">
          <label>Email:</label>
          <span>{{ order.address.email }}</span>
        </div>
        <div class="info-row">
          <label>Created:</label>
          <span>{{ formatDate(order.created_at) }}</span>
        </div>
        <div class="info-row">
          <label>Updated:</label>
          <span>{{ formatDate(order.updated_at) }}</span>
        </div>
      </div>

      <div v-if="addressAvailable(order.address)" class="address-section">
        <h3>Shipping Address</h3>
        <div class="address">
          <div v-if="order.address.name">{{ order.address.name }}</div>
          <div>{{ order.address.line1 }}</div>
          <div v-if="order.address.line2">{{ order.address.line2 }}</div>
          <div>
            {{ order.address.city }}, {{ order.address.state }}
            {{ order.address.postal_code }}
          </div>
          <div v-if="order.address.country">{{ order.address.country }}</div>
        </div>
      </div>

      <div class="items-section">
        <h3>Order Items</h3>
        <div class="items-list">
          <div v-for="item in order.items" :key="item.product.id" class="order-item">
            <div class="item-info">
              <span class="item-name">{{ item.product.name }}</span>
              <span v-if="item.product.summary" class="item-summary">{{
                item.product.summary
              }}</span>
            </div>
            <div class="item-details">
              <span class="item-quantity">Qty: {{ item.quantity }}</span>
              <span class="item-price">{{ formatPrice(item.unit_price) }}</span>
              <span class="item-total">
                {{ formatPrice(item.unit_price * item.quantity) }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="totals-section">
        <div class="total-row">
          <label>Subtotal:</label>
          <span>{{ formatPrice(order.amount ?? 0) }}</span>
        </div>
        <div class="total-row">
          <label>Tax:</label>
          <span>{{ formatPrice(order.tax_amount ?? 0) }}</span>
        </div>
        <div class="total-row">
          <label>Shipping:</label>
          <span>{{ formatPrice(order.shipping_amount ?? 0) }}</span>
        </div>
        <div class="total-row total">
          <label>Total:</label>
          <span>{{ formatPrice(order.total_amount ?? 0) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getOrderOwner, getOrderPublic } from '@/services/api'
import type { Address, Order } from '@/types'
import { formatPrice, formatDate } from '@/utilities'

const route = useRoute()
const router = useRouter()

const order = ref<Order | null>(null)

const addressAvailable = (address: Address) => address.id !== ''

const tryGetOrder = async (orderId: string): Promise<Order | null> => {
  // Try owner endpoint first
  try {
    return await getOrderOwner(orderId)
  } catch (error: any) {
    const status = error.response?.status

    // If not authorized, try public endpoint
    if (status === 401 || status === 404) {
      return await getOrderPublic(orderId)
    }

    // Re-throw for other errors
    throw error
  }
}

const fetchOrder = async () => {
  const orderId = route.params['id'] as string

  try {
    order.value = await tryGetOrder(orderId)
  } catch (error: any) {
    const status = error.response?.status

    if (status === 404) {
      router.push('/not-found')
      return
    }

    if (status >= 500) {
      router.push('/error')
      return
    }

    // For other errors, just clear the order
    order.value = null
  }
}

onMounted(() => {
  fetchOrder()
})
</script>

<style scoped>
.order-detail-container {
  width: 100%;
  max-width: 1100px;
  margin: 0 auto;
  padding: 36px 32px 64px;
}
.order-detail {
  display: grid;
  grid-template-columns: minmax(0, 1.6fr) minmax(0, 1fr);
  grid-template-areas: 'header header' 'info address' 'items totals';
  gap: 24px;
  align-items: start;
}
.order-header {
  grid-area: header;
  padding: 24px 28px;
  border-radius: var(--panel-radius);
  background: var(--highlight-background);
}
.order-header h1 {
  font-size: 26px;
  font-weight: 700;
  letter-spacing: -0.7px;
  margin-bottom: 14px;
}
.order-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px 20px;
}
.order-id {
  font-family: 'Roboto Mono', monospace;
  font-size: 12px;
  color: var(--highlight-text-color);
  overflow-wrap: anywhere;
}
.order-status {
  padding: 4px 10px;
  border: 1px solid var(--highlight-border-color);
  border-radius: 6px;
  background: var(--surface-background);
  color: var(--highlight-text-color);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.5px;
}
.order-info,
.address-section,
.items-section,
.totals-section {
  min-width: 0;
  padding: 24px;
  border: 1px solid var(--border-color);
  border-radius: var(--panel-radius);
}
.order-info {
  grid-area: info;
}
.info-row {
  display: grid;
  grid-template-columns: 80px minmax(0, 1fr);
  gap: 16px;
  font-size: 13px;
  overflow-wrap: anywhere;
}
.info-row + .info-row {
  margin-top: 16px;
}
.info-row label {
  color: var(--muted-text-color);
}
.address-section {
  grid-area: address;
}
.address-section h3,
.items-section h3 {
  font-size: 16px;
  font-weight: 600;
  letter-spacing: 0;
  margin-bottom: 18px;
}
.address {
  font-size: 13px;
  line-height: 1.8;
  color: var(--muted-text-color);
  overflow-wrap: anywhere;
}
.address > div:first-child {
  font-weight: 600;
  color: var(--text-color);
}
.items-section {
  grid-area: items;
}
.order-item {
  display: flex;
  justify-content: space-between;
  align-items: start;
  gap: 24px;
  padding: 20px 0;
}
.order-item + .order-item {
  border-top: 1px solid var(--border-color);
}
.order-item:last-child {
  padding-bottom: 0;
}
.item-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
  overflow-wrap: anywhere;
}
.item-name {
  font-size: 14px;
  font-weight: 600;
}
.item-summary {
  font-size: 12px;
  color: var(--muted-text-color);
  line-height: 1.7;
}
.item-details {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 5px;
  font-size: 12px;
  white-space: nowrap;
}
.item-quantity,
.item-price {
  color: var(--muted-text-color);
}
.item-total {
  font-size: 14px;
  font-weight: 600;
  margin-top: 4px;
}
.totals-section {
  grid-area: totals;
  background: var(--highlight-background);
}
.total-row {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  font-size: 14px;
}
.total-row + .total-row {
  margin-top: 14px;
}
.total-row label {
  color: var(--highlight-text-color);
}
.total-row.total {
  font-size: 20px;
  font-weight: 700;
  border-top: 1px solid var(--highlight-border-color);
  padding-top: 20px;
  margin-top: 20px;
}
.total-row.total label {
  color: var(--text-color);
}
@media (max-width: 768px) {
  .order-detail {
    grid-template-columns: minmax(0, 1fr);
    grid-template-areas: 'header' 'info' 'address' 'items' 'totals';
    gap: 20px;
  }
}
@media (max-width: 600px) {
  .order-detail-container {
    padding: 24px 16px 40px;
  }
  .order-header,
  .order-info,
  .address-section,
  .items-section,
  .totals-section {
    padding: 20px;
  }
  .order-header h1 {
    font-size: 23px;
  }
  .order-item {
    flex-direction: column;
    gap: 14px;
  }
  .item-details {
    width: 100%;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 12px;
  }
  .item-total {
    margin-top: 0;
  }
}
</style>
