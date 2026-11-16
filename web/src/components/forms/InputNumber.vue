<template>
  <div class="input-container">
    <input
      v-model="value"
      type="number"
      :step="step"
      :required="required || false"
      class="input-field"
    />
    <span v-if="label" class="input-label" :class="{ 'input-label--optional': !required }">
      {{ label }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue: string | number | undefined
  label?: string
  required?: boolean
  step?: string | number
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const value = computed({
  get: () => String(props.modelValue),
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
}

/* Hide number input arrows/spinners */
.input-field::-webkit-outer-spin-button,
.input-field::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.input-field[type='number'] {
  -moz-appearance: textfield;
  appearance: textfield;
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
