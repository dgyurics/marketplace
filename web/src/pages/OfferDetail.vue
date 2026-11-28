<template>
  <div class="detail-container">
    <div v-if="offer" class="detail">
      <div class="header">
        <h1>Offer Details</h1>
        <div class="meta">
          <span class="id">ID: {{ offer.id }}</span>
          <span class="status">
            {{ offer.status.toUpperCase() }}
          </span>
        </div>
      </div>

      <div class="info-section">
        <div class="info-row">
          <label>Product:</label>
          <span>{{ offer.product.name }}</span>
        </div>
        <div class="info-row">
          <label>Amount:</label>
          <span>{{ formatPrice(offer.amount) }}</span>
        </div>
        <div v-if="offer.comment" class="info-row">
          <label>Comment:</label>
          <span class="pickup-notes">{{ offer.comment }}</span>
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

import { getOfferOwner } from '@/services/api'
import type { Offer } from '@/types'
import { formatPrice } from '@/utilities/currency'
import { formatDate } from '@/utilities/date'

const route = useRoute()

const offer = ref<Offer | null>(null)

const fetchOffer = async () => {
  try {
    const id = route.params['id'] as string
    offer.value = await getOfferOwner(id)
  } catch {
    offer.value = null
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
.status {
  padding: 4px 10px;
  border: 1px solid var(--highlight-border-color);
  border-radius: 6px;
  background: var(--surface-background);
  color: var(--highlight-text-color);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.5px;
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
