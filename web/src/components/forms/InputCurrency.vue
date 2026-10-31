<template>
  <div class="input-container" :class="{ 'input-container--inside': labelPosition === 'inside' }">
    <label
      :for="id"
      :class="{ 'input-label--optional': labelPosition === 'inside' && !required }"
      >{{ label }}</label
    >
    <div class="price-input-wrapper">
      <span class="price-symbol">$</span>
      <input
        :id="id"
        v-model.number="value"
        type="number"
        step="0.01"
        min="0"
        :required="required || false"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue: number | ''
  id: string
  label: string
  required?: boolean
  labelPosition?: 'above' | 'inside'
}>()

const emit = defineEmits<{
  'update:modelValue': [value: number | '']
}>()

const value = computed({
  get: () => props.modelValue,
  set: (newValue) => emit('update:modelValue', newValue),
})
</script>

<style scoped>
.input-container {
  position: relative;
  display: flex;
  flex-direction: column;
}

.input-container label {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 8px;
}

.price-input-wrapper {
  display: flex;
  align-items: center;
  border: 1px solid #ddd;
  border-radius: 4px;
  overflow: hidden;
}

.price-symbol {
  padding: 10px 8px;
  background-color: #f5f5f5;
  border-right: 1px solid #ddd;
  font-weight: 500;
  color: #333;
}

.price-input-wrapper input {
  flex: 1;
  min-width: 0;
  border: none;
  padding: 10px;
  font-size: 14px;
}

.price-input-wrapper input::-webkit-outer-spin-button,
.price-input-wrapper input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.price-input-wrapper input[type='number'] {
  appearance: textfield;
  -moz-appearance: textfield;
}

.price-input-wrapper input:focus {
  outline: none;
  background-color: #fafafa;
}
.input-container--inside label {
  position: absolute;
  top: 4px;
  left: 32px;
  z-index: 1;
  margin-bottom: 0;
  font-size: 12px;
  font-weight: 400;
  color: #666;
  text-transform: capitalize;
}

.input-container--inside .input-label--optional {
  font-style: italic;
}

.input-container--inside .price-input-wrapper {
  align-items: stretch;
  border-color: #ccc;
  border-radius: 1px;
}

.input-container--inside .price-symbol {
  display: flex;
  align-items: center;
}

.input-container--inside input {
  padding: 24px 12px 6px;
  font-size: 16px;
}
</style>
