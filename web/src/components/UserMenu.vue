<template>
  <div class="account-menu" :class="{ 'account-menu--admin': isStaff }">
    <header class="account-header">
      <div>
        <p class="account-eyebrow">{{ isStaff ? 'Store management' : 'Marketplace' }}</p>
        <h1>{{ isStaff ? 'Your store, at a glance' : 'Your account' }}</h1>
        <p class="account-description">
          {{
            isStaff
              ? 'Everything you need to keep things running.'
              : 'Your details and conversations, in one place.'
          }}
        </p>
      </div>
      <BuildingStorefrontIcon v-if="isStaff" class="header-icon" />
      <UserCircleIcon v-else class="header-icon" />
    </header>

    <nav v-if="isStaff" class="management-grid" aria-label="Store management">
      <router-link
        v-for="item in managementLinks"
        :key="item.to"
        :to="item.to"
        class="management-card"
      >
        <component :is="item.icon" class="card-icon" />
        <span class="card-title">{{ item.title }}</span>
        <span class="card-description">{{ item.description }}</span>
        <ArrowUpRightIcon class="card-arrow" />
      </router-link>
    </nav>

    <nav class="account-links" aria-label="Your account">
      <router-link to="/profile" class="account-link">
        <UserCircleIcon class="link-icon" />
        <span><strong>Profile</strong><small>Manage your account details</small></span>
        <ChevronRightIcon class="link-arrow" />
      </router-link>
      <router-link to="/inbox" class="account-link">
        <EnvelopeIcon class="link-icon" />
        <span><strong>Inbox</strong><small>Messages and order updates</small></span>
        <ChevronRightIcon class="link-arrow" />
      </router-link>
    </nav>
    <footer class="account-footer">
      <button type="button" class="logout-button" @click="handleLogout">
        <ArrowRightOnRectangleIcon /> Log out
      </button>
    </footer>
  </div>
</template>

<script setup lang="ts">
import {
  ArrowRightOnRectangleIcon,
  ArrowUpRightIcon,
  BuildingStorefrontIcon,
  ChevronRightIcon,
  CubeIcon,
  EnvelopeIcon,
  MapPinIcon,
  RectangleStackIcon,
  ShoppingBagIcon,
  TagIcon,
  UserCircleIcon,
  UsersIcon,
} from '@heroicons/vue/24/outline'
import { computed } from 'vue'

import { logout as apiLogout } from '@/services/api'
import { useAuthStore } from '@/store/auth'
import { useCartStore } from '@/store/cart'
import { useInboxStore } from '@/store/inbox'

const authStore = useAuthStore()
const isStaff = computed(() => authStore.hasMinimumRole('staff'))
const managementLinks = [
  {
    to: '/admin/products',
    title: 'Products',
    description: 'Listings and inventory',
    icon: CubeIcon,
  },
  {
    to: '/admin/orders',
    title: 'Orders',
    description: 'Purchases and fulfillment',
    icon: ShoppingBagIcon,
  },
  { to: '/admin/offers', title: 'Offers', description: 'Review and negotiate', icon: TagIcon },
  {
    to: '/admin/categories',
    title: 'Categories',
    description: 'Organize your catalog',
    icon: RectangleStackIcon,
  },
  { to: '/admin/users', title: 'Users', description: 'Customers and permissions', icon: UsersIcon },
  {
    to: '/admin/shipping-zones',
    title: 'Shipping',
    description: 'Delivery zones and coverage',
    icon: MapPinIcon,
  },
]

const handleLogout = async () => {
  try {
    await apiLogout()
    authStore.clearTokens()
    useCartStore().clearCart()
    useInboxStore().clearInbox()
  } catch {
    // logout failed silently
  }
}
</script>

<style scoped>
.account-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 24px;
  background: var(--highlight-background);
  border-radius: var(--panel-radius);
  padding: 32px;
  margin-bottom: 24px;
}
.account-eyebrow {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 1.5px;
  color: var(--accent-text-color);
  margin-bottom: 10px;
}
.account-header h1 {
  font-size: clamp(24px, 3vw, 32px);
  font-weight: 700;
  letter-spacing: -0.8px;
  line-height: 1.25;
}
.account-description {
  font-size: 13px;
  color: var(--muted-text-color);
  margin-top: 12px;
  line-height: 1.7;
}
.header-icon {
  width: 64px;
  height: 64px;
  flex-shrink: 0;
  stroke-width: 1.2;
  color: var(--highlight-icon-color);
}
.management-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
  margin-bottom: 28px;
}
.management-card {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  padding: 24px;
  border: 1px solid var(--border-color);
  border-radius: var(--panel-radius);
  transition:
    background 0.2s,
    border-color 0.2s;
}
.management-card:hover {
  background: var(--subtle-background);
  border-color: var(--highlight-border-color);
}
.card-icon {
  width: 38px;
  height: 38px;
  padding: 8px;
  border-radius: 10px;
  background: var(--accent-background);
  color: var(--accent-color);
  margin-bottom: 18px;
}
.card-title {
  font-size: 15px;
  font-weight: 600;
}
.card-description {
  font-size: 12px;
  color: var(--muted-text-color);
  margin-top: 6px;
}
.card-arrow {
  position: absolute;
  top: 24px;
  right: 20px;
  width: 16px;
  height: 16px;
  color: var(--subtle-text-color);
}
.account-links {
  border: 1px solid var(--border-color);
  border-radius: var(--panel-radius);
}
.account-link {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 22px;
}
.account-link + .account-link {
  border-top: 1px solid var(--border-color);
}
.account-link:hover {
  background: var(--subtle-background);
}
.account-link:first-child {
  border-radius: 12px 12px 0 0;
}
.account-link:last-child {
  border-radius: 0 0 12px 12px;
}
.link-icon {
  width: 24px;
  height: 24px;
  color: var(--accent-color);
  flex-shrink: 0;
}
.account-link strong {
  display: block;
  font-size: 14px;
  font-weight: 600;
}
.account-link small {
  display: block;
  font-size: 12px;
  color: var(--muted-text-color);
  margin-top: 4px;
}
.link-arrow {
  margin-left: auto;
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  color: var(--subtle-text-color);
}
.account-menu--admin .account-links {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}
.account-menu--admin .account-link + .account-link {
  border-top: none;
  border-left: 1px solid var(--border-color);
}
.account-menu--admin .account-link:first-child {
  border-radius: 12px 0 0 12px;
}
.account-menu--admin .account-link:last-child {
  border-radius: 0 12px 12px 0;
}
.account-footer {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 16px;
  margin-top: 24px;
}
.logout-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}
.account-footer svg {
  width: 16px;
  height: 16px;
}
.logout-button {
  flex: none;
  background: transparent;
  color: var(--muted-text-color);
  padding: 10px;
  border-radius: 6px;
}
.logout-button:hover {
  background: var(--accent-background);
  color: var(--accent-color);
}
@media (max-width: 700px) {
  .management-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }
  .account-header {
    padding: 24px;
  }
  .header-icon {
    width: 44px;
    height: 44px;
  }
  .management-card {
    padding: 20px;
  }
  .account-menu--admin .account-links {
    display: block;
  }
  .account-menu--admin .account-link + .account-link {
    border-left: none;
    border-top: 1px solid var(--border-color);
  }
  .account-menu--admin .account-link:first-child {
    border-radius: 12px 12px 0 0;
  }
  .account-menu--admin .account-link:last-child {
    border-radius: 0 0 12px 12px;
  }
}
@media (max-width: 380px) {
  .management-grid {
    grid-template-columns: minmax(0, 1fr);
  }
  .header-icon {
    display: none;
  }
}
</style>
