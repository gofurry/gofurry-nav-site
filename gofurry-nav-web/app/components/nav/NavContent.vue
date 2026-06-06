<template>
  <div class="relative min-h-screen overflow-hidden p-6">
    <div v-if="loading" class="py-8 text-center text-gray-500 dark:text-slate-300">
      {{ t('common.loading') }}...
    </div>

    <NavSpotlightPanels
      :spotlight="spotlight"
      :display-mode="displayMode"
    />

    <div
      v-for="group in groups"
      :key="group.id"
      class="relative"
      :class="filteredSites(group).length > 30 ? 'mb-0' : 'mb-10'"
    >
      <div
        ref="groupRefs"
        class="relative inline-block cursor-pointer"
        @mouseenter="(event) => onGroupMouseEnter(event, group)"
        @mouseleave="scheduleGroupHide"
      >
        <h2 class="text-xl font-semibold text-stone-900 transition-colors hover:text-amber-600 dark:text-slate-100 dark:hover:text-sky-300">
          {{ group.name }}
        </h2>
      </div>

      <div class="mt-2 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
        <div
          v-for="site in displaySites(group)"
          :key="`${group.id}-${site.id}`"
          ref="siteRefs"
          class="flex cursor-pointer gap-3 rounded-xl border border-orange-100/70 bg-orange-50/94 p-4 transition-shadow duration-200 dark:border-white/8 dark:bg-[rgba(18,30,48,0.82)] dark:shadow-[0_10px_24px_rgba(2,6,23,0.18)]"
          @click="goDomain(domainList(site), site)"
          @mouseenter="(event) => onSiteMouseEnter(event, site)"
          @mouseleave="scheduleSiteHide"
        >
          <div class="h-12 w-12 flex-shrink-0 overflow-hidden rounded">
            <img
              :key="siteLogoKey(site)"
              :src="siteLogoSrc(site)"
              class="h-full w-full object-cover"
              :alt="site.name"
            />
          </div>

          <div class="flex flex-1 flex-col overflow-hidden">
            <div class="flex items-center gap-1">
              <h3 class="truncate text-base font-medium text-stone-900 dark:text-slate-100">
                {{ site.name }}
              </h3>
            </div>

            <p class="mt-1 truncate text-xs text-gray-500 dark:text-slate-400">
              {{ site.info }}
            </p>
          </div>
        </div>
      </div>

      <div v-if="filteredSites(group).length > 30" class="my-4 flex justify-center">
        <button
          class="rounded-sm bg-orange-100 px-6 text-sm text-orange-700 transition hover:bg-orange-200 dark:bg-[rgba(30,41,59,0.82)] dark:text-slate-100 dark:hover:bg-[rgba(51,65,85,0.92)]"
          @click="toggleGroup(group.id)"
        >
          {{ expandedGroups[group.id] ? t('common.collapse') : t('common.expand') }}
        </button>
      </div>
    </div>

    <Teleport to="body">
      <GroupPopover
        v-if="!!activeGroup"
        :group="activeGroup"
        :target-element="activeGroupTarget"
        :visible="!!activeGroup"
        @mouseenter="cancelGroupHide"
        @mouseleave="scheduleGroupHide"
      />
    </Teleport>

    <Teleport to="body">
      <SitePopover
        v-if="!!activeSite"
        :site="activeSite"
        :visible="activeSiteVisible"
        :position="activeSitePosition"
        :placement="activeSitePlacement"
        :ping-data="pingData"
        @get-popover-height="handleGetPopoverHeight"
        @mouseenter="cancelSiteHide"
        @mouseleave="scheduleSiteHide"
      />
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useLangStore } from '@/store/langStore'
import { getNavHome, getNavHomePing, touchSiteView } from '~/services/nav'
import type { Delay, Group, NavHomeSpotlight, Site } from '~/types/nav'
import { recordRecentSite, toExternalUrl } from '@/utils/recentSites'
import { readDisplayMode, subscribeModeChange, type DisplayMode } from '@/utils/modeStorage'
import GroupPopover from './GroupPopover.vue'
import NavSpotlightPanels from './NavSpotlightPanels.vue'
import SitePopover from './SitePopover.vue'

