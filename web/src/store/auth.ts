import { defineStore } from 'pinia'

import type { AuthTokens, JwtUser, Role } from '@/types'
import {
  decodeJWT,
  isTokenExpired as checkTokenExpired,
  storeRefreshToken,
  removeRefreshToken,
  getRefreshToken,
} from '@/utilities/auth'

const hierarchy: Record<Role, number> = {
  guest: 0,
  user: 1,
  member: 2,
  staff: 3,
  admin: 4,
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: {
      user_id: '',
      email: '',
      role: 'guest',
      exp: 0,
      iat: 0,
    } as JwtUser,
    accessToken: '',
    refreshToken: getRefreshToken(),
    loading: false,
    error: '',
  }),

  getters: {
    isAuthenticated: (state) => Boolean(state.user.user_id),
  },

  actions: {
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

    hasMinimumRole(requiredRole: Role): boolean {
      return hierarchy[this.user.role] >= hierarchy[requiredRole]
    },
  },
})
