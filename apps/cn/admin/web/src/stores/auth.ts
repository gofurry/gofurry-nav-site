import { defineStore } from 'pinia'
import { getJSON, resetCsrf, sendJSON } from '../api'
import type { AuthIdentity, AuthState } from '../types'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    initialized: false,
    authenticated: false,
    identity: null as AuthIdentity | null,
    loaded: false,
  }),
  actions: {
    async loadState(force = false) {
      if (this.loaded && !force) {
        return
      }
      const state = await getJSON<AuthState>('/api/v1/auth/state')
      this.initialized = state.initialized
      this.authenticated = state.authenticated
      this.identity = state.identity ?? null
      this.loaded = true
      if (!this.authenticated) {
        resetCsrf()
      }
    },
    async bootstrap(username: string, displayName: string, password: string) {
      await sendJSON('/api/v1/auth/bootstrap', 'POST', { username, display_name: displayName, password })
      this.initialized = true
      this.authenticated = false
      this.loaded = true
    },
    async login(username: string, password: string) {
      const state = await sendJSON<AuthState>('/api/v1/auth/login', 'POST', { username, password })
      this.authenticated = true
      this.initialized = true
      this.identity = state.identity ?? null
      this.loaded = true
    },
    async logout() {
      await sendJSON('/api/v1/auth/logout', 'POST', {})
      this.authenticated = false
      this.identity = null
      resetCsrf()
    },
    has(capability: string) {
      return this.identity?.capabilities.includes(capability) ?? false
    },
  },
})
