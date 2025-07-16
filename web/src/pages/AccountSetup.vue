<template>
  <div class="setup-container">
    <h2>Update your credentials</h2>
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
        <button type="button" class="btn" @click="handleUpdate">Update</button>
      </div>

      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { updateCredentials as apiUpdateCredentials } from '@/services/api'
import { useAuthStore } from '@/store/auth'
import type { ApiError } from '@/types'

const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const errorMessage = ref<string | null>(null)

// TODO move to util; duplicated logic
const isValidEmail = (email: string) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

// TODO move to util; duplicated logic
const isValidPassword = (password: string) => {
  return password.length >= 3 && password.length <= 50
}

const router = useRouter()

const handleUpdate = async () => {
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
    const authTokens = await apiUpdateCredentials(email.value, password.value)
    authStore.setTokens(authTokens)
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
      case 500:
        return 'Server error. Please try again later.'
      default:
        return apiError.message || 'An error occurred updating your credentials.'
    }
  }

  if (error instanceof Error) {
    return error.message
  }

  return 'An unexpected error occurred. Please try again.'
}
</script>

<style scoped>
.setup-container {
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
  max-width: 100%; /* Ensures the input is full-width */
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
  margin: 0 8px;
}

.btn:hover {
  background: #333;
}
</style>
