<template>
  <div class="product-action-buttons">
    <div class="top-row">
      <button class="btn-lg" :tabindex="0" @click="() => {}">Make an Offer</button>
      <button
        class="btn-lg"
        :disabled="isOutOfStock || hasReachedCartLimit"
        :tabindex="0"
        @click="handleAddToCart"
      >
        {{ addToCartLabel }}
      </button>
    </div>
    <button class="btn-lg btn-full-width btn-outline" :tabindex="0" @click="() => {}">
      Buy Now
    </button>
    <p v-if="showLowStockWarning" class="low-stock-warning">
      Only {{ product.inventory }} left in stock
    </p>
    <p v-else-if="hasReachedCartLimit" class="limit-reached-warning">
      Limit {{ product.cart_limit }} per customer
    </p>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed } from 'vue'

import { createGuestUser as apiCreateGuestUser } from '@/services/api'
import { useAuthStore } from '@/store/auth'
import { useCartStore } from '@/store/cart'
import type { AuthTokens, Product } from '@/types'

const props = defineProps<{
  product: Product
}>()

const authStore = useAuthStore()
const { isAuthenticated } = storeToRefs(authStore)
const { setTokens } = authStore
const cartStore = useCartStore()

const cartLimit = computed(() => props.product.cart_limit ?? 0)
const cartQuantity = computed(() => cartStore.itemCountByProductId(props.product.id))

const showLowStockWarning = computed(
  () => props.product.inventory > 0 && props.product.inventory <= 20
)
const hasReachedCartLimit = computed(
  () => cartLimit.value > 0 && cartQuantity.value >= cartLimit.value
)
const isOutOfStock = computed(() => cartQuantity.value >= props.product.inventory)
const addToCartLabel = computed(() => (isOutOfStock.value ? 'Out of Stock' : 'Add to Cart'))

const handleAddToCart = async () => {
  try {
    if (!isAuthenticated.value) {
      const authTokens: AuthTokens = await apiCreateGuestUser()
      setTokens(authTokens)
    }

    // TODO: show toast notification when item is added to cart
    await cartStore.addToCart(props.product.id, 1)
  } catch (error: unknown) {
    const err = error as { response?: { status?: number } }
    if (err.response?.status === 409) {
      // eslint-disable-next-line vue/no-mutating-props
      props.product.inventory = 0
    }
  }
}
</script>

<style scoped>
.product-action-buttons {
  display: flex;
  flex-direction: column;
  gap: 14px;
  width: 100%;
}

.top-row {
  display: flex;
  gap: 14px;
}

.top-row button {
  flex: 1;
}

.product-action-buttons button {
  font-size: 15px;
}

.low-stock-warning,
.limit-reached-warning {
  text-align: center;
  font-size: 12px;
  color: #c00;
  margin-top: 8px;
}
</style>
