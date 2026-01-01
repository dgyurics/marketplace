<template>
  <div class="auth-container">
    <template v-if="authStore.hasMinimumRole('user')">
      <h2>You are logged in</h2>
      <button class="profile-button" :tabindex="0" @click="goToProfile">Profile</button>
      <button class="logout-button" :tabindex="0" @click="handleLogout">Logout</button>
      <div v-if="authStore.hasMinimumRole('staff')">
        <div class="button-group admin-buttons">
          <button :tabindex="0" @click="goToCategories">Categories</button>
          <button :tabindex="0" @click="goToProducts">Products</button>
          <button :tabindex="0" @click="goToOrders">Orders</button>
          <button :tabindex="0" @click="goToUsers">Users</button>
          <button :tabindex="0" @click="goToShippingZones">Shipping</button>
        </div>
      </div>
    </template>
    <template v-else>
      <h2>Sign In or Create an Account</h2>

      <form @submit.prevent>
        <div class="form-group">
          <label for="email">Email</label>
          <input id="email" v-model="email" type="email" required :tabindex="0" />
          <div class="underline"></div>
        </div>

        <div class="form-group">
          <label for="password">Password</label>
          <input id="password" v-model="password" type="password" required :tabindex="0" />
          <div class="underline"></div>
        </div>

        <div class="button-group">
          <button type="button" :tabindex="0" @click="handleLogin">Login</button>
          <button type="button" :tabindex="0" class="btn-outline" @click="handleRegister">
            Register
          </button>
        </div>

        <p v-if="errorMessage" class="error" v-html="errorMessage"></p>
      </form>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { login as apiLogin, register as apiRegister, logout as apiLogout } from '@/services/api'
import { useAuthStore } from '@/store/auth'

const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const errorMessage = ref<string | null>(null)

const isValidEmail = (email: string) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

const isValidPassword = (password: string) => {
  return password.length >= 3 && password.length <= 50
}

const router = useRouter()

const handleLogin = async () => {
  errorMessage.value = null

  if (!isValidEmail(email.value)) {
    errorMessage.value = 'Invalid email address'
    return
  }

  if (!isValidPassword(password.value)) {
    errorMessage.value = 'Password must be between 3 and 50 characters'
    return
  }

  try {
    const authTokens = await apiLogin(email.value, password.value)
    authStore.setTokens(authTokens)

    if (authTokens.requires_setup) {
      router.push('/auth/update')
    }
    // Clear email + password field after successful login
    email.value = ''
    password.value = ''
  } catch (error: any) {
    const status = error.response?.status
    if (status === 401) {
      errorMessage.value =
        `Invalid credentials<br>` +
        `<span class="reset-password-text">` +
        `Click <a href="/auth/email/${email.value}/password-reset" class="forgot-password-link">here</a> to reset password` +
        `</span>`
    } else if (status === 429) {
      errorMessage.value = 'Too many failed attempts'
    } else {
      errorMessage.value = 'Something went wrong'
    }
  }
}

const handleRegister = async () => {
  const emailCpy = email.value.trim()
  errorMessage.value = null

  if (!isValidEmail(emailCpy)) {
    errorMessage.value = 'Invalid email address'
    return
  }

  try {
    await apiRegister(emailCpy)
    // Clear email + password field after successful registration
    email.value = ''
    password.value = ''

    // Redirect with email as query parameter
    router.push({
      path: '/auth/register-confirm',
      query: { email: emailCpy },
    })
  } catch (error: any) {
    const status = error.response?.status
    if (status === 409) {
      errorMessage.value = 'Email already in use'
    } else if (status === 429) {
      errorMessage.value = 'You are doing that too much'
    } else {
      errorMessage.value = 'Something went wrong'
    }
  }
}

const goToCategories = () => {
  router.push('/admin/categories')
}

const goToProfile = () => {
  router.push('/profile')
}

const goToProducts = () => {
  router.push('/admin/products')
}

const goToOrders = () => {
  router.push('/admin/orders')
}

const goToUsers = () => {
  router.push('/admin/users')
}

const goToShippingZones = () => {
  router.push('/admin/shipping-zones')
}

const handleLogout = async () => {
  try {
    await apiLogout()
    authStore.clearTokens()
  } catch (error) {
    console.error('Logout error:', error)
  }
}
</script>

<style scoped>
.auth-container {
  max-width: 450px;
  margin: auto;
  padding: 20px;
  text-align: center;
  position: relative;
  top: -20px;
}

h2 {
  font-size: 22px;
  font-weight: 300;
  margin-bottom: 50px;
}

/* Labels */
label {
  font-weight: 500;
  font-size: 14px;
  display: block;
  margin-bottom: 5px;
}

/* Wider Input Fields */
input {
  width: 100%;
  max-width: 100%;
  padding: 10px 0;
  border: none;
  border-bottom: 2px solid #999;
  outline: none;
  background: transparent;
  font-size: 18px;
}

input:focus {
  border-bottom: 2px solid black;
}

/* Underline effect */
.underline {
  width: 100%;
  height: 2px;
  background: #999;
  position: absolute;
  bottom: 0;
  left: 0;
}

/* Button Group */
.button-group {
  display: flex;
  justify-content: space-between;
  margin-top: 30px;
  gap: 8px;
}

/* Admin button group - vertical layout */
.button-group.admin-buttons {
  flex-direction: column;
  align-items: center;
  max-width: 200px;
  margin: 30px auto 0;
}

.button-group.admin-buttons button {
  width: 100%;
}

/* Single logout button */
.logout-button,
.profile-button {
  width: 200px;
  margin: 20px auto 0;
  display: block;
}

.profile-button {
  margin: 10px auto 0;
}

/* Error message */
.error {
  color: #d32f2f;
  font-size: 14px;
  margin-top: 15px;
}

/* Success message */
.success {
  font-size: 14px;
  margin-top: 15px;
}

/* Reset password text - make entire line black */
.error :deep(.reset-password-text) {
  color: #000 !important;
}

/* Forgot password link styling - keep underline but inherit black color */
.error :deep(a) {
  color: inherit;
  text-decoration: underline;
}
</style>
