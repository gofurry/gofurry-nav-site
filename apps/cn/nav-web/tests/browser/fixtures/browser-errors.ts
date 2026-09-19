import type { Page } from '@playwright/test'

// Explicit per-page assertions keep failures with the scenario that caused them.
export function captureBrowserErrors(page: Page): string[] {
  const errors: string[] = []
  page.on('pageerror', error => errors.push(`pageerror: ${error.message}`))
  page.on('console', message => {
    if (message.type() === 'error' || /hydration.*mismatch|mismatch.*hydration/i.test(message.text())) {
      errors.push(`${message.type()}: ${message.text()}`)
    }
  })
  return errors
}
