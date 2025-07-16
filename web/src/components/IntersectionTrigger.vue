<template>
  <div ref="trigger" style="height: 1px"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const emit = defineEmits(['intersect'])
const trigger = ref<HTMLElement | null>(null)
let observer: IntersectionObserver | null = null

onMounted(() => {
  if (!trigger.value) return

  observer = new IntersectionObserver(([entry]) => {
    if (entry.isIntersecting) emit('intersect')
  })
  observer.observe(trigger.value)
})

onUnmounted(() => {
  if (observer && trigger.value) {
    observer.unobserve(trigger.value)
    observer.disconnect()
  }
})
</script>
