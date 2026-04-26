<template>
  <div class="inbox-container">
    <h2>Inbox</h2>

    <!-- Conversation List -->
    <div v-if="!selectedConversation" class="conversation-list">
      <div v-if="loading" class="loading">Loading conversations...</div>
      <div v-else-if="conversations.length === 0" class="no-conversations">
        No conversations found.
      </div>
      <div v-else>
        <div
          v-for="conversation in conversations"
          :key="conversation.id"
          class="conversation-item"
          tabindex="0"
          @click="selectConversation(conversation.id)"
          @keydown.enter="selectConversation(conversation.id)"
        >
          <div class="conversation-header">
            <div class="subject">
              <span class="subject-label">Subject: </span>
              <span class="subject-text">{{ conversation.subject }}</span>
            </div>
            <div class="date-time">
              <div class="date">{{ formatDate(conversation.updated_at) }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Conversation Detail -->
    <div v-else class="conversation-detail">
      <div class="detail-header">
        <h3 class="conversation-title">{{ selectedConversation.subject }}</h3>
      </div>

      <div class="messages">
        <div v-if="selectedConversation.messages.length === 0" class="no-messages">
          No messages in this conversation.
        </div>
        <div v-for="message in selectedConversation.messages" :key="message.id" class="message">
          <div class="message-header">
            <div class="message-time">{{ formatDate(message.created_at) }}</div>
          </div>
          <div class="message-body">{{ message.body }}</div>
        </div>
      </div>

      <button type="button" class="btn-full-width btn-outline mt-30" @click="goBack">
        Back to Inbox
      </button>
    </div>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

import { getConversations, getConversationById } from '@/services/api'
import type { Conversation } from '@/types/conversation'
import { formatDate } from '@/utilities'

const conversations = ref<Conversation[]>([])
const selectedConversation = ref<Conversation | null>(null)
const loading = ref(false)
const errorMessage = ref<string | null>(null)

const loadConversations = async () => {
  loading.value = true
  errorMessage.value = null
  try {
    conversations.value = await getConversations()
  } catch (error: unknown) {
    const status = (error as { response?: { status?: number } })?.response?.status
    if (status === 401) {
      errorMessage.value = 'Please log in to view your inbox'
    } else if (status === 403) {
      errorMessage.value = 'Access denied'
    } else {
      errorMessage.value = 'Failed to load conversations'
    }
  } finally {
    loading.value = false
  }
}

const selectConversation = async (id: string) => {
  try {
    selectedConversation.value = await getConversationById(id)
  } catch (error: unknown) {
    const status = (error as { response?: { status?: number } })?.response?.status
    if (status === 404) {
      errorMessage.value = 'Conversation not found'
    } else if (status === 403) {
      errorMessage.value = 'Access denied to this conversation'
    } else {
      errorMessage.value = 'Failed to load conversation details'
    }
  }
}

const goBack = () => {
  selectedConversation.value = null
  errorMessage.value = null
}

onMounted(loadConversations)
</script>

<style scoped>
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

  .conversation-item {
    padding: 12px;
  }

  .subject {
    font-size: 15px;
  }

  .date-time {
    font-size: 11px;
  }

  .from {
    font-size: 13px;
  }

  .detail-header h3 {
    font-size: 16px;
    margin: 15px 0;
  }

  .conversation-meta {
    font-size: 12px;
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

.conversation-title {
  text-transform: capitalize;
}

.loading {
  padding: 20px;
  font-style: italic;
  color: #666;
}

.no-conversations,
.no-messages {
  padding: 20px;
  color: #666;
  font-style: italic;
}

/* Conversation List Styles */
.conversation-list {
  margin-top: 20px;
}

.conversation-item {
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 10px;
  cursor: pointer;
  transition: background-color 0.2s;
  text-align: left;
}

.conversation-item:hover {
  background-color: #f5f5f5;
}

.conversation-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
}

.subject {
  font-weight: 500;
  font-size: 16px;
  flex: 1;
  margin-right: 10px;
  text-align: left;
}

.subject-label {
  color: #888;
  font-weight: 400;
  font-size: 14px;
}

.subject-text {
  text-transform: lowercase;
}

.date-time {
  text-align: right;
  font-size: 12px;
  color: #666;
}

.date {
  margin-bottom: 2px;
}

.time {
  color: #999;
}

.from {
  color: #666;
  font-size: 14px;
  margin-bottom: 8px;
}

/* Conversation Detail Styles */
.conversation-detail {
  margin-top: 20px;
}

.detail-header {
  margin-bottom: 30px;
}

.conversation-meta {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-top: 10px;
  font-size: 14px;
  color: #666;
}

.messages {
  text-align: left;
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

.sender {
  font-weight: 500;
}

.message-time {
  font-size: 12px;
}

.message-body {
  line-height: 1.5;
  white-space: pre-wrap;
}

.error {
  color: #e74c3c;
  margin-top: 15px;
  font-size: 14px;
}
</style>
