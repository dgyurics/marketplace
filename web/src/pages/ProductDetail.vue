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
        <div>
          <button class="btn-lg" :disabled="isOutOfStock || hasReachedCartLimit" @click="addToCart">
            <span v-if="!addedToCart && !isOutOfStock">Add to Cart</span>
            <span v-else-if="isOutOfStock">Out of Stock</span>
            <span v-else class="checkmark-animation">&#10003;</span>
          </button>
          <p v-if="showLowStockWarning" class="low-stock-warning">
            Only {{ product.inventory }} left in stock
          </p>
          <p v-else-if="hasReachedCartLimit" class="limit-reached-warning">
            Limit {{ product.cart_limit }} per customer
          </p>
        </div>
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
  cart_limit: 0,
})

const addedToCart = ref(false)

const isLowStock = computed(() => product.inventory > 0 && product.inventory <= 20)
const currentQuantityInCart = computed(() => cartStore.itemCountByProductId(product.id))
const hasReachedCartLimit = computed(
  () => product.cart_limit > 0 && currentQuantityInCart.value >= product.cart_limit
)
const isOutOfStock = computed(() => currentQuantityInCart.value >= product.inventory)
const showLowStockWarning = computed(
  () => isLowStock.value && !isOutOfStock.value && !hasReachedCartLimit.value
)

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

.gallery-container {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  max-width: 600px;
  height: 500px;
  margin-bottom: 35px;
}

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
  justify-content: space-between;
  align-items: flex-start;
  width: 100%;
  max-width: 1200px;
  padding: 0 80px;
}

.product-info {
  flex: 1;
  text-align: left;
  max-width: 50%;
  display: flex;
  flex-direction: column;
  justify-content: center; /* Vertically center content */
}

.product-actions {
  flex: 1;
  text-align: left;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
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
  font-family: 'Open Sans', sans-serif;
  color: #222;
}

.checkmark-animation {
  display: inline-block;
  animation: scaleIn 0.4s ease-in-out;
}

.limit-reached-warning,
.low-stock-warning {
  text-align: center;
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
