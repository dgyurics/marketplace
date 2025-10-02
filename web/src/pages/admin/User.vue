<template>
  <div class="user-container">
    <DataTable :columns="columns" :data="formattedUsers" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

import DataTable from '@/components/DataTable.vue'
import { getUsers } from '@/services/api'
import type { UserRecord } from '@/types'
import { formatDate } from '@/utilities/dateFormat'

const users = ref<UserRecord[]>([])

const columns = ['id', 'email', 'role', 'requires_setup', 'created', 'updated']

const formattedUsers = computed(() =>
  users.value.map((user) => ({
    id: user.id,
    email: user.email,
    role: user.role,
    requires_setup: user.requires_setup,
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

.page-title {
  font-size: 24px;
  font-weight: 300;
  margin-bottom: 20px;
  color: #333;
}
</style>
