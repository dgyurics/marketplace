<template>
  <div class="input-container">
    <select
      v-model="value"
      :required="required || false"
      v-bind="tabindex ? { tabindex } : {}"
      class="input-field"
    >
      <option value="" v-bind="required ? { disabled: true } : {}">
        <!-- {{ `Select a ${label}` }} -->
      </option>
      <option v-for="option in options" :key="option.value" :value="option.value">
        {{ option.label }}
      </option>
    </select>
    <span v-if="label" class="input-label" :class="{ 'input-label--optional': !required }">
      {{ label }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Option {
  value: string
  label: string
}

const props = defineProps<{
  modelValue: string
  label?: string
  required?: boolean
  options: Option[]
  tabindex?: number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const value = computed({
  get: () => props.modelValue,
  set: (newValue) => emit('update:modelValue', newValue),
})
</script>

<style scoped>
.input-container {
  position: relative;
}

.input-field {
  width: 100%;
  padding: 24px 12px 6px 12px;
  border: 1px solid var(--input-border-color);
  border-radius: var(--control-radius);
  font-size: 16px;
  font-family: inherit;
  background-color: var(--surface-background);
  color: var(--text-color);
  box-sizing: border-box;
  -webkit-appearance: none;
  appearance: none;
}

.input-field:focus-visible {
  border-color: var(--accent-color);
  outline-offset: 2px;
}

.input-label {
  position: absolute;
  top: 4px;
  left: 12px;
  font-size: 12px;
  color: var(--muted-text-color);
  text-transform: capitalize;
}

.input-label--optional {
  font-style: italic;
}
</style>
