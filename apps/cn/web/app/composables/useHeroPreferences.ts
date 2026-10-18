import { heroPreferenceKey, normalizeHeroPreference, type HeroPreference } from '~/utils/heroPreferences'

export function useHeroPreferences() {
  // Decode as strings: Nuxt's default JSON decoder would round bigint IDs.
  const options = { path: '/', sameSite: 'lax' as const, maxAge: 365 * 24 * 60 * 60, encode: (value: string | null) => encodeURIComponent(value ?? ''), decode: decodeURIComponent }
  const modeCookie = useCookie<string | null>('gf_hero_mode', options)
  const desktopCookie = useCookie<string | null>('gf_hero_desktop_id', options)
  const mobileCookie = useCookie<string | null>('gf_hero_mobile_id', options)
  const preference = useState<HeroPreference>('hero-preference', () => normalizeHeroPreference(modeCookie.value, desktopCookie.value, mobileCookie.value))
  const legacyPending = useState('hero-legacy-pending', () => modeCookie.value == null)
  const revision = useState('hero-preference-revision', () => 0)

  function save(value: HeroPreference, localChanged = false) {
    const next = normalizeHeroPreference(value.mode, value.desktopId, value.mobileId)
    const changed = heroPreferenceKey(next) !== heroPreferenceKey(preference.value) || (next.mode === 'local' && localChanged)
    modeCookie.value = next.mode
    desktopCookie.value = next.desktopId
    mobileCookie.value = next.mobileId
    preference.value = next
    legacyPending.value = false
    if (changed) revision.value++
  }
  return { preference, revision, legacyPending, save }
}
