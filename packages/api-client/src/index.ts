import axios, { AxiosInstance, AxiosRequestConfig } from 'axios'
import { QueryClient } from '@tanstack/react-query'
import { z } from 'zod'

// API Configuration
export interface APIConfig {
  baseURL: string
  timeout?: number
  headers?: Record<string, string>
  onTokenRefresh?: () => Promise<string>
}

// Base API Client Class
export class APIClient {
  private instance: AxiosInstance
  private config: APIConfig

  constructor(config: APIConfig) {
    this.config = config
    this.instance = axios.create({
      baseURL: config.baseURL,
      timeout: config.timeout || 30000,
      headers: {
        'Content-Type': 'application/json',
        ...config.headers,
      },
    })

    this.setupInterceptors()
  }

  private setupInterceptors() {
    // Request interceptor for auth
    this.instance.interceptors.request.use(
      async (config) => {
        const token = await this.getAuthToken()
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => Promise.reject(error)
    )

    // Response interceptor for error handling
    this.instance.interceptors.response.use(
      (response) => response,
      async (error) => {
        if (error.response?.status === 401 && this.config.onTokenRefresh) {
          const newToken = await this.config.onTokenRefresh()
          error.config.headers.Authorization = `Bearer ${newToken}`
          return this.instance(error.config)
        }
        return Promise.reject(this.handleError(error))
      }
    )
  }

  private async getAuthToken(): Promise<string | null> {
    // Implementation depends on your auth strategy
    return localStorage.getItem('auth_token')
  }

  private handleError(error: any): Error {
    if (error.response) {
      const message = error.response.data?.message || error.message
      return new APIError(message, error.response.status, error.response.data)
    }
    if (error.request) {
      return new APIError('Network error - no response received', 0)
    }
    return new APIError(error.message || 'Unknown error')
  }

  // Generic request methods with type safety
  async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.instance.get<T>(url, config)
    return response.data
  }

  async post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.instance.post<T>(url, data, config)
    return response.data
  }

  async put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.instance.put<T>(url, data, config)
    return response.data
  }

  async patch<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.instance.patch<T>(url, data, config)
    return response.data
  }

  async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.instance.delete<T>(url, config)
    return response.data
  }
}

// Custom Error Class
export class APIError extends Error {
  constructor(
    message: string,
    public statusCode?: number,
    public data?: any
  ) {
    super(message)
    this.name = 'APIError'
  }
}

// Service-specific clients
export class AuthService {
  constructor(private client: APIClient) {}

  async login(credentials: LoginCredentials) {
    return this.client.post<AuthResponse>('/auth/login', credentials)
  }

  async logout() {
    return this.client.post('/auth/logout')
  }

  async refreshToken(refreshToken: string) {
    return this.client.post<AuthResponse>('/auth/refresh', { refreshToken })
  }

  async getProfile() {
    return this.client.get<UserProfile>('/auth/profile')
  }
}

export class ChatService {
  constructor(private client: APIClient) {}

  async getConversations(params?: ConversationParams) {
    return this.client.get<Conversation[]>('/chat/conversations', { params })
  }

  async getConversation(id: string) {
    return this.client.get<Conversation>(`/chat/conversations/${id}`)
  }

  async createConversation(data: CreateConversationDTO) {
    return this.client.post<Conversation>('/chat/conversations', data)
  }

  async sendMessage(conversationId: string, message: SendMessageDTO) {
    return this.client.post<Message>(`/chat/conversations/${conversationId}/messages`, message)
  }

  async deleteConversation(id: string) {
    return this.client.delete(`/chat/conversations/${id}`)
  }
}

export class ProductService {
  constructor(private client: APIClient) {}

  async searchProducts(query: string, filters?: ProductFilters) {
    return this.client.get<ProductSearchResult>('/products/search', {
      params: { q: query, ...filters }
    })
  }

  async getProduct(id: string) {
    return this.client.get<Product>(`/products/${id}`)
  }

  async getRecommendations(productId: string) {
    return this.client.get<Product[]>(`/products/${productId}/recommendations`)
  }

  async compareProducts(ids: string[]) {
    return this.client.post<ProductComparison>('/products/compare', { ids })
  }
}

// Type Definitions
export interface LoginCredentials {
  email: string
  password: string
}

export interface AuthResponse {
  accessToken: string
  refreshToken: string
  expiresIn: number
  user: UserProfile
}

export interface UserProfile {
  id: string
  email: string
  name: string
  avatar?: string
  preferences?: UserPreferences
}

export interface UserPreferences {
  theme: 'light' | 'dark' | 'system'
  language: string
  notifications: boolean
}

export interface Conversation {
  id: string
  title: string
  messages: Message[]
  createdAt: string
  updatedAt: string
}

export interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  attachments?: Attachment[]
  productRecommendations?: Product[]
  createdAt: string
}

export interface Attachment {
  id: string
  type: 'image' | 'document'
  url: string
  name: string
  size: number
}

export interface Product {
  id: string
  name: string
  description: string
  price: number
  currency: string
  images: string[]
  store: string
  rating: number
  reviews: number
  features: string[]
  availability: 'in_stock' | 'out_of_stock' | 'limited'
}

export interface ProductSearchResult {
  products: Product[]
  total: number
  page: number
  pageSize: number
  facets: SearchFacets
}

export interface SearchFacets {
  brands: FacetValue[]
  priceRanges: FacetValue[]
  categories: FacetValue[]
  ratings: FacetValue[]
}

export interface FacetValue {
  value: string
  count: number
}

export interface ProductFilters {
  brand?: string[]
  minPrice?: number
  maxPrice?: number
  category?: string[]
  minRating?: number
  sortBy?: 'price_asc' | 'price_desc' | 'rating' | 'popularity'
}

export interface ProductComparison {
  products: Product[]
  features: ComparisonFeature[]
}

export interface ComparisonFeature {
  name: string
  values: Record<string, any>
}

export interface ConversationParams {
  page?: number
  limit?: number
  search?: string
}

export interface CreateConversationDTO {
  title?: string
  initialMessage?: string
}

export interface SendMessageDTO {
  content: string
  attachments?: string[]
}

// Factory function to create API client
export function createAPIClient(config: APIConfig) {
  const client = new APIClient(config)
  
  return {
    auth: new AuthService(client),
    chat: new ChatService(client),
    products: new ProductService(client),
    client,
  }
}

// React Query integration
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 3,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      staleTime: 5 * 60 * 1000, // 5 minutes
      gcTime: 10 * 60 * 1000, // 10 minutes
    },
    mutations: {
      retry: 1,
    },
  },
})

// Export everything
export * from './hooks'
export * from './schemas'