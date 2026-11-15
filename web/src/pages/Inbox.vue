<template>
  <div class="inbox-container">
    <header class="inbox-heading">
      <div>
        <h1>Inbox</h1>
        <p>Messages and order updates, all in one place.</p>
      </div>
      <EnvelopeIcon class="inbox-heading-icon" />
    </header>

    <div class="conversation-list">
      <div v-if="loading" class="inbox-state" role="status">Loading conversations...</div>
      <div v-else-if="conversations.length === 0 && !errorMessage" class="inbox-state">
        <EnvelopeIcon class="empty-icon" />
        <h2>You're all caught up</h2>
        <p>Your messages and order updates will appear here.</p>
      </div>
      <ul v-else class="conversation-rows" aria-label="Conversations">
        <li
          v-for="conversation in conversations"
          :key="conversation.id"
          :class="['conversation-item', { 'conversation-item--unread': isUnread(conversation) }]"
        >
          <router-link :to="`/inbox/${conversation.id}`" class="conversation-link">
            <span class="conversation-symbol"><EnvelopeIcon /></span>
            <span class="conversation-content">
              <span class="subject-text">{{ conversation.subject }}</span>
              <span class="conversation-meta">
                <span v-if="isUnread(conversation)" class="unread-badge">New</span>
                <span class="date">{{ formatDate(conversation.updated_at) }}</span>
              </span>
            </span>
            <ChevronRightIcon class="conversation-arrow" />
          </router-link>
          <button
            type="button"
            class="delete-btn"
            :aria-label="`Delete conversation: ${conversation.subject}`"
            title="Delete conversation"
            @click="removeConversation(conversation.id)"
          >
            <TrashIcon class="delete-icon" />
          </button>
        </li>
      </ul>
    </div>
    <p v-if="errorMessage" class="inbox-error" role="alert">{{ errorMessage }}</p>
  </div>
</template>

<script setup lang="ts">
import { ChevronRightIcon, EnvelopeIcon, TrashIcon } from '@heroicons/vue/24/outline'
import { ref, computed, onMounted } from 'vue'

import { useInboxStore } from '@/store/inbox'
import { formatDate } from '@/utilities'

const inboxStore = useInboxStore()
const loading = ref(false)
const errorMessage = ref<string | null>(null)

const conversations = computed(() => inboxStore.conversations)

const loadConversations = async () => {
  loading.value = true
  errorMessage.value = null
  try {
    await inboxStore.fetchConversations()
  } catch (error: unknown) {
    const status = (error as { response?: { status?: number } })?.response?.status
    if (status === 401) {
      errorMessage.value = 'Log in to view your inbox'
    } else if (status === 403) {
      errorMessage.value = 'Access denied'
    } else {
      errorMessage.value = 'Failed to load conversations'
    }
  } finally {
    loading.value = false
  }
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
.conversation-rows {
  list-style: none;
  margin: 0;
  padding: 0;
}
.conversation-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px 12px 12px;
  border: 1px solid #e9e6e0;
  border-radius: 12px;
  margin-bottom: 12px;
  transition: background 0.2s;
}
.conversation-item:hover {
  background: #faf8f2;
}
.conversation-item--unread {
  background: #fffaf3;
  border-color: #e8d9b6;
}
.conversation-link {
  display: flex;
  align-items: center;
  gap: 18px;
  flex: 1;
  min-width: 0;
  padding: 10px;
  border-radius: 8px;
}
.conversation-symbol {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  flex-shrink: 0;
  border-radius: 12px;
  background: #f5f3ee;
  color: #7c7568;
}
.conversation-symbol svg {
  width: 22px;
  height: 22px;
  stroke-width: 1.5;
}
.conversation-item--unread .conversation-symbol {
  background: #fff0eb;
  color: var(--accent-color);
}
.conversation-content {
  flex: 1;
  min-width: 0;
}
.subject-text {
  display: block;
  font-size: 14px;
  overflow-wrap: anywhere;
}
.conversation-item--unread .subject-text {
  font-weight: 600;
}
.conversation-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 7px;
}
.date {
  font-size: 12px;
  color: var(--muted-text-color);
}
.unread-badge {
  color: #a3342b;
  background: #ffe8e2;
  font-size: 10px;
  font-weight: 600;
  border-radius: 4px;
  padding: 2px 6px;
}
.conversation-arrow {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  color: #888;
}
.delete-btn {
  flex: none;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  color: #777;
  padding: 10px;
  border-radius: 8px;
}
.delete-btn:hover {
  background: #fff0eb;
  color: var(--accent-color);
}
.delete-icon {
  width: 18px;
  height: 18px;
}
.empty-icon {
  width: 40px;
  height: 40px;
  stroke-width: 1.3;
  color: #8e7650;
  margin-bottom: 16px;
}
@media (max-width: 600px) {
  .conversation-item {
    gap: 4px;
    padding: 8px;
  }
  .conversation-link {
    gap: 12px;
  }
  .conversation-symbol {
    width: 36px;
    height: 36px;
  }
  .conversation-arrow {
    display: none;
  }
}
</style>
