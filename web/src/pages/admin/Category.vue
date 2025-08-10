<template>
  <div class="category-container">
    <div class="new-category-form">
      <form @submit.prevent="handleSubmit">
        <div class="form-row">
          <input v-model="newCategory.name" type="text" placeholder="Category Name" required />
          <input v-model="newCategory.slug" type="text" placeholder="Slug" required />
          <select v-model="newCategory.parent_id">
            <option value="">No Parent Category</option>
            <option v-for="category in categories" :key="category.id" :value="category.id">
              {{ category.name }}
            </option>
          </select>
          <input
            v-model="newCategory.description"
            type="text"
            placeholder="Description (optional)"
          />
          <button type="submit" class="submit-button">Add Category</button>
        </div>
      </form>
    </div>
    <div class="category-grid">
      <CategoryTile
        v-for="category in categories"
        :key="category.id"
        :category="category"
        @delete="handleDelete"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

import CategoryTile from '@/components/CategoryTile.vue'
import { getCategories, createCategory, removeCategory } from '@/services/api'

const categories = ref([])
const newCategory = ref({
  name: '',
  slug: '',
  parent_id: '',
  description: '',
})

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
    const categoryData = {
      name: newCategory.value.name,
      slug: newCategory.value.slug,
      parent_id: newCategory.value.parent_id || undefined,
      description: newCategory.value.description || undefined,
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

const handleDelete = async (categoryId) => {
  try {
    await removeCategory(categoryId)
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
}

.form-row input,
.form-row select {
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 16px;
  background-color: transparent;
  min-width: 200px;
}

.submit-button {
  padding: 10px 20px;
  background-color: #000;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  cursor: pointer;
  transition: background-color 0.2s ease-in-out;
}

.submit-button:hover {
  background-color: #333;
}

.category-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  font-family: 'Inter', sans-serif;
  margin-top: 20px;
}
</style>
