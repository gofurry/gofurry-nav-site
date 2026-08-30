import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createContext, useContext, useEffect, type ReactNode } from 'react'
import { getJSON, resetCsrf, sendJSON } from '../../lib/api'
import type { AuthState } from '../../lib/types'

type AuthContextValue = {
  state?: AuthState
  loading: boolean
  error: Error | null
  can: (capability: string) => boolean
  reload: () => Promise<unknown>
  login: (input: { username: string; password: string }) => Promise<AuthState>
  bootstrap: (input: { username: string; display_name: string; password: string }) => Promise<unknown>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const client = useQueryClient()
  const stateQuery = useQuery({ queryKey: ['auth-state'], queryFn: () => getJSON<AuthState>('/api/v1/auth/state'), staleTime: 10_000 })
  const loginMutation = useMutation({ mutationFn: (input: { username: string; password: string }) => sendJSON<AuthState>('/api/v1/auth/login', 'POST', input) })
  const bootstrapMutation = useMutation({ mutationFn: (input: { username: string; display_name: string; password: string }) => sendJSON('/api/v1/auth/bootstrap', 'POST', input) })

  useEffect(() => {
    const onUnauthorized = () => {
      resetCsrf()
      client.setQueryData<AuthState>(['auth-state'], (current) => ({ initialized: current?.initialized ?? true, authenticated: false }))
    }
    window.addEventListener('gofurry:unauthorized', onUnauthorized)
    return () => window.removeEventListener('gofurry:unauthorized', onUnauthorized)
  }, [client])

  const value: AuthContextValue = {
    state: stateQuery.data,
    loading: stateQuery.isLoading,
    error: stateQuery.error,
    can: (capability) => stateQuery.data?.identity?.capabilities.includes(capability) ?? false,
    reload: () => stateQuery.refetch(),
    login: async (input) => {
      const state = await loginMutation.mutateAsync(input)
      client.setQueryData(['auth-state'], state)
      return state
    },
    bootstrap: async (input) => {
      const result = await bootstrapMutation.mutateAsync(input)
      client.setQueryData<AuthState>(['auth-state'], { initialized: true, authenticated: false })
      return result
    },
    logout: async () => {
      await sendJSON('/api/v1/auth/logout', 'POST', {})
      resetCsrf()
      client.clear()
      client.setQueryData<AuthState>(['auth-state'], { initialized: true, authenticated: false })
    },
  }
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) throw new Error('useAuth must be used within AuthProvider')
  return context
}
