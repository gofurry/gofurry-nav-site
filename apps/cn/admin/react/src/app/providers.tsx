import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useState, type ReactNode } from 'react'
import { AuthProvider } from '../features/auth/auth-context'
import { ThemeProvider } from './theme'
import { ToastProvider } from './toast'

export function AppProviders({ children }: { children: ReactNode }) {
  const [queryClient] = useState(() => new QueryClient({
    defaultOptions: {
      queries: { staleTime: 20_000, retry: 1, refetchOnWindowFocus: false },
      mutations: { retry: false },
    },
  }))
  return <QueryClientProvider client={queryClient}>
    <ThemeProvider><ToastProvider><AuthProvider>{children}</AuthProvider></ToastProvider></ThemeProvider>
  </QueryClientProvider>
}