const props = defineProps<{
  initialGroups?: Group[]
  initialSpotlight?: NavHomeSpotlight
  initialPingData?: Record<string, Delay>
}>()

const { t } = useI18n()

const groups = ref<Group[]>(props.initialGroups ?? [])
const spotlight = ref<NavHomeSpotlight>(props.initialSpotlight ?? { page_size: 6, featured: [], popular: [], latest: [], random: [] })
const pingData = ref<Record<string, Delay>>(props.initialPingData ?? {})
const loading = ref(false)
const expandedGroups = ref<Record<string, boolean>>({})

const groupRefs = ref<HTMLElement[]>([])
const siteRefs = ref<HTMLElement[]>([])

const langStore = useLangStore()

const logoPrefix = import.meta.env.VITE_SITE_LOGO_PREFIX_URL || ''
const defaultLogo = 'defaultLogo.svg'

const displayMode = ref<DisplayMode>('sfw')
let stopModeSubscription: (() => void) | null = null

function parsePingData(data: Record<string, string | undefined>) {
  const result: Record<string, Delay> = {}

  for (const key in data) {
    const value = data[key]
    if (typeof value === 'string') {
      try {
        result[key] = JSON.parse(value) as Delay
      } catch {
        result[key] = { status: 'down', delay: '-', loss: '-', time: '-' }
      }
    } else {
      result[key] = { status: 'down', delay: '-', loss: '-', time: '-' }
    }
  }

  pingData.value = result
}

function isSiteVisible(site: Site) {
  return displayMode.value === 'nsfw' || String(site.nsfw) !== '1'
}

async function loadData() {
  loading.value = true
  try {
    const lang = langStore.lang
    const home = await getNavHome(lang)

    groups.value = home.groups.sort((a, b) => Number(a.priority) - Number(b.priority))
    spotlight.value = home.spotlight
    parsePingData(home.ping)
  } catch (error) {
    console.error('Failed to load nav content:', error)
  } finally {
    loading.value = false
  }
}

function filteredSites(group: Group) {
  return group.sites.filter((site): site is Site => !!site && isSiteVisible(site))
}

function displaySites(group: Group) {
  const filtered = filteredSites(group)
  if (filtered.length <= 30) {
    return filtered
  }

  return expandedGroups.value[group.id] ? filtered : filtered.slice(0, 30)
}

