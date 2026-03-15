<template>
  <div class="purchase-intent-container">
    <h1 class="page-title">Purchase Intents</h1>
    <DataTable :columns="columns" :data="formattedPurchaseIntents" :on-row-click="handleRowClick" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import DataTable from '@/components/DataTable.vue'
import { getPurchaseIntents, updatePurchaseIntent } from '@/services/api'
import type { PurchaseIntent, PurchaseIntentStatus } from '@/types'
import { displayPrice } from '@/utilities/currency'
import { formatShortDate } from '@/utilities/dateFormat'

const router = useRouter()
const purchaseIntents = ref<PurchaseIntent[]>([])

const columns = ['id', 'product_id', 'user_id', 'offer_price', 'status', 'created_at']

const statusOptions: PurchaseIntentStatus[] = [
  'pending',
  'accepted',
  'rejected',
  'canceled',
  'completed',
]

const formattedPurchaseIntents = computed(() =>
  purchaseIntents.value.map((pi) => ({
    id: pi.id,
    product_id: pi.product.id,
    user_id: pi.user_id,
    offer_price: displayPrice(pi.offer_price),
    status: pi.status,
    created_at: formatShortDate(new Date(pi.created_at)),
    _statusDropdown: createStatusDropdown(pi),
    _original: pi,
  }))
)

const createStatusDropdown = (pi: PurchaseIntent) => {
  return {
    currentStatus: pi.status,
    options: statusOptions,
    onChange: async (newStatus: PurchaseIntentStatus) => {
      try {
        await updatePurchaseIntent(pi.id, newStatus)
        pi.status = newStatus
      } catch (error) {
        console.error('Error updating purchase intent:', error)
      }
    },
  }
}

const handleRowClick = (row: { [key: string]: unknown }) => {
  const pi = row['_original'] as PurchaseIntent
  router.push(`/admin/purchase-intents/${pi.id}`)
}

const fetchPurchaseIntents = async () => {
  try {
    const data = await getPurchaseIntents()
    purchaseIntents.value = data
  } catch {
    // Handle error silently
  }
}

onMounted(() => {
  fetchPurchaseIntents()
})
</script>

<style scoped>
.purchase-intent-container {
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
