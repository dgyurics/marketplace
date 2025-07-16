import { jwtDecode } from 'jwt-decode'

import type { User } from '@/types/user'

/**
 * Retrieves the refresh token from local storage.
 * @returns {string | null} The refresh token if it exists, otherwise null.
 */
export function getRefreshToken(): string | null {
  return localStorage.getItem('token')
}

/**
 * Stores the refresh token in local storage.
 * @param {string} token - The refresh token to store.
 */
export function storeRefreshToken(token: string): void {
  localStorage.setItem('token', token)
}

/**
 * Removes the refresh token from local storage.
 */
export function removeRefreshToken(): void {
  localStorage.removeItem('token')
}

/**
 * Decodes a JWT token into a User object.
 * @throws {Error} If the token is invalid or malformed.
 */
export function decodeJWT(token: string): User {
  return jwtDecode<User>(token)
}

/**
 * Checks if the JWT token is expired.
 * @param {User} jwt - The decoded JWT token.
 * @param {number} [buffer=60] - Optional buffer in seconds to check before expiration.
 * @returns {boolean} True if the token is expired, false otherwise.
 */
export function isTokenExpired(jwt: User, buffer: number = 60): boolean {
  const now = Math.floor(Date.now() / 1000)
  return jwt.exp - now < buffer
}
