<template>
  <footer :class="['footer', { 'transparent-footer': route.path === '/' }]">
    <p>© {{ currentYear }} Marketplace. All rights reserved.</p>
  </footer>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
const currentYear = new Date().getFullYear()
</script>

<style scoped>
.footer {
  z-index: 1000; /* Ensure it's above other content */
  width: 100%; /* Changed from 100vw */
  box-sizing: border-box; /* Ensures padding doesn’t add extra width */
  padding: 15px 0;
  text-align: center;
  font-size: 14px;
  color: #000;
  bottom: 0;
  left: 0;
  transition: background 0.3s ease-in-out;
}

.transparent-footer {
  background: transparent; /* Removes background on Home */
  color: #fff; /* Changes text color on Home */
  position: fixed;
}
</style>
