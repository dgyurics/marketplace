<template>
  <div class="inbox-container">
    <router-link class="inbox-back" to="/inbox">← Back to inbox</router-link>
    <header class="inbox-heading conversation-heading">
      <h1>{{ conversation?.subject || 'Conversation' }}</h1>
      <EnvelopeIcon class="inbox-heading-icon" />
    </header>

    <div v-if="loading" class="inbox-state" role="status">Loading conversation...</div>

    <div v-else-if="conversation" class="conversation-detail">
      <div class="messages">
        <div v-if="conversation.messages.length === 0" class="inbox-state">
          No messages in this conversation.
        </div>
        <div v-for="message in conversation.messages" :key="message.id" class="message">
          <div class="message-header">
            <div class="message-time">{{ formatDate(message.created_at) }}</div>
          </div>
          <div class="message-body" v-html="message.body"></div>
        </div>
      </div>
    </div>

    <p v-if="errorMessage" class="inbox-error" role="alert">{{ errorMessage }}</p>
  </div>
</template>

<script setup lang="ts">
import { EnvelopeIcon } from '@heroicons/vue/24/outline'
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import { getConversationById } from '@/services/api'
import type { Conversation } from '@/types/conversation'
import { formatDate } from '@/utilities'

const route = useRoute()
const conversation = ref<Conversation | null>(null)
const loading = ref(false)
const errorMessage = ref<string | null>(null)

const loadConversation = async (id: string) => {
  loading.value = true
  errorMessage.value = null

  try {
    conversation.value = await getConversationById(id)
  } catch (error: unknown) {
    conversation.value = null
    const status = (error as { response?: { status?: number } })?.response?.status
    if (status === 404) {
      errorMessage.value = 'Conversation not found'
    } else if (status === 403) {
      errorMessage.value = 'Access denied to this conversation'
    } else {
      errorMessage.value = 'Failed to load conversation details'
    }
  } finally {
    loading.value = false
  }
}

watch(
  () => route.params['id'],
  (id) => {
    if (typeof id === 'string' && id.length > 0) {
      loadConversation(id)
    } else {
      conversation.value = null
      errorMessage.value = 'Conversation not found'
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.conversation-heading {
  padding: 16px 20px;
  margin-bottom: 20px;
}
.conversation-heading h1 {
  font-size: 18px;
  font-weight: 600;
  letter-spacing: 0;
  line-height: 1.4;
  overflow-wrap: anywhere;
}
.conversation-heading .inbox-heading-icon {
  width: 26px;
  height: 26px;
}
.inbox-back {
  display: inline-block;
  margin-bottom: 20px;
  font-size: 13px;
  color: var(--muted-text-color);
}
.messages {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.message {
  border: 1px solid var(--border-color);
  border-radius: var(--panel-radius);
  padding: 28px;
  background: var(--surface-background);
  min-width: 0;
}
.message-header {
  border-bottom: 1px solid var(--subtle-border-color);
  padding-bottom: 14px;
  margin-bottom: 18px;
}
.message-time {
  font-size: 12px;
  color: var(--muted-text-color);
}
.message-body {
  line-height: 1.8;
  white-space: pre-wrap;
  font-size: 14px;
  color: var(--muted-text-color);
  overflow-wrap: anywhere;
}
.message-body :deep(a) {
  color: var(--accent-color);
  text-decoration: underline;
  text-underline-offset: 3px;
}
.message-body :deep(a:hover) {
  color: var(--accent-text-color);
}
.message-body :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: var(--control-radius);
}
.message-body :deep(ul),
.message-body :deep(ol) {
  padding-left: 20px;
}
@media (max-width: 600px) {
  .message {
    padding: 20px;
  }
}
</style>
