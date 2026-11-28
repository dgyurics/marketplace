<template>
  <div class="detail-container">
    <div v-if="offer" class="detail">
      <div class="header">
        <h1>Offer Details</h1>
        <div class="meta">
          <span class="id">ID: {{ offer.id }}</span>
          <select
            v-model="currentStatus"
            class="status-select"
            aria-label="Offer status"
            @change="handleStatusChange"
          >
            <option v-for="status in statusOptions" :key="status" :value="status">
              {{ status }}
            </option>
          </select>
        </div>
      </div>

      <div class="info-section">
        <div class="info-row">
          <label>Product ID:</label>
          <span>{{ offer.product.id }}</span>
        </div>
        <div class="info-row">
          <label>Product Name:</label>
          <span>{{ offer.product.name }}</span>
        </div>
        <div class="info-row">
          <label>User ID:</label>
          <span>{{ offer.user_id }}</span>
        </div>
        <div class="info-row">
          <label>Amount:</label>
          <span>{{ formatPrice(offer.amount) }}</span>
        </div>
        <div class="info-row">
          <label>Comment:</label>
          <span class="pickup-notes">{{ offer.comment || 'No comment' }}</span>
        </div>
        <div class="info-row">
          <label>Created:</label>
          <span>{{ formatDate(offer.created_at) }}</span>
        </div>
        <div class="info-row">
          <label>Updated:</label>
          <span>{{ formatDate(offer.updated_at) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'

import { getOfferAdmin, updateOffer } from '@/services/api'
import type { Offer, OfferStatus } from '@/types'
import { formatPrice } from '@/utilities/currency'
import { formatDate } from '@/utilities/date'

const route = useRoute()

const offer = ref<Offer | null>(null)
const currentStatus = ref<OfferStatus>('pending')

const statusOptions: OfferStatus[] = ['pending', 'accepted', 'rejected', 'canceled', 'completed']

const fetchOffer = async () => {
  try {
    const id = route.params['id'] as string
    const data = await getOfferAdmin(id)
    offer.value = data
    currentStatus.value = data.status
  } catch (error) {
    console.error('Error fetching offer:', error)
    offer.value = null
  }
}

const handleStatusChange = async () => {
  if (!offer.value) return

  try {
    await updateOffer(offer.value.id, currentStatus.value)
    offer.value.status = currentStatus.value
  } catch (error) {
    console.error('Error updating offer:', error)
    currentStatus.value = offer.value.status
  }
}

onMounted(() => {
  fetchOffer()
})
</script>

<style scoped>
.detail-container {
  width: 100%;
  max-width: 1100px;
  margin: 0 auto;
  padding: 36px 32px 64px;
}
.header {
  padding: 24px 28px;
  margin-bottom: 24px;
  border-radius: var(--panel-radius);
  background: var(--highlight-background);
}
.header h1 {
  font-size: 26px;
  font-weight: 700;
  letter-spacing: -0.7px;
  margin-bottom: 14px;
}
.meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px 20px;
}
.id {
  font-family: 'Roboto Mono', monospace;
  font-size: 12px;
  color: var(--highlight-text-color);
  overflow-wrap: anywhere;
}
.info-section {
  padding: 24px;
  border: 1px solid var(--border-color);
  border-radius: var(--panel-radius);
}
.info-row {
  display: grid;
  grid-template-columns: 100px minmax(0, 1fr);
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
.pickup-notes {
  white-space: pre-wrap;
}
.status-select {
  padding: 8px 12px;
  border: 1px solid var(--highlight-border-color);
  border-radius: var(--control-radius);
  background: var(--surface-background);
  color: var(--highlight-text-color);
  font-family: inherit;
  font-size: 13px;
  text-transform: capitalize;
  cursor: pointer;
}
.status-select:hover {
  border-color: var(--highlight-icon-color);
}
.status-select:focus-visible {
  border-color: var(--accent-color);
  outline-offset: 2px;
}
@media (max-width: 600px) {
  .detail-container {
    padding: 24px 16px 40px;
  }
  .header,
  .info-section {
    padding: 20px;
  }
  .header h1 {
    font-size: 23px;
  }
}
</style>
