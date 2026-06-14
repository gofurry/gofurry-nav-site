<template>
  <div class="gf-static-page legal-page relative isolate flex w-full flex-1 flex-col overflow-hidden transition-colors duration-500">
    <GoFurryGridBackground :fixed="false" palette="nav-content" />
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
import { computed } from 'vue'
import { i18n } from '@/main'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'

const isZh = computed(() => i18n.global.locale.value === 'zh')

const content = computed(() => (
  isZh.value
    ? {
        kicker: 'Terms of Service',
        title: '服务条款',
        summary: '使用兽人控导航站即表示您理解本站提供的是公开站点导航、收录信息与用户视角的网络观测数据。本站不提供用户系统，不中转访问流量，也不保证外部站点内容始终可用或适合所有访问者。',
        updated: '最后更新：2026年6月6日',
        sections: [
          {
            title: '服务范围',
            body: '本站主要提供兽圈相关站点的导航、分类、搜索、收录说明、访问统计展示与公开网络观测信息。站点信息可能来自公开页面、站点提交、维护者整理或自动化采集，展示内容仅用于帮助用户发现和理解相关入口。'
          },
          {
            title: '外部链接与第三方内容',
            body: '本站收录的链接会跳转到第三方网站。第三方网站的内容、可用性、安全性、年龄限制、账号体系、隐私策略和后续变更均由对应站点自行负责。访问外部链接前，请自行判断风险并遵守对方站点规则。'
          },
          {
            title: '收录、纠错与下架',
            body: '如果站点信息不准确、Logo 或描述需要修正，或您希望补充、变更、下架相关内容，可以通过 GitHub Issue 联系维护者。我们会根据公开性、相关性、可访问性和社区安全等因素进行处理。'
          },
          {
            title: '浏览量与观测数据',
            body: '本站可能记录站点卡片点击、详情页访问等用于排序和展示的聚合浏览数据。这些数据用于改善导航体验和观察站点热度，不代表本站对外部站点内容的认可、担保或推荐。'
          },
          {
            title: '可用性与变更',
            body: '本站仍处于小规模持续迭代阶段，页面、接口、排序、缓存、收录规则和展示方式都可能调整。我们会尽量保持服务稳定，但不承诺任何功能在所有时间都持续可用。'
          },
          {
            title: '联系与反馈',
            body: '如您对服务条款、站点收录、版权归属或其他事项有疑问，欢迎前往 GitHub Issues 提交反馈。'
          }
        ]
      }
    : {
        kicker: 'Terms of Service',
        title: 'Terms of Service',
        summary: 'By using GoFurry Navigation, you understand that this site provides public site navigation, listing information, and user-perspective web observation data. It does not provide user accounts, proxy traffic, or guarantee that external sites remain available or suitable for every visitor.',
        updated: 'Last updated: June 6, 2026',
        sections: [
          {
            title: 'Scope of Service',
            body: 'The site provides navigation, categories, search, listing notes, visit statistics, and public observation data for furry-related websites. Information may come from public pages, submissions, maintainer curation, or automated collection.'
          },
          {
            title: 'External Links',
            body: 'Listed links lead to third-party websites. Their content, availability, safety, age restrictions, accounts, privacy policies, and future changes are controlled by those websites. Please evaluate risks and follow their rules before visiting.'
          },
          {
            title: 'Listings, Corrections, and Removal',
            body: 'If listing information, logos, or descriptions need changes, or if you want to submit, update, or remove content, please contact the maintainer through GitHub Issues. Requests are handled based on public availability, relevance, accessibility, and community safety.'
          },
          {
            title: 'Visits and Observation Data',
            body: 'The site may record aggregated card clicks and detail-page visits for ordering and display. These signals improve navigation and reflect observed popularity, but they are not endorsements or guarantees of third-party content.'
          },
          {
            title: 'Availability and Changes',
            body: 'The project is still under small-scale continuous iteration. Pages, APIs, ranking, cache behavior, listing rules, and display logic may change. We try to keep the service stable, but no feature is guaranteed to be continuously available.'
          },
          {
            title: 'Contact and Feedback',
            body: 'For questions about these terms, listings, copyright ownership, or related matters, please submit feedback through GitHub Issues.'
          }
        ]
      }
))

const pageSeo = computed(() => (
  isZh.value
    ? {
        title: 'GoFurry 服务条款 - 导航收录、外部链接与观测数据说明',
        description: '阅读 GoFurry 兽人控导航站服务条款，了解导航收录范围、外部链接责任、站点纠错与下架流程、浏览量和用户视角观测数据的使用方式。'
      }
    : {
        title: 'GoFurry Terms of Service - Listings, external links, and observation data',
        description: 'Read the GoFurry Navigation terms covering listing scope, external links, correction and removal requests, visit signals, observation data, availability, and feedback channels.'
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
