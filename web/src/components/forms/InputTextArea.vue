<template>
  <div class="input-container">
    <textarea
      v-model="value"
      :required="required || false"
      :disabled="disabled || false"
      :placeholder="placeholder"
      :rows="rows"
      :maxlength="maxlength"
      class="input-field"
      :class="{ 'input-field--not-resizable': resizable === false }"
    ></textarea>
    <span v-if="label" class="input-label" :class="{ 'input-label--optional': !required }">
      {{ label }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue: string
  label?: string
  required?: boolean
  resizable?: boolean
  disabled?: boolean
  placeholder?: string
  rows?: number
  maxlength?: number
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
  padding: 24px 12px 12px 12px;
  border: 1px solid var(--input-border-color);
  border-radius: var(--control-radius);
  font-size: 16px;
  font-family: inherit;
  background-color: var(--surface-background);
  color: var(--text-color);
  box-sizing: border-box;
  resize: vertical;
  min-height: 80px;
}

.input-field--not-resizable {
  resize: none;
}

.input-field:disabled {
  background-color: var(--subtle-background);
  color: var(--subtle-text-color);
  cursor: not-allowed;
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
