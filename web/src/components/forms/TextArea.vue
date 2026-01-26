<template>
  <div class="input-container">
    <textarea
      v-model="value"
      :required="required || false"
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
  border: 1px solid #ccc;
  border-radius: 1px;
  font-size: 16px;
  box-sizing: border-box;
  resize: vertical;
  min-height: 80px;
}

.input-field--not-resizable {
  resize: none;
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
