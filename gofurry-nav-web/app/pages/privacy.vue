<template>
  <div :class="pageClass" :style="pageVars">
    <GoFurryGridBackground />
    <div class="pointer-events-none absolute inset-x-0 top-0 h-48 bg-[var(--legal-top-veil)]" />

    <main class="relative mx-auto flex w-full max-w-5xl flex-1 flex-col px-4 py-8 sm:px-6 md:px-8 md:py-12">
      <article class="legal-panel">
        <div class="text-xs uppercase tracking-[0.28em] text-[var(--legal-accent)]">
          {{ content.kicker }}
        </div>
        <h1 class="mt-4 text-3xl font-semibold leading-tight text-[var(--legal-heading)] md:text-5xl">
          {{ content.title }}
        </h1>
        <p class="mt-5 max-w-3xl text-sm leading-7 text-[var(--legal-muted)] md:text-base">
          {{ content.summary }}
        </p>
        <div class="mt-6 text-xs text-[var(--legal-subtle)]">
          {{ content.updated }}
        </div>

        <div class="mt-8 divide-y divide-[var(--legal-rule)]">
          <section
            v-for="section in content.sections"
            :key="section.title"
            class="py-6 first:pt-0 last:pb-0"
        >
          <h2 class="text-xl font-semibold text-[var(--legal-heading)]">
            {{ section.title }}
          </h2>
          <p class="mt-3 text-sm leading-7 text-[var(--legal-muted)]">
            {{ section.body }}
          </p>
          </section>
        </div>
      </article>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { i18n } from '@/main'
import { useThemeStore } from '@/stores/theme'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'

const themeStore = useThemeStore()
const isZh = computed(() => i18n.global.locale.value === 'zh')
const isDark = computed(() => themeStore.theme === 'dark')

onMounted(() => {
  themeStore.initTheme()
})

const pageClass = computed(() => [
  'relative isolate flex w-full flex-1 flex-col overflow-hidden transition-colors duration-500',
  isDark.value ? 'bg-[#08101b] text-slate-100' : 'bg-[#f6ebdc] text-slate-950'
])

const pageVars = computed(() => isDark.value
  ? {
      '--legal-surface': 'rgba(9, 16, 27, 0.74)',
      '--legal-border': 'rgba(123, 154, 189, 0.2)',
      '--legal-rule': 'rgba(123, 154, 189, 0.16)',
      '--legal-heading': 'rgb(239 246 255)',
      '--legal-muted': 'rgb(179 195 214)',
      '--legal-subtle': 'rgb(100 116 139)',
      '--legal-accent': 'rgb(142 214 255)',
      '--legal-top-veil': 'linear-gradient(180deg, rgba(7, 13, 23, 0.46), rgba(7, 13, 23, 0))'
    }
  : {
      '--legal-surface': 'rgba(255, 249, 241, 0.76)',
      '--legal-border': 'rgba(168, 112, 46, 0.18)',
      '--legal-rule': 'rgba(168, 112, 46, 0.16)',
      '--legal-heading': 'rgb(15 23 42)',
      '--legal-muted': 'rgb(71 85 105)',
      '--legal-subtle': 'rgb(100 116 139)',
      '--legal-accent': 'rgb(190 112 28)',
      '--legal-top-veil': 'linear-gradient(180deg, rgba(255, 248, 239, 0.54), rgba(255, 248, 239, 0))'
    }
)

