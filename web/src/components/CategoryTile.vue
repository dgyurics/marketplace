<template>
  <div class="category-tile" :tabindex="0" @click="handleClick">
    <div class="category-info">
      <h3 class="category-title">{{ category.name }}</h3>
      <p v-if="category.description" class="category-description">{{ category.description }}</p>
      <p class="category-slug">{{ category.slug }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Category } from '@/types'

defineProps<{ category: Category }>()
const emit = defineEmits<{
  click: []
}>()

const handleClick = () => {
  emit('click')
}
</script>

<style scoped>
.category-tile {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 20px;
  border: var(--tile-border);
  box-shadow: var(--tile-shadow);
  border-radius: var(--tile-radius);
  background-color: var(--surface-background);
  cursor: pointer;
  transition:
    transform 0.2s ease-in-out,
    box-shadow 0.2s ease-in-out;
}

.category-tile:hover {
  opacity: 1;
}

.category-tile:hover {
  transform: scale(1.02);
  box-shadow: var(--tile-hover-shadow);
}

.category-info {
  width: 100%;
}

.category-title {
  font-size: 16px;
  font-weight: 500;
  letter-spacing: 2px;
  text-transform: uppercase;
  margin-bottom: 5px;
}

.category-slug {
  font-size: 11px;
  font-weight: 700;
  color: var(--text-color);
  margin: 5px 0;
  letter-spacing: 0.5px;
}

.category-description {
  font-size: 14px;
  color: var(--muted-text-color);
  margin-top: 5px;
  padding: 0 10px;
}
</style>
