<template>
  <div class="offer-container">
    <h1 class="page-title">Offers</h1>
    <DataTable :columns="columns" :data="formattedOffers" :on-row-click="handleRowClick" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import DataTable from '@/components/DataTable.vue'
import { getOffers, updateOffer } from '@/services/api'
import type { Offer, OfferStatus } from '@/types'
import { displayPrice } from '@/utilities/currency'
import { formatShortDate } from '@/utilities/dateFormat'

const router = useRouter()
const offers = ref<Offer[]>([])

const columns = ['id', 'product_id', 'user_id', 'amount', 'status', 'created_at']

const statusOptions: OfferStatus[] = ['pending', 'accepted', 'rejected', 'canceled', 'completed']

const formattedOffers = computed(() =>
  offers.value.map((pi) => ({
    id: pi.id,
    product_id: pi.product.id,
    user_id: pi.user_id,
    amount: displayPrice(pi.amount),
    status: pi.status,
    created_at: formatShortDate(new Date(pi.created_at)),
    _statusDropdown: createStatusDropdown(pi),
    _original: pi,
  }))
)

const createStatusDropdown = (pi: Offer) => {
  return {
    currentStatus: pi.status,
    options: statusOptions,
    onChange: async (newStatus: OfferStatus) => {
      try {
        await updateOffer(pi.id, newStatus)
        pi.status = newStatus
      } catch (error) {
        console.error('Error updating offer:', error)
      }
    },
  }
}

const handleRowClick = (row: { [key: string]: unknown }) => {
  const pi = row['_original'] as Offer
  router.push(`/admin/offers/${pi.id}`)
}

const fetchOffers = async () => {
  try {
    const data = await getOffers()
    offers.value = data
  } catch {
    // Handle error silently
  }
}

onMounted(() => {
  fetchOffers()
})
</script>

<style scoped>
.offer-container {
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
