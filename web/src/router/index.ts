import type { NavigationGuardNext, RouteLocationNormalized, RouteRecordRaw } from 'vue-router'
import { createRouter, createWebHistory } from 'vue-router'

import AccountSetup from '@/pages/AccountSetup.vue'
import AdminCategories from '@/pages/admin/Category.vue'
import AdminOrders from '@/pages/admin/Order.vue'
import AdminOrderDetail from '@/pages/admin/OrderDetail.vue'
import AdminProducts from '@/pages/admin/Product.vue'
import AdminProductEdit from '@/pages/admin/ProductDetail.vue'
import AdminUsers from '@/pages/admin/User.vue'
import Cart from '@/pages/Cart.vue'
import Error from '@/pages/Error.vue'
import Home from '@/pages/Home.vue'
import LoginRegister from '@/pages/LoginRegister.vue'
import NotFound from '@/pages/NotFound.vue'
import OrderConfirmation from '@/pages/OrderConfirmation.vue'
import OrderDetail from '@/pages/OrderDetail.vue'
import Payment from '@/pages/Payment.vue'
import Product from '@/pages/Product.vue'
import ProductDetails from '@/pages/ProductDetail.vue'
import Register from '@/pages/Register.vue'
import RegisterConfirmation from '@/pages/RegisterConfirmation.vue'
import ShippingAddress from '@/pages/ShippingAddress.vue'
import Unsupported from '@/pages/Unsupported.vue'
import { getCategories } from '@/services/api'
import { useAuthStore } from '@/store/auth'

async function initRoutes(): Promise<RouteRecordRaw[]> {
  const baseRoutes: RouteRecordRaw[] = [
    { path: '/', component: Home },
    { path: '/auth/update', component: AccountSetup },
    { path: '/auth', component: LoginRegister },
    {
      path: '/auth/email/:email(.*)/registration-code/:registrationCode',
      component: Register,
      props: true,
    },
    { path: '/auth/register-confirm', component: RegisterConfirmation },
    { path: '/cart', component: Cart },
    { path: '/error', component: Error },
    { path: '/not-found', component: NotFound },
    { path: '/products/:id', component: ProductDetails, props: true },
    { path: '/checkout/shipping', component: ShippingAddress },
    { path: '/checkout/payment', component: Payment },
    { path: '/checkout/confirmation', component: OrderConfirmation },
    { path: '/orders/:id', component: OrderDetail },
    { path: '/unsupported', component: Unsupported },
    { path: '/admin/products', component: AdminProducts, beforeEnter: requireAdmin },
    { path: '/admin/products/:id', component: AdminProductEdit, beforeEnter: requireAdmin },
    { path: '/admin/categories', component: AdminCategories, beforeEnter: requireAdmin },
    { path: '/admin/orders', component: AdminOrders, beforeEnter: requireAdmin },
    { path: '/admin/orders/:id', component: AdminOrderDetail, beforeEnter: requireAdmin },
    { path: '/admin/users', component: AdminUsers, beforeEnter: requireAdmin },
    { path: '/new', component: Product, props: { sortBy: 'newest' }, name: 'NewProducts' },
    {
      path: '/popular',
      component: Product,
      props: { sortBy: 'popularity' },
      name: 'PopularProducts',
    },
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
  } catch {
    // Fail silently if categories cannot be fetched
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
