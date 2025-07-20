import axios from 'axios'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
const KEYCLOAK_URL = process.env.NEXT_PUBLIC_KEYCLOAK_URL || 'http://localhost:8082'
const KEYCLOAK_REALM = process.env.NEXT_PUBLIC_KEYCLOAK_REALM || 'shopmindai'
const KEYCLOAK_CLIENT_ID = process.env.NEXT_PUBLIC_KEYCLOAK_CLIENT_ID || 'shopmindai-web'

const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

export const authApi = {
  // Login via Keycloak
  login: async (email: string, password: string) => {
    const { data } = await axios.post(
      `${KEYCLOAK_URL}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token`,
      new URLSearchParams({
        client_id: KEYCLOAK_CLIENT_ID,
        grant_type: 'password',
        username: email,
        password: password,
      }),
      {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
      }
    )

    // Get user info
    const userInfo = await authApi.getUserInfo(data.access_token)

    return {
      user: {
        id: userInfo.sub,
        email: userInfo.email,
        name: userInfo.name || userInfo.preferred_username,
        role: userInfo.realm_access?.roles?.includes('admin') ? 'admin' : 'user',
        createdAt: new Date().toISOString(),
      },
      token: data.access_token,
      refreshToken: data.refresh_token,
    }
  },

  // Register new user
  register: async (email: string, password: string, name: string) => {
    // First register via our API (which will create user in Keycloak)
    const { data } = await api.post('/api/auth/register', {
      email,
      password,
      name,
    })

    // Then login to get tokens
    return authApi.login(email, password)
  },

  // Refresh token
  refreshToken: async (refreshToken: string) => {
    const { data } = await axios.post(
      `${KEYCLOAK_URL}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token`,
      new URLSearchParams({
        client_id: KEYCLOAK_CLIENT_ID,
        grant_type: 'refresh_token',
        refresh_token: refreshToken,
      }),
      {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
      }
    )

    return {
      token: data.access_token,
      refreshToken: data.refresh_token,
    }
  },

  // Logout
  logout: async (refreshToken?: string) => {
    if (refreshToken) {
      try {
        await axios.post(
          `${KEYCLOAK_URL}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/logout`,
          new URLSearchParams({
            client_id: KEYCLOAK_CLIENT_ID,
            refresh_token: refreshToken,
          }),
          {
            headers: {
              'Content-Type': 'application/x-www-form-urlencoded',
            },
          }
        )
      } catch (error) {
        console.error('Logout error:', error)
      }
    }
  },

  // Get user info from token
  getUserInfo: async (token: string) => {
    const { data } = await axios.get(
      `${KEYCLOAK_URL}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/userinfo`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    )
    return data
  },

  // Get current user
  getMe: async () => {
    const { data } = await api.get('/api/auth/me')
    return data
  },

  // Update profile
  updateProfile: async (updates: any) => {
    const { data } = await api.patch('/api/auth/profile', updates)
    return data
  },

  // Change password
  changePassword: async (currentPassword: string, newPassword: string) => {
    const { data } = await api.post('/api/auth/change-password', {
      currentPassword,
      newPassword,
    })
    return data
  },

  // OAuth providers
  getOAuthUrl: (provider: 'google' | 'github') => {
    return `${KEYCLOAK_URL}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/auth?` +
      new URLSearchParams({
        client_id: KEYCLOAK_CLIENT_ID,
        redirect_uri: `${window.location.origin}/auth/callback`,
        response_type: 'code',
        scope: 'openid email profile',
        kc_idp_hint: provider,
      }).toString()
  },

  // Handle OAuth callback
  handleOAuthCallback: async (code: string) => {
    const { data } = await api.post('/api/auth/oauth/callback', { code })
    return data
  },
}