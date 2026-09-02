import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'

export type ThemeMode = 'system' | 'light' | 'dark'
type ThemeContextValue = { mode: ThemeMode; setMode: (mode: ThemeMode) => void }
const ThemeContext = createContext<ThemeContextValue | null>(null)

function applyTheme(mode: ThemeMode) {
  const dark = mode === 'dark' || (mode === 'system' && matchMedia('(prefers-color-scheme: dark)').matches)
  document.documentElement.classList.toggle('dark', dark)
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [mode, setModeState] = useState<ThemeMode>(() => {
    const stored = localStorage.getItem('gofurry-admin-theme')
    return stored === 'light' || stored === 'dark' ? stored : 'system'
  })
  useEffect(() => {
    applyTheme(mode)
    const media = matchMedia('(prefers-color-scheme: dark)')
    const onChange = () => mode === 'system' && applyTheme(mode)
    media.addEventListener('change', onChange)
    return () => media.removeEventListener('change', onChange)
  }, [mode])
  const value = useMemo(() => ({ mode, setMode: (next: ThemeMode) => {
    setModeState(next)
    localStorage.setItem('gofurry-admin-theme', next)
  } }), [mode])
  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}

export function useTheme() {
  const context = useContext(ThemeContext)
  if (!context) throw new Error('useTheme must be used within ThemeProvider')
  return context
}
