<template>
  <section
    v-if="hasAnySpotlight"
    class="mb-4"
    :class="{ 'spotlight-panels--dark': isDarkTheme }"
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
            <button type="button" :aria-label="`${panel.title} 上一页`" @click="changePage(panel.key, -1)">‹</button>
            <span>{{ panel.page + 1 }}/{{ panel.totalPages }}</span>
            <button type="button" :aria-label="`${panel.title} 下一页`" @click="changePage(panel.key, 1)">›</button>
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
              :aria-label="visitedSiteIds.has(site.id) ? '已浏览' : undefined"
            >
              <svg v-if="visitedSiteIds.has(site.id)" viewBox="0 0 16 16" aria-hidden="true">
                <path d="M3 8.3 6.4 11.2 13 4.6" />
              </svg>
              <template v-else>{{ panel.page * pageSize + index + 1 }}</template>
            </span>
            <span class="spotlight-site__logo">
              <img :src="siteLogoSrc(site)" :alt="site.name" />
            </span>
            <span class="spotlight-site__body">
              <strong>{{ site.name }}</strong>
              <small>{{ metaText(panel.key, site) }}</small>
            </span>
          </button>

          <div v-if="!panel.items.length" class="spotlight-panel__empty">暂无站点</div>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import type { NavHomeSpotlight, Site } from '~/types/nav'
import { touchSiteView } from '~/services/nav'
import { loadRecentSites, recordRecentSite, RECENT_SITES_EVENT, toExternalUrl } from '@/utils/recentSites'
import type { DisplayMode } from '@/utils/modeStorage'
import { useThemeStore } from '@/stores/theme'

type PanelKey = 'featured' | 'popular' | 'latest' | 'random'

const props = defineProps<{
  spotlight: NavHomeSpotlight
  displayMode: DisplayMode
}>()

const logoPrefix = import.meta.env.VITE_SITE_LOGO_PREFIX_URL || ''
const defaultLogo = 'defaultLogo.svg'
const themeStore = useThemeStore()
const isDarkTheme = computed(() => themeStore.theme === 'dark')
const pages = ref<Record<PanelKey, number>>({
  featured: 0,
  popular: 0,
  latest: 0,
  random: 0,
})
const visitedSiteIds = ref<Set<string>>(new Set())
const spotlightPageSize = 6
const pageSize = computed(() => spotlightPageSize)

const panelConfigs: Array<{
  key: PanelKey
  title: string
  visibilityClass: string
}> = [
  { key: 'featured', title: '精选站点', visibilityClass: '' },
  { key: 'popular', title: '热门站点', visibilityClass: 'hidden sm:block' },
  { key: 'latest', title: '最新收录', visibilityClass: 'hidden lg:block' },
  { key: 'random', title: '随机站点', visibilityClass: 'hidden xl:block' },
]

const hasAnySpotlight = computed(() => {
  return panelConfigs.some(panel => visibleSites(props.spotlight?.[panel.key] ?? []).length > 0)
})

const visiblePanels = computed(() => panelConfigs.map((config) => {
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
    return `${formatNumber(site.view_count)} 次浏览`
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
  return num.toLocaleString('zh-CN')
}

function formatDate(value?: string | null) {
  if (!value) {
    return '最近收录'
  }
  return value.slice(0, 10)
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

<style scoped>
.spotlight-panel {
  min-height: 17.6rem;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.62);
  border-radius: 0.5rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.32), rgba(255, 255, 255, 0.14)),
    rgba(255, 255, 255, 0.08);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.65),
    0 16px 42px rgba(124, 45, 18, 0.08);
  backdrop-filter: blur(20px) saturate(1.16);
}

.spotlight-panel__header {
  display: flex;
  min-height: 3.05rem;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.46);
  padding: 0.62rem 0.8rem;
}

.spotlight-panel__header h2 {
  margin: 0;
  color: #1f2937;
  font-size: 1rem;
  font-weight: 650;
  line-height: 1.2;
}

.spotlight-panel__pager {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.35rem;
}

.spotlight-panel__pager button {
  display: grid;
  width: 1.6rem;
  height: 1.6rem;
  place-items: center;
  border: 1px solid rgba(251, 146, 60, 0.30);
  border-radius: 0.4rem;
  background: rgba(255, 255, 255, 0.24);
  color: #9a3412;
  cursor: pointer;
}

