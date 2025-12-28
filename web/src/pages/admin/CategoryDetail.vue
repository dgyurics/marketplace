<template>
  <div class="category-detail-container">
    <div v-if="loading" class="loading">Loading category...</div>
    <div v-else-if="category" class="edit-form">
      <form @submit.prevent="handleSubmit">
        <div class="form-row">
          <InputText v-model="editCategory.name" label="name" required />
          <InputText v-model="editCategory.slug" label="slug" required />
          <SelectInput
            v-model="editCategory.parent_id"
            label="parent category"
            :options="parentCategoryOptions"
          />
          <InputText v-model="editCategory.description" label="description" />
        </div>

        <div class="form-actions">
          <button v-auth="'admin'" type="submit" :tabindex="0" class="btn-full-width mt-15">
            Save
          </button>
          <button type="button" :tabindex="0" class="btn-full-width btn-outline" @click="goBack">
            Cancel
          </button>
          <button
            v-auth="'admin'"
            type="button"
            :tabindex="0"
            class="btn-full-width btn-subtle"
            @click="handleDelete"
          >
            Remove
          </button>
        </div>
      </form>
    </div>

    <div v-else class="error">Category not found or failed to load.</div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { InputText, SelectInput } from '@/components/forms'
import { getCategoryById, getCategories, updateCategory, removeCategory } from '@/services/api'

const route = useRoute()
const router = useRouter()

const category = ref(null)
const categories = ref([])
const loading = ref(true)

const editCategory = ref({
  name: '',
  slug: '',
  parent_id: '',
  description: '',
})

const parentCategoryOptions = computed(() =>
  categories.value
    .filter((cat) => cat.id !== category.value?.id) // Don't allow selecting self as parent
    .map((cat) => ({ value: cat.id, label: cat.name }))
)

const fetchCategory = async () => {
  try {
    loading.value = true
    const categoryId = route.params.id
    const data = await getCategoryById(categoryId)
    category.value = data

    // Populate form with existing data
    editCategory.value = {
      name: data.name,
      slug: data.slug,
      parent_id: data.parent_id ?? '',
      description: data.description ?? '',
    }
  } catch {
    // Handle error silently
    category.value = null
  } finally {
    loading.value = false
  }
}

const fetchCategories = async () => {
  try {
    const data = await getCategories()
    categories.value = data
  } catch {
    // Handle error silently
  }
}

const handleSubmit = async () => {
  try {
    const updateData = {
      id: category.value.id,
      name: editCategory.value.name,
      slug: editCategory.value.slug,
      parent_id: editCategory.value.parent_id || undefined,
      description: editCategory.value.description || undefined,
    }

    await updateCategory(updateData)

    // For now, just go back
    goBack()
  } catch {
    // Handle error silently
  }
}

const goBack = () => {
  router.back()
}

const handleDelete = async () => {
  try {
    await removeCategory(category.value.id)
    router.push('/admin/categories')
  } catch {
    // Handle error silently
  }
}

onMounted(() => {
  fetchCategory()
  fetchCategories()
})
</script>

<style scoped>
.category-detail-container {
  max-width: 800px;
  margin: auto;
  padding: 20px;
}

.header {
  margin-bottom: 30px;
}

.header h1 {
  margin: 0;
  font-size: 24px;
  color: #333;
}

.loading,
.error {
  text-align: center;
  padding: 40px;
  font-size: 16px;
  color: #666;
}

.error {
  color: #e74c3c;
}

.edit-form {
  padding: 30px;
  border-radius: 8px;
}

.form-row {
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.form-row :deep(.input-container) {
  flex: 1 1 calc(50% - 10px);
}

.form-actions {
  display: flex;
  flex-direction: column;
  gap: 15px;
  margin-top: 30px;
}
</style>
