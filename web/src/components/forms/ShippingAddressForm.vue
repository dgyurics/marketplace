<template>
  <form @submit.prevent="handleSubmit">
    <div class="form-group-flex">
      <InputText v-model="addressee" label="full name" :tabindex="1" />
    </div>

    <div class="form-group-flex">
      <InputText v-model="formData.line1" label="address" required :tabindex="2" />
    </div>

    <div class="form-group-flex">
      <InputText v-model="line2" label="Apt, Suite, Building" :tabindex="3" />
    </div>

    <div class="form-row">
      <div class="form-group-flex">
        <InputText v-model="formData.city" label="city" required :tabindex="4" />
      </div>
      <div v-if="locale.state_required" class="form-group-flex">
        <SelectInput
          v-model="state"
          :label="locale.state_label"
          :options="states"
          :required="locale.state_required"
          :tabindex="5"
        />
      </div>
      <div class="form-group-flex">
        <InputText
          v-model="formData.postal_code"
          :label="locale.postal_code_label"
          :pattern="locale.postal_code_pattern"
          title="Invalid format"
          required
          :tabindex="6"
        />
      </div>
    </div>

    <div class="form-group-flex">
      <InputText v-model="formData.email" label="email" type="email" required :tabindex="7" />
      <small class="receipt-note">A receipt will be sent to this email.</small>
    </div>

    <button type="submit" class="btn-full-width mt-15" :tabindex="8">Continue</button>
  </form>
</template>

<script setup lang="ts">
import { computed, reactive } from 'vue'

import { InputText, SelectInput } from '@/components/forms'
import type { Address, Locale } from '@/types'
import { getLocale } from '@/utilities'

const locale: Locale = getLocale()
const props = defineProps<{ modelValue?: Address }>()

const states = Object.entries(locale.state_codes || []).map(([k, v]) => {
  return {
    value: k,
    label: v,
  }
})

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
  country: getLocale().country_code,
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
