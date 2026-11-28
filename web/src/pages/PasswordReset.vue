<template>
  <div class="setup-container">
    <h2>Password Reset</h2>
    <form @submit.prevent>
      <div class="form-group">
        <label for="password">New Password</label>
        <input id="password" v-model="password" type="password" required />
      </div>

      <div class="button-group">
        <button type="button" @click="handleSubmit">Submit</button>
      </div>

      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'

import { passwordUpdate } from '@/services/api'
import { isValidPassword } from '@/utilities/auth'

const route = useRoute()
const router = useRouter()

const password = ref('')
const errorMessage = ref<string | null>(null)

const handleSubmit = async () => {
  errorMessage.value = null

  if (!isValidPassword(password.value)) {
    errorMessage.value = 'Password must be between 3 and 50 characters'
    return
  }

  try {
    const resetCode = route.params['resetCode'] as string
    const email = route.params['email'] as string
    await passwordUpdate(email, password.value, resetCode)
    router.push('/auth')
  } catch {
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

/* Labels */
label {
  font-weight: 500;
  font-size: 14px;
  display: block;
  margin-bottom: 5px;
}

input {
  width: 100%;
  padding: 14px 12px;
  border: 1px solid var(--input-border-color);
  border-radius: var(--control-radius);
  background: var(--surface-background);
  color: var(--text-color);
  font-size: 16px;
}

input:focus-visible {
  border-color: var(--accent-color);
  outline-offset: 2px;
}

/* Button Group */
.button-group {
  display: flex;
  justify-content: space-between;
  margin-top: 30px;
}
</style>
