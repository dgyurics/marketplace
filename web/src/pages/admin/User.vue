<template>
  <div class="user-container">
    <h2>Users</h2>
    <DataTable :columns="columns" :data="formattedUsers" :on-row-click="handleRowClick" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import DataTable from '@/components/DataTable.vue'
import { getUsers } from '@/services/api'
import type { UserRecord } from '@/types'
import { formatDate } from '@/utilities/dateFormat'

const router = useRouter()

const users = ref<UserRecord[]>([])

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
</style>
