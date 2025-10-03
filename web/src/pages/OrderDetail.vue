<template>
  <div class="order-detail-container">
    <div v-if="loading" class="loading">Loading order...</div>
    <div v-else-if="order" class="order-detail">
      <div class="order-header">
        <h1>Order Details</h1>
        <div class="order-meta">
          <span class="order-id">ID: {{ order.id }}</span>
          <span class="order-status">
            {{ order.status.toUpperCase() }}
          </span>
        </div>
      </div>

      <div class="order-info">
        <div v-if="order.email" class="info-row">
          <label>Email:</label>
          <span>{{ order.email }}</span>
        </div>
        <div class="info-row">
          <label>Currency:</label>
          <span>{{ order.currency.toUpperCase() }}</span>
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

      <div v-if="order.address" class="address-section">
        <h3>Shipping Address</h3>
        <div class="address">
          <div v-if="order.address.addressee">{{ order.address.addressee }}</div>
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
              <span class="item-summary">{{ item.product.summary }}</span>
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
          <span>{{ formatPrice(order.amount) }}</span>
        </div>
        <div class="total-row">
          <label>Tax:</label>
          <span>{{ formatPrice(order.tax_amount) }}</span>
        </div>
        <div class="total-row">
          <label>Shipping:</label>
          <span>{{ formatPrice(order.shipping_amount) }}</span>
        </div>
        <div class="total-row total">
          <label>Total:</label>
          <span>{{ formatPrice(order.total_amount) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getOrderOwner, getOrderPublic } from '@/services/api'
import type { Order } from '@/types'
import { formatPrice, formatDate } from '@/utilities'

const route = useRoute()
const router = useRouter()

const order = ref<Order | null>(null)
const loading = ref(true)

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
    loading.value = true
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
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchOrder()
})
</script>

<style scoped>
.order-detail-container {
  max-width: 800px;
  width: 800px;
  margin: auto;
  padding: 20px;
}

.loading {
  text-align: center;
  padding: 40px;
  font-size: 16px;
  color: #666;
}

.order-header {
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 1px solid #ddd;
}

.order-header h1 {
  margin: 0 0 10px 0;
  font-size: 24px;
  color: #333;
}

.order-meta {
  display: flex;
  align-items: center;
  gap: 20px;
}

.order-id {
  font-family: 'Roboto Mono', monospace;
  font-size: 14px;
  color: #666;
}

.order-status {
  padding: 4px 12px;
  border-radius: 4px;
  border-width: 1px;
  border-style: solid;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
}

.order-info {
  margin-bottom: 30px;
}

.info-row {
  display: flex;
  margin-bottom: 10px;
}

.info-row label {
  min-width: 100px;
  font-weight: 500;
  color: #666;
}

.address-section {
  margin-bottom: 30px;
}

.address-section h3 {
  margin: 0 0 15px 0;
  font-size: 18px;
  color: #333;
}

.address {
  padding: 15px;
  background-color: #f9f9f9;
  border-radius: 4px;
  line-height: 1.4;
}

.items-section {
  margin-bottom: 30px;
}

.items-section h3 {
  margin: 0 0 15px 0;
  font-size: 18px;
  color: #333;
}

.order-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  border-bottom: 1px solid #eee;
}

.order-item:last-child {
  border-bottom: none;
}

.item-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.item-name {
  font-weight: 500;
  margin-bottom: 4px;
}

.item-summary {
  font-size: 14px;
  color: #666;
}

.item-details {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
  font-size: 14px;
}

.item-quantity {
  color: #666;
}

.item-price,
.item-total {
  font-weight: 500;
}

.totals-section {
  border-top: 1px solid #ddd;
  padding-top: 15px;
}

.total-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.total-row.total {
  font-weight: 600;
  font-size: 16px;
  border-top: 1px solid #ddd;
  padding-top: 8px;
  margin-top: 8px;
}

.total-row label {
  color: #666;
}
</style>
