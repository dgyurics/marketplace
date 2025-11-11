<template>
  <form @submit.prevent="handleSubmit">
    <div class="form-group-flex">
      <InputText v-model="addressee" label="full name" />
    </div>

    <div class="form-group-flex">
      <InputText v-model="formData.line1" label="address" required />
    </div>

    <div class="form-group-flex">
      <InputText v-model="line2" label="Apt, Suite, Building" />
    </div>

    <div class="form-row">
      <div class="form-group-flex">
        <InputText v-model="formData.city" label="city" required />
      </div>
      <div class="form-group-flex">
        <InputText v-model="state" label="state" />
      </div>
      <div class="form-group-flex">
        <InputText v-model="formData.postal_code" label="zip code" required />
      </div>
    </div>

    <div class="form-group-flex">
      <InputText v-model="formData.email" label="email" required />
      <small class="receipt-note">A receipt will be sent to this email.</small>
    </div>

    <button type="submit" class="btn-full-width mt-15">Continue</button>
  </form>
</template>

<script setup lang="ts">
import { computed, reactive } from 'vue'

import { InputText } from '@/components/forms'
import type { Address } from '@/types'
import { getCountryForLocale, getAppLocale } from '@/utilities'

const props = defineProps<{ modelValue?: Address }>()

const addressee = computed({
  get: () => formData.addressee ?? '',
  set: (val: string) => (formData.addressee = val),
})

const line2 = computed({
  get: () => formData.line2 ?? '',
  set: (val: string) => (formData.line2 = val),
})

const state = computed({
  get: () => formData.state ?? '',
  set: (val: string) => (formData.state = val),
})

const emit = defineEmits<{
  submit: [address: Address]
}>()

const formData = reactive<Address>({
  line1: '',
  city: '',
  postal_code: '',
  email: '',
  country: getCountryForLocale(getAppLocale()),
  ...props.modelValue,
})

function handleSubmit() {
  emit('submit', formData)
}
</script>

<style scoped>
.form-row {
  display: flex;
  gap: 10px;
}

.form-row .form-group-flex {
  flex: 1;
}

.receipt-note {
  font-size: 10px;
  color: #666;
  margin-top: 2px;
}
</style>
