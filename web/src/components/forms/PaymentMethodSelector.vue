<template>
  <div class="payment-method-selector">
    <button
      type="button"
      class="method-btn"
      :class="{ 'method-btn--active': modelValue === 'stripe' }"
      :disabled="!enabled.includes('stripe')"
      @click="$emit('update:modelValue', 'stripe')"
    >
      pay now
    </button>
    <button
      type="button"
      class="method-btn"
      :class="{ 'method-btn--active': modelValue === 'on_delivery' }"
      :disabled="!enabled.includes('on_delivery')"
      @click="$emit('update:modelValue', 'on_delivery')"
    >
      pay on delivery
    </button>
  </div>
</template>

<script setup lang="ts">
import type { PaymentMethod } from '@/types'

withDefaults(
  defineProps<{
    modelValue: PaymentMethod | null
    enabled?: PaymentMethod[]
  }>(),
  { enabled: () => ['stripe', 'on_delivery'] }
)

defineEmits<{
  'update:modelValue': [value: PaymentMethod]
}>()
</script>

<style scoped>
.payment-method-selector {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.method-btn {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0;
  text-transform: capitalize;
  padding: 14px 16px;
  border-radius: var(--control-radius);
  cursor: pointer;
  transition: all 0.3s ease;
  background: transparent;
  color: var(--muted-text-color);
  border: 1px solid var(--input-border-color);
}

.method-btn:hover:not(:disabled) {
  border-color: var(--accent-color);
  color: var(--text-color);
}

.method-btn--active {
  background: var(--accent-color);
  color: var(--on-accent-color);
  border-color: var(--accent-color);
}

.method-btn--active:hover:not(:disabled) {
  background: var(--accent-hover-color);
  border-color: var(--accent-hover-color);
  color: var(--on-accent-color);
}

.method-btn:disabled {
  background: var(--disabled-background);
  color: var(--disabled-text-color);
  border-color: var(--border-color);
}
</style>
