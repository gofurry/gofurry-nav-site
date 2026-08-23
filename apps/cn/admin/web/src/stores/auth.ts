import { defineStore } from 'pinia'
import { getJSON, resetCsrf, sendJSON } from '../api'
import type { AuthState } from '../types'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    initialized: false,
    authenticated: false,
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
      this.loaded = true
      if (!this.authenticated) {
        resetCsrf()
      }
    },
    async bootstrap(password: string) {
      await sendJSON('/api/v1/auth/bootstrap', 'POST', { password })
      this.initialized = true
      this.authenticated = false
      this.loaded = true
    },
    async login(password: string) {
      await sendJSON('/api/v1/auth/login', 'POST', { password })
      this.authenticated = true
      this.initialized = true
      this.loaded = true
    },
    async logout() {
      await sendJSON('/api/v1/auth/logout', 'POST', {})
      this.authenticated = false
      resetCsrf()
    },
  },
})
