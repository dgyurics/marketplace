<template>
  <div class="auth-container">
    <template v-if="authStore.user?.role === 'user' || authStore.user?.role === 'admin'">
      <h2>You are logged in</h2>
      <button class="btn logout-btn" @click="handleLogout">Logout</button>
    </template>
    <template v-else>
      <h2>Sign In or Create an Account</h2>

      <!-- TODO move this to form component -->
      <form @submit.prevent>
        <div class="form-group">
          <label for="email">Email</label>
          <input id="email" v-model="email" type="email" required />
          <div class="underline"></div>
        </div>

        <div class="form-group">
          <label for="password">Password</label>
          <input id="password" v-model="password" type="password" required />
          <div class="underline"></div>
        </div>

        <div class="button-group">
          <button type="button" class="btn" @click="handleLogin">Login</button>
          <button type="button" class="btn register-btn" @click="handleRegister">Register</button>
        </div>

        <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
      </form>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { login as apiLogin, register as apiRegister, logout as apiLogout } from '@/services/api'
import { useAuthStore } from '@/store/auth'
import type { ApiError } from '@/types'

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
    errorMessage.value = 'Invalid email address.'
    return
  }

  if (!isValidPassword(password.value)) {
    errorMessage.value = 'Password must be between 3 and 50 characters.'
    return
  }

  try {
    const authTokens = await apiLogin(email.value, password.value)
    authStore.setTokens(authTokens)

    const redirectUri = authTokens.requires_setup ? '/auth/update' : '/'
    router.push(redirectUri)
  } catch (error) {
    errorMessage.value = handleApiError(error)
  }
}

const handleRegister = async () => {
  errorMessage.value = null

  if (!isValidEmail(email.value)) {
    errorMessage.value = 'Invalid email address.'
    return
  }

  if (!isValidPassword(password.value)) {
    errorMessage.value = 'Password must be between 3 and 50 characters.'
    return
  }

  try {
    const authTokens = await apiRegister(email.value, password.value)
    authStore.setTokens(authTokens)
    router.push('/')
  } catch (error) {
    errorMessage.value = handleApiError(error)
  }
}

const handleLogout = async () => {
  try {
    await apiLogout()
    authStore.clearTokens()
    router.push('/')
  } catch (error) {
    errorMessage.value = handleApiError(error)
  }
}

const handleApiError = (error: unknown): string => {
  if (error && typeof error === 'object' && 'response' in error) {
    const apiError = error as ApiError

    switch (apiError.response?.status) {
      case 409:
        return 'Email already in use.'
      case 401:
        return 'Invalid credentials. Try again.' // TODO have this return a button, disguised as a link, to reset password
      case 500:
        return 'Server error. Please try again later.'
      default:
        return apiError.message || 'An error occurred during registration.'
    }
  }

  if (error instanceof Error) {
    return error.message
  }

  return 'An unexpected error occurred. Please try again.'
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

.logout-btn {
  background: black;
  color: white;
  border: none;
  cursor: pointer;
  padding: 12px;
  font-size: 14px;
  width: 100%;
  margin-top: 20px;
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
  /* Ensures the input is full-width */
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
}

.btn {
  flex: 1;
  padding: 12px;
  background: black;
  color: white;
  border: none;
  cursor: pointer;
  transition: background 0.3s ease;
  font-size: 14px;
  margin: 0;
}

.register-btn {
  background: white;
  color: black;
  border: 1px solid black;
  margin-left: 8px;
}

.register-btn:hover {
  background: black;
  color: white;
}

.btn:hover {
  background: #333;
}
</style>
