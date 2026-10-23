<template>
  <button type="button" class="copy-btn" @click="copyUrl">
    <span class="copy-circle">
      <CheckIcon v-if="copied" class="copy-icon" />
      <ClipboardDocumentIcon v-else class="copy-icon" />
    </span>
    <span class="copy-label">{{ copied ? 'Copied' : 'Copy link' }}</span>
  </button>
</template>

<script setup lang="ts">
import { CheckIcon, ClipboardDocumentIcon } from '@heroicons/vue/24/outline'
import { ref } from 'vue'

const copied = ref(false)

const copyUrl = async () => {
  await navigator.clipboard.writeText(window.location.href)
  copied.value = true
  setTimeout(() => {
    copied.value = false
  }, 2000)
}
</script>

<style scoped>
.copy-btn {
  position: absolute;
  bottom: 15px;
  right: 15px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 0;
  min-width: 0;
  background: none;
  color: #666;
  border: none;
  cursor: pointer;
}

.copy-circle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid #ccc;
  border-radius: 50%;
  background: #fff;
}

.copy-btn:hover {
  background: none;
  color: #000;
}

.copy-btn:hover .copy-circle {
  border-color: #000;
  background: #f0f0f0;
}

.copy-icon {
  width: 18px;
  height: 18px;
}

.copy-label {
  font-size: 11px;
  font-family: 'Open Sans', sans-serif;
}
</style>
