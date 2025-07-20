import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { v4 as uuidv4 } from 'uuid'
import { chatApi } from '@/lib/api/chat'

interface Message {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: string
  isStreaming?: boolean
}

interface Chat {
  id: string
  title: string | null
  createdAt: string
  updatedAt: string
}

interface ChatStore {
  chats: Chat[]
  messages: Record<string, Message[]>
  currentChatId: string | null
  
  // Actions
  createChat: () => Chat
  selectChat: (chatId: string) => void
  deleteChat: (chatId: string) => void
  updateChatTitle: (chatId: string, title: string) => void
  
  // Message actions
  sendMessage: (chatId: string, content: string) => Promise<void>
  addMessage: (chatId: string, message: Message) => void
  updateMessage: (chatId: string, messageId: string, content: string) => void
  
  // Clear all data
  clearAll: () => void
}

export const useChatStore = create<ChatStore>()(
  persist(
    (set, get) => ({
      chats: [],
      messages: {},
      currentChatId: null,

      createChat: () => {
        const newChat: Chat = {
          id: uuidv4(),
          title: null,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        }
        
        set((state) => ({
          chats: [newChat, ...state.chats],
          messages: { ...state.messages, [newChat.id]: [] },
        }))
        
        return newChat
      },

      selectChat: (chatId) => {
        set({ currentChatId: chatId })
      },

      deleteChat: (chatId) => {
        set((state) => {
          const { [chatId]: deleted, ...remainingMessages } = state.messages
          return {
            chats: state.chats.filter((chat) => chat.id !== chatId),
            messages: remainingMessages,
            currentChatId: state.currentChatId === chatId ? null : state.currentChatId,
          }
        })
      },

      updateChatTitle: (chatId, title) => {
        set((state) => ({
          chats: state.chats.map((chat) =>
            chat.id === chatId
              ? { ...chat, title, updatedAt: new Date().toISOString() }
              : chat
          ),
        }))
      },

      sendMessage: async (chatId, content) => {
        const { addMessage, updateChatTitle } = get()
        
        // Add user message
        const userMessage: Message = {
          id: uuidv4(),
          role: 'user',
          content,
          timestamp: new Date().toISOString(),
        }
        addMessage(chatId, userMessage)

        // Update chat title if it's the first message
        const messages = get().messages[chatId] || []
        if (messages.length === 1) {
          // Generate title from first message
          const title = content.slice(0, 50) + (content.length > 50 ? '...' : '')
          updateChatTitle(chatId, title)
        }

        try {
          // Send to API and get streaming response
          const assistantMessage: Message = {
            id: uuidv4(),
            role: 'assistant',
            content: '',
            timestamp: new Date().toISOString(),
            isStreaming: true,
          }
          addMessage(chatId, assistantMessage)

          // Simulate streaming response (replace with actual WebSocket/SSE)
          await chatApi.sendMessage(chatId, content, (chunk) => {
            set((state) => ({
              messages: {
                ...state.messages,
                [chatId]: state.messages[chatId].map((msg) =>
                  msg.id === assistantMessage.id
                    ? { ...msg, content: msg.content + chunk, isStreaming: true }
                    : msg
                ),
              },
            }))
          })

          // Mark streaming as complete
          set((state) => ({
            messages: {
              ...state.messages,
              [chatId]: state.messages[chatId].map((msg) =>
                msg.id === assistantMessage.id
                  ? { ...msg, isStreaming: false }
                  : msg
              ),
            },
          }))
        } catch (error) {
          console.error('Failed to send message:', error)
          // Remove the streaming message on error
          set((state) => ({
            messages: {
              ...state.messages,
              [chatId]: state.messages[chatId].filter(
                (msg) => msg.id !== assistantMessage.id
              ),
            },
          }))
          throw error
        }
      },

      addMessage: (chatId, message) => {
        set((state) => ({
          messages: {
            ...state.messages,
            [chatId]: [...(state.messages[chatId] || []), message],
          },
        }))
      },

      updateMessage: (chatId, messageId, content) => {
        set((state) => ({
          messages: {
            ...state.messages,
            [chatId]: state.messages[chatId].map((msg) =>
              msg.id === messageId ? { ...msg, content } : msg
            ),
          },
        }))
      },

      clearAll: () => {
        set({
          chats: [],
          messages: {},
          currentChatId: null,
        })
      },
    }),
    {
      name: 'chat-storage',
    }
  )
)