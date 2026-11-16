<template>
  <div class="product-detail">
    <nav class="product-breadcrumb" aria-label="Breadcrumb">
      <router-link to="/">Marketplace</router-link>
      <span aria-hidden="true">/</span>
      <span>{{ product.name || 'Item details' }}</span>
    </nav>

    <div class="product-layout">
      <div class="gallery-container">
        <Swiper
          v-if="product.images.length"
          :modules="[Navigation, Pagination, Keyboard]"
          :navigation="product.images.length > 1"
          :pagination="{ clickable: true }"
          :keyboard="true"
          class="product-gallery"
        >
          <SwiperSlide v-for="img in product.images" :key="img.id">
            <img :src="img.url" :alt="product.name" class="gallery-image" />
          </SwiperSlide>
        </Swiper>
        <p v-else class="image-placeholder">No photo available</p>
      </div>

      <div class="product-info">
        <h1 class="product-title">{{ product.name }}</h1>
        <p v-if="product.summary" class="product-summary">{{ product.summary }}</p>
        <p v-if="product.id" class="product-price">{{ displayPrice(product.price) }}</p>

        <div class="purchase-panel">
          <ProductActions
            v-if="product.id"
            :product="product"
            :has-pending-offer="offers.some((offer) => offer.status === 'pending')"
          />
        </div>

        <section v-if="product.description" class="product-section">
          <h2>About this item</h2>
          <div class="product-description" v-html="product.description"></div>
        </section>

        <section v-if="Object.entries(product.details).length > 0" class="product-section">
          <h2>Item details</h2>
          <dl class="details">
            <div v-for="(value, key) in product.details" :key="key" class="detail-row">
              <dt>{{ key }}</dt>
              <dd>{{ value }}</dd>
            </div>
          </dl>
        </section>
        <div class="share-row"><CopyToClipboard /></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { Keyboard, Navigation, Pagination } from 'swiper/modules'
import { Swiper, SwiperSlide } from 'swiper/vue'
import { ref, onMounted, reactive } from 'vue'
import { useRoute } from 'vue-router'

import CopyToClipboard from '@/components/CopyToClipboard.vue'
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
    const galleryImages = productData.images.filter((img) => img.type === 'gallery')
    productData.images = galleryImages.length
      ? galleryImages
      : productData.images.filter((img) => img.type === 'hero')

    // Load existing offers before displaying product actions.
    if (isAuthenticated.value && authStore.hasMinimumRole('member') && productData.negotiable) {
      try {
        offers.value = await getOffersByProductId(String(route.params['id']))
      } catch (error) {
        console.error('Error fetching offers:', error)
      }
    }
    Object.assign(product, productData)
  } catch (error) {
    console.error('Error fetching product:', error)
  }
})
</script>

<style scoped>
.product-detail {
  width: 100%;
  max-width: 1280px;
  margin: 0 auto;
  padding: 28px 32px 64px;
}
.product-breadcrumb {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 28px;
  font-size: 12px;
  color: var(--muted-text-color);
}
.product-breadcrumb > span:last-child {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.product-breadcrumb a {
  flex-shrink: 0;
}
.product-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(0, 1fr);
  align-items: start;
  gap: 56px;
}
.gallery-container {
  min-width: 0;
  background: var(--image-background);
  border-radius: var(--panel-radius);
  overflow: hidden;
}
.product-gallery {
  width: 100%;
  --swiper-navigation-size: 20px;
  --swiper-navigation-color: var(--accent-color);
  --swiper-pagination-color: var(--accent-color);
}
:deep(.swiper-slide) {
  display: flex;
  align-items: center;
  justify-content: center;
  aspect-ratio: 1;
  height: auto;
  padding: 36px;
}
.gallery-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.image-placeholder {
  display: grid;
  place-items: center;
  aspect-ratio: 1;
  color: var(--muted-text-color);
  font-size: 14px;
}
.product-info {
  min-width: 0;
}
.product-title {
  font-size: clamp(26px, 3vw, 36px);
  font-weight: 700;
  letter-spacing: -1px;
  line-height: 1.25;
  overflow-wrap: anywhere;
}
.product-summary {
  font-size: 14px;
  color: var(--muted-text-color);
  line-height: 1.7;
  margin-top: 14px;
}
.product-price {
  font-size: 28px;
  font-weight: 700;
  margin-top: 22px;
}
.purchase-panel {
  margin: 24px 0 30px;
}
.product-section {
  border-top: 1px solid var(--border-color);
  padding: 24px 0;
}
.product-section h2 {
  font-size: 16px;
  font-weight: 600;
  text-align: left;
  margin-bottom: 14px;
}
.product-description {
  font-size: 14px;
  line-height: 1.8;
  color: var(--muted-text-color);
  overflow-wrap: anywhere;
}
.product-description :deep(p + p) {
  margin-top: 12px;
}
.product-description :deep(ul),
.product-description :deep(ol) {
  padding-left: 20px;
}
.product-description :deep(img) {
  max-width: 100%;
  height: auto;
}
.details {
  font-size: 13px;
}
.detail-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.4fr);
  gap: 20px;
  padding: 10px 0;
  overflow-wrap: anywhere;
}
.detail-row + .detail-row {
  border-top: 1px solid var(--subtle-border-color);
}
.detail-row dt {
  color: var(--muted-text-color);
}
.share-row {
  border-top: 1px solid var(--border-color);
  padding-top: 18px;
}
.share-row :deep(.copy-btn) {
  position: static;
  flex-direction: row;
  gap: 10px;
}
.share-row :deep(.copy-label) {
  font-size: 12px;
}
@media (max-width: 900px) {
  .product-layout {
    gap: 28px;
  }
}
@media (max-width: 768px) {
  .product-layout {
    grid-template-columns: minmax(0, 1fr);
  }
  .gallery-container {
    max-width: 560px;
    width: 100%;
    margin: 0 auto;
  }
}
@media (max-width: 600px) {
  .product-detail {
    padding: 20px 16px 40px;
  }
  .product-breadcrumb {
    margin-bottom: 20px;
  }
  :deep(.swiper-slide) {
    padding: 28px;
  }
}
</style>
