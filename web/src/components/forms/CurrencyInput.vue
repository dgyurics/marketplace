<template>
  <input :value="formatted" @input="onInput" @blur="onBlur" @focus="onFocus" />
</template>

<script setup>
import { ref, computed } from 'vue'

const rawValue = ref('') // this is stored as "1999", etc.
const currencySymbol = '$'

const formatted = computed(() => {
  if (!rawValue.value) return ''
  const num = parseInt(rawValue.value, 10) / 100
  return (
    currencySymbol +
    num.toLocaleString('en-US', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    })
  )
})

function onInput(event) {
  const digitsOnly = event.target.value.replace(/\D/g, '') // strip non-digits
  rawValue.value = digitsOnly
}

function onBlur() {
  // optional: format to $x,xxx.xx on blur
}

function onFocus(event) {
  // optional: strip formatting for easier editing
  event.target.value = rawValue.value
}
</script>
