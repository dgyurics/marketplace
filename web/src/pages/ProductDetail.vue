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
        <p class="product-price">{{ displayPrice(product.price) }}</p>
      </div>

      <div class="product-actions">
        <div v-if="Object.entries(product.details).length > 0" class="product-details">
          <h3>Details</h3>
          <div class="details">
            <p v-for="(value, key) in product.details" :key="key">
              <b class="detail-item">{{ key }}:</b> {{ value }}
            </p>
          </div>
        </div>
        <div>
          <ProductActions :product="product" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { Navigation, Pagination } from 'swiper/modules'
import { Swiper, SwiperSlide } from 'swiper/vue'
import { ref, onMounted, reactive } from 'vue'
import { useRoute } from 'vue-router'

import ProductActions from '@/components/ProductActions.vue'
import { getProductById, getOffersByProductId } from '@/services/api'
import { useAuthStore } from '@/store/auth'
import type { Offer, Product } from '@/types'
import { displayPrice } from '@/utilities/currency'

// @ts-ignore
import 'swiper/css'
// @ts-ignore
import 'swiper/css/navigation'
// @ts-ignore
import 'swiper/css/pagination'

const route = useRoute()

const authStore = useAuthStore()
const { isAuthenticated } = storeToRefs(authStore)

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
  featured: false,
  negotiable: false,
})

const offers = ref<Offer[]>([])

onMounted(async () => {
  try {
    const productData = await getProductById(String(route.params['id']))
    productData.images = productData.images.filter((img) => img.type === 'gallery')
    Object.assign(product, productData)

    // Check if existing offers exists
    if (isAuthenticated.value) {
      try {
        offers.value = await getOffersByProductId(String(route.params['id']))
      } catch (error) {
        console.error('Error fetching offers:', error)
      }
    }
  } catch (error) {
    console.error('Error fetching product:', error)
  }
})
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
  max-height: 500px;
  object-fit: contain;
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
  gap: 40px;
  margin-bottom: 50px;
}

.product-info {
  flex: 1;
  text-align: left;
  max-width: 50%;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
}

.product-actions {
  flex: 1;
  text-align: left;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  gap: 20px;
}

/* Tablet styles */
@media (max-width: 768px) {
  .product-detail-bottom {
    flex-direction: column;
    align-items: center;
    padding: 0 40px;
    gap: 15px;
  }

  .product-info {
    max-width: 100%;
    text-align: left;
    width: 100%;
  }

  .product-actions {
    width: 100%;
    align-items: flex-start;
  }
}

/* Mobile styles */
@media (max-width: 480px) {
  .product-detail {
    padding: 15px;
  }

  .product-detail-bottom {
    padding: 0 20px;
    gap: 25px;
  }

  .gallery-container {
    max-width: 100%;
    height: 350px;
    margin-bottom: 25px;
  }

  :deep(.swiper-slide) {
    height: 350px;
  }

  .gallery-image {
    max-height: 350px;
  }
}

.product-details h3 {
  margin-bottom: 20px;
}

.product-details {
  font-size: 14px;
  color: #555;
}

.product-details p,
.details p {
  margin-bottom: 8px;
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
  text-transform: uppercase;
}

.product-title,
.product-description,
.product-summary,
.product-price,
.product-details {
  font-size: 14px;
  font-family: 'Open Sans', sans-serif;
  color: #222;
}

.detail-item {
  text-transform: capitalize;
}
</style>
