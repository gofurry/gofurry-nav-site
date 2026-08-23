<template>
  <div class="gf-static-page legal-page relative isolate flex w-full flex-1 flex-col overflow-hidden transition-colors duration-500">
    <GoFurryGridBackground :fixed="false" palette="nav-content" />
    <div class="gf-static-page__top-veil" />

    <main class="gf-static-page__main gf-static-page__main--legal">
      <article class="legal-panel">
        <div class="gf-static-kicker">
          {{ content.kicker }}
        </div>
        <h1 class="gf-static-title">
          {{ content.title }}
        </h1>
        <p class="gf-static-summary">
          {{ content.summary }}
        </p>
        <div class="gf-static-updated">
          {{ content.updated }}
        </div>

        <div class="gf-static-section-list">
          <section
            v-for="section in content.sections"
            :key="section.title"
            class="gf-static-section"
        >
          <h2 class="gf-static-section-title">
            {{ section.title }}
          </h2>
          <p class="gf-static-section-body">
            {{ section.body }}
          </p>
          </section>
        </div>
      </article>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'

const { locale } = useI18n()
const isZh = computed(() => locale.value === 'zh')

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
