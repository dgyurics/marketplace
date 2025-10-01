<template>
  <div class="container">
    <h2>Registration Successful</h2>
    <div class="confirmation-message mt-45">
      <p class="confirmation-note">
        Confirmation link sent to <strong>{{ email }}</strong>
      </p>
      <p class="confirmation-footnote">
        <i>(Check your junk mail if you do not see it)</i>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const email = ref('')

onMounted(() => {
  // Get email from query parameters
  email.value = route.query.email || ''
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
