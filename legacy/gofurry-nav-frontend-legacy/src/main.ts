import { createApp } from 'vue'
import App from '@/App.vue'
import '@/style.css'
import { router } from '@/router'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'



import zh from '@/locales/zh.json'
import en from '@/locales/en.json'

const savedLang = localStorage.getItem('lang') as 'zh' | 'en' | null
const initialLocale = savedLang || 'zh'

const messages = { zh, en }

const i18n = createI18n({
    legacy: false,
    locale: initialLocale,
    fallbackLocale: 'en',
    messages,
    globalInjection: true,
})




const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(i18n)
app.mount('#app')

export { i18n }
