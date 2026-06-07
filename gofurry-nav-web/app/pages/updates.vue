<template>
  <div class="updates-page" :class="{ 'is-dark-theme': isDarkTheme }">
    <GoFurryGridBackground profile="light" />

    <main class="relative z-[1] mx-auto w-[min(1100px,calc(100%-40px))] py-9 pb-24">
      <h1 class="sr-only">{{ copy.pageHeading }}</h1>
      <UpdatesSummaryBar
        :label="copy.summaryAriaLabel"
        :latest-label="copy.latestLabel"
        :latest-value="latestDateLabel"
        :count-label="copy.countLabel"
        :count-value="items.length"
        :divider-src="summaryDividerSrc"
        :dark="isDarkTheme"
      />

      <section class="mx-auto min-w-0 max-w-[920px]" :aria-busy="pending">
        <div v-if="pending" class="updates-state">
          <span class="state-line" />
          <p>{{ copy.loading }}</p>
        </div>

        <div v-else-if="error || responseState === 'error'" class="updates-state is-error">
          <span class="state-line" />
          <p>{{ errorMessage }}</p>
        </div>

        <div v-else-if="items.length === 0" class="updates-state">
          <span class="state-line" />
          <p>{{ copy.empty }}</p>
        </div>

        <ol v-else class="timeline-feed">
          <li
            v-for="(group, groupIndex) in yearGroups"
            :key="group.year"
            class="timeline-year-group"
            :style="{ '--delay': `${Math.min(groupIndex, 10) * 55}ms` }"
          >
            <UpdatesTimelineYearGroup
              :group="group"
              :expanded="isYearExpanded(group.year)"
              :visible-items="visibleItemsForYear(group.year, group.items)"
              :has-more="hasMoreInYear(group.year, group.items)"
              :latest-id="latestId"
              :latest-tag="copy.latestTag"
              :load-more-label="copy.loadMore"
              :year-summary="formatYearSummary(group.items.length)"
              :locale-code="localeCode"
              :unavailable-label="copy.unavailable"
              :dark="isDarkTheme"
              @toggle="toggleYear(group.year)"
              @load-more="loadMoreForYear(group.year)"
            />
          </li>
        </ol>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useLangStore } from '@/store/langStore'
import { useThemeStore } from '@/stores/theme'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
import updatesDividerUrl from '@/assets/svgs/updates-divider.svg'
import updatesDividerDarkUrl from '@/assets/svgs/updates-divider-dark.svg'
import { getNavUpdates } from '~/services/nav'
import type { NavUpdateNotice, NavUpdatesResponse, NavUpdatesState } from '~/types/nav'
import {
  formatUpdatesFullDate,
  formatUpdatesYear,
} from '~/utils/updatesDate'

interface YearGroup {
  year: string
  items: NavUpdateNotice[]
}

const YEAR_BATCH_SIZE = 6

const emptyUpdatesResponse = (): NavUpdatesResponse => ({
  schema_version: 1,
  generated_at: '',
  state: 'empty',
  items: [],
})

const langStore = useLangStore()
const themeStore = useThemeStore()
const lang = computed(() => langStore.lang)
const localeCode = computed(() => (lang.value === 'en' ? 'en-US' : 'zh-CN'))
const isDarkTheme = computed(() => themeStore.theme === 'dark')
const summaryDividerSrc = computed(() => (isDarkTheme.value ? updatesDividerDarkUrl : updatesDividerUrl))

const copy = computed(() => {
  if (lang.value === 'en') {
    return {
      summaryAriaLabel: 'Updates summary',
      latestLabel: 'Latest',
      countLabel: 'Entries',
      latestTag: 'Latest',
      loadMore: 'Load more',
      loading: 'Loading update notices.',
      empty: 'No update notices yet.',
      unavailable: 'Unavailable',
      errorFallback: 'Update notices are temporarily unavailable.',
      pageHeading: 'GoFurry Updates',
      seoTitle: 'GoFurry Updates - Product, navigation, and monitoring change log',
      seoDescription: 'Follow GoFurry product updates, navigation changes, site monitoring improvements, maintenance notices, and newly shipped features for the furry resource discovery platform.',
    }
  }

  return {
    summaryAriaLabel: '更新公告概览',
    latestLabel: '最新更新',
    countLabel: '公告数量',
    latestTag: '最新',
    loadMore: '加载更多',
    loading: '正在读取更新公告。',
    empty: '暂时还没有更新公告。',
    unavailable: '暂无',
    errorFallback: '更新公告暂时不可用。',
    pageHeading: 'GoFurry 更新公告',
    seoTitle: 'GoFurry 更新公告 - 产品、导航与站点监测变更记录',
    seoDescription: '查看 GoFurry 的产品更新、导航收录变化、站点监测能力改进、维护公告与近期上线功能，了解兽人资源发现平台的持续迭代。'
  }
})

const { data, pending, error } = await useAsyncData<NavUpdatesResponse>(
  () => `updates-v2-page:${lang.value}`,
  () => getNavUpdates(lang.value),
  {
    default: emptyUpdatesResponse,
    watch: [lang],
  }
)

