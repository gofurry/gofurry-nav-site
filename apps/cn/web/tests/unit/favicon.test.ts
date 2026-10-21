import { describe, expect, it } from 'vitest'
import { getSiteFaviconUrl } from '../../app/utils/favicon'

describe('getSiteFaviconUrl', () => {
  it.each([
    ['https://example.com', 'example.com'],
    ['http://example.com/path/to/page', 'example.com'],
    ['https://user:pass@WWW.Example.com:8443/path?q=value#fragment', 'www.example.com'],
  ])('uses only the hostname of %s with explicit failure reporting', (input, hostname) => {
    expect(getSiteFaviconUrl(input)).toBe(`https://a.favicon.im/${hostname}?larger=true&throw-error-on-404=true`)
  })

  it.each(['', 'not a URL', 'https://', '/relative/path', 'javascript:alert(1)', 'data:image/png;base64,AA==', 'ftp://example.com/file'])('rejects invalid or non-website input %j', input => {
    expect(getSiteFaviconUrl(input)).toBeUndefined()
  })
})