function toggleGroup(groupId: string) {
  expandedGroups.value[groupId] = !expandedGroups.value[groupId]
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

function siteLogoKey(site: Site) {
  return `${site.id}:${site.icon || defaultLogo}:${site.update_time || ''}`
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

function goDomain(domains: string[], site?: Site) {
  if (!domains.length) {
    return
  }

  const firstDomain = domains[0]
  if (!firstDomain) {
    return
  }

  const targetUrl = toExternalUrl(firstDomain)
  if (site) {
    void updateSiteViewCount(site)
    recordRecentSite({
      id: site.id,
      name: site.name,
      url: targetUrl,
    })
  }

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

const activeGroup = ref<Group | null>(null)
const activeGroupTarget = ref<HTMLElement | null>(null)
let groupHideTimer: number | null = null

function onGroupMouseEnter(event: MouseEvent, group: Group) {
  cancelGroupHide()
  activeGroup.value = group
  activeGroupTarget.value = event.currentTarget as HTMLElement
}

function scheduleGroupHide() {
  groupHideTimer = window.setTimeout(() => {
    activeGroup.value = null
    activeGroupTarget.value = null
  }, 200)
}

function cancelGroupHide() {
  if (groupHideTimer) {
    clearTimeout(groupHideTimer)
    groupHideTimer = null
  }
}

const activeSite = ref<Site | null>(null)
const activeSiteVisible = ref(false)
const activeSiteTarget = ref<HTMLElement | null>(null)
const activeSitePosition = ref<{ left: number; top: number } | null>(null)
const activeSitePlacement = ref<'top' | 'bottom'>('bottom')
const popoverHeight = ref(0)
let siteHideTimer: number | null = null
let siteCleanupTimer: number | null = null

function onSiteMouseEnter(event: MouseEvent, site: Site) {
  cancelSiteHide()
  const currentTarget = event.currentTarget as HTMLElement
  const isSwitchingSite = activeSite.value?.id !== site.id

  activeSite.value = site
  activeSiteTarget.value = currentTarget
  popoverHeight.value = 0
  updateSitePopoverPosition(220)

  if (isSwitchingSite && !activeSiteVisible.value) {
    requestAnimationFrame(() => {
      activeSiteVisible.value = true
    })
    return
  }

  activeSiteVisible.value = true
}

function handleGetPopoverHeight(height: number) {
  popoverHeight.value = height
  updateSitePopoverPosition(height)
}

function updateSitePopoverPosition(measuredHeight = popoverHeight.value || 220) {
  if (!activeSite.value || !activeSiteTarget.value) {
    return
  }

  const targetRect = activeSiteTarget.value.getBoundingClientRect()
  const popoverWidth = 288
  const viewportHeight = window.innerHeight
  const viewportWidth = window.innerWidth
  const gap = 10
  const safeInset = 12

  let left = targetRect.left + (targetRect.width - popoverWidth) / 2
  left = Math.max(safeInset, Math.min(left, viewportWidth - popoverWidth - safeInset))

  const canPlaceBelow = targetRect.bottom + gap + measuredHeight <= viewportHeight - safeInset
  const placement = canPlaceBelow ? 'bottom' : 'top'
  let top = placement === 'bottom'
    ? targetRect.bottom + gap
    : targetRect.top - measuredHeight - gap
  top = Math.max(safeInset, Math.min(top, viewportHeight - measuredHeight - safeInset))

  activeSitePlacement.value = placement
  activeSitePosition.value = { left, top }
}

function scheduleSiteHide() {
  if (siteHideTimer) {
    clearTimeout(siteHideTimer)
  }
  if (siteCleanupTimer) {
    clearTimeout(siteCleanupTimer)
    siteCleanupTimer = null
  }

  siteHideTimer = window.setTimeout(() => {
    activeSiteVisible.value = false

    siteCleanupTimer = window.setTimeout(() => {
      activeSite.value = null
      activeSiteTarget.value = null
      activeSitePosition.value = null
      popoverHeight.value = 0
      siteCleanupTimer = null
    }, 320)

    siteHideTimer = null
  }, 520)
}

function cancelSiteHide() {
  if (siteHideTimer) {
    clearTimeout(siteHideTimer)
    siteHideTimer = null
  }
  if (siteCleanupTimer) {
    clearTimeout(siteCleanupTimer)
    siteCleanupTimer = null
  }

  activeSiteVisible.value = true
}

function handleScrollOrResize() {
  if (activeSite.value && activeSiteTarget.value && popoverHeight.value > 0) {
    updateSitePopoverPosition()
  }
}

let pingTimer: ReturnType<typeof setInterval> | null = null

watch(
  () => langStore.lang,
  async (newLang, oldLang) => {
    if (newLang !== oldLang) {
      await loadData()
    }
  }
)

onMounted(() => {
  displayMode.value = readDisplayMode()
  stopModeSubscription = subscribeModeChange(({ displayMode: nextMode }) => {
    displayMode.value = nextMode
  })

  window.addEventListener('scroll', handleScrollOrResize, { passive: true, capture: true })
  document.addEventListener('scroll', handleScrollOrResize, { passive: true, capture: true })
  window.addEventListener('resize', handleScrollOrResize)

  pingTimer = setInterval(async () => {
    try {
      const response = await getNavHomePing()
      parsePingData(response.ping)
    } catch (error) {
      console.error('Failed to refresh ping data:', error)
    }
  }, 60000)
})

onUnmounted(() => {
  stopModeSubscription?.()
  if (pingTimer) {
    clearInterval(pingTimer)
  }

  window.removeEventListener('scroll', handleScrollOrResize, { capture: true })
  document.removeEventListener('scroll', handleScrollOrResize, { capture: true })
  window.removeEventListener('resize', handleScrollOrResize)
  cancelSiteHide()
  cancelGroupHide()
})
</script>
