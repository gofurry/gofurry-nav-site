<template>
  <div class="p-6 overflow-hidden relative min-h-screen">
    <!-- 加载状态 -->
    <div v-if="loading" class="text-center py-8 text-gray-500">
      {{ t("common.loading") }}...
    </div>

    <!-- 站点分组 -->
    <div
        v-for="group in groups"
        :key="group.id"
        class="relative"
        :class='
          filteredSites(group).length > 30 ? "mb-0" : "mb-10"
        '
    >
      <!-- 分组标题 -->
      <div
          ref="groupRefs"
          class="relative inline-block cursor-pointer"
          @mouseenter="(e) => onGroupMouseEnter(e, group)"
          @mouseleave="scheduleGroupHide"
      >
        <h2 class="text-xl font-semibold hover:text-amber-600 transition-colors">
          {{ group.name }}
        </h2>
      </div>

      <!-- 站点网格 -->
      <div
          class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4 mt-2"
      >
        <div
            v-for="site in displaySites(group)"
            :key="`${group.id}-${site.id}`"
            class="bg-orange-50 rounded-xl transition-shadow duration-200 p-4 cursor-pointer flex gap-3"
            @click="goDomain(domainsMap[site.id] || [], site)"
            @mouseenter="(e) => onSiteMouseEnter(e, site)"
            @mouseleave="scheduleSiteHide"
            ref="siteRefs"
        >
          <!-- Logo -->
          <div class="w-12 h-12 rounded flex-shrink-0 overflow-hidden">
            <img
                :src="`${logoPrefix ? logoPrefix + '/' : ''}${site.icon || defaultLogo}`"
                class="w-full h-full object-cover"
                alt="Site logo"
            />
          </div>

          <!-- 内容 -->
          <div class="flex-1 flex flex-col overflow-hidden">
            <div class="flex items-center gap-1">
              <h3 class="text-base font-medium truncate">
                {{ site.name }}
              </h3>
            </div>

            <p class="text-xs text-gray-500 truncate mt-1">
              {{ site.info }}
            </p>
          </div>
        </div>
      </div>

      <!-- 按钮 -->
      <div
        v-if="filteredSites(group).length > 30"
        class="flex justify-center my-4"
      >
        <button
          class="px-6 text-sm rounded-sm bg-orange-100 hover:bg-orange-200 text-orange-700 transition"
          @click="toggleGroup(group.id)"
        >
          {{ expandedGroups[group.id] ? t("common.collapse") : t("common.expand") }}
        </button>
      </div>
    </div>

    <!-- 分组悬浮卡片组件 -->
    <GroupPopover
        v-if="!!activeGroup"
        :group="activeGroup"
        :target-element="activeGroupTarget"
        :visible="!!activeGroup"
        @mouseenter="cancelGroupHide"
        @mouseleave="scheduleGroupHide"
    />

    <!-- 站点悬浮卡片组件 -->
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
import { ref, onMounted, onUnmounted, watch, computed, nextTick } from 'vue'
import { useLangStore } from '@/store/langStore'
import { getGroups, getSites, getPing } from '@/utils/api/nav'
import type { Group, Site, Delay } from '@/types/nav'
import { i18n } from "@/main";
import { recordRecentSite, toExternalUrl } from '@/utils/recentSites'
import { readDisplayMode, subscribeModeChange } from '@/utils/modeStorage'

// 导入悬浮组件
import GroupPopover from './GroupPopover.vue'
import SitePopover from './SitePopover.vue'
import FloatingSearch from "@/components/nav/FloatingSearch.vue";

const { t } = i18n.global

// 核心数据
const groups = ref<Group[]>([])
const sites = ref<Site[]>([])
const pingData = ref<Record<string, Delay>>({})
const loading = ref(true)
const expandedGroups = ref<Record<string, boolean>>({})

// DOM引用
const groupRefs = ref<HTMLElement[]>([])
const siteRefs = ref<HTMLElement[]>([])

// 路由和语言
const langStore = useLangStore()

// 配置
const logoPrefix = import.meta.env.VITE_SITE_LOGO_PREFIX_URL || ''
const defaultLogo = 'defaultLogo.svg'

// ---------- 显示模式 ----------
const displayMode = ref<'sfw' | 'nsfw'>(
    readDisplayMode()
)

// 可选：监听 localStorage 变化（当其他组件修改 mode 时同步显示）
let stopModeSubscription: (() => void) | null = null

// ---------- 缓存域名列表 ----------
const domainsMap = computed(() => {
  const map: Record<string, string[]> = {}
  sites.value.forEach(site => {
    if (Array.isArray(site.domain)) {
      map[site.id] = site.domain
    } else {
      try {
        const obj = JSON.parse(site.domain);
        map[site.id] = Array.isArray(obj?.domain) ? obj.domain : []
      } catch {
        map[site.id] = site.domain ? [site.domain] : []
      }
    }
  })
  return map
})

