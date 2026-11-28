<template>
  <div
    class="auth-container"
    :class="{
      'auth-container--account': authStore.hasMinimumRole('user'),
      'auth-container--admin': authStore.hasMinimumRole('staff'),
    }"
  >
    <template v-if="authStore.hasMinimumRole('user')">
      <UserMenu />
    </template>
    <template v-else>
      <header class="auth-heading">
        <h1>Sign in or register</h1>
        <p>Enter your email and password.</p>
      </header>

      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label for="email">Email</label>
          <input id="email" v-model="email" type="email" required :tabindex="0" />
        </div>

        <div class="form-group">
          <label for="password">Password</label>
          <input id="password" v-model="password" type="password" required :tabindex="0" />
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
import { useInboxStore } from '@/store/inbox'
import { isValidPassword } from '@/utilities/auth'

const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const errorMessage = ref<string | null>(null)

const isValidEmail = (email: string) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
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
    useInboxStore().fetchConversations()

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
  width: calc(100% - 64px);
  max-width: 520px;
  margin: 48px auto 64px;
  padding: 36px;
  border: 1px solid var(--border-color);
  border-radius: var(--panel-radius);
  background: var(--surface-background);
}
.auth-container--account {
  max-width: 620px;
  padding: 0;
  border: none;
}
.auth-container--admin {
  max-width: 1000px;
}
.auth-heading {
  padding: 24px;
  border-radius: 10px;
  background: var(--highlight-background);
  margin-bottom: 28px;
}
.auth-heading h1 {
  font-size: 26px;
  font-weight: 700;
  line-height: 1.3;
  letter-spacing: -0.7px;
}
.auth-heading p {
  font-size: 13px;
  line-height: 1.7;
  color: var(--muted-text-color);
  margin-top: 12px;
}
.form-group {
  margin-bottom: 20px;
}
label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 8px;
}
input {
  width: 100%;
  padding: 14px 12px;
  border: 1px solid var(--input-border-color);
  border-radius: var(--control-radius);
  background: var(--surface-background);
  color: var(--text-color);
  font-family: inherit;
  font-size: 16px;
}
input:focus-visible {
  border-color: var(--accent-color);
  outline-offset: 2px;
}
.button-group {
  display: flex;
  gap: 12px;
  margin-top: 28px;
}
.auth-container form button {
  min-width: 0;
}
.error {
  background: var(--error-background);
  border-radius: var(--control-radius);
  padding: 12px 16px;
  color: var(--error-color);
  font-size: 13px;
  text-align: left;
  line-height: 1.7;
  margin-top: 20px;
  overflow-wrap: anywhere;
}
.error :deep(.reset-password-text) {
  color: var(--text-color);
}
.error :deep(a) {
  color: inherit;
  text-decoration: underline;
  text-underline-offset: 3px;
}
@media (max-width: 600px) {
  .auth-container {
    width: calc(100% - 32px);
    margin: 24px auto 40px;
    padding: 20px;
  }
  .auth-container--account {
    padding: 0;
  }
  .auth-heading {
    padding: 20px;
  }
  .auth-heading h1 {
    font-size: 23px;
  }
}
</style>
