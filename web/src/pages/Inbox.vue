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
            <div class="conversation-actions">
              <div class="date">{{ formatDate(conversation.updated_at) }}</div>
              <button
                class="delete-btn"
                title="Delete conversation"
                @click.stop="removeConversation(conversation.id)"
              >
                <TrashIcon class="delete-icon" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
  </div>
</template>

<script setup lang="ts">
import { TrashIcon } from '@heroicons/vue/24/outline'
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { useInboxStore } from '@/store/inbox'
import { formatDate } from '@/utilities'

const inboxStore = useInboxStore()
const loading = ref(false)
const errorMessage = ref<string | null>(null)
const router = useRouter()

const conversations = computed(() => inboxStore.conversations)

const loadConversations = async () => {
  loading.value = true
  errorMessage.value = null
  try {
    await inboxStore.fetchConversations()
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

const removeConversation = async (id: string) => {
  try {
    await inboxStore.removeConversation(id)
  } catch {
    errorMessage.value = 'Failed to delete conversation'
  }
}

const isUnread = inboxStore.isUnread

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
  background-color: #fff;
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

.conversation-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  color: #666;
}

.delete-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  opacity: 0.4;
  transition: opacity 0.2s;
}

.delete-btn:hover {
  opacity: 1;
}

.delete-icon {
  width: 16px;
  height: 16px;
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
