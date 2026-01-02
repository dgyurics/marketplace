export type Role = 'admin' | 'user' | 'guest' | 'staff' | 'member'

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
  created_at: string
  updated_at: string
}

export interface AuthTokens {
  token: string
  refresh_token: string
}
