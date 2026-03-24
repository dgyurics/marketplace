<template>
  <main class="app-container">
    <Banner v-if="TEST_MODE" :message="bannerMessage" />
    <NavBar v-if="!isNotFound" />
    <div class="content" :class="{ 'home-content': route.path === '/' }">
      <router-view />
    </div>
    <Footer />
  </main>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import Banner from '@/components/Banner.vue'
import Footer from '@/components/Footer.vue'
import NavBar from '@/components/NavBar.vue'
import { TEST_MODE } from '@/config'

const route = useRoute()

const isNotFound = computed(() => {
  return route.matched.length === 1 && route.matched[0].path === '/:pathMatch(.*)*'
})

const bannerMessage = 'DEMO ONLY - Products not for sale'
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
}

.home-content {
  padding-top: 0;
}
</style>
