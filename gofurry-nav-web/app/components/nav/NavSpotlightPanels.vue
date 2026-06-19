<template>
  <section
    v-if="hasAnySpotlight"
    class="mb-4"
  >
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <article
        v-for="panel in visiblePanels"
        :key="panel.key"
        class="spotlight-panel"
        :class="panel.visibilityClass"
      >
        <header class="spotlight-panel__header">
          <h2>{{ panel.title }}</h2>
          <div v-if="panel.totalPages > 1" class="spotlight-panel__pager">
            <button type="button" :aria-label="`${panel.title} ${label('上一页', 'previous page')}`" @click="changePage(panel.key, -1)">‹</button>
            <span>{{ panel.page + 1 }}/{{ panel.totalPages }}</span>
            <button type="button" :aria-label="`${panel.title} ${label('下一页', 'next page')}`" @click="changePage(panel.key, 1)">›</button>
          </div>
        </header>

        <div class="spotlight-panel__list">
          <button
            v-for="(site, index) in panel.items"
            :key="`${panel.key}-${site.id}`"
            type="button"
            class="spotlight-site"
            @click="openSite(site)"
          >
            <span
              class="spotlight-site__rank"
              :class="{ 'spotlight-site__rank--visited': visitedSiteIds.has(site.id) }"
              :aria-label="visitedSiteIds.has(site.id) ? label('已浏览', 'Visited') : undefined"
            >
              <svg v-if="visitedSiteIds.has(site.id)" viewBox="0 0 16 16" aria-hidden="true">
                <path d="M3 8.3 6.4 11.2 13 4.6" />
              </svg>
              <template v-else>{{ panel.page * pageSize + index + 1 }}</template>
            </span>
            <span class="spotlight-site__logo">
              <img
                :src="siteLogoSrc(site)"
                :alt="site.name"
                width="29"
                height="29"
                loading="lazy"
                decoding="async"
                fetchpriority="low"
              />
            </span>
            <span class="spotlight-site__body">
              <strong>{{ site.name }}</strong>
              <small>{{ metaText(panel.key, site) }}</small>
            </span>
          </button>

          <div v-if="!panel.items.length" class="spotlight-panel__empty">{{ label('暂无站点', 'No sites') }}</div>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { NavHomeSpotlight, Site } from '~/types/nav'
import { touchSiteView } from '~/services/nav'
import { loadRecentSites, recordRecentSite, RECENT_SITES_EVENT, toExternalUrl } from '@/utils/recentSites'
import type { DisplayMode } from '@/utils/modeStorage'

type PanelKey = 'featured' | 'popular' | 'latest' | 'random'

const props = defineProps<{
  spotlight: NavHomeSpotlight
  displayMode: DisplayMode
}>()

const logoPrefix = import.meta.env.VITE_SITE_LOGO_PREFIX_URL || ''
const defaultLogo = 'defaultLogo.svg'
const { locale } = useI18n()
const isEnglish = computed(() => locale.value === 'en')
const pages = ref<Record<PanelKey, number>>({
  featured: 0,
  popular: 0,
  latest: 0,
  random: 0,
})
const visitedSiteIds = ref<Set<string>>(new Set())
const spotlightPageSize = 6
const pageSize = computed(() => spotlightPageSize)

const panelConfigs = computed<Array<{
  key: PanelKey
  title: string
  visibilityClass: string
}>>(() => [
  { key: 'featured', title: label('精选站点', 'Featured'), visibilityClass: '' },
  { key: 'popular', title: label('热门站点', 'Popular'), visibilityClass: 'hidden sm:block' },
  { key: 'latest', title: label('最新收录', 'Latest'), visibilityClass: 'hidden lg:block' },
  { key: 'random', title: label('随机站点', 'Random'), visibilityClass: 'hidden xl:block' },
])

const hasAnySpotlight = computed(() => {
  return panelConfigs.value.some(panel => visibleSites(props.spotlight?.[panel.key] ?? []).length > 0)
})

