<template>
  <div class="input-container">
    <button
      type="button"
      class="step-btn step-btn--left"
      aria-label="decrease value"
      @click="decrement"
    >
      −
    </button>

    <input
      :value="displayValue"
      type="text"
      inputmode="numeric"
      :required="required || false"
      class="input-field"
      :class="{ 'input-field--with-prefix': !!prefix }"
      @input="handleInput"
      @keydown="handleKeydown"
    />

    <button
      type="button"
      class="step-btn step-btn--right"
      aria-label="increase value"
      @click="increment"
    >
      +
    </button>

    <span v-if="label" class="input-label" :class="{ 'input-label--optional': !required }">
      {{ label }}
    </span>
    <span v-if="prefix" class="input-prefix" aria-hidden="true">{{ prefix }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue: string | number | undefined
  label?: string
  required?: boolean
  prefix?: string
  step?: number | string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const displayValue = computed(() => String(props.modelValue ?? ''))

const stepSize = computed(() => {
  const parsed = Number.parseFloat(String(props.step ?? ''))
  if (Number.isFinite(parsed) && parsed > 0) return parsed
  return 1
})

const toNumber = (value: string | number | undefined): number => {
  const parsed = Number.parseInt(String(value ?? '0'), 10)
  return Number.isFinite(parsed) ? parsed : 0
}

const emitFromNumber = (value: number) => {
  const safe = Math.max(0, value)
  emit('update:modelValue', String(Math.trunc(safe)))
}

const handleInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', target.value.replace(/\D/g, ''))
}

const increment = () => {
  emitFromNumber(toNumber(props.modelValue) + stepSize.value)
}

const decrement = () => {
  emitFromNumber(toNumber(props.modelValue) - stepSize.value)
}

const handleKeydown = (event: { key: string; preventDefault: () => void }) => {
  if (event.key === 'ArrowUp') {
    event.preventDefault()
    increment()
    return
  }

  if (event.key === 'ArrowDown') {
    event.preventDefault()
    decrement()
  }
}
</script>

<style scoped>
.input-container {
  position: relative;
}

.input-field {
  width: 100%;
  padding: 24px 44px 6px 44px;
  border: 1px solid #ccc;
  border-radius: 1px;
  font-size: 16px;
  font-weight: 400;
  text-align: center;
  font-variant-numeric: tabular-nums;
  box-sizing: border-box;
}

.input-field:focus {
  outline: none;
  border-color: #ccc;
}

.input-field--with-prefix {
  padding-left: 70px;
}

.input-label {
  position: absolute;
  top: 4px;
  left: 44px;
  font-size: 12px;
  color: #666;
  text-transform: capitalize;
}

.input-label--optional {
  font-style: italic;
}

.input-prefix {
  position: absolute;
  left: 37px;
  top: 0;
  bottom: 0;
  width: 24px;
  display: inline-flex;
  align-items: flex-end;
  justify-content: center;
  padding-bottom: 6px;
  color: #666;
  font-size: 14px;
  font-weight: 500;
}

.step-btn {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 36px;
  border: 0;
  border-top: 1px solid #ccc;
  border-bottom: 1px solid #ccc;
  border-radius: 0;
  background: #f5f5f5;
  color: #333;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 500;
  line-height: 1;
  cursor: pointer;
}

.step-btn:hover {
  background: #ebebeb;
}

.step-btn--left {
  left: 1px;
  border-right: 1px solid #ddd;
}

.step-btn--right {
  right: 1px;
  border-left: 1px solid #ddd;
}
</style>
