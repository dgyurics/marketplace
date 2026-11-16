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
          <span>{{ offer.product.id }}</span>
        </div>
        <div class="info-row">
          <label>Amount:</label>
          <span>{{ formatPrice(offer.amount) }}</span>
        </div>
        <div class="info-row">
          <label>Status:</label>
          <span>{{ offer.status }}</span>
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
  max-width: 800px;
  margin: auto;
  padding: 20px;
}

.header {
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--border-color);
}

.header h1 {
  margin: 0 0 10px 0;
  font-size: 24px;
  color: var(--text-color);
}

.meta {
  display: flex;
  align-items: center;
  gap: 20px;
}

.id {
  font-family: 'Roboto Mono', monospace;
  font-size: 14px;
  color: var(--muted-text-color);
}

.status {
  padding: 4px 12px;
  border-radius: var(--control-radius);
  border-width: 1px;
  border-style: solid;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
}

.info-section {
  margin-bottom: 30px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--subtle-border-color);
}

.info-row label {
  font-weight: 600;
  color: var(--text-color);
  min-width: 150px;
}

.info-row span {
  color: var(--muted-text-color);
  flex: 1;
  text-align: right;
}

.pickup-notes {
  white-space: pre-wrap;
  word-break: break-word;
}

.mt-30 {
  margin-top: 30px;
}
</style>
