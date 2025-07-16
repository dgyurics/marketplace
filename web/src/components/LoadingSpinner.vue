<template>
  <div class="loading-container">
    <ArrowPathIcon :class="spinnerClasses" />
    <span v-if="message" class="loading-text">{{ message }}</span>
  </div>
</template>

<script setup lang="ts">
import { ArrowPathIcon } from '@heroicons/vue/24/outline'
import { computed } from 'vue'

interface Props {
  message?: string
  size?: 'sm' | 'md' | 'lg'
}

const props = withDefaults(defineProps<Props>(), {
  message: '',
  size: 'md',
})

const spinnerClasses = computed(() => ['loading-spinner', `size-${props.size}`])
</script>

<style scoped>
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.loading-spinner {
  animation: spin 10s linear infinite;
  color: #ffffff;
}

/* Size variants */
.loading-spinner.size-sm {
  width: 16px;
  height: 16px;
}

.loading-spinner.size-md {
  width: 24px;
  height: 24px;
}

.loading-spinner.size-lg {
  width: 32px;
  height: 32px;
}

.loading-text {
  font-size: 14px;
  color: #6b7280;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
