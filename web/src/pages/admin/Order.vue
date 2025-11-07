<template>
  <div class="order-container">
    <DataTable :columns="columns" :data="formattedOrders" :on-row-click="handleRowClick" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import DataTable from '@/components/DataTable.vue'
import { getOrders } from '@/services/api'
import type { Order } from '@/types'
import { formatPrice } from '@/utilities/currency'
import { formatDate } from '@/utilities/dateFormat'

const router = useRouter()
const orders = ref<Order[]>([])

const columns = [
  'id',
  'email',
  'status',
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
    email: order.address.email,
    status: order.status,
    location: order.address
      ? [order.address.city, order.address.state].filter(Boolean).join(', ')
      : '',
    amount: formatPrice(order.amount),
    tax: formatPrice(order.tax_amount),
    shipping: formatPrice(order.shipping_amount),
    total: formatPrice(order.total_amount),
    updated: formatDate(new Date(order.updated_at)),
    created: formatDate(new Date(order.created_at)),
    _originalOrder: order, // Keep reference to original order
  }))
)

const handleRowClick = (row: { [key: string]: unknown }) => {
  const originalOrder = row['_originalOrder'] as Order
  router.push(`/admin/orders/${originalOrder.id}`)
}

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
