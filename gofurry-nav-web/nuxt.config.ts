import tailwindcss from '@tailwindcss/vite'

const siteUrl = process.env.NUXT_PUBLIC_SITE_URL || 'http://localhost:3000'
const normalizedSiteUrl = siteUrl.replace(/\/$/, '')
const siteName = 'GoFurry 兽人控导航站 - GoFurry Navigation'
const siteTitle = 'GoFurry 兽人控导航站 - 发现兽人文化相关资源与社区'
const siteDescription = 'GoFurry 是面向兽人文化爱好者的双语导航站，收录 Furry 社区、艺术、小说、游戏、工具与站点监测资源。GoFurry is a bilingual Furry navigation hub for communities, art, fiction, games, tools, and site monitoring.'
const siteKeywords = [
  'gofurry',
  'furry',
  'furries',
  'furry navigation',
  'furry community',
  'furry art',
  'furry games',
  'furry fiction',
  'anthro',
  'kemono',
  'scalies',
  '兽人',
  '兽人控',
  '福瑞',
  '毛茸茸',
  '兽人导航',
  '兽人文化',
  '兽人社区',
  '兽人资源',
  '兽游',
  '兽人游戏',
  '兽人小说',
  '兽人艺术'
].join(', ')
const ogImage = `${normalizedSiteUrl}/og-image.jpg`
const ragApiInternalBase = process.env.RAG_API_INTERNAL_BASE
  || process.env.NUXT_RAG_API_INTERNAL_BASE
  || (process.env.NODE_ENV === 'production' ? 'http://10.6.0.11:9997' : 'http://192.168.153.1:9997')
const publicNavApiBase = process.env.NUXT_PUBLIC_NAV_API_BASE || '/api/v1'
const navApiInternalBase = process.env.NAV_API_INTERNAL_BASE || process.env.NUXT_NAV_API_INTERNAL_BASE || 'http://192.168.153.1:9999/api/v1'

function deriveNavV2ApiBase(base: string) {
  if (base.includes('/api/v1')) {
    return base.replace('/api/v1', '/api/v2')
  }
  return `${base.replace(/\/$/, '')}/api/v2`
}

const publicNavV2ApiBase = process.env.NUXT_PUBLIC_NAV_V2_API_BASE || deriveNavV2ApiBase(publicNavApiBase)
const navV2ApiInternalBase = process.env.NAV_V2_API_INTERNAL_BASE || process.env.NUXT_NAV_V2_API_INTERNAL_BASE || deriveNavV2ApiBase(navApiInternalBase)

