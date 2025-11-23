<template>
  <div class="billing-address-form">
    <div class="form-group-flex">
      <InputText
        label="full name"
        required
        :model-value="modelValue.name ?? ''"
        @update:model-value="(val) => updateField('name', val)"
      />
    </div>
    <div class="form-group-flex">
      <InputText
        label="address"
        required
        :model-value="modelValue.line1 ?? ''"
        @update:model-value="(val) => updateField('line1', val)"
      />
    </div>
    <div class="form-group-flex">
      <InputText
        label="Apt, Suite, Building"
        :model-value="modelValue.line2 ?? ''"
        @update:model-value="(val) => updateField('line2', val)"
      />
    </div>
    <div class="form-row">
      <div class="form-group-flex">
        <InputText
          label="city"
          required
          :model-value="modelValue.city ?? ''"
          @update:model-value="(val) => updateField('city', val)"
        />
      </div>
      <div v-if="locale.state_required" class="form-group-flex">
        <SelectInput
          :label="locale.state_label"
          :options="states"
          :required="locale.state_required"
          :model-value="modelValue.state ?? ''"
          @update:model-value="(val) => updateField('state', val)"
        />
      </div>
      <div class="form-group-flex">
        <InputText
          :label="locale.postal_code_label"
          :pattern="locale.postal_code_pattern"
          title="Invalid format"
          required
          :model-value="modelValue.postal_code ?? ''"
          @update:model-value="(val) => updateField('postal_code', val)"
        />
      </div>
    </div>
    <div class="form-group-flex">
      <InputText
        label="email"
        type="email"
        required
        :model-value="modelValue.email ?? ''"
        @update:model-value="(val) => updateField('email', val)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { InputText, SelectInput } from '@/components/forms'
import type { Address, Locale } from '@/types'
import { getLocale } from '@/utilities'

const locale: Locale = getLocale()
const states = Object.entries(locale.state_codes || []).map(([k, v]) => ({ value: k, label: v }))

const props = defineProps<{ modelValue: Address }>()
const emit = defineEmits<{ 'update:modelValue': [value: Address] }>()

function updateField<K extends keyof Address>(field: K, value: Address[K]) {
  emit('update:modelValue', { ...props.modelValue, [field]: value })
}
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
