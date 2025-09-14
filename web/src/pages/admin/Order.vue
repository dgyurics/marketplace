<template>
  <div class="order-container">
    <DataTable :columns="columns" :data="formattedOrders" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

import DataTable from '@/components/DataTable.vue'
import { getOrders } from '@/services/api'
import type { Order } from '@/types'

const orders = ref<Order[]>([])

const columns = [
  'id',
  'email',
  'status',
  'items',
  'location',
  'amount',
  'tax',
  'shipping',
  'total',
  'updated',
  'created',
]

const formattedOrders = computed(() =>
  orders.value.map((order) => ({
    id: order.id,
    email: order.email,
    status: order.status,
    items: order.items.reduce((sum, item) => sum + item.quantity, 0),
    location: order.address
      ? [order.address.city, order.address.state].filter(Boolean).join(', ')
      : '',
    amount: `$${(order.amount / 100).toFixed(2)}`,
    tax: `$${(order.tax_amount / 100).toFixed(2)}`,
    shipping: `$${(order.shipping_amount / 100).toFixed(2)}`,
    total: `$${(order.total_amount / 100).toFixed(2)}`,
    updated: new Date(order.updated_at).toLocaleString('en-US', {
      year: 'numeric',
      month: 'numeric',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    }),
    created: new Date(order.created_at).toLocaleString('en-US', {
      year: 'numeric',
      month: 'numeric',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    }),
  }))
)

const fetchOrders = async () => {
  try {
    const data = await getOrders()
    orders.value = data
  } catch {
    // Handle error silently
  }
}

onMounted(() => {
  fetchOrders()
})
</script>

<style scoped>
.order-container {
  max-width: 1200px;
  margin: auto;
  padding: 20px;
}

.page-title {
  font-size: 24px;
  font-weight: 300;
  margin-bottom: 20px;
  color: #333;
}
</style>
