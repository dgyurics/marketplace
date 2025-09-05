<template>
  <footer class="footer">
    <p>
      © Open Marketplace Project
      <BuildingStorefrontIcon class="footer-icon" />
      <span class="separator">|</span>
      <a href="https://github.com/dgyurics/marketplace">View Source</a>
    </p>
  </footer>
</template>

<script setup lang="ts">
import { BuildingStorefrontIcon } from '@heroicons/vue/24/outline'
</script>

<style scoped>
.footer {
  width: 100%;
  box-sizing: border-box;
  padding: 15px 0;
  text-align: center;
  font-size: 14px;
  background: #fff;
  color: #000;
  border-top: 1px solid rgba(0, 0, 0, 0.1);
  margin-top: auto;
  z-index: 1000;
  position: relative;
}

.separator {
  margin: 0 1rem;
  opacity: 0.5;
}

.footer-icon {
  width: 16px;
  height: 16px;
  display: inline-block;
  vertical-align: middle;
  margin-left: 0.5rem;
  stroke-width: 1.5;
}

a {
  color: #000;
  text-decoration: none;
  transition: color 0.3s ease;
}

a:hover {
  color: #333;
}
</style>
