<template>
  <div class="success-container">
    <CheckIcon :class="checkClasses" />
    <span v-if="message" class="loading-text">{{ message }}</span>
  </div>
</template>

<script setup lang="ts">
import { CheckIcon } from '@heroicons/vue/24/outline'
import { computed } from 'vue'

interface Props {
  message?: string
  size?: 'sm' | 'md' | 'lg'
}

const props = withDefaults(defineProps<Props>(), {
  message: '',
  size: 'md',
})

const checkClasses = computed(() => ['success-check', `size-${props.size}`])
</script>

<style scoped>
.success-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.success-check {
  animation: checkmark 1s ease-in-out;
  color: #ffffff;
}

/* Size variants */
.success-check.size-sm {
  width: 16px;
  height: 16px;
}

.success-check.size-md {
  width: 24px;
  height: 24px;
}

.success-check.size-lg {
  width: 32px;
  height: 32px;
}

.loading-text {
  font-size: 14px;
  color: #6b7280;
}

@keyframes checkmark {
  0% {
    opacity: 0;
    transform: scale(0.3);
  }
  50% {
    opacity: 1;
    transform: scale(1.1);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
}

/* @keyframes checkmark {
  0% {
    opacity: 0;
    transform: scale(0.8);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
} */
</style>