.spotlight-panel__pager span {
  min-width: 2.5rem;
  color: rgba(68, 64, 60, 0.72);
  font-size: 0.72rem;
  text-align: center;
}

.spotlight-panel__list {
  display: grid;
  gap: 0.22rem;
  padding: 0.42rem;
}

.spotlight-site {
  display: grid;
  width: 100%;
  min-height: 2.12rem;
  grid-template-columns: 1.45rem 2rem minmax(0, 1fr);
  align-items: center;
  gap: 0.42rem;
  border: 1px solid rgba(255, 255, 255, 0.38);
  border-radius: 0.4rem;
  background: rgba(255, 255, 255, 0.12);
  color: inherit;
  cursor: pointer;
  padding: 0.14rem 0.3rem;
  text-align: left;
  transition: border-color 220ms ease, background 220ms ease, transform 220ms ease;
}

.spotlight-site:hover {
  border-color: rgba(251, 146, 60, 0.38);
  background: rgba(255, 255, 255, 0.36);
  transform: translateY(-1px);
}

.spotlight-site__rank {
  display: grid;
  width: 1.45rem;
  height: 1.45rem;
  place-items: center;
  color: rgba(68, 64, 60, 0.7);
  font-size: 0.72rem;
  font-variant-numeric: tabular-nums;
  text-align: center;
}

.spotlight-site__rank--visited {
  color: #16a34a;
}

.spotlight-site__rank svg {
  width: 0.92rem;
  height: 0.92rem;
  fill: none;
  stroke: currentColor;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 3.2;
  filter: drop-shadow(0 0 3px rgba(22, 163, 74, 0.25));
}

.spotlight-site__logo {
  display: block;
  width: 1.78rem;
  height: 1.78rem;
  overflow: hidden;
  border-radius: 0.38rem;
  background: rgba(255, 237, 213, 0.58);
}

.spotlight-site__logo img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.spotlight-site__body {
  min-width: 0;
}

.spotlight-site__body strong,
.spotlight-site__body small {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.spotlight-site__body strong {
  color: #1f2937;
  font-size: 0.78rem;
  font-weight: 620;
  line-height: 1.05;
}

.spotlight-site__body small {
  margin-top: 0.08rem;
  color: rgba(68, 64, 60, 0.7);
  font-size: 0.64rem;
}

.spotlight-panel__empty {
  display: grid;
  min-height: 12rem;
  place-items: center;
  color: rgba(68, 64, 60, 0.7);
  font-size: 0.8rem;
}

:global(.dark) .spotlight-panel {
  border-color: rgba(148, 163, 184, 0.38);
  background:
    linear-gradient(180deg, rgba(15, 23, 42, 0.30), rgba(15, 23, 42, 0.14)),
    rgba(15, 23, 42, 0.10);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.05), 0 18px 42px rgba(2, 6, 23, 0.16);
}

:global(.dark) .spotlight-panel__header {
  border-bottom-color: rgba(148, 163, 184, 0.22);
}

:global(.dark) .spotlight-panel__header h2,
:global(.dark) .spotlight-site__body strong {
  color: #f8fafc;
}

:global(.dark) .spotlight-site__rank,
:global(.dark) .spotlight-site__body small,
:global(.dark) .spotlight-panel__empty {
  color: rgba(203, 213, 225, 0.72);
}

.spotlight-panels--dark .spotlight-site__rank--visited {
  color: #fb923c;
}

.spotlight-panels--dark .spotlight-site__rank--visited svg {
  filter: drop-shadow(0 0 3px rgba(251, 146, 60, 0.34));
}

:global(.dark) .spotlight-site:hover {
  background: rgba(148, 163, 184, 0.14);
  border-color: rgba(148, 163, 184, 0.28);
}

:global(.dark) .spotlight-panel__pager button {
  border-color: rgba(148, 163, 184, 0.22);
  background: rgba(15, 23, 42, 0.24);
  color: #e2e8f0;
}

:global(.dark) .spotlight-panel__pager span {
  color: rgba(203, 213, 225, 0.72);
}

:global(.dark) .spotlight-site__logo {
  background: rgba(15, 23, 42, 0.72);
}

:global(.dark) .spotlight-site {
  border-color: rgba(148, 163, 184, 0.16);
  background: rgba(15, 23, 42, 0.16);
}
</style>
