import type { InsightChange } from '@/types/insights'

const publicChangeI18nKeys: Record<string, string> = {
  'site.ipv6.enabled': 'siteIpv6Enabled',
  'site.ipv6.disabled': 'siteIpv6Disabled',
  'site.tls13.enabled': 'siteTls13Enabled',
  'site.tls13.disabled': 'siteTls13Disabled',
  'site.http2.enabled': 'siteHttp2Enabled',
  'site.http2.disabled': 'siteHttp2Disabled',
  'site.hsts.added': 'siteHstsAdded',
  'site.hsts.removed': 'siteHstsRemoved',
  'site.csp.added': 'siteCspAdded',
  'site.csp.removed': 'siteCspRemoved',
  'site.security_txt.added': 'siteSecurityTxtAdded',
  'site.security_txt.removed': 'siteSecurityTxtRemoved',
  'site.primary_target.changed': 'sitePrimaryTargetChanged',
  'site.tls_certificate.changed': 'siteTlsCertificateChanged',
  'site.tls_certificate.verification_failed': 'siteTlsCertificateVerificationFailed',
  'site.tls_certificate.verification_restored': 'siteTlsCertificateVerificationRestored',
  'game.free.enabled': 'gameFreeEnabled',
  'game.free.disabled': 'gameFreeDisabled',
  'game.windows.added': 'gameWindowsAdded',
  'game.windows.removed': 'gameWindowsRemoved',
  'game.linux.added': 'gameLinuxAdded',
  'game.linux.removed': 'gameLinuxRemoved',
  'game.mac.added': 'gameMacAdded',
  'game.mac.removed': 'gameMacRemoved',
  'game.release.available': 'gameReleaseAvailable',
  'game.release.withdrawn': 'gameReleaseWithdrawn',
  'game.release.changed': 'gameReleaseChanged',
  'game.price.increased': 'gamePriceIncreased',
  'game.price.decreased': 'gamePriceDecreased',
  'game.price.state_changed': 'gamePriceStateChanged',
  'game.price.currency_changed': 'gamePriceCurrencyChanged',
  'game.discount.started': 'gameDiscountStarted',
  'game.discount.ended': 'gameDiscountEnded',
  'game.discount.changed': 'gameDiscountChanged',
}

export function insightChangeI18nKey(type: string) {
  return `insights.changes.events.${publicChangeI18nKeys[type] ?? 'unknown'}`
}

export function insightChangeOrder(change: InsightChange) {
  const timestamp = Date.parse(change.occurred_at || `${change.date}T00:00:00Z`)
  return Number.isFinite(timestamp) ? timestamp : 0
}

export function formatInsightChangeWhen(change: InsightChange, locale: string) {
  if (!change.occurred_at) {
    return change.date
  }

  const value = new Date(change.occurred_at)
  if (Number.isNaN(value.getTime())) {
    return change.date
  }

  return new Intl.DateTimeFormat(locale === 'en' ? 'en-US' : 'zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(value)
}
