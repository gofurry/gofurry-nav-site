import { ASSET_CDN_TTL_SECONDS, isAssetCDN, normalizeAssetMode, probeAssetCDNs, readAssetDiagnostics, resolveAssetPreferred, type AssetMode, type AssetOrigins } from '~/utils/managedAssets'
import { createRouteProbe, readRouteStorage, ROUTE_MODE_MAX_AGE, writeRouteStorage } from '~/utils/resourceRouting'

export default defineNuxtPlugin((nuxtApp) => {
  const config = useRuntimeConfig()
  const origins: AssetOrigins = { primary: String(config.public.assetPrimaryBase), mirror: String(config.public.assetMirrorBase) }
  const options = { path: '/', sameSite: 'lax' as const, secure: !import.meta.dev }
  const cookie = useCookie('gf_asset_cdn', { ...options, maxAge: ASSET_CDN_TTL_SECONDS })
  const modeCookie = useCookie('gf_asset_cdn_mode', { ...options, maxAge: ROUTE_MODE_MAX_AGE })
  const mode = useState<AssetMode>('managed-asset-mode', () => normalizeAssetMode(modeCookie.value))
  const recommendation = useState('managed-asset-recommendation', () => isAssetCDN(cookie.value) ? cookie.value : null)
  const storageKey = 'gf_asset_cdn_diagnostics'
  const runner = createRouteProbe(async () => {
    const result = await probeAssetCDNs(origins, fetch, () => performance.now(), recommendation.value || 'primary')
    recommendation.value = result.selected
    cookie.value = result.selected
    return result
  }, result => writeRouteStorage(storageKey, result))
  if (import.meta.client) nuxtApp.hook('app:mounted', () => {
    const stored = readAssetDiagnostics(readRouteStorage(storageKey))
    // Keep history when the short-lived cookie has expired. This cannot mutate
    // the already hydrated resource snapshots.
    if (!recommendation.value && stored) recommendation.value = stored.selected
    runner.start(stored)
  })
  return { provide: { assetCDN: {
    ...runner, mode, recommendation, origins,
    resolvePreferred: () => resolveAssetPreferred(mode.value, recommendation.value),
    saveMode(value: AssetMode) { mode.value = normalizeAssetMode(value); modeCookie.value = mode.value },
    reportFailure: () => runner.markStale(),
  } } }
})
