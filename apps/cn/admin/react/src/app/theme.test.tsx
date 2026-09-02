import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it } from 'vitest'
import { ThemeProvider, useTheme } from './theme'

function ThemeControls() {
  const { mode, setMode } = useTheme()
  return <><span>{mode}</span><button onClick={() => setMode('dark')}>dark</button><button onClick={() => setMode('light')}>light</button><button onClick={() => setMode('system')}>system</button></>
}

describe('Admin theme readiness', () => {
  beforeEach(() => { localStorage.clear(); document.documentElement.classList.remove('dark') })

  it('switches System, Light, and Dark through the shared semantic theme provider', async () => {
    render(<ThemeProvider><ThemeControls /></ThemeProvider>)
    expect(screen.getByText('system', { selector: 'span' })).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'dark' }))
    expect(document.documentElement).toHaveClass('dark')
    await userEvent.click(screen.getByRole('button', { name: 'light' }))
    expect(document.documentElement).not.toHaveClass('dark')
    await userEvent.click(screen.getByRole('button', { name: 'system' }))
    expect(localStorage.getItem('gofurry-admin-theme')).toBe('system')
  })
})
