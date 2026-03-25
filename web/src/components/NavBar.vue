<template>
  <nav class="navbar">
    <div class="nav-container">
      <!-- Left: Logo -->
      <div class="nav-left">
        <div class="logo">
          <router-link to="/" :tabindex="0">marketplace</router-link>
        </div>
      </div>

      <!-- Center: Navigation Links -->
      <div class="nav-center">
        <div class="nav-links">
          <router-link
            v-for="category in categories"
            :key="category.id"
            :to="`/${category.slug}`"
            :tabindex="0"
          >
            <span class="nav-text">{{ category.name }}</span>
          </router-link>
        </div>
      </div>

      <!-- Right: Icons -->
      <div class="nav-right">
        <div class="nav-icons">
          <!-- Mobile Menu Button -->
          <button :tabindex="0" class="mobile-menu-btn" @click="toggleMobileMenu">
            <Bars3Icon class="icon" />
          </button>

          <router-link to="/cart" :tabindex="0">
            <ShoppingBagIcon class="icon" />
          </router-link>
          <router-link to="/auth" :tabindex="0">
            <UserIcon class="icon" />
          </router-link>
        </div>
      </div>
    </div>

    <!-- Mobile Menu Overlay -->
    <div v-if="isMobileMenuOpen" class="mobile-menu">
      <div class="mobile-nav-links">
        <router-link
          v-for="category in categories"
          :key="category.id"
          :to="`/${category.slug}`"
          :tabindex="0"
          @click="closeMobileMenu"
        >
          {{ category.name }}
        </router-link>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { Bars3Icon, ShoppingBagIcon, UserIcon } from '@heroicons/vue/24/outline'
import { onMounted, ref } from 'vue'

import { getCategories } from '@/services/api'
import type { Category } from '@/types'

const categories = ref<Category[]>([])
const isMobileMenuOpen = ref(false)

const toggleMobileMenu = () => {
  isMobileMenuOpen.value = !isMobileMenuOpen.value
}

const closeMobileMenu = () => {
  isMobileMenuOpen.value = false
}

onMounted(async () => {
  // Fetch categories from API
  try {
    categories.value = (await getCategories()).filter((category) => !category.parent_id)
  } catch {
    // Handle error silently
    categories.value = []
  }
})
</script>

<style scoped>
nav {
  width: 100%;
  z-index: 1000;
  background: #fff;
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
}

.logo {
  text-transform: uppercase;
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
}

.nav-links {
  display: flex;
  gap: 2.5rem;
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

/* Icons */
.nav-icons {
  display: flex;
  gap: 1.5rem;
}

.icon {
  width: 22px;
  height: 22px;
  color: #000;
  stroke-width: 1.5;
  cursor: pointer;
  transition: color 0.3s ease;
}

/* Mobile Menu Button */
.mobile-menu-btn {
  display: none;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
}

/* Mobile Menu Overlay */
.mobile-menu {
  position: fixed;
  top: 70px;
  left: 0;
  width: 100%;
  height: calc(100vh - 70px);
  background: white;
  z-index: 999;
  animation: slideDown 0.3s ease;
}

@keyframes slideDown {
  from {
    transform: translateY(-100%);
  }
  to {
    transform: translateY(0);
  }
}

.mobile-nav-links {
  display: flex;
  flex-direction: column;
  padding: 30px 2rem;
  gap: 0;
}

.mobile-nav-links a {
  color: #000;
  text-decoration: none;
  padding: 20px 0;
  border-bottom: 1px solid rgba(0, 0, 0, 0.1);
  text-transform: uppercase;
  font-size: 0.9rem;
  letter-spacing: 0.05em;
  cursor: pointer;
  transition: color 0.2s ease;
}

.mobile-nav-links a:hover {
  color: #666;
}

.mobile-nav-links a:last-child {
  border-bottom: none;
}

/* Mobile Breakpoints */
@media (max-width: 768px) {
  .nav-center {
    display: none;
  }

  .mobile-menu-btn {
    display: block;
  }

  .nav-container {
    padding: 0 1rem;
  }
}

@media (max-width: 480px) {
  .logo {
    font-size: 0.9rem;
  }

  .nav-icons {
    gap: 1rem;
  }

  .nav-icons .icon {
    width: 20px;
    height: 20px;
  }

  .mobile-menu-btn .icon {
    width: 20px;
    height: 20px;
  }
}
</style>
