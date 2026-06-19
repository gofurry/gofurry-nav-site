<template>
  <div class="nav-content">
    <NavSpotlightPanels
      :spotlight="spotlight"
      :display-mode="displayMode"
    />

    <div
      v-for="group in groups"
      :key="group.id"
      class="nav-group-section relative"
    >
      <div class="mb-3 flex items-center justify-between gap-4">
        <div
          ref="groupRefs"
          class="relative inline-block cursor-pointer"
          @mouseenter="(event) => onGroupMouseEnter(event, group)"
          @mouseleave="scheduleGroupHide"
        >
          <h2 class="nav-group-title">
            {{ group.name }}
          </h2>
        </div>

        <button
          v-if="group.has_more || Number(group.site_count || 0) > group.sites.length"
          type="button"
          class="nav-group-toggle"
          @click="goToGroupDetail(group)"
        >
          {{ t('common.showMore') }}
        </button>
      </div>
      <NavSiteGrid :sites="group.sites" :ping-data="pingData" />
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

  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getNavHomePing } from '~/services/nav'
import type { Delay, Group, NavHomeSpotlight } from '~/types/nav'
import { readDisplayMode, subscribeModeChange, type DisplayMode } from '@/utils/modeStorage'
import GroupPopover from './GroupPopover.vue'
import NavSpotlightPanels from './NavSpotlightPanels.vue'
import NavSiteGrid from './NavSiteGrid.vue'

const props = defineProps<{
  initialGroups?: Group[]
  initialSpotlight?: NavHomeSpotlight
  initialPingData?: Record<string, Delay>
}>()

const { t } = useI18n()

const groups = ref<Group[]>(props.initialGroups ?? [])
const spotlight = ref<NavHomeSpotlight>(props.initialSpotlight ?? { page_size: 6, featured: [], popular: [], latest: [], random: [] })
const pingData = ref<Record<string, Delay>>(props.initialPingData ?? {})
const displayMode = ref<DisplayMode>('sfw')

const groupRefs = ref<HTMLElement[]>([])
const router = useRouter()
const localePath = useLocalePath()
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

function goToGroupDetail(group: Group) {
  const targetPath = group.detail_path || `/site-groups/${group.id}`
  void router.push(localePath(targetPath))
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

let pingTimer: ReturnType<typeof setInterval> | null = null

watch(
  () => props.initialGroups,
  (value) => {
    groups.value = [...(value ?? [])].sort((a, b) => Number(a.priority) - Number(b.priority))
  },
  { immediate: true }
)

watch(
  () => props.initialSpotlight,
  (value) => {
    spotlight.value = value ?? { page_size: 6, featured: [], popular: [], latest: [], random: [] }
  },
  { immediate: true }
)

watch(
  () => props.initialPingData,
  (value) => {
    pingData.value = value ?? {}
  },
  { immediate: true }
)

onMounted(() => {
  displayMode.value = readDisplayMode()
  stopModeSubscription = subscribeModeChange(({ displayMode: nextMode }) => {
    displayMode.value = nextMode
  })

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
  cancelGroupHide()
})
</script>
