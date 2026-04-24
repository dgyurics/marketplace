export interface Conversation {
  id: string
  type: 'support' | 'notification'
  subject: string
  recipient_id: string
  recipient_last_read_at: string
  messages: Message[]
  updated_at: string
  created_at: string
}

export interface Message {
  id: string
  sender_id: string
  conversation_id: string
  body: string
  created_at: string
}
