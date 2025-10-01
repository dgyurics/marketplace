export type Role = 'admin' | 'user' | 'guest'

export interface JwtUser {
  user_id: string
  email: string
  role: Role
  exp: number
  iat: number
}

export interface UserRecord {
  id: string
  email: string
  role: Role
  requires_setup?: boolean
  created_at: string
  updated_at: string
}

export interface AuthTokens {
  token: string
  refresh_token: string
  requires_setup: boolean
}
