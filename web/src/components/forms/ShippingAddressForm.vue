<template>
  <form @submit.prevent="handleSubmit">
    <div class="form-group-flex">
      <InputText
        label="full name"
        :model-value="formData.name ?? ''"
        @update:model-value="(val) => updateField('name', val)"
      />
    </div>

    <div class="form-group-flex">
      <InputText
        label="address"
        required
        :model-value="formData.line1"
        @update:model-value="(val) => updateField('line1', val)"
      />
    </div>

    <div class="form-group-flex">
      <InputText
        label="Apt, Suite, Building"
        :model-value="formData.line2 ?? ''"
        @update:model-value="(val) => updateField('line2', val)"
      />
    </div>

    <div class="form-row">
      <div class="form-group-flex">
        <InputText
          label="city"
          required
          :model-value="formData.city"
          @update:model-value="(val) => updateField('city', val)"
        />
      </div>
      <div v-if="locale.state_required" class="form-group-flex">
        <SelectInput
          :label="locale.state_label"
          :options="states"
          :required="locale.state_required"
          :model-value="formData.state ?? ''"
          @update:model-value="(val) => updateField('state', val)"
        />
      </div>
      <div class="form-group-flex">
        <InputText
          :label="locale.postal_code_label"
          :pattern="locale.postal_code_pattern"
          title="Invalid format"
          required
          :model-value="formData.postal_code"
          @update:model-value="(val) => updateField('postal_code', val)"
        />
      </div>
    </div>

    <div class="form-group-flex">
      <InputText
        label="email"
        type="email"
        required
        :model-value="formData.email"
        @update:model-value="(val) => updateField('email', val)"
      />
      <small class="receipt-note">A receipt will be sent to this email.</small>
    </div>

    <button type="submit" class="btn-full-width mt-15" :tabindex="0">Continue</button>
  </form>
</template>

<script setup lang="ts">
import { computed } from 'vue'

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

const emit = defineEmits<{
  submit: [address: Address]
  'update:modelValue': [address: Address]
}>()

function updateField<K extends keyof Address>(field: K, value: Address[K]) {
  emit('update:modelValue', {
    line1: '',
    city: '',
    postal_code: '',
    email: '',
    country: getLocale().country_code,
    ...props.modelValue,
    [field]: value,
  })
}

const formData = computed(() => ({
  line1: '',
  city: '',
  postal_code: '',
  email: '',
  country: getLocale().country_code,
  ...props.modelValue,
}))

function handleSubmit() {
  emit('submit', formData.value)
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
