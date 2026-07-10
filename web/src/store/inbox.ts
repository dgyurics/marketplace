import { defineStore } from 'pinia'

import {
  getConversations as apiGetConversations,
  getConversationById as apiGetConversationById,
} from '@/services/api'
import type { Conversation } from '@/types/conversation'

export const useInboxStore = defineStore('inbox', {
  state: () => ({
    conversations: [] as Conversation[],
  }),

  getters: {
    unreadCount: (state) =>
      state.conversations.filter((c) => {
        const lastRead = new Date(c.recipient_last_read_at).getTime()
        const updatedAt = new Date(c.updated_at).getTime()
        return updatedAt + 50 > lastRead
      }).length,

    isUnread: () => (conversation: Conversation) => {
      const lastRead = new Date(conversation.recipient_last_read_at).getTime()
      const updatedAt = new Date(conversation.updated_at).getTime()
      return updatedAt + 50 > lastRead
    },
  },

  actions: {
    async fetchConversations() {
      try {
        this.conversations = await apiGetConversations()
        console.log('Fetched conversations:', this.conversations)
        return this.conversations
      } catch (err) {
        console.error('Error fetching conversations:', err)
        this.conversations = []
        return []
      }
    },

    async fetchConversationById(id: string) {
      try {
        return await apiGetConversationById(id)
      } catch (err) {
        console.error('Error fetching conversation:', err)
        throw err
      }
    },

    clearInbox() {
      this.conversations = []
    },
  },
})
