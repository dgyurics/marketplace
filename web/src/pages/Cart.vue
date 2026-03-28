<template>
  <div class="container">
    <h2>Shopping Cart</h2>

    <!-- Display a message if the cart is empty -->
    <div v-if="items.length === 0" class="empty-cart">Your cart is empty.</div>

    <!-- Display the cart items -->
    <ul v-else>
      <li v-for="item in items" :key="item.product.id" class="cart-item">
        <div class="item-info">
          <ThumbnailImage :images="item.product.images" />
          <div class="item-text">
            <span class="item-name">{{ item.product.name }}</span>
            <p class="item-summary">{{ item.product.summary }}</p>
          </div>
        </div>
        <div class="divider"></div>
        <div class="item-actions">
          <div class="item-details">
            <span class="item-price">{{ formatPrice(item.unit_price) }}</span>
            <span class="item-quantity">Qty: {{ item.quantity }}</span>
          </div>
          <button class="remove-button" :tabindex="0" @click="removeFromCart(item.product.id)">
            <TrashIcon class="remove-icon" />
          </button>
        </div>
      </li>
    </ul>

    <!-- Display the subtotal and checkout button -->
    <div v-if="items.length > 0" class="cart-total">
      <strong>Subtotal:</strong> {{ formatPrice(subtotal) }}
    </div>
    <button
      v-if="items.length > 0"
      class="btn-full-width mt-15"
      :tabindex="0"
      @click="goToCheckout"
    >
      Proceed to Checkout
    </button>
  </div>
</template>

<script setup lang="ts">
import { TrashIcon } from '@heroicons/vue/24/outline'
import { storeToRefs } from 'pinia'
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

import ThumbnailImage from '@/components/ThumbnailImage.vue'
import { createGuestUser as apiCreateGuestUser } from '@/services/api'
import { useAuthStore } from '@/store/auth'
import { useCartStore } from '@/store/cart'
import type { AuthTokens } from '@/types'
import { formatPrice } from '@/utilities/currency'

const authStore = useAuthStore()
const { isAuthenticated } = storeToRefs(authStore)
const { setTokens } = authStore

const cartStore = useCartStore()
const router = useRouter()

const { items, subtotal } = storeToRefs(cartStore)
const { removeFromCart } = cartStore

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

const goToCheckout = () => {
  router.push('/checkout/shipping')
}
</script>

<style scoped>
.container {
  max-width: 1200px;
}

.remove-button {
  display: flex;
  flex-direction: column;
  align-items: center;
  background: transparent;
  color: inherit; /* Match text color */
  border: none;
  padding: 8px;
  font-size: 14px;
  cursor: pointer;
  border-radius: 5px;
  transition: opacity 0.3s;
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .remove-button {
    align-self: center;
  }
}

.remove-button:hover {
  opacity: 0.7;
}

.remove-icon {
  width: 24px;
  height: 24px;
}

h2 {
  text-align: center;
  font-size: 24px;
  margin-bottom: 20px;
}

.empty-cart {
  text-align: center;
  font-size: 18px;
  color: #777;
  padding: 20px;
}

/* Cart List */
ul {
  list-style: none;
  padding: 0;
}

/* Cart Item */
.cart-item {
  display: flex;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #ddd;
  justify-content: space-between;
  gap: 15px;
}

@media (max-width: 768px) {
  .cart-item {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
  }
}

/* Item Info */
.item-info {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0; /* Prevents text from shrinking */
}

@media (max-width: 768px) {
  .item-info {
    width: 100%;
  }
}

.item-text {
  display: flex;
  flex-direction: column;
  justify-content: center;
  margin-left: 20px;
  flex: 1;
}

@media (max-width: 768px) {
  .item-text {
    margin-left: 15px;
  }
}

.item-name {
  font-size: 18px;
  color: #333;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

@media (max-width: 768px) {
  .item-name {
    white-space: normal;
    font-size: 16px;
  }
}

.item-summary {
  font-size: 16px;
  color: #555;
  margin-top: 6px;
  line-height: 1.4;
  max-width: 100%;
  overflow-wrap: break-word;
  word-wrap: break-word;
}

/* Divider */
.divider {
  width: 2px;
  height: 80%;
  background-color: #ddd;
  margin: 0 15px;
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .divider {
    display: none;
  }
}

/* Price & Quantity & Actions Group */
.item-actions {
  display: flex;
  align-items: center;
  gap: 15px;
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .item-actions {
    width: 100%;
    justify-content: space-between;
    margin: 10px 0;
  }
}

/* Price & Quantity */
.item-details {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  min-width: 120px;
  font-size: 14px;
}

@media (max-width: 768px) {
  .item-details {
    flex-direction: row;
    align-items: center;
    gap: 10px;
  }
}

.item-price {
  font-weight: 500;
}

.item-quantity {
  color: #444;
  margin-top: 2px;
}

@media (max-width: 768px) {
  .item-quantity {
    margin-top: 0;
  }
}

.cart-total {
  text-align: right;
  font-size: 14px;
  color: #444;
  margin-top: 20px;
  margin-bottom: 10px;
}
</style>
