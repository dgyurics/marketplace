<template>
  <nav class="navbar">
    <div class="nav-container">
      <!-- Left: Logo -->
      <div class="nav-left">
        <div class="logo">
          <router-link to="/">marketplace</router-link>
        </div>
      </div>

      <!-- Center: Navigation Links -->
      <div class="nav-center">
        <div class="nav-links">
          <router-link v-for="category in categories" :key="category.id" :to="`/${category.slug}`">
            <span class="nav-text">{{ category.name }}</span>
          </router-link>
        </div>
      </div>

      <!-- Right: Icons -->
      <div class="nav-right">
        <div class="nav-icons">
          <router-link to="/cart">
            <ShoppingBagIcon class="icon" />
          </router-link>
          <router-link to="/auth">
            <UserIcon class="icon" />
          </router-link>
        </div>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { ShoppingBagIcon, UserIcon } from '@heroicons/vue/24/outline'
import { onMounted, ref } from 'vue'

import { getCategories } from '@/services/api'
import type { Category } from '@/types'

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
  background: #fff;
  color: #000;
  box-sizing: border-box;
  border-bottom: 1px solid rgba(0, 0, 0, 0.1);
}

/* Main Nav Container */
.nav-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 2rem;
  height: 70px;
  position: relative;
}

/* Left Section - Logo */
.nav-left {
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.logo {
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-size: 1rem;
  letter-spacing: 0.05em;
  color: #000;
}

.logo a {
  color: #000;
  text-decoration: none;
  transition: color 0.3s ease;
  position: relative;
}

/* Center Section - Navigation Links */
.nav-center {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  justify-content: center;
}

.nav-links {
  display: flex;
  gap: 2.5rem;
  justify-content: center;
}

.nav-text {
  text-transform: uppercase;
  font-size: 0.85rem;
  letter-spacing: 0.05em;
}

.nav-links a {
  color: #000;
  text-decoration: none;
  transition: color 0.3s ease;
  position: relative;
}

.nav-links a::after {
  content: '';
  position: absolute;
  bottom: -4px;
  left: 0;
  width: 0;
  height: 1px;
  background-color: #000;
  transition: width 0.3s ease;
}

.nav-links a:hover::after {
  width: 100%;
}

/* Right Section - Icons */
.nav-right {
  display: flex;
  justify-content: flex-end;
}

.nav-icons {
  display: flex;
  gap: 1.5rem;
  justify-content: flex-end;
}

.nav-icons .icon {
  width: 22px;
  height: 22px;
  cursor: pointer;
  transition: color 0.3s ease;
  color: #000;
  stroke-width: 1.5;
}
</style>
