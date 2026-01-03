<template>
  <div class="unsupported">
    <h1>Unsupported Device</h1>
    <p>Mobile devices not yet supported</p>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

onMounted(() => {
  const isStillSmall = window.innerWidth < 768
  if (!isStillSmall) {
    router.replace('/')
  }
})
</script>

<style scoped>
.unsupported {
  text-align: center;
  padding: 4rem 2rem;
}
</style>
