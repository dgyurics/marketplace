<template>
  <input v-model="code" type="text" maxlength="6" class="code-input" @input="handleInput" />
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

interface Props {
  modelValue?: string
}

interface Emits {
  (e: 'update:modelValue', value: string): void
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
})

const emit = defineEmits<Emits>()

const code = ref(props.modelValue)

const handleInput = () => {
  code.value = code.value
    .toUpperCase()
    .replace(/[^A-Z0-9]/g, '')
    .slice(0, 6)
  emit('update:modelValue', code.value)
}

watch(
  () => props.modelValue,
  (newValue) => {
    code.value = newValue || ''
  }
)
</script>

<style scoped>
.code-input {
  width: 100%;
  height: 100%;
  text-align: center;
  font-size: 14px;
  letter-spacing: 0.2rem;
  text-transform: uppercase;
  border: 1px solid #ddd;
  border-radius: 4px;
  outline: none;
  padding: 10px;
}
</style>
