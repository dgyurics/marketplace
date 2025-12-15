<template>
  <div class="shipping-zone-container">
    <!-- Regular Shipping Zones -->
    <div class="section">
      <h2>Shipping Zones</h2>
      <div class="new-zone-form">
        <form @submit.prevent="handleShippingZoneSubmit">
          <div class="form-row">
            <SelectInput
              v-model="newShippingZone.state"
              :label="locale.state_label"
              :options="states"
            />
            <InputText v-model="newShippingZone.postal_code" label="postal code" />
          </div>
          <button type="submit" class="btn-full-width mt-15" :tabindex="0">Add Zone</button>
        </form>
        <p v-if="shippingZoneError" class="error">{{ shippingZoneError }}</p>
      </div>
      <DataTable
        :columns="shippingZoneColumns"
        :data="formattedShippingZones"
        :on-row-click="removeShippingZoneHandler"
      />
    </div>

    <!-- Excluded Shipping Zones -->
    <div class="section">
      <h2>Shipping Exclusions</h2>
      <div class="new-zone-form">
        <form @submit.prevent="handleExcludedZoneSubmit">
          <div class="form-row">
            <InputText v-model="newExcludedZone.postal_code" label="postal code" required />
          </div>
          <button type="submit" class="btn-full-width mt-15" :tabindex="0">Add Exclusion</button>
        </form>
        <p v-if="excludedZoneError" class="error">{{ excludedZoneError }}</p>
      </div>
      <DataTable
        :columns="excludedZoneColumns"
        :data="formattedExcludedZones"
        :on-row-click="removeExcludedZoneHandler"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

import DataTable from '@/components/DataTable.vue'
import { InputText, SelectInput } from '@/components/forms'
import {
  getShippingZones,
  createShippingZone,
  removeShippingZone,
  getExcludedShippingZones,
  createExcludedShippingZone,
  removeExcludedShippingZone,
} from '@/services/api'
import type { ShippingZone, ExcludedShippingZone, Locale } from '@/types'
import { getLocale } from '@/utilities'

const locale: Locale = getLocale()
const states = Object.entries(locale.state_codes || []).map(([k, v]) => ({ value: k, label: v }))
const shippingZones = ref<ShippingZone[]>([])
const excludedZones = ref<ExcludedShippingZone[]>([])

const shippingZoneError = ref<string | null>(null)
const excludedZoneError = ref<string | null>(null)

const newShippingZone = ref({
  state: '',
  postal_code: '',
})

const newExcludedZone = ref({
  postal_code: '',
})

const shippingZoneColumns = ['country', 'state', 'postal_code', 'action']
const excludedZoneColumns = ['country', 'postal_code', 'action']

const formattedShippingZones = computed(() =>
  shippingZones.value.map((zone) => ({
    id: zone.id,
    country: zone.country,
    state: zone.state || '-',
    postal_code: zone.postal_code || '-',
    action: 'Click to remove',
  }))
)

const formattedExcludedZones = computed(() =>
  excludedZones.value.map((zone) => ({
    id: zone.id,
    country: zone.country,
    postal_code: zone.postal_code,
    action: 'Click to remove',
  }))
)

const fetchShippingZones = async () => {
  try {
    const response = await getShippingZones()
    shippingZones.value = response
  } catch {
    // Handle error silently
  }
}

const fetchExcludedZones = async () => {
  try {
    const response = await getExcludedShippingZones()
    excludedZones.value = response
  } catch {
    // Handle error silently
  }
}

const handleShippingZoneSubmit = async () => {
  try {
    const zoneData = {
      country: locale.country_code,
      state: newShippingZone.value.state || null,
      postal_code: newShippingZone.value.postal_code || null,
    }

    await createShippingZone(zoneData)

    // Clear any previous errors
    shippingZoneError.value = null

    // Reset form
    newShippingZone.value = { state: '', postal_code: '' }

    // Refresh zones
    await fetchShippingZones()
  } catch (error: any) {
    const status = error.response?.status
    if (status === 400) {
      shippingZoneError.value = 'Invalid shipping zone data'
    } else if (status === 409) {
      shippingZoneError.value = 'Shipping zone already exists'
    } else if (status === 422) {
      shippingZoneError.value = 'Conflicting shipping exclusion exists'
    } else {
      shippingZoneError.value = 'Something went wrong'
    }
  }
}

const handleExcludedZoneSubmit = async () => {
  try {
    const zoneData = {
      country: locale.country_code,
      postal_code: newExcludedZone.value.postal_code,
    }

    await createExcludedShippingZone(zoneData)

    // Clear any previous errors
    excludedZoneError.value = null

    // Reset form
    newExcludedZone.value = { postal_code: '' }

    // Refresh zones
    await fetchExcludedZones()
  } catch (error: any) {
    const status = error.response?.status
    if (status === 400) {
      excludedZoneError.value = 'Invalid excluded zone data'
    } else if (status === 409) {
      excludedZoneError.value = 'Excluded zone already exists'
    } else if (status === 422) {
      excludedZoneError.value = 'Conflicting shipping zone exists'
    } else {
      excludedZoneError.value = 'Something went wrong'
    }
  }
}

const removeShippingZoneHandler = async (row: any) => {
  if (row.id) {
    try {
      await removeShippingZone(row.id)
      await fetchShippingZones()
    } catch {
      // Handle error silently
    }
  }
}

const removeExcludedZoneHandler = async (row: any) => {
  if (row.id) {
    try {
      await removeExcludedShippingZone(row.id)
      await fetchExcludedZones()
    } catch {
      // Handle error silently
    }
  }
}

onMounted(() => {
  fetchShippingZones()
  fetchExcludedZones()
})
</script>

<style scoped>
.shipping-zone-container {
  max-width: 1200px;
  margin: auto;
  padding: 20px;
}

.section {
  margin-bottom: 50px;
}

.section h2 {
  font-size: 20px;
  font-weight: 300;
  margin-bottom: 20px;
  color: #333;
  text-align: center;
}

.new-zone-form {
  margin-bottom: 30px;
}

.form-row {
  display: flex;
  gap: 10px;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  margin-bottom: 15px;
}

.form-row :deep(.input-container) {
  flex: 1 1 calc(33% - 10px);
}

.error {
  color: #e74c3c;
  text-align: center;
  margin-top: 10px;
  font-size: 14px;
}
</style>
