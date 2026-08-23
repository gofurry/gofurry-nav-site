<template>
  <div class="nav-site-grid grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
    <div
      v-for="site in visibleSites"
      :key="site.id"
      ref="siteRefs"
      class="nav-site-card"
      @click="goDomain(domainList(site), site)"
      @mouseenter="(event) => onSiteMouseEnter(event, site)"
      @mouseleave="scheduleSiteHide"
    >
      <div class="nav-site-card__logo">
        <img
          :key="siteLogoKey(site)"
          :src="siteLogoSrc(site)"
          class="h-full w-full object-contain"
          :alt="site.name"
          width="48"
          height="48"
          loading="lazy"
          decoding="async"
          fetchpriority="low"
        />
      </div>

      <div class="flex flex-1 flex-col overflow-hidden">
        <div class="flex items-center gap-1">
          <h3 class="nav-site-card__title truncate text-base font-medium">
            {{ site.name }}
          </h3>
        </div>

        <p class="nav-site-card__desc mt-1 text-xs">
          {{ site.info }}
        </p>
      </div>
    </div>
  </div>

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
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import type { Delay, Site } from '~/types/nav'
import { touchSiteView } from '~/services/nav'
import { recordRecentSite, toExternalUrl } from '@/utils/recentSites'
import { readDisplayMode, subscribeModeChange, type DisplayMode } from '@/utils/modeStorage'
import SitePopover from './SitePopover.vue'

const props = defineProps<{
  sites: Site[]
  pingData?: Record<string, Delay>
}>()

const logoPrefix = import.meta.env.VITE_SITE_LOGO_PREFIX_URL || ''
const defaultLogo = 'defaultLogo.svg'
const pingData = computed(() => props.pingData ?? {})
const displayMode = ref<DisplayMode>('sfw')
const siteRefs = ref<HTMLElement[]>([])
let stopModeSubscription: (() => void) | null = null

function isSiteVisible(site: Site) {
  return displayMode.value === 'nsfw' || String(site.nsfw) !== '1'
}

const visibleSites = computed(() => props.sites.filter((site): site is Site => !!site && isSiteVisible(site)))

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

onMounted(() => {
  displayMode.value = readDisplayMode()
  stopModeSubscription = subscribeModeChange(({ displayMode: nextMode }) => {
    displayMode.value = nextMode
  })

  window.addEventListener('scroll', handleScrollOrResize, { passive: true, capture: true })
  document.addEventListener('scroll', handleScrollOrResize, { passive: true, capture: true })
  window.addEventListener('resize', handleScrollOrResize)
})

onUnmounted(() => {
  stopModeSubscription?.()
  window.removeEventListener('scroll', handleScrollOrResize, { capture: true })
  document.removeEventListener('scroll', handleScrollOrResize, { capture: true })
  window.removeEventListener('resize', handleScrollOrResize)
  cancelSiteHide()
})
</script>
