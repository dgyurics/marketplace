<template>
  <div class="input-container">
    <select v-model="value" :required="required || false" class="input-field">
      <option value="">{{ placeholder || 'Select...' }}</option>
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
  placeholder?: string
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
  border: 1px solid #ccc;
  border-radius: 1px;
  font-size: 16px;
  box-sizing: border-box;
  background-color: white;
}

.input-label {
  position: absolute;
  top: 4px;
  left: 12px;
  font-size: 12px;
  color: #666;
  text-transform: capitalize;
}

.input-label--optional {
  font-style: italic;
}
</style>
