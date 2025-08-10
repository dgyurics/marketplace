<template>
  <div class="key-value-editor">
    <h3 v-if="title">{{ title }}</h3>
    <div class="editor-container">
      <div v-for="(pair, index) in pairs" :key="index" class="pair-row">
        <input
          v-model="pair.key"
          type="text"
          :placeholder="keyPlaceholder"
          class="pair-input"
          :class="{ error: errors[index] }"
          @input="validateKey(index)"
        />
        <input
          v-model="pair.value"
          type="text"
          :placeholder="valuePlaceholder"
          class="pair-input"
          @input="emitValue"
        />
        <button
          type="button"
          class="remove-btn"
          :disabled="pairs.length === 1"
          :title="pairs.length === 1 ? 'Cannot remove the last pair' : 'Remove this pair'"
          @click="removePair(index)"
        >
          ×
        </button>
        <div v-if="errors[index]" class="error-message">{{ errors[index] }}</div>
      </div>
      <button type="button" class="add-btn" @click="addPair">+ Add {{ pairName || 'Pair' }}</button>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, defineEmits, defineProps } from 'vue'

const props = defineProps({
  title: {
    type: String,
    default: '',
  },
  keyPlaceholder: {
    type: String,
    default: 'Key',
  },
  valuePlaceholder: {
    type: String,
    default: 'Value',
  },
  pairName: {
    type: String,
    default: 'Pair',
  },
  modelValue: {
    type: Object,
    default: () => ({}),
  },
  allowDuplicateKeys: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:modelValue'])

const pairs = ref([{ key: '', value: '' }])
const errors = ref([])

// Initialize pairs from modelValue
const initializePairs = () => {
  const entries = Object.entries(props.modelValue)
  if (entries.length === 0) {
    pairs.value = [{ key: '', value: '' }]
  } else {
    pairs.value = entries.map(([key, value]) => ({ key, value }))
  }
  errors.value = new Array(pairs.value.length).fill('')
}

// Initialize on mount
initializePairs()

const addPair = () => {
  pairs.value.push({ key: '', value: '' })
  errors.value.push('')
}

const removePair = (index) => {
  if (pairs.value.length > 1) {
    pairs.value.splice(index, 1)
    errors.value.splice(index, 1)
    emitValue()
  }
}

const validateKey = (index) => {
  const currentKey = pairs.value[index].key.trim()

  if (!currentKey) {
    errors.value[index] = ''
    emitValue()
    return
  }

  if (!props.allowDuplicateKeys) {
    const duplicateIndex = pairs.value.findIndex(
      (pair, i) => i !== index && pair.key.trim().toLowerCase() === currentKey.toLowerCase()
    )

    if (duplicateIndex !== -1) {
      errors.value[index] = 'Duplicate key'
      return
    }
  }

  errors.value[index] = ''
  emitValue()
}

const emitValue = () => {
  const result = {}
  pairs.value.forEach((pair) => {
    const key = pair.key.trim()
    const value = pair.value.trim()
    if (key && value) {
      result[key] = value
    }
  })
  emit('update:modelValue', result)
}

// Watch pairs for value changes
watch(
  pairs,
  () => {
    emitValue()
  },
  { deep: true }
)

// Expose validation state
const hasErrors = () => {
  return errors.value.some((error) => Boolean(error))
}

defineExpose({
  hasErrors,
  reset: () => {
    pairs.value = [{ key: '', value: '' }]
    errors.value = ['']
    emitValue()
  },
})
</script>

<style scoped>
.key-value-editor {
  margin: 20px 0;
}

.key-value-editor h3 {
  font-size: 18px;
  margin-bottom: 15px;
  text-align: center;
}

.editor-container {
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 15px;
  background-color: #f9f9f9;
}

.pair-row {
  position: relative;
  display: flex;
  gap: 10px;
  align-items: flex-start;
  margin-bottom: 10px;
}

.pair-input {
  flex: 1;
  padding: 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 14px;
  background-color: white;
  transition: border-color 0.2s;
}

.pair-input:focus {
  outline: none;
  border-color: #007bff;
}

.pair-input.error {
  border-color: #e74c3c;
}

.remove-btn {
  width: 30px;
  height: 30px;
  border: none;
  background-color: #e74c3c;
  color: white;
  border-radius: 50%;
  cursor: pointer;
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
  flex-shrink: 0;
}

.remove-btn:disabled {
  background-color: #bdc3c7;
  cursor: not-allowed;
}

.remove-btn:not(:disabled):hover {
  background-color: #c0392b;
}

.add-btn {
  width: 100%;
  padding: 10px;
  border: 1px dashed #666;
  background-color: transparent;
  color: #666;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.add-btn:hover {
  background-color: #f0f0f0;
  border-color: #333;
  color: #333;
}

.error-message {
  position: absolute;
  top: 100%;
  left: 0;
  color: #e74c3c;
  font-size: 12px;
  margin-top: 2px;
}
</style>
