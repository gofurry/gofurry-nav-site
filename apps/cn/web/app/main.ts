import { computed } from 'vue'

function getComposer() {
  return useNuxtApp().$i18n as any
}

export const i18n = {
  get global() {
    return {
      locale: computed({
        get() {
          const composer = getComposer()
          return typeof composer.locale === 'string' ? composer.locale : composer.locale.value
        },
        set(value: string) {
          const composer = getComposer()
          if (typeof composer.setLocale === 'function') {
            composer.setLocale(value)
            return
          }

          if (typeof composer.locale === 'string') {
            composer.locale = value
          } else {
            composer.locale.value = value
          }
        }
      }),
      t: (...args: any[]) => getComposer().t(...args)
    }
  }
}
