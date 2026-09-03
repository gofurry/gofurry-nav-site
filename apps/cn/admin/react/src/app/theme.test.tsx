import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it } from 'vitest'
import { ThemeProvider, useTheme } from './theme'

function ThemeControls() {
  const { mode, resolvedTheme, setMode } = useTheme()
  return <><span data-testid="mode">{mode}</span><span data-testid="resolved">{resolvedTheme}</span><button onClick={() => setMode('dark')}>dark</button><button onClick={() => setMode('light')}>light</button><button onClick={() => setMode('system')}>system</button></>
}

describe('Admin theme readiness', () => {
  beforeEach(() => { localStorage.clear(); document.documentElement.classList.remove('dark') })

  it('starts from the system preference and persists the first explicit Light or Dark choice', async () => {
    render(<ThemeProvider><ThemeControls /></ThemeProvider>)
    expect(screen.getByTestId('mode')).toHaveTextContent('system')
    expect(screen.getByTestId('resolved')).toHaveTextContent('light')
    expect(localStorage.getItem('gofurry-admin-theme')).toBeNull()
    await userEvent.click(screen.getByRole('button', { name: 'dark' }))
    expect(document.documentElement).toHaveClass('dark')
    expect(localStorage.getItem('gofurry-admin-theme')).toBe('dark')
    await userEvent.click(screen.getByRole('button', { name: 'light' }))
    expect(document.documentElement).not.toHaveClass('dark')
    expect(localStorage.getItem('gofurry-admin-theme')).toBe('light')
    await userEvent.click(screen.getByRole('button', { name: 'system' }))
    expect(localStorage.getItem('gofurry-admin-theme')).toBeNull()
  })
})
