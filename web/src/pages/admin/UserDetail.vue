<template>
  <div class="user-detail-container">
    <div v-if="user" class="user-detail">
      <div class="user-header">
        <h1>User Details</h1>
        <div class="user-meta">
          <span class="user-id">ID: {{ user.id }}</span>
          <span class="user-status">{{ user.role.toUpperCase() }}</span>
        </div>
      </div>

      <div class="user-info">
        <div class="info-row">
          <label>Email:</label>
          <span>{{ user.email || 'N/A' }}</span>
        </div>
        <div class="info-row">
          <label>Role:</label>
          <span>{{ user.role }}</span>
        </div>
        <div class="info-row">
          <label>Created:</label>
          <span>{{ formatDate(user.created_at) }}</span>
        </div>
        <div class="info-row">
          <label>Updated:</label>
          <span>{{ formatDate(user.updated_at) }}</span>
        </div>
      </div>

      <div class="info-row">
        <label>Registration Link:</label>
        <div class="code-section">
          <span v-if="generatedCode" class="registration-link">
            {{ registrationUrl }}
          </span>
          <span
            v-else-if="!generatedCode"
            v-auth="'admin'"
            class="clickable-generate"
            @click="handleGenerateCode"
          >
            click to generate
          </span>
        </div>
      </div>

      <div class="user-actions">
        <button type="button" class="btn-full-width btn-outline" @click="goBack">
          Back to Users
        </button>
        <button
          v-auth="'admin'"
          type="button"
          class="btn-full-width btn-subtle"
          @click="handleRemoveUser"
        >
          Remove User
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getUser, removeUser, createRegistrationCode } from '@/services/api'
import type { UserRecord, RegistrationCode } from '@/types'
import { formatDate } from '@/utilities/dateFormat'

const route = useRoute()
const router = useRouter()

const user = ref<UserRecord | null>(null)
const generatedCode = ref<RegistrationCode | null>(null)

const registrationUrl = computed(() => {
  if (!generatedCode.value) return ''
  const { protocol, host } = window.location
  return `${protocol}//${host}?registration-code=${generatedCode.value.registration_code}`
})

const fetchUser = async () => {
  try {
    const userId = route.params['id'] as string
    const data = await getUser(userId)
    user.value = data
  } catch {
    user.value = null
  }
}

const handleRemoveUser = async () => {
  if (!user.value) return

  try {
    await removeUser(user.value.id.toString())
    router.push('/admin/users')
  } catch {
    // Handle error silently
  }
}

const handleGenerateCode = async () => {
  if (!user.value) return

  try {
    const response = await createRegistrationCode(user.value.id)
    generatedCode.value = response
  } catch {
    // Handle error silently
  }
}

const goBack = () => {
  router.push('/admin/users')
}

onMounted(() => {
  fetchUser()
})
</script>

<style scoped>
.user-detail-container {
  max-width: 800px;
  width: 800px;
  margin: auto;
  padding: 20px;
}

@media (max-width: 768px) {
  .user-detail-container {
    width: calc(100% - 20px);
    padding: 15px;
  }
}

.user-header {
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 1px solid #ddd;
}

.user-header h1 {
  margin: 0 0 10px 0;
  font-size: 24px;
  color: #333;
}

.user-meta {
  display: flex;
  align-items: center;
  gap: 20px;
}

.user-id {
  font-family: 'Roboto Mono', monospace;
  font-size: 14px;
  color: #666;
}

.user-status {
  padding: 4px 12px;
  border-radius: 4px;
  border: 1px solid #ddd;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  background-color: #f8f9fa;
}

.user-info {
  margin-bottom: 30px;
}

.info-row {
  display: flex;
  margin-bottom: 15px;
  align-items: center;
}

.info-row label {
  min-width: 100px;
  font-weight: 500;
  color: #666;
}

.code-section {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  margin-left: 20px;
}

.code-display {
  font-family: 'Roboto Mono', monospace;
  font-size: 14px;
  font-weight: 600;
  color: #333;
  letter-spacing: 1px;
  background-color: #f8f9fa;
  padding: 6px 10px;
  border-radius: 4px;
  border: 1px solid #e1e5e9;
  line-height: 1;
}

.registration-link {
  font-family: 'Roboto Mono', monospace;
  font-size: 13px;
  color: #333;
  letter-spacing: 1px;
  background-color: #f8f9fa;
  padding: 6px 10px;
  border-radius: 4px;
  border: 1px solid #e1e5e9;
  line-height: 1;
  word-break: break-all;
  display: inline-block;
}

.clickable-generate {
  font-size: 14px;
  cursor: pointer;
  text-decoration: underline;
  line-height: 1;
  transition: color 0.2s;
}

.user-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
</style>
