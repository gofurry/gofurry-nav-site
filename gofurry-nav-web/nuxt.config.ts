import tailwindcss from '@tailwindcss/vite'

const siteUrl = process.env.NUXT_PUBLIC_SITE_URL || 'http://localhost:3000'
const ragApiInternalBase = process.env.RAG_API_INTERNAL_BASE
  || process.env.NUXT_RAG_API_INTERNAL_BASE
  || (process.env.NODE_ENV === 'production' ? 'http://10.6.0.11:9997' : 'http://192.168.153.1:9997')

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
      'import.meta.env.VITE_NAV_API_BASE_URL': JSON.stringify(process.env.NUXT_PUBLIC_NAV_API_BASE || '/api'),
      'import.meta.env.VITE_GAME_API_BASE_URL': JSON.stringify(process.env.NUXT_PUBLIC_GAME_API_BASE || '/api'),
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
      htmlAttrs: {
        lang: 'zh-CN'
      },
      title: 'gofurry 兽人控导航站 - 发现兽人文化相关资源与社区',
      meta: [
        // SEO
        { name: 'description', content: 'gofurry 兽人控导航站是一个专注于兽人文化整合的导航站点，为兽人爱好者提供便捷的资源发现与社区入口。' },
        { name: 'keywords', content: 'furry, 兽人, 兽人控, 兽人导航, 兽人文化, 兽人社区, gofurry, 兽人资源, fur, furries, game, anthro, scalies, kemono' },
        // 搜索引擎 爬虫
        { name: 'robots', content: 'index, follow' },
        { name: 'googlebot', content: 'index, follow' },
        // 社交平台
        { property: 'og:site_name', content: 'gofurry 兽人控导航站' },
        { property: 'og:title', content: 'gofurry 兽人控导航站' },
        { property: 'og:description', content: 'gofurry 兽人控导航站是一个专注于兽人文化整合的导航站点，为兽人爱好者提供便捷的资源发现与社区入口。' },
        { property: 'og:type', content: 'website' },
        { property: 'og:url', content: siteUrl },
        { property: 'og:image', content: `${siteUrl.replace(/\/$/, '')}/og-image.jpg` },
        { name: 'theme-color', content: '#f97316' },
        // 移动端
        { name: 'format-detection', content: 'telephone=no' },
        { name: 'mobile-web-app-capable', content: 'yes' },
        { name: 'apple-mobile-web-app-capable', content: 'yes' },
        { name: 'apple-mobile-web-app-title', content: 'gofurry 兽人控导航站' }
      ],
      link: [
        { rel: 'icon', type: 'image/png', href: '/logo-mini.png' },
        { rel: 'apple-touch-icon', sizes: '180x180', href: '/logo-mini.png' },
        { rel: 'shortcut icon', href: '/logo-mini.png' }
      ]
    }
  },
  runtimeConfig: {
    navApiInternalBase: process.env.NAV_API_INTERNAL_BASE || process.env.NUXT_NAV_API_INTERNAL_BASE || 'http://192.168.153.1:9999/api',
    gameApiInternalBase: process.env.GAME_API_INTERNAL_BASE || process.env.NUXT_GAME_API_INTERNAL_BASE || 'http://192.168.153.1:9998/api',
    ragApiInternalBase,
    public: {
      siteUrl: process.env.NUXT_PUBLIC_SITE_URL || 'http://localhost:3000',
      navApiBase: process.env.NUXT_PUBLIC_NAV_API_BASE || '/api',
      gameApiBase: process.env.NUXT_PUBLIC_GAME_API_BASE || '/api',
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
    '/settings/**': { ssr: false },
    '/panel': { ssr: false }
  },
  i18n: {
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
