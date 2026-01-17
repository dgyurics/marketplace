<template>
  <div class="setup-container">
    <h2>Set Password</h2>
    <h3>Set your password to finish</h3>
    <form @submit.prevent>
      <div class="form-group">
        <label for="password">Password</label>
        <input id="password" v-model="password" type="password" required />
        <div class="underline"></div>
      </div>

      <div class="button-group">
        <button type="button" :tabindex="0" @click="handleSubmit">Submit</button>
      </div>

      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'

import { registerInviteConfirm } from '@/services/api'
import { useAuthStore } from '@/store/auth'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const password = ref('')
const errorMessage = ref<string | null>(null)

const isValidPassword = (password: string) => {
  return password.length >= 3 && password.length <= 50
}

const handleSubmit = async () => {
  errorMessage.value = null

  if (!isValidPassword(password.value)) {
    errorMessage.value = 'Password must be between 3 and 50 characters'
    return
  }

  try {
    const registrationCode = route.query['registration-code'] as string
    const authTokens = await registerInviteConfirm(registrationCode, password.value)
    authStore.setTokens(authTokens)
    router.push('/auth')
  } catch (error: any) {
    const status = error.response?.status
    if (status === 400 || status === 409) {
      errorMessage.value = 'Registration code expired or invalid'
      return
    }
    errorMessage.value = 'Something went wrong'
  }
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
  margin-bottom: 10px;
}

h3 {
  font-size: 16px;
  font-weight: 400;
  color: #666;
  margin-bottom: 40px;
  margin-top: 0;
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

input.readonly {
  color: #666;
  background-color: #f5f5f5;
  cursor: not-allowed;
}

input.readonly:focus {
  border-bottom: 2px solid #999;
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
</style>