export default defineNuxtConfig({
  compatibilityDate: '2026-05-01',
  experimental: {
    appManifest: false
  },
  modules: ['@pinia/nuxt', '@nuxtjs/i18n'],
  css: ['~/assets/css/main.css'],
  vite: {
    plugins: [tailwindcss()],
    define: {
      'import.meta.env.VITE_NAV_API_BASE_URL': JSON.stringify(process.env.NUXT_PUBLIC_NAV_API_BASE || '/api/v1'),
      'import.meta.env.VITE_GAME_API_BASE_URL': JSON.stringify(process.env.NUXT_PUBLIC_GAME_API_BASE || '/api/v1'),
      'import.meta.env.VITE_SITE_LOGO_PREFIX_URL': JSON.stringify(process.env.NUXT_PUBLIC_SITE_LOGO_PREFIX_URL || 'https://qcdn.go-furry.com/nav/static/SiteLogos/'),
      'import.meta.env.VITE_SITE_DEFAULT_LOGO': JSON.stringify(process.env.NUXT_PUBLIC_SITE_DEFAULT_LOGO || 'https://qcdn.go-furry.com/nav/static/SiteLogos/defaultLogo.svg'),
      'import.meta.env.VITE_GAME_SITE_LOGO_PREFIX_URL': JSON.stringify(process.env.NUXT_PUBLIC_GAME_SITE_LOGO_PREFIX_URL || 'https://qcdn.go-furry.com/game/icons/'),
      'import.meta.env.VITE_GAME_PREFIX_URL': JSON.stringify(process.env.NUXT_PUBLIC_GAME_PREFIX_URL || 'https://qcdn.go-furry.com/game/'),
      'import.meta.env.VITE_STEAM_APP_PREFIX_URL': JSON.stringify(process.env.NUXT_PUBLIC_STEAM_APP_PREFIX_URL || 'https://store.steampowered.com/app/'),
      'import.meta.env.VITE_STEAM_COVER_PREFIX_URL': JSON.stringify(process.env.NUXT_PUBLIC_STEAM_COVER_PREFIX_URL || 'https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/')
    }
  },
  app: {
    head: {
      charset: 'utf-8',
      viewport: 'width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no',
      title: siteTitle,
      meta: [
        { name: 'description', content: siteDescription },
        { name: 'keywords', content: siteKeywords },
        { name: 'robots', content: 'index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1' },
        { name: 'googlebot', content: 'index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1' },
        { name: 'bingbot', content: 'index, follow, max-image-preview:large' },
        { name: 'author', content: 'GoFurry' },
        { name: 'application-name', content: 'GoFurry' },
        { name: 'apple-mobile-web-app-title', content: 'GoFurry' },
        { name: 'theme-color', content: '#f97316' },
        { name: 'color-scheme', content: 'light' },
        { name: 'format-detection', content: 'telephone=no' },
        { name: 'mobile-web-app-capable', content: 'yes' },
        { name: 'apple-mobile-web-app-capable', content: 'yes' },
        { name: 'apple-mobile-web-app-status-bar-style', content: 'default' },
        { property: 'og:site_name', content: siteName },
        { property: 'og:title', content: siteTitle },
        { property: 'og:description', content: siteDescription },
        { property: 'og:type', content: 'website' },
        { property: 'og:url', content: normalizedSiteUrl },
        { property: 'og:image', content: ogImage },
        { property: 'og:image:secure_url', content: ogImage },
        { property: 'og:image:type', content: 'image/jpeg' },
        { property: 'og:image:width', content: '1200' },
        { property: 'og:image:height', content: '630' },
        { property: 'og:image:alt', content: siteTitle },
        { property: 'og:locale', content: 'zh_CN' },
        { property: 'og:locale:alternate', content: 'en_US' },
        { name: 'twitter:card', content: 'summary_large_image' },
        { name: 'twitter:title', content: siteTitle },
        { name: 'twitter:description', content: siteDescription },
        { name: 'twitter:image', content: ogImage },
        { name: 'twitter:image:alt', content: siteTitle }
      ],
      link: [
        { rel: 'icon', type: 'image/png', href: '/logo-mini.png' },
        { rel: 'apple-touch-icon', sizes: '180x180', href: '/logo-mini.png' },
        { rel: 'shortcut icon', href: '/logo-mini.png' },
        { rel: 'manifest', href: '/manifest.webmanifest' }
      ],
      script: [
        {
          type: 'application/ld+json',
          textContent: JSON.stringify({
            '@context': 'https://schema.org',
            '@type': 'WebSite',
            name: 'GoFurry',
            alternateName: ['GoFurry 兽人控导航站', 'GoFurry Navigation'],
            url: normalizedSiteUrl,
            inLanguage: ['zh-CN', 'en-US'],
            description: siteDescription,
            image: ogImage,
            potentialAction: {
              '@type': 'SearchAction',
              target: `${normalizedSiteUrl}/nav?q={search_term_string}`,
              'query-input': 'required name=search_term_string'
            }
          })
        },
        {
          type: 'application/ld+json',
          textContent: JSON.stringify({
            '@context': 'https://schema.org',
            '@type': 'Organization',
            name: 'GoFurry',
            url: normalizedSiteUrl,
            logo: `${normalizedSiteUrl}/logo-mini.png`,
            sameAs: [
              'https://github.com/gofurry'
            ]
          })
        }
      ]
    }
  },
  runtimeConfig: {
    navApiInternalBase,
    navV2ApiInternalBase,
    gameApiInternalBase: process.env.GAME_API_INTERNAL_BASE || process.env.NUXT_GAME_API_INTERNAL_BASE || 'http://192.168.153.1:9998/api/v1',
    ragApiInternalBase,
    public: {
      siteUrl: process.env.NUXT_PUBLIC_SITE_URL || 'http://localhost:3000',
      navApiBase: publicNavApiBase,
      navV2ApiBase: publicNavV2ApiBase,
      gameApiBase: process.env.NUXT_PUBLIC_GAME_API_BASE || '/api/v1',
      siteLogoPrefixUrl: process.env.NUXT_PUBLIC_SITE_LOGO_PREFIX_URL || 'https://qcdn.go-furry.com/nav/static/SiteLogos/',
      siteDefaultLogo: process.env.NUXT_PUBLIC_SITE_DEFAULT_LOGO || 'https://qcdn.go-furry.com/nav/static/SiteLogos/defaultLogo.svg',
      gameSiteLogoPrefixUrl: process.env.NUXT_PUBLIC_GAME_SITE_LOGO_PREFIX_URL || 'https://qcdn.go-furry.com/game/icons/',
      gamePrefixUrl: process.env.NUXT_PUBLIC_GAME_PREFIX_URL || 'https://qcdn.go-furry.com/game/',
      steamAppPrefixUrl: process.env.NUXT_PUBLIC_STEAM_APP_PREFIX_URL || 'https://store.steampowered.com/app/',
      steamCoverPrefixUrl: process.env.NUXT_PUBLIC_STEAM_COVER_PREFIX_URL || 'https://shared.fastly.steamstatic.com/store_item_assets/steam/apps/'
    }
  },
  routeRules: {
    '/': { prerender: true },
    '/about': { prerender: true },
    '/nav': { ssr: true },
    '/sites': { ssr: true },
    '/sites/**': { ssr: true },
    '/site/**': { ssr: true },
    '/games': { ssr: true },
    '/games/**': { ssr: true },
    '/updates': { ssr: true },
    '/games/news/more': { ssr: true },
    '/games/search': { ssr: false },
    '/games/prize/**': { ssr: false },
    '/archive': { ssr: false },
    '/user/**': { ssr: false },
    '/settings/**': { ssr: false }
  },
  i18n: {
    baseUrl: normalizedSiteUrl,
    defaultLocale: 'zh',
    strategy: 'prefix_except_default',
    langDir: 'locales',
    compilation: {
      strictMessage: false,
      escapeHtml: false
    },
    locales: [
      {
        code: 'zh',
        name: '简体中文',
        language: 'zh-CN',
        file: 'zh.json'
      },
      {
        code: 'en',
        name: 'English',
        language: 'en-US',
        file: 'en.json'
      }
    ]
  }
})
