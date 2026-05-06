<template>
  <div class="relative min-h-screen overflow-hidden p-6">
    <div v-if="loading" class="py-8 text-center text-gray-500">
      {{ t('common.loading') }}...
    </div>

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
        <h2 class="text-xl font-semibold transition-colors hover:text-amber-600">
          {{ group.name }}
        </h2>
      </div>

      <div class="mt-2 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5">
        <div
          v-for="site in displaySites(group)"
          :key="`${group.id}-${site.id}`"
          ref="siteRefs"
          class="flex cursor-pointer gap-3 rounded-xl bg-orange-50 p-4 transition-shadow duration-200"
          @click="goDomain(domainsMap[site.id] || [], site)"
          @mouseenter="(event) => onSiteMouseEnter(event, site)"
          @mouseleave="scheduleSiteHide"
        >
          <div class="h-12 w-12 flex-shrink-0 overflow-hidden rounded">
            <img
              :src="`${logoPrefix ? `${logoPrefix}/` : ''}${site.icon || defaultLogo}`"
              class="h-full w-full object-cover"
              alt="Site logo"
            />
          </div>

          <div class="flex flex-1 flex-col overflow-hidden">
            <div class="flex items-center gap-1">
              <h3 class="truncate text-base font-medium">
                {{ site.name }}
              </h3>
            </div>

            <p class="mt-1 truncate text-xs text-gray-500">
              {{ site.info }}
            </p>
          </div>
        </div>
      </div>

      <div v-if="filteredSites(group).length > 30" class="my-4 flex justify-center">
        <button
          class="rounded-sm bg-orange-100 px-6 text-sm text-orange-700 transition hover:bg-orange-200"
          @click="toggleGroup(group.id)"
        >
          {{ expandedGroups[group.id] ? t('common.collapse') : t('common.expand') }}
        </button>
      </div>
    </div>

    <GroupPopover
      v-if="!!activeGroup"
      :group="activeGroup"
      :target-element="activeGroupTarget"
      :visible="!!activeGroup"
      @mouseenter="cancelGroupHide"
      @mouseleave="scheduleGroupHide"
    />

    <Teleport to="body">
      <SitePopover
        v-if="!!activeSite"
        :site="activeSite"
        :target-element="activeSiteTarget"
        :visible="!!activeSite"
        :display-mode="displayMode"
        :ping-data="pingData"
        @get-popover-height="handleGetPopoverHeight"
        @mouseenter="cancelSiteHide"
        @mouseleave="scheduleSiteHide"
      />
    </Teleport>

    <FloatingSearch :items="sites" />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useLangStore } from '@/store/langStore'
import { getGroups, getPing, getSites } from '~/services/nav'
import type { Delay, Group, Site } from '~/types/nav'
import { recordRecentSite, toExternalUrl } from '@/utils/recentSites'
import { readDisplayMode, subscribeModeChange } from '@/utils/modeStorage'
import GroupPopover from './GroupPopover.vue'
import SitePopover from './SitePopover.vue'
import FloatingSearch from '@/components/nav/FloatingSearch.vue'

const props = defineProps<{
  initialGroups?: Group[]
  initialSites?: Site[]
  initialPingData?: Record<string, Delay>
}>()

const { t } = useI18n()

const groups = ref<Group[]>(props.initialGroups ?? [])
const sites = ref<Site[]>(props.initialSites ?? [])
const pingData = ref<Record<string, Delay>>(props.initialPingData ?? {})
const loading = ref(false)
const expandedGroups = ref<Record<string, boolean>>({})

const groupRefs = ref<HTMLElement[]>([])
const siteRefs = ref<HTMLElement[]>([])

const langStore = useLangStore()

const logoPrefix = import.meta.env.VITE_SITE_LOGO_PREFIX_URL || ''
const defaultLogo = 'defaultLogo.svg'

const displayMode = ref<'sfw' | 'nsfw'>(readDisplayMode())
let stopModeSubscription: (() => void) | null = null

const domainsMap = computed(() => {
  const map: Record<string, string[]> = {}

  sites.value.forEach((site) => {
    if (Array.isArray(site.domain)) {
      map[site.id] = site.domain
      return
    }

    try {
      const domainObject = JSON.parse(site.domain)
      map[site.id] = Array.isArray(domainObject?.domain) ? domainObject.domain : []
    } catch {
      map[site.id] = site.domain ? [site.domain] : []
    }
  })

  return map
})

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

async function loadData() {
  loading.value = true
  try {
    const lang = langStore.lang
    const [nextGroups, nextSites, nextPing] = await Promise.all([
      getGroups(lang),
      getSites(lang),
      getPing(),
    ])

    groups.value = nextGroups.sort((a, b) => Number(a.priority) - Number(b.priority))
    sites.value = nextSites
    parsePingData(nextPing)
  } catch (error) {
    console.error('Failed to load nav content:', error)
  } finally {
    loading.value = false
  }
}

function filteredSites(group: Group) {
  return sites.value.filter(
    (site) => group.sites.includes(site.id) && (displayMode.value === 'nsfw' || site.nsfw !== '1')
  )
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
    recordRecentSite({
      id: site.id,
      name: site.name,
      url: targetUrl,
    })
  }

  window.open(targetUrl, '_blank')
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
const activeSiteTarget = ref<HTMLElement | null>(null)
const popoverHeight = ref(0)
let siteHideTimer: number | null = null

function onSiteMouseEnter(event: MouseEvent, site: Site) {
  cancelSiteHide()
  activeSite.value = site
  activeSiteTarget.value = event.currentTarget as HTMLElement
  popoverHeight.value = 0
  nextTick(() => updateSitePopoverPosition())
}

function handleGetPopoverHeight(height: number) {
  popoverHeight.value = height
  updateSitePopoverPosition()
}

function updateSitePopoverPosition() {
  if (!activeSite.value || !activeSiteTarget.value || popoverHeight.value === 0) {
    return
  }

  const targetRect = activeSiteTarget.value.getBoundingClientRect()
  const popoverWidth = 288
  const viewportHeight = window.innerHeight
  const viewportWidth = window.innerWidth

  let left = targetRect.left + (targetRect.width - popoverWidth) / 2
  left = Math.max(8, Math.min(left, viewportWidth - popoverWidth - 8))

  const bottomPositionIfDown = targetRect.top + 90 + popoverHeight.value
  let top = bottomPositionIfDown > viewportHeight
    ? targetRect.top - popoverHeight.value - 12
    : targetRect.top + 90
  top = Math.max(8, top)

  const navWindow = window as typeof window & {
    sitePopoverUpdate?: (position: { left: number; top: number }) => void
  }

  if (navWindow.sitePopoverUpdate) {
    navWindow.sitePopoverUpdate({ left, top })
  }
}

function scheduleSiteHide() {
  siteHideTimer = window.setTimeout(() => {
    activeSite.value = null
    activeSiteTarget.value = null
    popoverHeight.value = 0
  }, 200)
}

function cancelSiteHide() {
  if (siteHideTimer) {
    clearTimeout(siteHideTimer)
    siteHideTimer = null
  }
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
  stopModeSubscription = subscribeModeChange(({ displayMode: nextMode }) => {
    displayMode.value = nextMode
  })

  window.addEventListener('scroll', handleScrollOrResize, { passive: true, capture: true })
  document.addEventListener('scroll', handleScrollOrResize, { passive: true, capture: true })
  window.addEventListener('resize', handleScrollOrResize)

  pingTimer = setInterval(async () => {
    try {
      parsePingData(await getPing())
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
  delete (window as typeof window & { sitePopoverUpdate?: unknown }).sitePopoverUpdate
})
</script>
