import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import type { Ref } from 'vue'
import { i18n } from '@/main'

export const useLangStore = defineStore('lang', () => {
    const lang: Ref<'zh' | 'en'> = ref(i18n.global.locale.value as 'zh' | 'en')

    function setLang(newLang: 'zh' | 'en') {
        lang.value = newLang
        i18n.global.locale.value = newLang
        localStorage.setItem('lang', newLang)
    }

    watch(lang, (newLang) => {
        i18n.global.locale.value = newLang
        localStorage.setItem('lang', newLang)
    })

    return { lang, setLang }
})
