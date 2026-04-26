<template>
  <div class="user-container">
    <h2>Users</h2>

    <!-- Create User Form -->
    <div class="new-user-form">
      <form @submit.prevent="handleCreateUser">
        <div class="form-row">
          <InputText v-model="newUserEmail" label="email" type="email" />
          <SelectInput v-model="newUserRole" label="role" :options="roleOptions" required />
        </div>
        <button type="submit" class="btn-full-width mt-15" :disabled="!newUserRole">
          Create User
        </button>
      </form>
      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    </div>

    <DataTable :columns="columns" :data="formattedUsers" :on-row-click="handleRowClick" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import DataTable from '@/components/DataTable.vue'
import { InputText, SelectInput } from '@/components/forms'
import { getUsers, createUser } from '@/services/api'
import type { UserRecord } from '@/types'
import { formatDate } from '@/utilities/dateFormat'

const router = useRouter()

const users = ref<UserRecord[]>([])
const newUserEmail = ref('')
const newUserRole = ref('')
const errorMessage = ref('')

const roleOptions = [
  { value: 'admin', label: 'Admin' },
  { value: 'user', label: 'User' },
  { value: 'guest', label: 'Guest' },
  { value: 'staff', label: 'Staff' },
  { value: 'member', label: 'Member' },
]

const columns = ['id', 'email', 'role', 'created', 'updated']

const formattedUsers = computed(() =>
  users.value.map((user) => ({
    id: user.id,
    email: user.email,
    role: user.role,
    created: formatDate(new Date(user.created_at)),
    updated: formatDate(new Date(user.updated_at)),
  }))
)

const handleRowClick = (row: { [key: string]: unknown }) => {
  router.push(`/admin/users/${row['id']}`)
}

const handleCreateUser = async () => {
  errorMessage.value = ''

  try {
    const newUser: Partial<UserRecord> = {
      role: newUserRole.value as 'admin' | 'user' | 'guest' | 'staff' | 'member',
    }

    if (newUserEmail.value.trim()) {
      newUser.email = newUserEmail.value.trim()
    }

    await createUser(newUser)

    // Clear form
    newUserEmail.value = ''
    newUserRole.value = ''

    // Refresh users list
    await fetchUsers()
  } catch {
    errorMessage.value = 'Failed to create user'
  }
}

const fetchUsers = async () => {
  try {
    const data = await getUsers()
    users.value = data
  } catch {
    // Handle error silently
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
  text-align: center;
}

.new-user-form {
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

@media (max-width: 768px) {
  .user-container {
    width: calc(100% - 20px);
    padding: 10px;
  }

  .form-row {
    flex-direction: column;
    gap: 15px;
  }

  .form-row :deep(.input-container) {
    flex: 1 1 100%;
  }
}
</style>
