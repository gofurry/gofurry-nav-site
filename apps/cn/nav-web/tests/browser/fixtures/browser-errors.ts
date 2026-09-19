import type { Page } from '@playwright/test'

// Explicit per-page assertions keep failures with the scenario that caused them.
export function captureBrowserErrors(page: Page, expectedNetworkFailures: ReadonlySet<string> = new Set()): string[] {
  const errors: string[] = []
  page.on('pageerror', error => errors.push(`pageerror: ${error.message}`))
  page.on('console', message => {
    // Fault-injection scenarios may abort specific requests. Only Chromium's
    // exact network diagnostic for an injected URL is expected; never suppress
    // application errors or hydration messages, even at the same URL.
    if (message.type() === 'error' && message.text() === 'Failed to load resource: net::ERR_FAILED'
      && expectedNetworkFailures.has(message.location().url)) return
    if (message.type() === 'error' || /hydration.*mismatch|mismatch.*hydration/i.test(message.text())) {
      errors.push(`${message.type()}: ${message.text()}`)
    }
  })
  return errors
}
