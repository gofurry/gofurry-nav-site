import { isSteamGroup, normalizeSteamMode, resolveSteamPreferred, type SteamAssetMode, type SteamSharedCdnGroup } from '~/utils/steamAssets'
import { legacySteamRecommendation, probeSteamGroups, readSteamDiagnostics, STEAM_DIAGNOSTICS_KEY, STEAM_LEGACY_KEY } from '~/utils/steamAssetRouting'
import { createRouteProbe, readRouteStorage, ROUTE_MODE_MAX_AGE, ROUTE_TTL_MS, writeRouteStorage } from '~/utils/resourceRouting'
import type { Composer } from 'vue-i18n'

export default defineNuxtPlugin((nuxtApp) => {
  const options = { path: '/', sameSite: 'lax' as const, secure: !import.meta.dev }
  const modeCookie = useCookie('gf_steam_asset_mode', { ...options, maxAge: ROUTE_MODE_MAX_AGE })
  const groupCookie = useCookie('gf_steam_asset_group', { ...options, maxAge: ROUTE_TTL_MS / 1000 })
  const mode = useState<SteamAssetMode>('steam-asset-mode', () => normalizeSteamMode(modeCookie.value))
  const recommendation = useState<SteamSharedCdnGroup | null>('steam-asset-recommendation', () => isSteamGroup(groupCookie.value) ? groupCookie.value : null)
  const locale = () => (nuxtApp.$i18n as Composer).locale.value
  const runner = createRouteProbe(async manual => {
    const previous = resolveSteamPreferred('auto', recommendation.value, locale())
    const result = await probeSteamGroups(previous, undefined, Date.now, manual)
    recommendation.value = result.selected
    groupCookie.value = result.selected
    return result
  }, result => writeRouteStorage(STEAM_DIAGNOSTICS_KEY, result))
  if (import.meta.client) nuxtApp.hook('app:mounted', () => {
    // Do this after hydration so the SSR resource snapshots remain unchanged.
    const stored = readSteamDiagnostics(readRouteStorage(STEAM_DIAGNOSTICS_KEY))
    if (!recommendation.value && stored) recommendation.value = stored.selected
    const legacy = legacySteamRecommendation(readRouteStorage(STEAM_LEGACY_KEY))
    if (!recommendation.value && legacy) { recommendation.value = legacy; groupCookie.value = legacy }
    try { localStorage.removeItem(STEAM_LEGACY_KEY) } catch { /* Storage may be disabled. */ }
    runner.start(stored)
  })
  return { provide: { steamAssetRoute: {
    ...runner, mode, recommendation,
    resolvePreferred: (language?: string) => resolveSteamPreferred(mode.value, recommendation.value, language || locale()),
    saveMode(value: SteamAssetMode) { mode.value = normalizeSteamMode(value); modeCookie.value = mode.value },
    reportFailure: () => runner.markStale(),
  } } }
})
