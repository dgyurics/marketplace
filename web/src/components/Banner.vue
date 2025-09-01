<template>
  <div v-if="message" class="banner">
    <span>{{ message }}</span>
  </div>
</template>

<script setup lang="ts">
defineProps<{ message?: string }>()
</script>

<style scoped>
.banner {
  background-color: #f0f0f0;
  width: 100%;
  z-index: 999;
  text-align: center;
  color: #000;
  font-size: 12px;
  padding: 2px 0;
}
</style>
