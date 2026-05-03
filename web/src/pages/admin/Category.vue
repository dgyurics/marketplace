<template>
  <div class="category-container">
    <div class="new-category-form">
      <form @submit.prevent="handleSubmit">
        <div class="form-row">
          <InputText v-model="newCategory.name" label="name" required />
          <InputText v-model="newCategory.slug" label="slug" required />
          <SelectInput
            v-model="newCategory.parent_id"
            label="parent category"
            :options="parentCategoryOptions"
          />
          <InputText v-model="newCategory.description" label="description" />
        </div>
        <button v-auth="'admin'" type="submit" :tabindex="0" class="btn-full-width mt-15">
          Create Category
        </button>
      </form>
    </div>
    <div class="category-grid">
      <CategoryTile
        v-for="category in categories"
        :key="category.id"
        :category="category"
        @click="goToDetail(category.id)"
        @keydown.enter="goToDetail(category.id)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'

import CategoryTile from '@/components/CategoryTile.vue'
import { InputText, SelectInput } from '@/components/forms'
import { getCategories, createCategory } from '@/services/api'
import type { Category } from '@/types/category'

const router = useRouter()

const categories = ref<Category[]>([])
const newCategory = ref({
  name: '',
  slug: '',
  parent_id: '',
  description: '',
})

const parentCategoryOptions = computed(() =>
  categories.value.map((category) => ({ value: category.id, label: category.name }))
)

const goToDetail = (categoryId: string) => {
  router.push(`/admin/categories/${categoryId}`)
}

const fetchCategories = async () => {
  try {
    const response = await getCategories()
    categories.value = response
  } catch {
    // Handle error silently
  }
}

const handleSubmit = async () => {
  try {
    const categoryData: Partial<Category> = {
      name: newCategory.value.name,
      slug: newCategory.value.slug,
    }

    // Only add optional fields if they have values
    if (newCategory.value.parent_id) {
      categoryData.parent_id = newCategory.value.parent_id
    }

    if (newCategory.value.description) {
      categoryData.description = newCategory.value.description
    }

    await createCategory(categoryData)

    // Reset form
    newCategory.value = { name: '', slug: '', parent_id: '', description: '' }

    // Refresh categories
    await fetchCategories()
  } catch {
    // Handle error silently
  }
}

onMounted(() => {
  fetchCategories()
})
</script>

<style scoped>
.category-container {
  max-width: 1200px;
  margin: auto;
  padding: 20px;
  text-align: center;
}

.new-category-form {
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
  flex: 1 1 calc(25% - 10px);
  min-width: 0;
}

.form-row :deep(.input-field),
.form-row :deep(.select-field) {
  width: 100%;
  box-sizing: border-box;
}

.category-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 20px;
  margin-top: 20px;
}

@media (max-width: 768px) {
  .category-container {
    width: calc(100% - 20px);
    padding: 10px;
  }

  .form-row {
    flex-direction: column;
    gap: 15px;
  }

  .form-row :deep(.input-container) {
    flex: 1 1 100%;
    width: 100%;
  }

  .category-grid {
    grid-template-columns: 1fr;
    gap: 15px;
  }
}
</style>
