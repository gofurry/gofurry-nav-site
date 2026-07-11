import { defineStore } from 'pinia'
import * as api from '../api'

export const useSessionStore = defineStore('session', {
  state: () => ({
    authenticated: false,
    initialized: false,
    loading: false,
    error: '',
  }),
  actions: {
    async refresh() {
      this.loading = true
      this.error = ''
      try {
        const state = await api.authState()
        this.authenticated = state.authenticated
        this.initialized = state.initialized
      } catch (error) {
        this.authenticated = false
        this.error = error instanceof Error ? error.message : '认证状态不可用'
      } finally {
        this.loading = false
      }
    },
    async login(passcode: string) {
      this.loading = true
      this.error = ''
      try {
        const state = await api.login(passcode)
        this.authenticated = state.authenticated
        this.initialized = true
      } catch (error) {
        this.error = error instanceof Error ? error.message : '登录失败'
        throw error
      } finally {
        this.loading = false
      }
    },
    async logout() {
      await api.logout()
      this.authenticated = false
    },
  },
})