// ---------- 数据加载 ----------
async function loadData() {
  loading.value = true
  try {
    const lang = langStore.lang
    const [g, s, p] = await Promise.all([getGroups(lang), getSites(lang), getPing()])
    groups.value = g.sort((a, b) => Number(a.priority) - Number(b.priority))
    sites.value = s
    parsePingData(p)
  } catch (error) {
    console.error('加载站点数据失败:', error)
  } finally {
    loading.value = false
  }
}

function parsePingData(data: Record<string, string | undefined>) {
  const result: Record<string, Delay> = {}
  for (const k in data) {
    const value = data[k]
    if (typeof value === 'string') {
      try {
        result[k] = JSON.parse(value)
      } catch {
        result[k] = { status: 'down', delay: '-', loss: '-', time: '-' }
      }
    } else {
      result[k] = { status: 'down', delay: '-', loss: '-', time: '-' }
    }
  }
  pingData.value = result
}

// ---------- 站点显示逻辑 ----------
function filteredSites(group: Group) {
  return sites.value.filter(
      s => group.sites.includes(s.id) && (displayMode.value === 'nsfw' || s.nsfw !== '1')
  )
}

function displaySites(group: Group) {
  const list = filteredSites(group)
  if (list.length <= 30) return list
  return expandedGroups.value[group.id] ? list : list.slice(0, 30)
}

function toggleGroup(groupId: string) {
  expandedGroups.value[groupId] = !expandedGroups.value[groupId]
}

function goDomain(domains: string[], site?: Site) {
  if (!domains.length) return

  const firstDomain = domains[0]
  if (!firstDomain) return

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

// ---------- 分组悬浮逻辑 ----------
const activeGroup = ref<Group | null>(null)
const activeGroupTarget = ref<HTMLElement | null>(null)
let groupHideTimer: number | null = null

function onGroupMouseEnter(e: MouseEvent, group: Group) {
  cancelGroupHide()
  activeGroup.value = group
  activeGroupTarget.value = e.currentTarget as HTMLElement
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

// ---------- 站点悬浮逻辑 ----------
const activeSite = ref<Site | null>(null)
const activeSiteTarget = ref<HTMLElement | null>(null)
const popoverHeight = ref(0)
let siteHideTimer: number | null = null

function onSiteMouseEnter(e: MouseEvent, site: Site) {
  cancelSiteHide()
  activeSite.value = site
  activeSiteTarget.value = e.currentTarget as HTMLElement
  popoverHeight.value = 0
  nextTick(() => updateSitePopoverPosition())
}

function handleGetPopoverHeight(height: number) {
  popoverHeight.value = height
  updateSitePopoverPosition()
}

function updateSitePopoverPosition() {
  if (!activeSite.value || !activeSiteTarget.value || popoverHeight.value === 0) return

  const target = activeSiteTarget.value
  const targetRect = target.getBoundingClientRect()
  const popoverWidth = 288
  const viewportHeight = window.innerHeight
  const viewportWidth = window.innerWidth

  let left = targetRect.left + (targetRect.width - popoverWidth) / 2
  left = Math.max(8, Math.min(left, viewportWidth - popoverWidth - 8))

  let top = 0
  const bottomPositionIfDown = targetRect.top + 90 + popoverHeight.value
  top = bottomPositionIfDown > viewportHeight
      ? targetRect.top - popoverHeight.value - 12
      : targetRect.top + 90
  top = Math.max(8, top)

  if ((window as any).sitePopoverUpdate) {
    (window as any).sitePopoverUpdate({ left, top })
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

// ---------- 滚动/缩放更新 ----------
function handleScrollOrResize() {
  if (activeSite.value && activeSiteTarget.value && popoverHeight.value > 0) {
    updateSitePopoverPosition()
  }
}

// ---------- 生命周期 ----------
let pingTimer: any = null

watch(
    () => langStore.lang,
    async (newLang, oldLang) => {
      if (newLang !== oldLang) {
        await loadData()
      }
    }
)

onMounted(() => {
  loadData()
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
      console.error('更新Ping数据失败:', error)
    }
  }, 60000)
})

onUnmounted(() => {
  stopModeSubscription?.()
  clearInterval(pingTimer)
  window.removeEventListener('scroll', handleScrollOrResize, { capture: true })
  document.removeEventListener('scroll', handleScrollOrResize, { capture: true })
  window.removeEventListener('resize', handleScrollOrResize)
  cancelSiteHide()
  cancelGroupHide()
  delete (window as any).sitePopoverUpdate
})
</script>
