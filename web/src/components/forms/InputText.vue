<template>
  <div class="input-container">
    <input
      v-model="value"
      :type="type || 'text'"
      :required="required || false"
      :readonly="readonly || false"
      v-bind="{
        ...(pattern ? { pattern } : {}),
        ...(tabindex ? { tabindex } : {}),
      }"
      class="input-field"
      :class="{ 'input-field--readonly': readonly }"
    />
    <span v-if="label" class="input-label" :class="{ 'input-label--optional': !required }">
      {{ label }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue: string
  label: string
  required?: boolean
  pattern?: string
  type?: string
  tabindex?: number
  readonly?: boolean
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
}

.input-field--readonly {
  background-color: var(--subtle-background);
  color: var(--muted-text-color);
  cursor: not-allowed;
  pointer-events: none;
}

.input-field--readonly:focus {
  outline: none;
  border-color: var(--border-color);
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
