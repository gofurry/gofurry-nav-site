import type { SiteInsightCapabilityKey } from '../types/insights'

export type SiteCapabilityCategory = 'network' | 'transport' | 'web_policy'

interface SiteCapabilityPresentation {
  key: SiteInsightCapabilityKey
  category: SiteCapabilityCategory
  order: number
  labelKey: string
}

export const siteCapabilityRegistry = [
  { key: 'ipv6', category: 'network', order: 0, labelKey: 'insights.metrics.ipv6.name' },
  { key: 'tls13', category: 'transport', order: 1, labelKey: 'insights.metrics.tls13.name' },
  { key: 'http2', category: 'network', order: 2, labelKey: 'insights.metrics.http2.name' },
  { key: 'hsts', category: 'web_policy', order: 3, labelKey: 'insights.metrics.hsts.name' },
  { key: 'csp', category: 'web_policy', order: 4, labelKey: 'insights.metrics.csp.name' },
  { key: 'security_txt', category: 'web_policy', order: 5, labelKey: 'insights.metrics.security_txt.name' },
  { key: 'certificate_verified', category: 'transport', order: 6, labelKey: 'insights.metrics.certificate_verified.name' },
] as const satisfies readonly SiteCapabilityPresentation[]

export const siteCapabilityKeys = siteCapabilityRegistry.map(capability => capability.key)
