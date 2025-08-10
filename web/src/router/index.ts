import type { NavigationGuardNext, RouteLocationNormalized, RouteRecordRaw } from 'vue-router'
import { createRouter, createWebHistory } from 'vue-router'

import AccountSetup from '@/pages/AccountSetup.vue'
import Category from '@/pages/admin/Category.vue'
import NewImage from '@/pages/admin/NewImage.vue'
import NewProduct from '@/pages/admin/NewProduct.vue'
import Cart from '@/pages/Cart.vue'
import Home from '@/pages/Home.vue'
import LoginRegister from '@/pages/LoginRegister.vue'
import NotFound from '@/pages/NotFound.vue'
import OrderConfirmation from '@/pages/OrderConfirmation.vue'
import Payment from '@/pages/Payment.vue'
import Product from '@/pages/Product.vue'
import ProductDetails from '@/pages/ProductDetail.vue'
import ShippingAddress from '@/pages/ShippingAddress.vue'
import Unsupported from '@/pages/Unsupported.vue'
import { getCategories } from '@/services/api'
import { useAuthStore } from '@/store/auth'

async function initRoutes(): Promise<RouteRecordRaw[]> {
  const baseRoutes: RouteRecordRaw[] = [
    { path: '/', component: Home },
    { path: '/auth/update', component: AccountSetup },
    { path: '/auth', component: LoginRegister },
    { path: '/cart', component: Cart },
    { path: '/products/:id', component: ProductDetails, props: true },
    { path: '/checkout/shipping', component: ShippingAddress },
    { path: '/checkout/payment', component: Payment },
    { path: '/checkout/confirmation', component: OrderConfirmation },
    { path: '/unsupported', component: Unsupported },
    { path: '/admin/products', component: NewProduct, beforeEnter: requireAdmin },
    { path: '/admin/categories', component: Category, beforeEnter: requireAdmin },
    { path: '/admin/products/:id/images', component: NewImage, beforeEnter: requireAdmin },
  ]

  try {
    const categories = await getCategories()
    const parentCategories = categories.filter((category) => !category.parent_id)
    const categorySlugs = parentCategories.map((category) => category.slug)

    if (categorySlugs.length > 0) {
      const categoryPattern = `/:category(${categorySlugs.join('|')})`
      baseRoutes.push({
        path: categoryPattern,
        component: Product,
        props: true,
        name: 'Category',
      })
    }
  } catch (error) {
    console.error('Failed to load categories:', error)
  }

  // NotFound route should always be last
  baseRoutes.push({ path: '/:pathMatch(.*)*', component: NotFound })

  return baseRoutes
}

// Admin route guard
function requireAdmin(
  _to: RouteLocationNormalized,
  _from: RouteLocationNormalized,
  next: NavigationGuardNext
) {
  useAuthStore().isAdmin ? next() : next('/')
}

// Export function to create router (to be called after Pinia is initialized)
export async function createAppRouter() {
  const routes = await initRoutes()
  return createRouter({
    history: createWebHistory(),
    routes,
  })
}
