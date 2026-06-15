<template>
  <div class="button-group menu">
    <button :tabindex="0" @click="goToProfile">Profile</button>
    <template v-if="authStore.hasMinimumRole('staff')">
      <button :tabindex="0" @click="goToCategories">Categories</button>
      <button :tabindex="0" @click="goToProducts">Products</button>
      <button :tabindex="0" @click="goToOrders">Orders</button>
      <button :tabindex="0" @click="goToOffers">Offers</button>
      <button :tabindex="0" @click="goToUsers">Users</button>
      <button :tabindex="0" @click="goToShippingZones">Shipping</button>
    </template>
    <button :tabindex="0" @click="handleLogout">Logout</button>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

import { logout as apiLogout } from '@/services/api'
import { useAuthStore } from '@/store/auth'

const authStore = useAuthStore()
const router = useRouter()

const goToProfile = () => router.push('/profile')
const goToCategories = () => router.push('/admin/categories')
const goToProducts = () => router.push('/admin/products')
const goToOrders = () => router.push('/admin/orders')
const goToOffers = () => router.push('/admin/offers')
const goToUsers = () => router.push('/admin/users')
const goToShippingZones = () => router.push('/admin/shipping-zones')

const handleLogout = async () => {
  try {
    await apiLogout()
    authStore.clearTokens()
  } catch {
    // logout failed silently
  }
}
</script>

<style scoped>
.button-group {
  display: flex;
  justify-content: space-between;
  gap: 8px;
}

.button-group.menu {
  flex-direction: column;
  align-items: center;
  max-width: 200px;
  margin: 30px auto 0;
}

.button-group.menu button {
  width: 100%;
}
</style>
