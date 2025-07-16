export type Role = 'admin' | 'user' | 'guest'

export interface User {
  user_id: string
  email: string
  role: Role
  exp: number
  iat: number
}

export interface AuthTokens {
  token: string
  refresh_token: string
  requires_setup: boolean
}
