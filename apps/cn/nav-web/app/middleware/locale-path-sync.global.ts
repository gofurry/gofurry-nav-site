type SupportedLocale = 'zh' | 'en'

type I18nRuntime = {
  locale?: string | { value?: string }
  loadLocaleMessages?: (locale: SupportedLocale) => Promise<unknown>
  setLocaleCookie?: (locale: SupportedLocale) => void
}

export default defineNuxtRouteMiddleware(async (to) => {
  const routeLocale = localeFromPath(to.path)
  const i18n = useNuxtApp().$i18n as I18nRuntime

  if (readRuntimeLocale(i18n) !== routeLocale) {
    if (typeof i18n.loadLocaleMessages === 'function') {
      try {
        await i18n.loadLocaleMessages(routeLocale)
      } catch {
        // Do not block routing; the URL locale is still the source of truth.
      }
    }
    writeRuntimeLocale(i18n, routeLocale)
  }

  i18n.setLocaleCookie?.(routeLocale)
  syncLegacyLangStorage(routeLocale)
})

function localeFromPath(path: string): SupportedLocale {
  return path === '/en' || path.startsWith('/en/') ? 'en' : 'zh'
}

function readRuntimeLocale(i18n: I18nRuntime): string {
  if (typeof i18n.locale === 'string') {
    return i18n.locale
  }
  return i18n.locale?.value ?? ''
}

function writeRuntimeLocale(i18n: I18nRuntime, locale: SupportedLocale) {
  if (typeof i18n.locale === 'string') {
    i18n.locale = locale
    return
  }

  if (i18n.locale && typeof i18n.locale === 'object') {
    i18n.locale.value = locale
  }
}

function syncLegacyLangStorage(locale: SupportedLocale) {
  if (!import.meta.client) {
    return
  }

  try {
    window.localStorage.setItem('lang', locale)
  } catch {
    // Storage may be unavailable in private mode; route locale remains authoritative.
  }
}
