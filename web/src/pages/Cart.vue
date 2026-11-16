<template>
  <div class="container checkout-page checkout-page--wide">
    <header class="checkout-heading">
      <p class="checkout-eyebrow">Your finds, all in one place</p>
      <h1>Shopping cart</h1>
      <router-link class="checkout-back" to="/new">← Keep shopping</router-link>
    </header>

    <!-- Insufficient stock warning -->
    <div v-if="insufficientStock.length">
      <p v-for="item in insufficientStock" :key="item.product.id" class="error">
        <template v-if="item.inventory === 0">
          {{ item.product.name }} is no longer in stock</template
        >
        <template v-else>
          Only {{ item.inventory }} of {{ item.product.name }} left; your cart has been updated
        </template>
      </p>
    </div>

    <!-- Display a message if the cart is empty -->
    <div v-if="items.length === 0" class="checkout-panel empty-cart">
      <h2>Your cart is empty</h2>
      <router-link class="checkout-link-button" to="/new">Explore new finds</router-link>
    </div>

    <!-- Display the cart items -->
    <div v-else class="checkout-layout">
      <ul class="checkout-panel cart-list">
        <li v-for="item in items" :key="item.product.id" class="cart-item">
          <div class="item-info">
            <ThumbnailImage :images="item.product.images" />
            <div class="item-text">
              <span class="item-name">{{ item.product.name }}</span>
              <p v-if="item.product.summary" class="item-summary">{{ item.product.summary }}</p>
            </div>
          </div>
          <div class="item-actions">
            <div class="item-details">
              <span class="item-price">{{ formatPrice(item.unit_price) }}</span>
              <span class="item-quantity">Qty: {{ item.quantity }}</span>
            </div>
            <button
              class="remove-button"
              :aria-label="`Remove ${item.product.name} from cart`"
              :tabindex="0"
              @click="removeFromCart(item.product.id)"
            >
              <TrashIcon class="remove-icon" />
            </button>
          </div>
        </li>
      </ul>

      <!-- Display the subtotal and checkout button -->
      <aside class="checkout-panel cart-summary">
        <h2>Order summary</h2>
        <div class="cart-total">
          <span>Subtotal</span><strong>{{ formatPrice(subtotal) }}</strong>
        </div>
        <p class="cart-note">Tax is calculated at checkout.</p>
        <button
          v-if="items.length > 0"
          class="btn-full-width mt-15"
          :tabindex="0"
          @click="goToCheckout"
        >
          Proceed to Checkout
        </button>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import { TrashIcon } from '@heroicons/vue/24/outline'
import { storeToRefs } from 'pinia'
import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'

import ThumbnailImage from '@/components/ThumbnailImage.vue'
import { createGuestUser as apiCreateGuestUser } from '@/services/api'
import { useAuthStore } from '@/store/auth'
import { useCartStore } from '@/store/cart'
import { useCheckoutStore } from '@/store/checkout'
import type { AuthTokens } from '@/types'
import { formatPrice } from '@/utilities/currency'

const authStore = useAuthStore()
const { isAuthenticated } = storeToRefs(authStore)
const { setTokens } = authStore

const cartStore = useCartStore()
const checkoutStore = useCheckoutStore()
const router = useRouter()

const { items, subtotal } = storeToRefs(cartStore)
const { removeFromCart } = cartStore
const { insufficientStock } = storeToRefs(checkoutStore)

onMounted(async () => {
  // If the user is not authenticated, create a guest user
  if (!isAuthenticated.value) {
    try {
      const authTokens: AuthTokens = await apiCreateGuestUser()
      setTokens(authTokens)
    } catch (error) {
      console.error('Failed to create guest user:', error)
      return
    }
  }
  await cartStore.fetchCart()
})

onUnmounted(() => {
  checkoutStore.insufficientStock = []
})

const goToCheckout = () => {
  checkoutStore.insufficientStock = []
  router.push('/checkout/shipping')
}
</script>

<style scoped>
.cart-list {
  list-style: none;
  margin: 0;
}
.cart-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 22px 0;
}
.cart-item:first-child {
  padding-top: 0;
}
.cart-item:last-child {
  padding-bottom: 0;
}
.cart-item + .cart-item {
  border-top: 1px solid var(--border-color);
}
.item-info {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
  gap: 18px;
}
.item-text {
  min-width: 0;
}
.item-name {
  font-size: 15px;
  font-weight: 600;
  overflow-wrap: anywhere;
}
.item-summary {
  font-size: 12px;
  color: var(--muted-text-color);
  margin-top: 6px;
  overflow-wrap: anywhere;
}
.item-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}
.item-details {
  display: flex;
  flex-direction: column;
  gap: 4px;
  white-space: nowrap;
}
.item-price {
  font-size: 15px;
  font-weight: 600;
}
.item-quantity {
  font-size: 12px;
  color: var(--muted-text-color);
}
.remove-button {
  flex: 0;
  background: transparent;
  color: var(--disabled-text-color);
  padding: 8px;
  display: flex;
}
.remove-button:hover {
  background: var(--highlight-background);
  color: var(--accent-color);
}
.remove-icon {
  width: 20px;
  height: 20px;
}
.empty-cart {
  text-align: center;
}
.empty-cart h2 {
  text-align: center;
}
.empty-cart p {
  font-size: 14px;
  color: var(--muted-text-color);
}
.cart-summary {
  background: var(--highlight-background);
}
.cart-total {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  font-size: 15px;
  margin-top: 24px;
}
.cart-note {
  font-size: 12px;
  color: var(--muted-text-color);
  margin: 12px 0 24px;
}
@media (max-width: 600px) {
  .cart-item {
    flex-wrap: wrap;
  }
  .item-info {
    flex-basis: 100%;
  }
  .item-actions {
    width: 100%;
    justify-content: space-between;
  }
  .item-details {
    flex-direction: row;
    align-items: center;
    gap: 14px;
  }
}
</style>
