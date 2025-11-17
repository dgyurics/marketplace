<template>
  <div class="input-container">
    <input
      v-model="value"
      :type="type || 'text'"
      :required="required || false"
      v-bind="pattern ? { pattern } : {}"
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
  modelValue: string
  label: string
  required?: boolean
  pattern?: string
  type?: string
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
