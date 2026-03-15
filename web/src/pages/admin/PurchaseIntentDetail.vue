<template>
  <div class="detail-container">
    <div v-if="purchaseIntent" class="detail">
      <div class="header">
        <h1>Purchase Intent Details</h1>
        <div class="meta">
          <span class="id">ID: {{ purchaseIntent.id }}</span>
        </div>
      </div>

      <div class="info-section">
        <div class="info-row">
          <label>Product ID:</label>
          <span>{{ purchaseIntent.product.id }}</span>
        </div>
        <div class="info-row">
          <label>Product Name:</label>
          <span>{{ purchaseIntent.product.name }}</span>
        </div>
        <div class="info-row">
          <label>User ID:</label>
          <span>{{ purchaseIntent.user_id }}</span>
        </div>
        <div class="info-row">
          <label>Offer Price:</label>
          <span>{{ displayPrice(purchaseIntent.offer_price) }}</span>
        </div>
        <div class="info-row">
          <label>Status:</label>
          <select v-model="currentStatus" @change="handleStatusChange" class="status-select">
            <option v-for="status in statusOptions" :key="status" :value="status">
              {{ status }}
            </option>
          </select>
        </div>
        <div class="info-row">
          <label>Pickup Notes:</label>
          <span class="pickup-notes">{{ purchaseIntent.pickup_notes }}</span>
        </div>
        <div class="info-row">
          <label>Created:</label>
          <span>{{ formatDate(purchaseIntent.created_at) }}</span>
        </div>
        <div class="info-row">
          <label>Updated:</label>
          <span>{{ formatDate(purchaseIntent.updated_at) }}</span>
        </div>
      </div>

      <button type="button" class="btn-full-width btn-outline mt-30" @click="goBack">
        Back to Purchase Intents
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getPurchaseIntentById, updatePurchaseIntent } from '@/services/api'
import type { PurchaseIntent, PurchaseIntentStatus } from '@/types'
import { displayPrice } from '@/utilities/currency'
import { formatDate } from '@/utilities/dateFormat'

const route = useRoute()
const router = useRouter()

const purchaseIntent = ref<PurchaseIntent | null>(null)
const currentStatus = ref<PurchaseIntentStatus>('pending')

const statusOptions: PurchaseIntentStatus[] = [
  'pending',
  'accepted',
  'rejected',
  'canceled',
  'completed',
]

const fetchPurchaseIntent = async () => {
  try {
    const id = route.params['id'] as string
    const data = await getPurchaseIntentById(id)
    purchaseIntent.value = data
    currentStatus.value = data.status
  } catch (error) {
    console.error('Error fetching purchase intent:', error)
    purchaseIntent.value = null
  }
}

const handleStatusChange = async () => {
  if (!purchaseIntent.value) return

  try {
    await updatePurchaseIntent(purchaseIntent.value.id, currentStatus.value)
    purchaseIntent.value.status = currentStatus.value
  } catch (error) {
    console.error('Error updating purchase intent:', error)
    currentStatus.value = purchaseIntent.value.status
  }
}

const goBack = () => {
  router.push('/admin/purchase-intents')
}

onMounted(() => {
  fetchPurchaseIntent()
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
  gap: 20px;
}

.id {
  font-family: 'Roboto Mono', monospace;
  font-size: 14px;
  color: #666;
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

.status-select {
  padding: 6px 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  text-transform: capitalize;
}

.status-select:hover {
  border-color: #999;
}

.status-select:focus {
  outline: none;
  border-color: #333;
  box-shadow: 0 0 3px rgba(0, 0, 0, 0.1);
}

.mt-30 {
  margin-top: 30px;
}
</style>
