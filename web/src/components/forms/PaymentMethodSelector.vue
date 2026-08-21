<template>
  <div class="payment-method-selector">
    <button
      type="button"
      class="method-btn"
      :class="{ 'method-btn--active': modelValue === 'card' }"
      @click="$emit('update:modelValue', 'card')"
    >
      pay now
    </button>
    <button
      type="button"
      class="method-btn"
      :class="{ 'method-btn--active': modelValue === 'delivery' }"
      @click="$emit('update:modelValue', 'delivery')"
    >
      pay on delivery
    </button>
  </div>
</template>

<script setup lang="ts">
export type PaymentMethod = 'card' | 'delivery'

defineProps<{
  modelValue: PaymentMethod
}>()

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
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 2px;
  text-transform: uppercase;
  padding: 8px 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  background: transparent;
  color: rgba(0, 0, 0, 0.5);
  border: 1px solid rgba(0, 0, 0, 0.15);
}

.method-btn:hover {
  border-color: rgba(0, 0, 0, 0.4);
  color: #0a0a0a;
}

.method-btn--active {
  background: #0a0a0a;
  color: #fff;
  border-color: #0a0a0a;
}

.method-btn--active:hover {
  background: #333;
  border-color: #333;
  color: #fff;
}
</style>
