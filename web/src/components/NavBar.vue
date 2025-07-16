<template>
  <nav :class="['navbar', { 'transparent-navbar': route.path === '/' }]">
    <div class="logo">
      <router-link to="/">marketplace</router-link>
    </div>
    <div class="nav-container">
      <div class="nav-side"></div>
      <!-- empty space to balance nav-icons -->
      <div class="nav-links">
        <router-link v-for="category in categories" :key="category.id" :to="`/${category.slug}`">
          <span class="nav-text">{{ category.name }}</span>
        </router-link>
      </div>
      <div class="nav-icons">
        <router-link to="/cart">
          <ShoppingCartIcon class="icon" />
        </router-link>
        <router-link to="/auth">
          <UserIcon class="icon" />
        </router-link>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { ShoppingCartIcon, UserIcon } from '@heroicons/vue/24/outline'
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { getCategories } from '@/services/api'
import type { Category } from '@/types'
const route = useRoute()

const categories = ref<Category[]>([])

onMounted(async () => {
  // Fetch categories from API
  try {
    categories.value = (await getCategories()).filter((category) => !category.parent_id)
  } catch (error) {
    console.error('Failed to fetch categories:', error)
  }
})
</script>

<style scoped>
nav {
  position: relative;
  top: 0;
  left: 0;
  width: 100vw;
  z-index: 1000;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 30px 0;
  background: #fff;
  color: #000;
  box-sizing: border-box;
}

/* Centered Logo */
.logo {
  font-size: 2.5rem;
  font-weight: 400;
  font-family: 'Playfair Display', serif;
  text-transform: lowercase;
  letter-spacing: 1px;
  text-align: center; /* Ensures full centering */
  padding-bottom: 15px; /* Extra spacing below logo */
}

.logo a {
  color: #000;
  text-decoration: none;
  transition: color 0.3s ease;
}

.logo a:hover {
  color: #555;
}

/* Nav Container */
.nav-container {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  width: 90%;
  max-width: 1200px;
}

.nav-side {
  height: 100%; /* just acts as left-side spacing */
}

.nav-text {
  text-transform: capitalize;
}

.nav-links {
  display: flex;
  gap: 20px;
  justify-content: center;
}

.nav-icons {
  display: flex;
  gap: 15px;
  justify-content: flex-end;
}

.nav-icons .icon {
  width: 24px;
  height: 24px;
  cursor: pointer;
  transition: color 0.3s ease;
}

.nav-icons .icon:hover {
  color: #555;
}

.transparent-navbar {
  background: transparent;
  color: #fff;
}

.transparent-navbar .logo a,
.transparent-navbar .nav-links a,
.transparent-navbar .nav-icons .icon {
  color: #fff;
}

.transparent-navbar .logo a:hover,
.transparent-navbar .nav-links a:hover,
.transparent-navbar .nav-icons .icon:hover {
  color: #ddd;
}
</style>