const content = computed(() => (
  isZh.value
    ? {
        kicker: 'Privacy Policy',
        title: '隐私政策',
        summary: '兽人控导航站没有用户注册、登录或个人资料系统，不中转外部站点访问流量，也不会主动收集您的敏感个人信息。本页说明本站可能处理的数据类型及其用途。',
        updated: '最后更新：2026年6月6日',
        sections: [
          {
            title: '我们不收集的内容',
            body: '本站不要求注册或登录，不收集真实姓名、身份证件、支付信息、精确地址、私信内容、账号密码或其他敏感个人信息。访问第三方站点时，相关数据处理由第三方站点自行负责。'
          },
          {
            title: '浏览与访问数据',
            body: '本站可能记录站点卡片点击、详情页访问、接口请求等聚合数据，用于浏览量展示、热门排序、缓存优化和问题排查。这些数据以改善公开导航体验为目的，不用于建立用户画像。'
          },
          {
            title: '本地偏好设置',
            body: '部分界面偏好可能保存在您的浏览器本地，例如主题、显示模式、首屏快捷区开关或自定义背景设置。这些内容主要存储在本地浏览器中，用于让页面保持您的显示偏好。'
          },
          {
            title: '缓存与基础日志',
            body: '为了提升访问速度和服务稳定性，本站可能使用 Redis、浏览器缓存或服务器基础日志记录公开接口结果、请求时间、错误信息等运行数据。缓存内容会随服务策略调整或过期清理。'
          },
          {
            title: '第三方链接',
            body: '本站包含前往外部网站的链接。点击外链后，您将离开兽人控导航站，外部网站可能有自己的 Cookie、账号、统计和隐私政策。请在访问前自行阅读并判断。'
          },
          {
            title: '联系与更正',
            body: '如果您认为本站展示的收录信息涉及隐私、版权或错误内容，请通过 GitHub Issues 联系维护者，我们会根据具体情况进行更正、隐藏或移除。'
          }
        ]
      }
    : {
        kicker: 'Privacy Policy',
        title: 'Privacy Policy',
        summary: 'GoFurry Navigation has no registration, login, or user profile system. It does not proxy external website traffic or intentionally collect sensitive personal information. This page explains what data may be processed and why.',
        updated: 'Last updated: June 6, 2026',
        sections: [
          {
            title: 'What We Do Not Collect',
            body: 'The site does not require accounts and does not collect real names, identity documents, payment data, precise addresses, private messages, passwords, or other sensitive personal information. Data handling after visiting third-party sites is controlled by those sites.'
          },
          {
            title: 'Browsing and Visit Data',
            body: 'The site may record aggregated card clicks, detail-page visits, and API requests for visit counts, popular ordering, cache optimization, and troubleshooting. These signals are used to improve the public navigation experience, not to build user profiles.'
          },
          {
            title: 'Local Preferences',
            body: 'Some interface preferences may be stored locally in your browser, such as theme, display mode, hero shortcut visibility, or custom background settings. These values are mainly used to keep your own display preferences.'
          },
          {
            title: 'Cache and Basic Logs',
            body: 'To improve performance and stability, the site may use Redis, browser cache, or basic server logs for public API results, request timing, and error information. Cached content may expire or be cleared as service policies change.'
          },
          {
            title: 'Third-Party Links',
            body: 'The site includes links to external websites. Once you follow an external link, you leave GoFurry Navigation. Those sites may have their own cookies, accounts, analytics, and privacy policies.'
          },
          {
            title: 'Contact and Correction',
            body: 'If you believe a listing involves privacy, copyright, or incorrect information, please contact the maintainer through GitHub Issues. We will review and correct, hide, or remove content when appropriate.'
          }
        ]
      }
))

const pageSeo = computed(() => (
  isZh.value
    ? {
        title: 'GoFurry 隐私政策 - 数据处理、本地偏好与第三方链接说明',
        description: '阅读 GoFurry 兽人控导航站隐私政策，了解本站无注册登录系统、不主动收集敏感个人信息，以及浏览数据、本地偏好、缓存日志和第三方链接的处理方式。'
      }
    : {
        title: 'GoFurry Privacy Policy - Data handling, local preferences, and third-party links',
        description: 'Read the GoFurry Navigation privacy policy covering the no-account design, sensitive data boundaries, browsing signals, local preferences, cache logs, and third-party links.'
      }
))

useSeoMeta({
  title: () => pageSeo.value.title,
  description: () => pageSeo.value.description,
  ogTitle: () => pageSeo.value.title,
  ogDescription: () => pageSeo.value.description,
})
</script>

<style scoped>
.legal-panel {
  border: 1px solid var(--legal-border);
  border-radius: 20px;
  background: var(--legal-surface);
  box-shadow: 0 20px 48px rgba(15, 23, 42, 0.1);
  padding: clamp(1.8rem, 3vw, 3rem);
  backdrop-filter: blur(14px);
  transition: background-color 0.5s ease, border-color 0.5s ease, box-shadow 0.5s ease;
}

@media (max-width: 767px) {
  .legal-panel {
    border-radius: 18px;
  }
}
</style>
