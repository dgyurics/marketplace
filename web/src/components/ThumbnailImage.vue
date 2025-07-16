<template>
  <div class="thumbnail-container">
    <img v-if="thumbnail" :src="thumbnail.url" :altText="altText" class="thumbnail-image" />
  </div>
</template>

<script setup lang="ts">
import type { Image } from '@/types'

const props = defineProps<{
  images: Image[]
  altText?: string
}>()

const thumbnail = props.images.find((img) => img.type === 'thumbnail')
</script>

<style scoped>
.thumbnail-container {
  width: 150px;
  height: auto; /* Allow flexibility */
  max-height: 150px;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden; /* Prevent unexpected stretching */
}

.thumbnail-image {
  width: 100%;
  height: 100%;
  object-fit: contain; /* Prevent cropping */
  border-radius: 5px;
}
</style>
