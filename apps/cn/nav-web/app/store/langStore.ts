import { defineStore } from 'pinia'
import { computed } from 'vue'

type SupportedLang = 'zh' | 'en'

const normalizeLang = (value: unknown): SupportedLang => value === 'en' ? 'en' : 'zh'

export const useLangStore = defineStore('lang', () => {
    const { locale, setLocale } = useI18n()
    const lang = computed<SupportedLang>(() => normalizeLang(locale.value))

    function setLang(newLang: SupportedLang) {
        const normalizedLang = normalizeLang(newLang)
        if (lang.value !== normalizedLang) {
            void setLocale(normalizedLang)
        }
    }

    return { lang, setLang }
})
