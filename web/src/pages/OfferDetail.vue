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

      <button type="button" class="btn-full-width btn-outline mt-30" @click="goBack">
        Back to Offers
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getOfferById } from '@/services/api'
import type { Offer } from '@/types'
import { formatPrice } from '@/utilities/currency'
import { formatDate } from '@/utilities/dateFormat'

const route = useRoute()
const router = useRouter()

const offer = ref<Offer | null>(null)

const fetchOffer = async () => {
  try {
    const id = route.params['id'] as string
    offer.value = await getOfferById(id)
  } catch {
    offer.value = null
  }
}

const goBack = () => {
  router.back()
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
  border-bottom: 1px solid #ddd;
}

.header h1 {
  margin: 0 0 10px 0;
  font-size: 24px;
  color: #333;
}

.meta {
  display: flex;
  align-items: center;
  gap: 20px;
}

.id {
  font-family: 'Roboto Mono', monospace;
  font-size: 14px;
  color: #666;
}

.status {
  padding: 4px 12px;
  border-radius: 4px;
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
  border-bottom: 1px solid #f0f0f0;
}

.info-row label {
  font-weight: 600;
  color: #333;
  min-width: 150px;
}

.info-row span {
  color: #666;
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
