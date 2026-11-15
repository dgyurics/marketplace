import type { RouteLocationNormalized, RouteRecordRaw } from 'vue-router'
import { createRouter, createWebHistory } from 'vue-router'

import Home from '@/pages/Home.vue'
import { getCategories, registerConfirm } from '@/services/api'
import { useAuthStore } from '@/store/auth'
import { useInboxStore } from '@/store/inbox'

const AdminCategories = () => import('@/pages/admin/Category.vue')
const AdminCategoryDetail = () => import('@/pages/admin/CategoryDetail.vue')
const AdminOffers = () => import('@/pages/admin/Offer.vue')
const AdminOfferDetail = () => import('@/pages/admin/OfferDetail.vue')
const AdminOrders = () => import('@/pages/admin/Order.vue')
const AdminOrderDetail = () => import('@/pages/admin/OrderDetail.vue')
const AdminProducts = () => import('@/pages/admin/Product.vue')
const AdminProductEdit = () => import('@/pages/admin/ProductDetail.vue')
const AdminShippingZones = () => import('@/pages/admin/ShippingZone.vue')
const AdminUsers = () => import('@/pages/admin/User.vue')
const AdminUserDetail = () => import('@/pages/admin/UserDetail.vue')
const Cart = () => import('@/pages/Cart.vue')
const Error = () => import('@/pages/Error.vue')
const Inbox = () => import('@/pages/Inbox.vue')
const InboxDetail = () => import('@/pages/InboxDetail.vue')
const LoginRegister = () => import('@/pages/LoginRegister.vue')
const NotFound = () => import('@/pages/NotFound.vue')
const Offer = () => import('@/pages/Offer.vue')
const OfferDetail = () => import('@/pages/OfferDetail.vue')
const OrderConfirmation = () => import('@/pages/OrderConfirmation.vue')
const OrderDetail = () => import('@/pages/OrderDetail.vue')
const PasswordReset = () => import('@/pages/PasswordReset.vue')
const PasswordResetRequest = () => import('@/pages/PasswordResetRequest.vue')
const Payment = () => import('@/pages/Payment.vue')
const Product = () => import('@/pages/Product.vue')
const ProductDetails = () => import('@/pages/ProductDetail.vue')
const Profile = () => import('@/pages/Profile.vue')
const RegisterConfirmation = () => import('@/pages/RegisterConfirmation.vue')
const ShippingAddress = () => import('@/pages/ShippingAddress.vue')

async function initRoutes(): Promise<RouteRecordRaw[]> {
  const baseRoutes: RouteRecordRaw[] = [
    { path: '/', component: Home },
    { path: '/auth', component: LoginRegister },
    { path: '/auth/register-confirm', component: RegisterConfirmation },
    { path: '/cart', component: Cart },
    { path: '/offer/:id', component: Offer, beforeEnter: requireMember },
    { path: '/offers/:id', component: OfferDetail, beforeEnter: requireMember },
    { path: '/error', component: Error },
    { path: '/not-found', component: NotFound },
    { path: '/products/:id', component: ProductDetails, props: true },
    { path: '/auth/email/:email(.*)/password-reset', component: PasswordResetRequest, props: true },
    {
      path: '/auth/email/:email(.*)/password-reset/:resetCode',
      component: PasswordReset,
      props: true,
    },
    { path: '/checkout/shipping', component: ShippingAddress },
    { path: '/checkout/payment', component: Payment },
    { path: '/checkout/confirmation', component: OrderConfirmation },
    { path: '/orders/:id', component: OrderDetail },
    { path: '/profile', component: Profile, beforeEnter: requireUser },
    { path: '/inbox', component: Inbox, beforeEnter: requireGuest },
    { path: '/inbox/:id', component: InboxDetail, beforeEnter: requireGuest, props: true },
    { path: '/admin/products', component: AdminProducts, beforeEnter: requireAdmin },
    { path: '/admin/products/:id', component: AdminProductEdit, beforeEnter: requireAdmin },
    { path: '/admin/categories', component: AdminCategories, beforeEnter: requireAdmin },
    { path: '/admin/categories/:id', component: AdminCategoryDetail, beforeEnter: requireAdmin },
    { path: '/admin/orders', component: AdminOrders, beforeEnter: requireAdmin },
    { path: '/admin/orders/:id', component: AdminOrderDetail, beforeEnter: requireAdmin },
    { path: '/admin/offers', component: AdminOffers, beforeEnter: requireAdmin },
    {
      path: '/admin/offers/:id',
      component: AdminOfferDetail,
      beforeEnter: requireAdmin,
    },
    { path: '/admin/shipping-zones', component: AdminShippingZones, beforeEnter: requireAdmin },
    { path: '/admin/users', component: AdminUsers, beforeEnter: requireAdmin },
    { path: '/admin/users/:id', component: AdminUserDetail, beforeEnter: requireAdmin },
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
function requireAdmin(): string | undefined {
  if (!useAuthStore().hasMinimumRole('staff')) return '/'
  return undefined
}

// User route guard
function requireUser(): string | undefined {
  if (!useAuthStore().hasMinimumRole('user')) return '/'
  return undefined
}

// Member route guard
function requireMember(): string | undefined {
  if (!useAuthStore().hasMinimumRole('member')) return '/'
  return undefined
}

// Guest route guard
function requireGuest(): string | undefined {
  if (!useAuthStore().hasMinimumRole('guest')) return '/'
  return undefined
}

// Export function to create router (to be called after Pinia is initialized)
export async function createAppRouter() {
  const routes = await initRoutes()
  const router = createRouter({
    history: createWebHistory(),
    routes,
  })

  // global query parameter check
  router.beforeEach(async (to) => {
    if (to.query['registration-code']) {
      return handleRegistrationCode(to)
    }
    return undefined
  })

  // fetch inbox on every navigation
  router.afterEach(() => {
    if (useAuthStore().isAuthenticated) {
      useInboxStore().fetchConversations()
    }
  })

  return router
}

// Handle registration code confirmation
async function handleRegistrationCode(to: RouteLocationNormalized) {
  try {
    const authTokens = await registerConfirm(to.query['registration-code'] as string)
    const authStore = useAuthStore()
    authStore.setTokens(authTokens)
  } catch (error) {
    console.error('Registration confirmation failed:', error)
  }

  // Remove registration-code from query params
  const { 'registration-code': _, ...cleanQuery } = to.query
  return { ...to, query: cleanQuery, replace: true }
}
