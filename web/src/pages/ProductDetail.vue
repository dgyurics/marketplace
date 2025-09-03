<template>
  <div v-if="product" class="product-detail">
    <div class="gallery-container">
      <Swiper
        :modules="[Navigation, Pagination]"
        :navigation="true"
        :pagination="{ clickable: true }"
        class="product-gallery"
      >
        <SwiperSlide v-for="(img, index) in product.images" :key="index">
          <img :src="img.url" :alt="product.name" class="gallery-image" />
        </SwiperSlide>
      </Swiper>
    </div>

    <div class="product-detail-bottom">
      <div class="product-info">
        <h1 class="product-title">{{ product.name }}</h1>
        <p class="product-summary">{{ product.summary }}</p>
        <p class="product-description" v-html="product.description"></p>
        <p class="product-price">${{ (product.price / 100).toFixed(2) }}</p>
      </div>

      <div class="product-actions">
        <div v-if="product.details" class="product-details">
          <h3>Details</h3>
          <div class="details">
            <p v-for="(value, key) in product.details" :key="key">
              <b class="detail-item">{{ key }}:</b> {{ value }}
            </p>
          </div>
        </div>
        <button class="add-to-cart" :disabled="isOutOfStock" @click="addToCart">
          <span v-if="!addedToCart && !isOutOfStock">Add to Cart</span>
          <span v-else-if="isOutOfStock">Out of Stock</span>
          <span v-else class="checkmark-animation">&#10003;</span>
        </button>
        <p v-if="isLowStock" class="low-stock-warning">
          Only {{ product.inventory }} left in stock
        </p>
      </div>
    </div>
  </div>
  <div v-else class="loading">Loading product...</div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { Navigation, Pagination } from 'swiper/modules'
import { Swiper, SwiperSlide } from 'swiper/vue'
import { ref, onMounted, computed, reactive } from 'vue'
import { useRoute } from 'vue-router'

import { getProductById, createGuestUser as apiCreateGuestUser } from '@/services/api'
import { useAuthStore } from '@/store/auth'
import { useCartStore } from '@/store/cart'
import type { AuthTokens, Product } from '@/types'

// @ts-ignore
import 'swiper/css'
// @ts-ignore
import 'swiper/css/navigation'
// @ts-ignore
import 'swiper/css/pagination'

const route = useRoute()

const authStore = useAuthStore()
const { isAuthenticated } = storeToRefs(authStore)
const { setTokens } = authStore

const cartStore = useCartStore()

const product = reactive<Product>({
  id: '',
  name: '',
  summary: '',
  description: '',
  price: 0,
  images: [],
  details: {},
  inventory: 0,
})

const addedToCart = ref(false)

const isOutOfStock = computed(() => product.inventory === 0)
const isLowStock = computed(() => product.inventory > 0 && product.inventory <= 20)

onMounted(async () => {
  try {
    const productData = await getProductById(String(route.params['id']))
    productData.images = productData.images.filter((img) => img.type === 'gallery')
    Object.assign(product, productData)
  } catch (error) {
    console.error('Error fetching product:', error)
  }
})

const addToCart = async () => {
  try {
    // If the user is not authenticated, create a guest user
    if (!isAuthenticated.value) {
      console.debug('User is not authenticated, creating guest user...')
      const authTokens: AuthTokens = await apiCreateGuestUser()
      console.debug('Guest user created, setting tokens:', authTokens)
      setTokens(authTokens)
    }

    await cartStore.addToCart(product.id, 1)
    addedToCart.value = true
    setTimeout(() => {
      addedToCart.value = false
    }, 1000)
  } catch (error) {
    console.error('Failed to add item to cart:', error)
  }
}
</script>

<style scoped>
/* Swiper Pagination (bubbles) */
:deep(.swiper-pagination-bullet) {
  background-color: black !important;
}

/* Ensure navigation arrows are visible and properly styled */
:deep(.swiper-button-prev),
:deep(.swiper-button-next) {
  color: black !important;
}

:deep(.swiper-button-prev) {
  left: 10px !important;
}

:deep(.swiper-button-next) {
  right: 10px !important;
}

/* Ensure gallery container maintains consistent height */
.gallery-container {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  max-width: 600px;
  height: 500px; /* Ensures the gallery container has a consistent height */
  margin-bottom: 35px;
}

/* Ensure all images maintain consistent height */
.gallery-image {
  width: auto;
  max-width: 100%;
  height: 100%;
  min-height: 100%;
  max-height: 500px;
  object-fit: contain;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Ensure slides maintain consistent height */
:deep(.swiper-slide) {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 500px;
}

/* Prevent images from floating to the top */
.product-detail {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 20px;
}

.product-detail-bottom {
  display: flex;
  justify-content: space-between; /* Ensures better spacing */
  align-items: flex-start;
  width: 100%;
  max-width: 1200px; /* Increase max width */
  padding: 0 80px; /* Increased padding for better positioning */
}

.product-info {
  flex: 1;
  text-align: left;
  max-width: 50%; /* Moves section slightly toward the center */
  display: flex;
  flex-direction: column;
  justify-content: center; /* Vertically center content */
}

.product-actions {
  flex: 1;
  text-align: left;
  display: flex;
  flex-direction: column;
  align-items: center; /* Centers content horizontally */
}

.add-to-cart {
  background-color: black;
  color: white;
  padding: 10px 20px;
  font-size: 16px;
  border: none;
  cursor: pointer;
  margin-top: 20px; /* Added margin for spacing */
  min-width: 160px; /* Added min-width for consistent width */
}

.product-details h3 {
  margin-bottom: 20px;
}

.product-details {
  font-size: 14px;
  color: #555;
}

.product-details p {
  margin-bottom: 8px; /* Add light space between detail rows */
}

.product-detail-info {
  width: 100%;
  max-width: 600px;
}

.product-title {
  text-transform: uppercase;
  letter-spacing: 2px;
  margin-bottom: 20px;
}

.product-summary,
.product-description {
  margin-bottom: 20px;
}

.product-price {
  margin-bottom: 30px;
}

.product-title,
.product-description,
.product-summary,
.product-price,
.product-details {
  font-size: 14px; /* Match details section */
  font-weight: normal;
  font-family: 'Arial', sans-serif;
  color: #222;
}

.checkmark-animation {
  display: inline-block;
  animation: scaleIn 0.4s ease-in-out;
}

.low-stock-warning {
  font-size: 12px;
  color: #c00;
  margin-top: 8px;
}

.detail-item {
  text-transform: capitalize;
}

.details p {
  margin-bottom: 8px;
}

@keyframes scaleIn {
  0% {
    transform: scale(0);
    opacity: 0;
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}
</style>
