<template>
  <div class="inbox-container">
    <h2>Inbox</h2>

    <div class="conversation-list">
      <div v-if="loading" class="loading">Loading conversations...</div>
      <div v-else-if="conversations.length === 0" class="no-conversations">
        No conversations found
      </div>
      <div v-else>
        <div
          v-for="conversation in conversations"
          :key="conversation.id"
          :class="['conversation-item', { 'conversation-item--unread': isUnread(conversation) }]"
          tabindex="0"
          @click="openConversation(conversation.id)"
          @keydown.enter="openConversation(conversation.id)"
        >
          <div class="conversation-header">
            <div class="subject">
              <span v-if="isUnread(conversation)" class="unread-badge">new</span>
              <span class="subject-label">Subject: </span>
              <span :class="['subject-text', { 'subject-text--unread': isUnread(conversation) }]">
                {{ conversation.subject }}
              </span>
            </div>
            <div class="date-time">
              <div class="date">{{ formatDate(conversation.updated_at) }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { getConversations } from '@/services/api'
import type { Conversation } from '@/types/conversation'
import { formatDate } from '@/utilities'

const conversations = ref<Conversation[]>([])
const loading = ref(false)
const errorMessage = ref<string | null>(null)
const router = useRouter()

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

const openConversation = (id: string) => {
  router.push(`/inbox/${id}`)
}

const isUnread = (conversation: Conversation): boolean => {
  const tolerance = 50 // ms
  const lastRead = new Date(conversation.recipient_last_read_at).getTime()
  const updatedAt = new Date(conversation.updated_at).getTime()
  return updatedAt + tolerance > lastRead
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
}

.loading {
  padding: 20px;
  font-style: italic;
  color: #666;
}

.no-conversations {
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
  /* font-weight: 500;
  font-size: 16px; */
  flex: 1;
  margin-right: 10px;
  text-align: left;

  font-weight: 400;
  font-size: 14px;
  text-transform: lowercase;
}

.subject-label {
  color: #888;
}

.conversation-item--unread {
  border-left: 3px solid #3d3d3d;
}

.subject-text--unread {
  font-weight: 600;
}

.unread-badge {
  display: inline-block;
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: #fff;
  background-color: #3d3d3d;
  border-radius: 3px;
  padding: 1px 5px;
  margin-right: 7px;
  vertical-align: middle;
  position: relative;
  top: -1px;
}

.subject-text {
  color: #3d3d3d;
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

.error {
  color: #e74c3c;
  margin-top: 15px;
  font-size: 14px;
}
</style>
