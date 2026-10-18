export type HeroMode = 'random' | 'fixed' | 'local'
export interface HeroPreference { mode: HeroMode; desktopId: string | null; mobileId: string | null }

export function normalizeHeroID(value: unknown): string | null {
  if (typeof value !== 'string' || !/^[1-9]\d{0,18}$/.test(value)) return null
  return BigInt(value) <= 9223372036854775807n ? value : null
}

export function normalizeHeroPreference(mode: unknown, desktopId: unknown, mobileId: unknown): HeroPreference {
  return { mode: mode === 'fixed' || mode === 'local' ? mode : 'random', desktopId: normalizeHeroID(desktopId), mobileId: normalizeHeroID(mobileId) }
}

export function heroQuery(preference: HeroPreference) {
  return preference.mode === 'local' ? { hero_mode: 'local' } : preference.mode === 'fixed'
    ? { hero_desktop_id: preference.desktopId ?? undefined, hero_mobile_id: preference.mobileId ?? undefined } : {}
}

export function heroPreferenceKey(preference: HeroPreference) {
  return `${preference.mode}:${preference.desktopId ?? ''}:${preference.mobileId ?? ''}`
}
