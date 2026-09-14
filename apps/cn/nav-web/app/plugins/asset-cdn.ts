import { ASSET_CDN_TTL_SECONDS, probeAssetCDNs, type AssetCDN, type AssetOrigins } from '~/utils/managedAssets'

export default defineNuxtPlugin((nuxtApp) => {
  const config = useRuntimeConfig()
  const origins: AssetOrigins = { primary: String(config.public.assetPrimaryBase), mirror: String(config.public.assetMirrorBase) }
  const cookie = useCookie<AssetCDN | null>('gf_asset_cdn', { path: '/', sameSite: 'lax', maxAge: ASSET_CDN_TTL_SECONDS, secure: !import.meta.dev })
  const provider = useState<AssetCDN>('managed-asset-cdn', () => cookie.value === 'mirror' ? 'mirror' : 'primary')
  let pending: Promise<void> | null = null
  let scheduled = false
  const probe = () => {
    if (!import.meta.client) return Promise.resolve()
    if (pending) return pending
    pending = probeAssetCDNs(origins).then((result) => {
      provider.value = result.selected
      cookie.value = result.selected
      try { localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify(result)) } catch { /* Browser storage may be disabled. */ }
    }).finally(() => { pending = null })
    return pending
  }
  const schedule = () => {
    if (!import.meta.client || scheduled || pending) return
    scheduled = true
    const run = () => { scheduled = false; void probe() }
    if ('requestIdleCallback' in window) window.requestIdleCallback(run, { timeout: 2000 })
    else setTimeout(run, 250)
  }
  const invalidate = () => { if (import.meta.client) { cookie.value = null; schedule() } }
  if (import.meta.client) nuxtApp.hook('app:mounted', () => { if (cookie.value !== 'primary' && cookie.value !== 'mirror') schedule() })
  return { provide: { assetCDN: { provider, origins, invalidate } } }
})
