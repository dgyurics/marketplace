import { defineStore } from 'pinia'

import type { AuthTokens, User } from '@/types'
import {
  decodeJWT,
  isTokenExpired as checkTokenExpired,
  storeRefreshToken,
  removeRefreshToken,
  getRefreshToken,
} from '@/utilities/auth'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: {
      user_id: '',
      email: '',
      role: 'guest',
      exp: 0,
      iat: 0,
    } as User,
    accessToken: '',
    refreshToken: getRefreshToken(),
    loading: false,
    error: '',
  }),

  getters: {
    isAuthenticated: (state) => Boolean(state.user.user_id),
    isAdmin: (state) => state.user.role === 'admin',
  },

  actions: {
    setUser(user: User) {
      this.user = user
    },

    isTokenExpired() {
      return checkTokenExpired(this.user)
    },

    setTokens({ token, refresh_token: refreshToken }: AuthTokens) {
      try {
        this.accessToken = token
        this.refreshToken = refreshToken
        const decoded = decodeJWT(token)
        Object.assign(this.user, decoded)
        storeRefreshToken(refreshToken)
      } catch {
        this.clearTokens()
        throw new Error('Invalid access token')
      }
    },

    clearTokens() {
      this.accessToken = ''
      this.refreshToken = null
      this.user = {
        user_id: '',
        email: '',
        role: 'guest',
        exp: 0,
        iat: 0,
      }
      removeRefreshToken()
    },
  },
})