const visiblePanels = computed(() => panelConfigs.value.map((config) => {
  const list = visibleSites(props.spotlight?.[config.key] ?? [])
  const totalPages = Math.max(1, Math.ceil(list.length / pageSize.value))
  const page = Math.min(pages.value[config.key], totalPages - 1)
  return {
    ...config,
    page,
    totalPages,
    items: list.slice(page * pageSize.value, page * pageSize.value + pageSize.value),
  }
}))

watch(
  () => props.spotlight,
  () => {
    pages.value = { featured: 0, popular: 0, latest: 0, random: 0 }
  }
)

function visibleSites(sites: Site[]) {
  return sites.filter(site => props.displayMode === 'nsfw' || String(site.nsfw) !== '1')
}

function changePage(key: PanelKey, delta: number) {
  const panel = visiblePanels.value.find(item => item.key === key)
  if (!panel || panel.totalPages <= 1) {
    return
  }
  pages.value[key] = (panel.page + delta + panel.totalPages) % panel.totalPages
}

function joinAssetUrl(prefix: string, path: string) {
  if (!prefix) {
    return path
  }
  return `${prefix.replace(/\/+$/, '')}/${path.replace(/^\/+/, '')}`
}

function withAssetVersion(url: string, version?: string | null) {
  const normalizedVersion = (version || '').trim()
  if (!normalizedVersion) {
    return url
  }
  const separator = url.includes('?') ? '&' : '?'
  return `${url}${separator}v=${encodeURIComponent(normalizedVersion)}`
}

function siteLogoSrc(site: Site) {
  const iconPath = site.icon || defaultLogo
  const assetURL = joinAssetUrl(logoPrefix, iconPath)
  if (!site.icon) {
    return assetURL
  }
  return withAssetVersion(assetURL, site.update_time)
}

function domainList(site: Site) {
  if (Array.isArray(site.domain)) {
    return site.domain
  }

  try {
    const domainObject = JSON.parse(site.domain)
    return Array.isArray(domainObject?.domain) ? domainObject.domain : []
  } catch {
    return site.domain ? [site.domain] : []
  }
}

function openSite(site: Site) {
  const targetUrl = toExternalUrl(domainList(site)[0] || '')
  if (!targetUrl) {
    return
  }

  void updateSiteViewCount(site)
  recordRecentSite({
    id: site.id,
    name: site.name,
    url: targetUrl,
  })
  syncVisitedSites()
  window.open(targetUrl, '_blank')
}

async function updateSiteViewCount(site: Site) {
  try {
    const response = await touchSiteView(site.id)
    if (Number.isFinite(response.view_count)) {
      site.view_count = response.view_count
    }
  } catch {
    // 浏览量统计是旁路副作用，失败不影响跳转。
  }
}

function metaText(key: PanelKey, site: Site) {
  if (key === 'popular') {
    return `${formatNumber(site.view_count)} ${label('次浏览', 'views')}`
  }
  if (key === 'latest') {
    return formatDate(site.create_time)
  }
  return site.info
}

function formatNumber(value: unknown) {
  const num = Number(value)
  if (!Number.isFinite(num)) {
    return '0'
  }
  return num.toLocaleString(isEnglish.value ? 'en-US' : 'zh-CN')
}

function formatDate(value?: string | null) {
  if (!value) {
    return label('最近收录', 'Recently added')
  }
  return value.slice(0, 10)
}

function label(zh: string, en: string) {
  return isEnglish.value ? en : zh
}

function syncVisitedSites() {
  try {
    visitedSiteIds.value = new Set(loadRecentSites().map(site => site.id))
  } catch {
    visitedSiteIds.value = new Set()
  }
}

onMounted(() => {
  syncVisitedSites()
  window.addEventListener(RECENT_SITES_EVENT, syncVisitedSites)
})

onUnmounted(() => {
  window.removeEventListener(RECENT_SITES_EVENT, syncVisitedSites)
})
</script>
