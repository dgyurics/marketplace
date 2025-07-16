<template>
  <form @submit.prevent="handleSubmit">
    <div class="form-group">
      <label for="productName">name</label>
      <input id="productName" v-model="product.name" type="text" required />
    </div>
    <div class="form-group">
      <label for="productCategory">category</label>
      <select id="productCategory" v-model="selectedCategorySlug" required>
        <option v-for="category in categories" :key="category.slug" :value="category.slug">
          {{ category.name }}
        </option>
      </select>
    </div>
    <div class="form-group">
      <label for="productPrice">price</label>
      <input id="productPrice" v-model="priceText" type="number" required />
      <!-- <InputCurrency /> -->
    </div>
    <div class="form-group">
      <label for="productDescription">description</label>
      <textarea id="productDescription" v-model="product.description" rows="2" required></textarea>
    </div>
    <div class="form-group">
      <label for="productDetails">details</label>
      <textarea id="productDetails" v-model="detailsText" rows="6" required></textarea>
    </div>
    <!-- <div class="form-group">
      <label for="productTaxCode">tax code</label>
      <input id="productTaxCode" v-model="product.tax_code" type="text" />
    </div> -->
    <!-- <div class="form-group">
      <label for="productEnable">enable</label>
      <input id="productEnable" type="checkbox" />
    </div> -->
    <button type="submit" class="submit-button">Submit</button>
  </form>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'

// import InputCurrency from '@/components/InputCurrency.vue'
import { getCategories } from '@/services/api'
import type { CreateProductRequest, Category } from '@/types'

const emit = defineEmits<{
  submit: [product: CreateProductRequest, categorySlug: string]
}>()

const product = ref<CreateProductRequest>({
  name: '',
  description: '',
  details: {},
  price: 0,
  tax_code: '',
})

const categories = ref<Category[]>([])
const selectedCategorySlug = ref('')

// Handle details as JSON string for easier editing
const detailsText = ref('')

// Handle price as a string to allow easier formatting
const priceText = ref('')

onMounted(async () => {
  try {
    categories.value = await getCategories()
  } catch (error) {
    console.error('Failed to load categories:', error)
  }
})

// Parse details when form is submitted
const parsedDetails = computed(() => {
  try {
    return detailsText.value ? JSON.parse(detailsText.value) : {}
  } catch {
    return {}
  }
})

// Format price as a number
const formattedPrice = computed(() => {
  // remove non-numeric characters except for decimal point
  // multiply by 100 to store as cents
  // FIXME needs to work with other currencies like JPY, KWD, etc.
  // Locale/currency back-end api in-progress
  //const price = parseFloat(priceText.value)
  const price = parseInt(priceText.value)
  return isNaN(price) ? 0 : price
})

const handleSubmit = () => {
  const productData: CreateProductRequest = {
    name: product.value.name,
    description: product.value.description,
    details: parsedDetails.value,
    price: formattedPrice.value,
  }

  emit('submit', productData, selectedCategorySlug.value)
}
</script>

<style scoped>
label {
  font-weight: 500;
  font-size: 14px;
  display: block;
  margin-bottom: 5px;
  text-transform: capitalize;
}

input {
  width: 100%;
  max-width: 100%;
  font-size: 18px;
}

input[type='text'],
input[type='email'],
input[type='password'],
input[type='tel'],
input[type='number'],
input[type='search'],
input[type='file'],
textarea,
select {
  width: 100%;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 18px;
  box-sizing: border-box;
  background-color: transparent;
}
</style>
