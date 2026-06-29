<template>
  <div class="auth-container">
    <template v-if="authStore.hasMinimumRole('user')">
      <h2>You are logged in</h2>
      <UserMenu />
    </template>
    <template v-else>
      <h2>Sign In or Create an Account</h2>

      <form @submit.prevent="handleLogin">
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
          <button type="submit" :tabindex="0">Login</button>
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

import UserMenu from '@/components/UserMenu.vue'
import { login as apiLogin, register as apiRegister } from '@/services/api'
import { useAuthStore } from '@/store/auth'
import { useCartStore } from '@/store/cart'

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
    useCartStore().fetchCart()

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

  if (!isValidPassword(password.value)) {
    errorMessage.value = 'Password must be between 3 and 50 characters'
    return
  }

  try {
    await apiRegister(emailCpy, password.value)
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

form {
  margin-top: 50px;
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

.button-group.menu {
  flex-direction: column;
  align-items: center;
  max-width: 200px;
  margin: 30px auto 0;
}

.button-group.menu button {
  width: 100%;
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
