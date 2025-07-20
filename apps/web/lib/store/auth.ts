import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { authApi } from '@/lib/api/auth'

interface User {
  id: string
  email: string
  name: string
  avatar?: string
  role: 'user' | 'admin'
  createdAt: string
}

interface AuthStore {
  user: User | null
  token: string | null
  isLoading: boolean
  error: string | null
  
  // Actions
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, name: string) => Promise<void>
  logout: () => void
  updateProfile: (data: Partial<User>) => Promise<void>
  checkAuth: () => Promise<void>
  clearError: () => void
}

export const useAuthStore = create<AuthStore>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isLoading: false,
      error: null,

      login: async (email, password) => {
        set({ isLoading: true, error: null })
        try {
          const { user, token } = await authApi.login(email, password)
          set({ user, token, isLoading: false })
        } catch (error: any) {
          set({ 
            error: error.message || 'Login failed', 
            isLoading: false 
          })
          throw error
        }
      },

      register: async (email, password, name) => {
        set({ isLoading: true, error: null })
        try {
          const { user, token } = await authApi.register(email, password, name)
          set({ user, token, isLoading: false })
        } catch (error: any) {
          set({ 
            error: error.message || 'Registration failed', 
            isLoading: false 
          })
          throw error
        }
      },

      logout: () => {
        authApi.logout()
        set({ user: null, token: null, error: null })
      },

      updateProfile: async (data) => {
        set({ isLoading: true, error: null })
        try {
          const updatedUser = await authApi.updateProfile(data)
          set({ user: updatedUser, isLoading: false })
        } catch (error: any) {
          set({ 
            error: error.message || 'Update failed', 
            isLoading: false 
          })
          throw error
        }
      },

      checkAuth: async () => {
        const token = get().token
        if (!token) return
        
        set({ isLoading: true })
        try {
          const user = await authApi.getMe()
          set({ user, isLoading: false })
        } catch (error) {
          set({ user: null, token: null, isLoading: false })
        }
      },

      clearError: () => {
        set({ error: null })
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ 
        user: state.user, 
        token: state.token 
      }),
    }
  )
)