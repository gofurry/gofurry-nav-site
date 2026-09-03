import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'

export type ThemeMode = 'system' | 'light' | 'dark'
export type ResolvedTheme = Exclude<ThemeMode, 'system'>
type ThemeContextValue = { mode: ThemeMode; resolvedTheme: ResolvedTheme; setMode: (mode: ThemeMode) => void }
const ThemeContext = createContext<ThemeContextValue | null>(null)

function systemPrefersDark() {
  return matchMedia('(prefers-color-scheme: dark)').matches
}

function applyTheme(theme: ResolvedTheme) {
  document.documentElement.classList.toggle('dark', theme === 'dark')
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [mode, setModeState] = useState<ThemeMode>(() => {
    const stored = localStorage.getItem('gofurry-admin-theme')
    return stored === 'light' || stored === 'dark' ? stored : 'system'
  })
  const [systemDark, setSystemDark] = useState(systemPrefersDark)
  const resolvedTheme: ResolvedTheme = mode === 'system' ? (systemDark ? 'dark' : 'light') : mode
  useEffect(() => {
    const media = matchMedia('(prefers-color-scheme: dark)')
    const onChange = () => setSystemDark(media.matches)
    media.addEventListener('change', onChange)
    return () => media.removeEventListener('change', onChange)
  }, [])
  useEffect(() => applyTheme(resolvedTheme), [resolvedTheme])
  const value = useMemo(() => ({ mode, resolvedTheme, setMode: (next: ThemeMode) => {
    setModeState(next)
    if (next === 'system') localStorage.removeItem('gofurry-admin-theme')
    else localStorage.setItem('gofurry-admin-theme', next)
  } }), [mode, resolvedTheme])
  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}

export function useTheme() {
  const context = useContext(ThemeContext)
  if (!context) throw new Error('useTheme must be used within ThemeProvider')
  return context
}
