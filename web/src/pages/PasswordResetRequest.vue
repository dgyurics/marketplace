<template>
  <div v-if="emailSent" class="container">
    <h2>Password Reset</h2>
    <div class="confirmation-message mt-45">
      <p class="confirmation-note">
        A password reset link has been sent to <strong>{{ email }}</strong>
      </p>
      <p class="confirmation-footnote">
        <i>(Check your junk mail if you do not see it)</i>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { passwordReset } from '@/services/api'

const router = useRouter()
const route = useRoute()
const email = ref('')
const emailSent = ref(false)

onMounted(async () => {
  try {
    // Get email from route parameters
    email.value = route.params['email'] as string
    await passwordReset(email.value)
    emailSent.value = true
  } catch (error: any) {
    const status = error.response?.status
    if (status === 404) {
      emailSent.value = true
    } else {
      router.push(`/error?status=${status || 500}`)
    }
  }
})
</script>

<style scoped>
h2,
h3 {
  text-align: center;
  margin-bottom: 10px;
}

.confirmation-footnote,
.confirmation-note {
  text-align: center;
}

.confirmation-footnote {
  font-size: 12px;
}
</style>
