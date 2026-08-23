export const useSiteAssets = () => {
  const config = useRuntimeConfig()

  return {
    siteLogoPrefixUrl: config.public.siteLogoPrefixUrl,
    siteDefaultLogo: config.public.siteDefaultLogo,
    gameSiteLogoPrefixUrl: config.public.gameSiteLogoPrefixUrl,
    gamePrefixUrl: config.public.gamePrefixUrl,
    steamAppPrefixUrl: config.public.steamAppPrefixUrl
  }
}
