<template>
  <main class="app-container">
    <NavBar v-if="!isMobile && !isNotFound" />
    <div class="content" :class="{ 'home-content': route.path === '/' }">
      <router-view />
    </div>
    <Footer />
  </main>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import Footer from '@/components/Footer.vue'
import NavBar from '@/components/NavBar.vue'
const route = useRoute()
const router = useRouter()
const isMobile = ref(false)

const isNotFound = computed(() => {
  return route.matched.length === 1 && route.matched[0].path === '/:pathMatch(.*)*'
})

onMounted(() => {
  isMobile.value = window.innerWidth < 768
  if (isMobile.value) {
    router.replace('/unsupported')
  }
})
</script>

<style scoped>
.app-container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  width: 100%;
  overflow-x: hidden;
}

.content {
  flex-grow: 1; /* Takes up remaining space between navbar and footer */
  display: flex;
  flex-direction: column;
  padding-bottom: 50px; /* Adjust based on footer height */
}

.home-content {
  padding-top: 0;
  padding-bottom: 0;
}
</style>
