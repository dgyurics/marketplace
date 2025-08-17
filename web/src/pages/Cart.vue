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
        <div class="item-details">
          <span class="item-price">${{ (item.unit_price / 100).toFixed(2) }}</span>
          <span class="item-quantity">Qty: {{ item.quantity }}</span>
        </div>
        <button class="remove-button" @click="removeFromCart(item.product.id)">
          <XMarkIcon class="remove-icon" />
        </button>
      </li>
    </ul>

    <!-- Display the subtotal and checkout button -->
    <div v-if="items.length > 0" class="cart-total">
      <strong>Subtotal:</strong> ${{ (subtotal / 100).toFixed(2) }}
    </div>
    <button v-if="items.length > 0" class="checkout-button" @click="goToCheckout">
      Proceed to Checkout
    </button>
  </div>
</template>

<script setup lang="ts">
import { XMarkIcon } from '@heroicons/vue/24/outline'
import { storeToRefs } from 'pinia'
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

import ThumbnailImage from '@/components/ThumbnailImage.vue'
import { createGuestUser as apiCreateGuestUser } from '@/services/api'
import { useAuthStore } from '@/store/auth'
import { useCartStore } from '@/store/cart'
import type { AuthTokens } from '@/types'

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
  padding: 0;
  font-size: 14px;
  cursor: pointer;
  border-radius: 5px;
  transition: opacity 0.3s;
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
}

/* Item Info */
.item-info {
  display: flex;
  align-items: center;
  flex: 3;
  min-width: 0; /* Prevents text from shrinking */
}

/* Thumbnail Image */
.item-image img {
  width: 220px;
  height: 220px;
  object-fit: cover;
  border-radius: 5px;
}

.item-text {
  display: flex;
  flex-direction: column;
  justify-content: center;
  margin-left: 20px;
  max-width: 500px; /* Allows space for a long summary */
}

.item-name {
  font-size: 18px;
  color: #333;
  font-weight: 500;
  white-space: nowrap; /* Prevents stacking */
  overflow: hidden;
  text-overflow: ellipsis;
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
  margin: 0 180px; /* Ensures wide spacing */
}

/* Price & Quantity */
.item-details {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  flex: 1;
  min-width: 120px; /* Prevents shrinking */
  font-family: 'Inter', sans-serif;
  font-size: 14px;
}

.item-price {
  font-weight: 500;
}

.item-quantity {
  color: #444;
  margin-top: 5px;
}

.checkout-button {
  width: 100%;
  padding: 10px;
  background-color: black;
  color: white;
  border: none;
  font-size: 16px;
  cursor: pointer;
  margin-top: 15px;
}

.checkout-button:hover {
  background-color: #333;
}

.cart-total {
  text-align: right;
  font-size: 14px;
  font-weight: 500;
  font-family: 'Inter', sans-serif;
  color: #444;
  margin-top: 20px;
  margin-bottom: 10px;
}
</style>
