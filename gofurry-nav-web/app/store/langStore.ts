import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'
import type { Ref } from 'vue'

type SupportedLang = 'zh' | 'en'

const normalizeLang = (value: unknown): SupportedLang => value === 'en' ? 'en' : 'zh'

export const useLangStore = defineStore('lang', () => {
    const { locale, setLocale } = useI18n()
    const localeLang = computed<SupportedLang>(() => normalizeLang(locale.value))
    const lang: Ref<SupportedLang> = ref(localeLang.value)

    function persistLang(value: SupportedLang) {
        if (import.meta.client) {
            localStorage.setItem('lang', value)
        }
    }

    function setLang(newLang: SupportedLang) {
        if (lang.value !== newLang) {
            lang.value = newLang
        }

        persistLang(newLang)

        if (localeLang.value !== newLang) {
            setLocale(newLang)
        }
    }

    watch(
        localeLang,
        (newLang) => {
            lang.value = newLang
            persistLang(newLang)
        },
        { immediate: true }
    )

    return { lang, setLang }
})
