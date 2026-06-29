<template>
  <div class="inbox-container">
    <h2>Inbox</h2>

    <div v-if="loading" class="loading">Loading conversation...</div>

    <div v-else-if="conversation" class="conversation-detail">
      <div class="detail-header">
        <h3 class="conversation-title">{{ conversation.subject }}</h3>
      </div>

      <div class="messages">
        <div v-if="conversation.messages.length === 0" class="no-messages">
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

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
  </div>
</template>

<script setup lang="ts">
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

<style>
.inbox-container {
  max-width: 900px;
  width: calc(100% - 40px);
  margin: auto;
  padding: 20px;
  text-align: center;
  position: relative;
  top: -20px;
}

@media (max-width: 768px) {
  .inbox-container {
    width: calc(100% - 20px);
    padding: 10px;
  }

  .detail-header h3 {
    font-size: 16px;
    margin: 15px 0;
  }

  .message {
    padding: 12px;
    margin-bottom: 10px;
  }

  .message-header {
    font-size: 12px;
  }

  .message-body {
    font-size: 14px;
    line-height: 1.4;
  }
}

.conversation-detail {
  margin-top: 20px;
}

.detail-header {
  margin-bottom: 30px;
}

.conversation-title {
  text-transform: capitalize;
}

.messages {
  text-align: left;
}

.no-messages {
  padding: 20px;
  color: #666;
  font-style: italic;
}

.loading {
  padding: 20px;
  font-style: italic;
  color: #666;
}

.message {
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 15px;
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  font-size: 14px;
  color: #666;
}

.message-time {
  font-size: 12px;
}

.message-body {
  line-height: 1.6;
  white-space: pre-wrap;
  font-size: 15px;
  color: #333;
  letter-spacing: 0.2px;
}

.message-body a {
  color: #2563eb;
  text-decoration: underline;
}

.message-body a:hover {
  color: #1d4ed8;
}

.error {
  color: #e74c3c;
  margin-top: 15px;
  font-size: 14px;
}
</style>
