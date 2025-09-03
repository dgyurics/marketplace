<template>
  <div class="billing-address-form">
    <div class="form-group-flex">
      <label for="billingLine1">Address</label>
      <input id="billingLine1" v-model="address.line1" type="text" required />
    </div>

    <div class="form-group-flex">
      <label for="billingLine2">Apt, Suite, Building (Optional)</label>
      <input id="billingLine2" v-model="address.line2" type="text" />
    </div>

    <div class="form-row">
      <div class="form-group-flex">
        <label for="billingCity">City</label>
        <input id="billingCity" v-model="address.city" type="text" required />
      </div>
      <div class="form-group-flex">
        <label for="billingState">State</label>
        <input id="billingState" v-model="address.state" type="text" required />
      </div>
      <div class="form-group-flex">
        <label for="billingZip">Zip Code</label>
        <input id="billingZip" v-model="address.postal_code" type="text" required />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import type { Address } from '@/types'

const props = defineProps<{
  modelValue: Address
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Address]
}>()

const address = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})
</script>

<style scoped>
.billing-address-form {
  display: flex;
  flex-direction: column;
}

.form-row {
  display: flex;
  gap: 10px;
}

.form-row .form-group-flex {
  flex: 1;
}

input[type='text'] {
  width: 100%;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 16px;
  box-sizing: border-box;
}
</style>
