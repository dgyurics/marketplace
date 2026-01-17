<template>
  <div class="user-container">
    <div class="invite-user-form">
      <form @submit.prevent="handleInviteSubmit">
        <div class="form-row">
          <InputText v-model="inviteForm.email" label="email" required type="email" />
          <SelectInput v-model="inviteForm.role" label="role" :options="roleOptions" required />
        </div>
        <button v-auth="'admin'" type="submit" class="btn-full-width mt-15" :tabindex="0">
          Invite User
        </button>
      </form>
      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    </div>
    <DataTable :columns="columns" :data="formattedUsers" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

import DataTable from '@/components/DataTable.vue'
import { InputText, SelectInput } from '@/components/forms'
import { getUsers, inviteUser } from '@/services/api'
import type { UserRecord } from '@/types'
import { AuthenticatedRole } from '@/types'
import { formatDate } from '@/utilities/dateFormat'

const users = ref<UserRecord[]>([])
const errorMessage = ref<string | null>(null)
const inviteForm = ref<{
  email: string
  role: AuthenticatedRole | ''
}>({
  email: '',
  role: '',
})

const columns = ['id', 'email', 'role', 'created', 'updated']

const roleOptions = [
  { value: AuthenticatedRole.ADMIN, label: 'Admin' },
  { value: AuthenticatedRole.USER, label: 'User' },
  { value: AuthenticatedRole.MEMBER, label: 'Member' },
  { value: AuthenticatedRole.STAFF, label: 'Staff' },
]

const formattedUsers = computed(() =>
  users.value.map((user) => ({
    id: user.id,
    email: user.email,
    role: user.role,
    created: formatDate(new Date(user.created_at)),
    updated: formatDate(new Date(user.updated_at)),
  }))
)

const fetchUsers = async () => {
  try {
    const data = await getUsers()
    users.value = data
  } catch {
    // Handle error silently
  }
}

const handleInviteSubmit = async () => {
  errorMessage.value = null

  try {
    if (inviteForm.value.role) {
      await inviteUser(inviteForm.value.email, inviteForm.value.role as AuthenticatedRole)

      // Reset form
      inviteForm.value = { email: '', role: '' }

      // Refresh users list
      await fetchUsers()
    }
  } catch (error: any) {
    const status = error.response?.status
    if (status === 409) {
      errorMessage.value = 'Email already in use'
      return
    }
    if (status === 400) {
      errorMessage.value = 'Invalid email or role'
      return
    }
    errorMessage.value = 'Something went wrong'
  }
}

onMounted(() => {
  fetchUsers()
})
</script>

<style scoped>
.user-container {
  max-width: 1200px;
  margin: auto;
  padding: 20px;
}

.invite-user-form {
  margin-bottom: 30px;
}

.form-row {
  display: flex;
  gap: 10px;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  margin-bottom: 15px;
}

.form-row :deep(.input-container) {
  flex: 1 1 calc(50% - 10px);
}

.error {
  color: #e74c3c;
  font-size: 14px;
  margin-top: 10px;
  text-align: center;
}

.page-title {
  font-size: 24px;
  font-weight: 300;
  margin-bottom: 20px;
  color: #333;
}
</style>
