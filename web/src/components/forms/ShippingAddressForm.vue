<template>
  <form @submit.prevent="handleSubmit">
    <div class="form-group-flex">
      <label for="fullName">Full Name</label>
      <input id="fullName" v-model="formData.addressee" type="text" required />
    </div>

    <div class="form-group-flex">
      <label for="street">Address</label>
      <input id="street" v-model="formData.line1" type="text" required />
    </div>

    <div class="form-group-flex">
      <label for="apt">Apt, Suite, Building (Optional)</label>
      <input id="apt" v-model="formData.line2" type="text" />
    </div>

    <div class="form-row">
      <div class="form-group-flex city">
        <label for="city">City</label>
        <input id="city" v-model="formData.city" type="text" required />
      </div>
      <div class="form-group-flex state">
        <label for="state">State</label>
        <input id="state" v-model="formData.state" type="text" required />
      </div>
      <div class="form-group-flex zip">
        <label for="zip">Zip Code</label>
        <input id="zip" v-model="formData.postal_code" type="text" required />
      </div>
    </div>

    <div class="form-group-flex">
      <label for="addressEmail">Email</label>
      <input id="addressEmail" v-model="formData.email" type="email" required />
      <small class="receipt-note">A receipt will be sent to this email.</small>
    </div>

    <button type="submit" class="btn-full-width mt-15">Continue</button>
  </form>
</template>

<script setup lang="ts">
import { reactive } from 'vue'

import type { Address } from '@/types'
import { getCountryForLocale, getAppLocale } from '@/utilities'

const props = defineProps<{ modelValue?: Address }>()

const emit = defineEmits<{
  submit: [address: Address]
}>()

const formData = reactive<Address>({
  addressee: '',
  line1: '',
  line2: '',
  city: '',
  state: '',
  postal_code: '',
  country: getCountryForLocale(getAppLocale()),
  email: '',
  ...props.modelValue,
})

function handleSubmit() {
  emit('submit', formData)
}
</script>

<style scoped>
input[type='text'],
input[type='email'],
input[type='password'],
input[type='tel'],
input[type='number'],
input[type='search'] {
  width: 100%;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 16px;
  box-sizing: border-box;
}

.form-row {
  display: flex;
  gap: 10px;
}

.form-row .form-group-flex {
  flex: 1;
}

.receipt-note {
  font-size: 10px;
  color: #666;
  margin-top: 2px;
}
</style>
