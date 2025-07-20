import axios from 'axios'
import { useAuthStore } from '@/lib/store/auth'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add auth token to requests
api.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Handle auth errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout()
    }
    return Promise.reject(error)
  }
)

export const chatApi = {
  // Get all chats
  getChats: async () => {
    const { data } = await api.get('/api/chat/conversations')
    return data
  },

  // Get chat by ID
  getChat: async (chatId: string) => {
    const { data } = await api.get(`/api/chat/conversations/${chatId}`)
    return data
  },

  // Create new chat
  createChat: async () => {
    const { data } = await api.post('/api/chat/conversations')
    return data
  },

  // Delete chat
  deleteChat: async (chatId: string) => {
    await api.delete(`/api/chat/conversations/${chatId}`)
  },

  // Get messages for a chat
  getMessages: async (chatId: string, limit = 50, offset = 0) => {
    const { data } = await api.get(`/api/chat/conversations/${chatId}/messages`, {
      params: { limit, offset },
    })
    return data
  },

  // Send message (with streaming support)
  sendMessage: async (
    chatId: string, 
    content: string, 
    onChunk?: (chunk: string) => void
  ) => {
    // For streaming, we'll use EventSource or WebSocket
    // This is a placeholder for standard API call
    const { data } = await api.post(`/api/chat/conversations/${chatId}/messages`, {
      content,
    })

    // Simulate streaming for now
    if (onChunk && data.content) {
      const words = data.content.split(' ')
      for (let i = 0; i < words.length; i++) {
        await new Promise((resolve) => setTimeout(resolve, 50))
        onChunk(words.slice(0, i + 1).join(' '))
      }
    }

    return data
  },

  // Update chat title
  updateChatTitle: async (chatId: string, title: string) => {
    const { data } = await api.patch(`/api/chat/conversations/${chatId}`, {
      title,
    })
    return data
  },

  // Search chats
  searchChats: async (query: string) => {
    const { data } = await api.get('/api/chat/search', {
      params: { q: query },
    })
    return data
  },

  // Export chat
  exportChat: async (chatId: string, format: 'json' | 'markdown' = 'json') => {
    const { data } = await api.get(`/api/chat/conversations/${chatId}/export`, {
      params: { format },
      responseType: format === 'json' ? 'json' : 'text',
    })
    return data
  },
}