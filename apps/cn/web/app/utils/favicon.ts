export function getSiteFaviconUrl(url: string): string | undefined {
  try {
    const { protocol, hostname } = new URL(url)
    if ((protocol !== 'https:' && protocol !== 'http:') || !hostname) return undefined
    return `https://a.favicon.im/${hostname}?larger=true&throw-error-on-404=true`
  } catch {
    return undefined
  }
}
