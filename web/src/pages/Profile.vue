<template>
  <div class="profile-container">
    <h2>Profile</h2>

    <!-- User Info Section -->
    <div class="user-info">
      <h3>Account Information</h3>
      <div class="form-group-flex">
        <InputText
          label="email"
          type="email"
          :readonly="!isAdmin"
          :model-value="userEmail"
          required
          @update:model-value="(val) => (editableEmail = val)"
        />
      </div>
      <div class="form-group-flex">
        <InputText label="account type" readonly :model-value="userRole" required />
      </div>
      <div class="form-group-flex">
        <InputText
          label="current password"
          type="password"
          :model-value="currentPassword"
          @update:model-value="(val) => (currentPassword = val)"
        />
      </div>
      <div class="form-group-flex">
        <InputText
          label="new password"
          type="password"
          :model-value="newPassword"
          @update:model-value="(val) => (newPassword = val)"
        />
      </div>
    </div>

    <!-- Submit Section -->
    <div class="submit-section">
      <form @submit.prevent="handleSubmit">
        <div class="button-group">
          <button type="submit" class="btn-full-width mt-15">Update Profile</button>
        </div>
      </form>

      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'

import { InputText } from '@/components/forms'
import { updatePassword, updateEmail } from '@/services/api'
import { useAuthStore } from '@/store/auth'

const authStore = useAuthStore()
const router = useRouter()

const currentPassword = ref('')
const newPassword = ref('')
const editableEmail = ref('')
const errorMessage = ref<string | null>(null)

const userEmail = computed(() => authStore.user.email || 'Not set')
const userRole = computed(() => authStore.user.role || 'guest')
const isAdmin = computed(() => authStore.hasMinimumRole('admin'))

const isValidPassword = (password: string) => {
  return password.length >= 3 && password.length <= 50
}

const handleSubmit = async () => {
  errorMessage.value = null

  // Check if admin changed email
  if (isAdmin.value && editableEmail.value && editableEmail.value !== userEmail.value) {
    try {
      // TODO make password required for email change
      authStore.setTokens(await updateEmail(editableEmail.value))
      router.push('/auth')
      return
    } catch (error: any) {
      const status = error.response?.status
      if (status === 400) {
        errorMessage.value = 'Invalid email address'
      } else if (status === 409) {
        errorMessage.value = 'Email already in use'
      } else if (status === 429) {
        errorMessage.value = 'Too many failed attempts'
      } else {
        errorMessage.value = 'Something went wrong'
      }
      return
    }
  }

  // Only attempt password change if both fields have input
  if (currentPassword.value && newPassword.value) {
    if (!isValidPassword(currentPassword.value)) {
      errorMessage.value = 'Current password must be between 3 and 50 characters'
      return
    }

    if (!isValidPassword(newPassword.value)) {
      errorMessage.value = 'New password must be between 3 and 50 characters'
      return
    }

    if (currentPassword.value === newPassword.value) {
      errorMessage.value = 'New password must be different from current password'
      return
    }

    try {
      authStore.setTokens(await updatePassword(currentPassword.value, newPassword.value))
      router.push('/auth')
      return
    } catch (error: any) {
      const status = error.response?.status
      if (status === 400) {
        errorMessage.value = 'Invalid password'
      } else if (status === 429) {
        errorMessage.value = 'Too many failed attempts'
      } else {
        errorMessage.value = 'Something went wrong'
      }
      return
    }
  }

  // If we get here and no password fields were filled, redirect to auth
  router.push('/auth')
}
</script>

<style scoped>
.profile-container {
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
  text-align: center;
}

h3 {
  font-size: 18px;
  font-weight: 400;
  margin: 30px 0 20px 0;
  text-align: center;
  text-transform: capitalize;
}

.user-info {
  margin-bottom: 40px;
}

.submit-section {
  margin-top: 20px;
}

.form-group-flex {
  margin-bottom: 20px;
}

/* Button Group */
.button-group {
  display: flex;
  justify-content: center;
  margin-top: 30px;
}

.error {
  color: #e74c3c;
  margin-top: 15px;
  font-size: 14px;
}
</style>