const items = computed<NavUpdateNotice[]>(() => data.value?.items ?? [])
const responseState = computed<NavUpdatesState>(() => data.value?.state ?? 'empty')
const latest = computed(() => items.value[0] ?? null)
const latestId = computed<number | null>(() => latest.value?.id ?? null)
const expandedYears = ref<string[]>([])
const visibleCounts = ref<Record<string, number>>({})

const yearGroups = computed<YearGroup[]>(() => {
  const groups: YearGroup[] = []
  let currentGroup: YearGroup | null = null

  items.value.forEach((item) => {
    const year = formatUpdatesYear(item.published_at)
    if (!currentGroup || currentGroup.year !== year) {
      currentGroup = {
        year,
        items: [],
      }
      groups.push(currentGroup)
    }
    currentGroup.items.push(item)
  })

  return groups
})

const latestDateLabel = computed(() => {
  if (!latest.value) {
    return copy.value.unavailable
  }

  return formatUpdatesFullDate(latest.value.published_at, localeCode.value, copy.value.unavailable)
})

const errorMessage = computed(() => {
  const reasons = data.value?.reason_messages?.filter(Boolean) ?? []
  if (reasons.length > 0) {
    return reasons.join(' / ')
  }
  return copy.value.errorFallback
})

watch(
  yearGroups,
  (groups) => {
    const nextCounts: Record<string, number> = {}
    const latestYear = groups[0]?.year ?? ''

    groups.forEach((group) => {
      const previous = visibleCounts.value[group.year]
      nextCounts[group.year] = previous
        ? Math.min(previous, group.items.length)
        : Math.min(YEAR_BATCH_SIZE, group.items.length)
    })

    visibleCounts.value = nextCounts
    expandedYears.value = latestYear ? [latestYear] : []
  },
  { immediate: true }
)

onMounted(() => {
  themeStore.initTheme()
})

useSeoMeta({
  title: () => copy.value.seoTitle,
  description: () => copy.value.seoDescription,
  ogTitle: () => copy.value.seoTitle,
  ogDescription: () => copy.value.seoDescription,
})

function isYearExpanded(year: string) {
  return expandedYears.value.includes(year)
}

function toggleYear(year: string) {
  if (isYearExpanded(year)) {
    expandedYears.value = expandedYears.value.filter((item) => item !== year)
    return
  }

  expandedYears.value = [...expandedYears.value, year]
}

function visibleItemsForYear(year: string, groupItems: NavUpdateNotice[]) {
  return groupItems.slice(0, visibleCounts.value[year] ?? YEAR_BATCH_SIZE)
}

function hasMoreInYear(year: string, groupItems: NavUpdateNotice[]) {
  return (visibleCounts.value[year] ?? YEAR_BATCH_SIZE) < groupItems.length
}

function loadMoreForYear(year: string) {
  visibleCounts.value = {
    ...visibleCounts.value,
    [year]: (visibleCounts.value[year] ?? YEAR_BATCH_SIZE) + YEAR_BATCH_SIZE,
  }
}

function formatYearSummary(count: number) {
  return lang.value === 'en' ? `${count} entries` : `${count} 条`
}
</script>

<style scoped>
.updates-page {
  position: relative;
  min-height: 100svh;
  overflow: clip;
  color: #201815;
}

.updates-page.is-dark-theme {
  color: #e5edf5;
}

.updates-state {
  display: grid;
  min-height: 320px;
  place-items: center;
  gap: 18px;
  color: rgba(32, 24, 21, 0.72);
  text-align: center;
}

.updates-state p {
  margin: 0;
  font-size: 1rem;
}

.updates-state.is-error {
  color: #b42347;
}

.state-line {
  width: min(240px, 56vw);
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(15, 118, 110, 0.72), transparent);
}

.timeline-feed {
  position: relative;
  margin: 0;
  padding: 0 0 0 34px;
  list-style: none;
}

.timeline-feed::before {
  content: "";
  position: absolute;
  top: 10px;
  bottom: 0;
  left: 8px;
  width: 1px;
  background: linear-gradient(180deg, rgba(15, 118, 110, 0.55), rgba(15, 118, 110, 0.1));
}

.timeline-year-group {
  opacity: 0;
  transform: translateY(14px);
  animation: feed-enter 520ms ease forwards;
  animation-delay: var(--delay, 0ms);
}

.updates-page.is-dark-theme .updates-state {
  color: rgba(204, 223, 228, 0.76);
}

.updates-page.is-dark-theme .state-line {
  background: linear-gradient(90deg, transparent, rgba(154, 248, 251, 0.78), transparent);
}

.updates-page.is-dark-theme .timeline-feed::before {
  background: linear-gradient(180deg, rgba(127, 240, 247, 0.62), rgba(127, 240, 247, 0.1));
}

.updates-page.is-dark-theme .updates-state.is-error {
  color: #fda4af;
}

@keyframes feed-enter {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 920px) {
  main {
    width: min(100% - 28px, 1100px);
    padding-top: 30px;
    padding-bottom: 72px;
  }
}

@media (max-width: 720px) {
  .timeline-feed {
    padding-left: 26px;
  }

  .timeline-feed::before {
    left: 4px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .timeline-year-group {
    animation: none;
    opacity: 1;
    transform: none;
  }
}
</style>
