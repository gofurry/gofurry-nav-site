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

        <div class="spotlight-panel__viewport">
          <div
            :key="panel.trackKey"
            class="spotlight-panel__track"
            :class="{
              'spotlight-panel__track--instant': panel.instantTrack,
              'spotlight-panel__track--next': panel.isSliding && panel.direction > 0,
              'spotlight-panel__track--prev': panel.isSliding && panel.direction < 0,
            }"
          >
            <div
              v-for="slide in panel.slides"
              :key="slide.key"
              class="spotlight-panel__slide"
              :aria-hidden="slide.page !== panel.page"
            >
              <div class="spotlight-panel__list">
                <button
                  v-for="(site, index) in slide.sites"
                  :key="`${panel.key}-${slide.page}-${site.id}`"
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
                    <template v-else>{{ slide.page * pageSize + index + 1 }}</template>
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

                <div v-if="!slide.sites.length" class="spotlight-panel__empty">{{ label('暂无站点', 'No sites') }}</div>
              </div>
            </div>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { NavHomeSpotlight, Site } from '~/types/nav'
import { touchSiteView } from '~/services/nav'
import { loadRecentSites, recordRecentSite, RECENT_SITES_EVENT, toExternalUrl } from '@/utils/recentSites'
import type { DisplayMode } from '@/utils/modeStorage'

type PanelKey = 'featured' | 'popular' | 'latest' | 'random'

type PanelMotionState = {
  renderedPage: number
  sourcePage: number
  targetPage: number
  direction: 1 | -1
  isSliding: boolean
  instantTrack: boolean
  nonce: number
}

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
const panelMotion = ref<Record<PanelKey, PanelMotionState>>(createPanelMotionStates())
const visitedSiteIds = ref<Set<string>>(new Set())
const spotlightPageSize = 6
const pageSize = computed(() => spotlightPageSize)
const panelKeys: PanelKey[] = ['featured', 'popular', 'latest', 'random']
const slideTimers: Partial<Record<PanelKey, ReturnType<typeof setTimeout>>> = {}
const slideFrames: Partial<Record<PanelKey, number>> = {}

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
  const motion = panelMotion.value[config.key]
  return {
    ...config,
    page,
    totalPages,
    slides: slidesForPanel(config.key, totalPages),
    trackKey: trackKeyFor(config.key, totalPages),
    direction: motion.direction,
    isSliding: motion.isSliding,
    instantTrack: motion.instantTrack,
  }
}))

watch(
  [() => props.spotlight, () => props.displayMode],
  () => {
    resetPanelPages()
  }
)

function createPanelMotionState(): PanelMotionState {
  return {
    renderedPage: 0,
    sourcePage: 0,
    targetPage: 0,
    direction: 1,
    isSliding: false,
    instantTrack: false,
    nonce: 0,
  }
}

function createPanelMotionStates(): Record<PanelKey, PanelMotionState> {
  return {
    featured: createPanelMotionState(),
    popular: createPanelMotionState(),
    latest: createPanelMotionState(),
    random: createPanelMotionState(),
  }
}

function visibleSites(sites: Site[]) {
  return sites.filter(site => props.displayMode === 'nsfw' || String(site.nsfw) !== '1')
}

function panelSites(key: PanelKey) {
  return visibleSites(props.spotlight?.[key] ?? [])
}

function sitesForPage(key: PanelKey, page: number) {
  const list = panelSites(key)
  const safePage = Math.max(0, page)
  return list.slice(safePage * pageSize.value, safePage * pageSize.value + pageSize.value)
}

function normalizedPanelPage(page: number, totalPages: number) {
  return Math.min(Math.max(page, 0), Math.max(totalPages - 1, 0))
}

function slidesForPanel(key: PanelKey, totalPages: number) {
  const motion = panelMotion.value[key]

  if (!motion.isSliding) {
    const page = normalizedPanelPage(motion.renderedPage, totalPages)
    return [{
      key: `${key}-page-${page}-${motion.nonce}`,
      page,
      sites: sitesForPage(key, page),
    }]
  }

  const sourcePage = normalizedPanelPage(motion.sourcePage, totalPages)
  const targetPage = normalizedPanelPage(motion.targetPage, totalPages)
  const source = {
    key: `${key}-source-${sourcePage}-${motion.nonce}`,
    page: sourcePage,
    sites: sitesForPage(key, sourcePage),
  }
  const target = {
    key: `${key}-target-${targetPage}-${motion.nonce}`,
    page: targetPage,
    sites: sitesForPage(key, targetPage),
  }

  return motion.direction > 0 ? [source, target] : [target, source]
}

function trackKeyFor(key: PanelKey, totalPages: number) {
  const motion = panelMotion.value[key]
  if (motion.isSliding) {
    const sourcePage = normalizedPanelPage(motion.sourcePage, totalPages)
    const targetPage = normalizedPanelPage(motion.targetPage, totalPages)
    return `${key}-motion-${sourcePage}-${targetPage}-${motion.nonce}`
  }

  const renderedPage = normalizedPanelPage(motion.renderedPage, totalPages)
  return `${key}-stable-${renderedPage}-${motion.nonce}`
}

function clearSlideTimer(key: PanelKey) {
  const timer = slideTimers[key]
  if (timer) {
    clearTimeout(timer)
    delete slideTimers[key]
  }

  const frame = slideFrames[key]
  if (frame) {
    cancelAnimationFrame(frame)
    delete slideFrames[key]
  }
}

function finishSlide(key: PanelKey) {
  const motion = panelMotion.value[key]
  motion.instantTrack = true
  motion.renderedPage = motion.targetPage
  motion.isSliding = false
  delete slideTimers[key]

  void nextTick().then(() => {
    slideFrames[key] = requestAnimationFrame(() => {
      motion.instantTrack = false
      delete slideFrames[key]
    })
  })
}

function changePage(key: PanelKey, delta: number) {
  const panel = visiblePanels.value.find(item => item.key === key)
  if (!panel || panel.totalPages <= 1) {
    return
  }

  const currentPage = panel.page
  const nextPage = (currentPage + delta + panel.totalPages) % panel.totalPages
  if (nextPage === currentPage) {
    return
  }

  clearSlideTimer(key)

  const motion = panelMotion.value[key]
  motion.instantTrack = false
  motion.direction = delta > 0 ? 1 : -1
  motion.renderedPage = currentPage
  motion.sourcePage = currentPage
  motion.targetPage = nextPage
  motion.isSliding = true
  motion.nonce += 1
  pages.value[key] = nextPage

  slideTimers[key] = setTimeout(() => finishSlide(key), 560)
}

function resetPanelPages() {
  panelKeys.forEach(clearSlideTimer)
  pages.value = { featured: 0, popular: 0, latest: 0, random: 0 }
  panelMotion.value = createPanelMotionStates()
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
  panelKeys.forEach(clearSlideTimer)
  window.removeEventListener(RECENT_SITES_EVENT, syncVisitedSites)
})
</script>
