import { expect, it } from 'vitest'
import { siteCapabilityKeys, siteCapabilityRegistry } from '../../app/utils/siteCapabilityRegistry'
import zh from '../../i18n/locales/zh.json'
import en from '../../i18n/locales/en.json'

it('owns all seven unique capabilities in stable presentation order', () => {
  expect(siteCapabilityKeys).toEqual(['ipv6', 'tls13', 'http2', 'hsts', 'csp', 'security_txt', 'certificate_verified'])
  expect(new Set(siteCapabilityKeys).size).toBe(7)
  expect(siteCapabilityRegistry.map(item => item.order)).toEqual([0, 1, 2, 3, 4, 5, 6])
})

it('provides valid categories and existing bilingual presentation metadata without API facts', () => {
  for (const item of siteCapabilityRegistry) {
    expect(['network', 'transport', 'web_policy']).toContain(item.category)
    for (const messages of [zh, en]) {
      const label = item.labelKey.split('.').reduce<unknown>((value, key) =>
        (value as Record<string, unknown>)[key], messages)
      expect(typeof label).toBe('string')
      expect(label).not.toBe('')
    }
    expect(Object.keys(item).sort()).toEqual(['category', 'key', 'labelKey', 'order', 'preview'])
  }
})

it('keeps the existing three-item preview explicitly separate from the complete catalog', () => {
  expect(siteCapabilityRegistry.filter(item => item.preview).map(item => item.key)).toEqual(['ipv6', 'tls13', 'security_txt'])
  expect(siteCapabilityRegistry.filter(item => !item.preview)).toHaveLength(4)
})
