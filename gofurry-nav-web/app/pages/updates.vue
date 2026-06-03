<template>
  <div class="updates-page" :class="{ 'is-dark-theme': isDarkTheme }">
    <GoFurryGridBackground />

    <main class="updates-shell">
      <section class="updates-summary-bar" :aria-label="copy.summaryAriaLabel">
        <div class="summary-inline">
          <div class="summary-inline-item">
            <span>{{ copy.latestLabel }}</span>
            <strong>{{ latestDateLabel }}</strong>
          </div>
          <span class="summary-divider" aria-hidden="true">
            <img class="summary-divider-image" :src="summaryDividerSrc" alt="" />
          </span>
          <div class="summary-inline-item">
            <span>{{ copy.countLabel }}</span>
            <strong>{{ items.length }}</strong>
          </div>
        </div>
      </section>

      <section class="timeline-section" :aria-busy="pending">
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
            <button
              type="button"
              class="year-divider year-toggle"
              :class="{ 'is-expanded': isYearExpanded(group.year) }"
              @click="toggleYear(group.year)"
            >
              <span class="year-divider-text">{{ group.year }}</span>
              <span class="year-divider-meta">{{ formatYearSummary(group.items.length) }}</span>
              <span class="year-divider-chevron" aria-hidden="true" />
            </button>

            <ol v-if="isYearExpanded(group.year)" class="year-entries">
              <li
                v-for="item in visibleItemsForYear(group.year, group.items)"
                :key="item.id"
                class="timeline-entry-wrap"
              >
                <article
                  class="timeline-entry"
                  :class="{ 'is-latest': item.id === latest?.id }"
                  tabindex="0"
                >
                  <div class="entry-marker" aria-hidden="true" />

                  <time class="entry-stamp" :datetime="item.published_at">
                    <span class="entry-month">{{ formatMonthDay(item.published_at) }}</span>
                    <span class="entry-time">{{ formatClock(item.published_at) }}</span>
                  </time>

                  <div class="entry-copy">
                    <div class="entry-heading">
                      <h2>{{ item.title }}</h2>
                      <span v-if="item.id === latest?.id" class="entry-tag">{{ copy.latestTag }}</span>
                    </div>
                    <p class="entry-body">{{ item.body }}</p>
                    <p class="entry-meta">{{ formatFullDate(item.published_at) }}</p>
                  </div>
                </article>
              </li>
              <li v-if="hasMoreInYear(group.year, group.items)" class="year-load-more-wrap">
                <button type="button" class="year-load-more" @click="loadMoreForYear(group.year)">
                  {{ copy.loadMore }}
                </button>
              </li>
            </ol>
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
      seoTitle: 'GoFurry Updates',
      seoDescription: 'Latest product and maintenance updates from GoFurry.',
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
    seoTitle: 'GoFurry Updates',
    seoDescription: 'GoFurry 的最新产品更新与维护记录。',
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
const expandedYears = ref<string[]>([])
const visibleCounts = ref<Record<string, number>>({})

const yearGroups = computed<YearGroup[]>(() => {
  const groups: YearGroup[] = []
  let currentGroup: YearGroup | null = null

  items.value.forEach((item) => {
    const year = formatYear(item.published_at)
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
  return formatFullDate(latest.value.published_at)
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

    groups.forEach((group, index) => {
      const previous = visibleCounts.value[group.year]
      const nextVisibleCount = previous
        ? Math.min(previous, group.items.length)
        : Math.min(YEAR_BATCH_SIZE, group.items.length)
      nextCounts[group.year] = nextVisibleCount
      if (index === 0 && nextVisibleCount < Math.min(YEAR_BATCH_SIZE, group.items.length)) {
        nextCounts[group.year] = Math.min(YEAR_BATCH_SIZE, group.items.length)
      }
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

function parseDate(value: string) {
  if (!value) {
    return null
  }
  const normalized = value.includes('T') ? value : value.replace(' ', 'T')
  const parsed = new Date(normalized)
  if (Number.isNaN(parsed.getTime())) {
    return null
  }
  return parsed
}

function formatMonthDay(value: string) {
  const date = parseDate(value)
  if (!date) {
    return '--.--'
  }
  return new Intl.DateTimeFormat(localeCode.value, {
    month: '2-digit',
    day: '2-digit',
  }).format(date)
}

function formatClock(value: string) {
  const date = parseDate(value)
  if (!date) {
    return '--:--'
  }
  return new Intl.DateTimeFormat(localeCode.value, {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(date)
}

function formatYear(value: string) {
  const date = parseDate(value)
  if (!date) {
    return '----'
  }
  return String(date.getFullYear())
}

function formatFullDate(value: string) {
  const date = parseDate(value)
  if (!date) {
    return value || copy.value.unavailable
  }
  return new Intl.DateTimeFormat(localeCode.value, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(date)
}
</script>

<style scoped>
.updates-page {
  position: relative;
  min-height: 100svh;
  overflow: clip;
  color: #201815;
}

.updates-shell {
  position: relative;
  z-index: 1;
  width: min(1100px, calc(100% - 40px));
  margin: 0 auto;
  padding: 36px 0 96px;
}

.updates-summary-bar {
  display: flex;
  justify-content: center;
  padding: 18px 0 30px;
}

.summary-inline {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 22px;
  padding: 12px 24px;
  border-bottom: 1px solid rgba(62, 50, 41, 0.12);
}

.summary-inline-item {
  display: inline-flex;
  align-items: baseline;
  gap: 12px;
}

.summary-inline-item span {
  color: rgba(32, 24, 21, 0.54);
  font-size: 0.76rem;
  text-transform: uppercase;
}

.summary-inline-item strong {
  overflow-wrap: anywhere;
  color: rgba(31, 23, 19, 0.8);
  font-size: clamp(1rem, 1.8vw, 1.12rem);
  font-weight: 780;
  line-height: 1.5;
}

.summary-divider {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 88px;
  height: 20px;
}

.summary-divider-image {
  width: 88px;
  height: 20px;
  display: block;
}

.timeline-section {
  min-width: 0;
  max-width: 920px;
  margin: 0 auto;
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

.timeline-year-group,
.timeline-entry-wrap {
  opacity: 0;
  transform: translateY(14px);
  animation: feed-enter 520ms ease forwards;
  animation-delay: var(--delay, 0ms);
}

.timeline-year-group {
  margin-bottom: 10px;
}

.year-divider {
  position: relative;
  display: flex;
  width: 100%;
  align-items: center;
  gap: 14px;
  margin: 8px 0 14px;
  padding-left: 28px;
}

.year-divider::before {
  content: "";
  position: absolute;
  top: 50%;
  left: -34px;
  width: 18px;
  height: 1px;
  background: rgba(15, 118, 110, 0.5);
}

.year-divider-text {
  color: rgba(32, 24, 21, 0.42);
  font-size: 0.76rem;
  font-weight: 700;
  text-transform: uppercase;
}

.year-divider-meta {
  color: rgba(32, 24, 21, 0.36);
  font-size: 0.78rem;
}

.year-toggle {
  border: 0;
  background: transparent;
  cursor: pointer;
  text-align: left;
}

.year-divider-chevron {
  position: relative;
  flex: 0 0 auto;
  width: 10px;
  height: 10px;
  margin-left: auto;
}

.year-divider-chevron::before,
.year-divider-chevron::after {
  content: "";
  position: absolute;
  top: 50%;
  width: 6px;
  height: 1px;
  background: rgba(32, 24, 21, 0.42);
  transition: transform 180ms ease;
}

.year-divider-chevron::before {
  left: 0;
  transform: translateY(-50%) rotate(45deg);
}

.year-divider-chevron::after {
  right: 0;
  transform: translateY(-50%) rotate(-45deg);
}

.year-toggle.is-expanded .year-divider-chevron::before {
  transform: translateY(-50%) rotate(-45deg);
}

.year-toggle.is-expanded .year-divider-chevron::after {
  transform: translateY(-50%) rotate(45deg);
}

.year-entries {
  margin: 0;
  padding: 0;
  list-style: none;
}

.timeline-entry {
  position: relative;
  display: grid;
  grid-template-columns: 126px minmax(0, 1fr);
  gap: clamp(18px, 4vw, 34px);
  padding: 0 0 28px;
  border-bottom: 1px solid rgba(62, 50, 41, 0.12);
  outline: none;
}

.timeline-entry-wrap + .timeline-entry-wrap .timeline-entry,
.timeline-year-group + .timeline-year-group .timeline-entry {
  padding-top: 18px;
}

.year-entries > .timeline-entry-wrap:first-child .timeline-entry {
  padding-top: 18px;
}

.entry-marker {
  position: absolute;
  top: 26px;
  left: -29px;
  width: 11px;
  height: 11px;
  border: 2px solid #0f766e;
  background: rgba(255, 251, 247, 0.92);
  transform: rotate(45deg);
  transition:
    transform 180ms ease,
    box-shadow 180ms ease,
    background-color 180ms ease;
}

.timeline-entry:hover .entry-marker,
.timeline-entry:focus-visible .entry-marker {
  background: #0f766e;
  box-shadow: 0 0 0 9px rgba(15, 118, 110, 0.12);
  transform: rotate(45deg) scale(1.08);
}

.timeline-entry.is-latest .entry-marker {
  background: #0f766e;
  animation: marker-pulse 2400ms ease-in-out infinite;
}

.entry-stamp {
  display: grid;
  align-content: start;
  gap: 6px;
  padding-top: 4px;
  padding-left: 14px;
  padding-right: 14px;
  color: rgba(32, 24, 21, 0.54);
  font-variant-numeric: tabular-nums;
}

.entry-month {
  font-size: clamp(1.3rem, 3vw, 1.9rem);
  font-weight: 820;
  line-height: 1;
}

.entry-time {
  font-size: 0.82rem;
}

.entry-copy {
  min-width: 0;
  padding-left: 18px;
  padding-right: min(8vw, 48px);
  transition: transform 180ms ease;
}

.timeline-entry:hover .entry-copy,
.timeline-entry:focus-visible .entry-copy {
  transform: translateX(4px);
}

.entry-heading {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 16px;
}

.entry-heading h2 {
  margin: 0;
  color: rgba(31, 23, 19, 0.82);
  font-size: clamp(1.38rem, 2.4vw, 2.2rem);
  font-weight: 830;
  line-height: 1.2;
  overflow-wrap: anywhere;
}

.timeline-entry.is-latest .entry-heading h2 {
  font-size: clamp(1.6rem, 2.8vw, 2.5rem);
}

.entry-tag {
  flex: 0 0 auto;
  border: 1px solid rgba(15, 118, 110, 0.22);
  background: rgba(15, 118, 110, 0.08);
  padding: 0.28rem 0.54rem;
  color: #0f766e;
  font-size: 0.72rem;
  font-weight: 900;
  text-transform: uppercase;
}

.entry-body {
  max-width: 820px;
  margin: 18px 0 0;
  color: rgba(32, 24, 21, 0.66);
  font-size: 1rem;
  line-height: 1.9;
  white-space: pre-line;
  overflow-wrap: anywhere;
}

.entry-meta {
  margin: 18px 0 0;
  color: rgba(32, 24, 21, 0.38);
  font-size: 0.84rem;
  font-variant-numeric: tabular-nums;
}

.year-load-more-wrap {
  padding: 18px 0 8px 28px;
}

.year-load-more {
  border: 1px solid rgba(15, 118, 110, 0.18);
  background: rgba(255, 251, 247, 0.5);
  padding: 0.58rem 0.9rem;
  color: #0f766e;
  font-size: 0.82rem;
  font-weight: 700;
  transition:
    border-color 180ms ease,
    background-color 180ms ease,
    transform 180ms ease;
}

.year-load-more:hover,
.year-load-more:focus-visible {
  border-color: rgba(15, 118, 110, 0.34);
  background: rgba(255, 251, 247, 0.82);
  transform: translateY(-1px);
}

@keyframes feed-enter {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes marker-pulse {
  0%,
  100% {
    box-shadow: 0 0 0 0 rgba(15, 118, 110, 0.22);
  }
  50% {
    box-shadow: 0 0 0 12px rgba(15, 118, 110, 0);
  }
}

.updates-page.is-dark-theme {
  color: #e5edf5;
}

.updates-page.is-dark-theme .summary-inline,
.updates-page.is-dark-theme .timeline-entry {
  border-color: rgba(151, 224, 236, 0.18);
}

.updates-page.is-dark-theme .summary-inline {
  background: rgba(6, 20, 29, 0.22);
  box-shadow:
    inset 0 -1px 0 rgba(171, 234, 243, 0.08),
    0 10px 30px rgba(0, 0, 0, 0.12);
  backdrop-filter: blur(10px);
}

.updates-page.is-dark-theme .timeline-entry {
  background: linear-gradient(90deg, rgba(6, 20, 29, 0.18), rgba(6, 20, 29, 0.04));
}

.updates-page.is-dark-theme .timeline-entry:hover,
.updates-page.is-dark-theme .timeline-entry:focus-visible {
  background: linear-gradient(90deg, rgba(9, 28, 40, 0.3), rgba(9, 28, 40, 0.1));
}

.updates-page.is-dark-theme .entry-tag {
  color: rgba(225, 242, 246, 0.84);
}

.updates-page.is-dark-theme .summary-inline-item strong,
.updates-page.is-dark-theme .entry-heading h2 {
  color: rgba(226, 238, 242, 0.84);
  text-shadow: 0 1px 18px rgba(0, 0, 0, 0.22);
}

.updates-page.is-dark-theme .entry-month {
  color: rgba(215, 233, 238, 0.82);
}

.updates-page.is-dark-theme .entry-time {
  color: rgba(182, 214, 221, 0.72);
}

.updates-page.is-dark-theme .updates-state,
.updates-page.is-dark-theme .entry-body {
  color: rgba(204, 223, 228, 0.76);
}

.updates-page.is-dark-theme .summary-inline-item span,
.updates-page.is-dark-theme .year-divider-text,
.updates-page.is-dark-theme .year-divider-meta,
.updates-page.is-dark-theme .entry-stamp,
.updates-page.is-dark-theme .entry-meta {
  color: rgba(174, 205, 212, 0.68);
}

.updates-page.is-dark-theme .year-divider-text {
  color: rgba(210, 231, 236, 0.8);
}

.updates-page.is-dark-theme .year-divider-meta {
  color: rgba(166, 199, 206, 0.64);
}

.updates-page.is-dark-theme .timeline-feed::before {
  background: linear-gradient(180deg, rgba(127, 240, 247, 0.62), rgba(127, 240, 247, 0.1));
}

.updates-page.is-dark-theme .year-divider::before {
  background: rgba(127, 240, 247, 0.56);
}

.updates-page.is-dark-theme .year-divider-chevron::before,
.updates-page.is-dark-theme .year-divider-chevron::after {
  background: rgba(224, 248, 252, 0.86);
}

.updates-page.is-dark-theme .entry-marker {
  border-color: rgba(153, 245, 250, 0.96);
  background: rgba(10, 24, 34, 0.96);
}

.updates-page.is-dark-theme .timeline-entry:hover .entry-marker,
.updates-page.is-dark-theme .timeline-entry:focus-visible .entry-marker,
.updates-page.is-dark-theme .timeline-entry.is-latest .entry-marker {
  background: #9af8fb;
}

.updates-page.is-dark-theme .timeline-entry:hover .entry-marker,
.updates-page.is-dark-theme .timeline-entry:focus-visible .entry-marker {
  box-shadow: 0 0 0 9px rgba(154, 248, 251, 0.14);
}

.updates-page.is-dark-theme .entry-tag {
  border-color: rgba(154, 248, 251, 0.3);
  background: rgba(154, 248, 251, 0.12);
  box-shadow: inset 0 0 0 1px rgba(247, 254, 255, 0.04);
}

.updates-page.is-dark-theme .updates-state.is-error {
  color: #fda4af;
}

.updates-page.is-dark-theme .year-load-more {
  border-color: rgba(154, 248, 251, 0.28);
  background: rgba(7, 24, 34, 0.74);
  color: rgba(223, 239, 243, 0.82);
  box-shadow: inset 0 0 0 1px rgba(247, 254, 255, 0.04);
}

.updates-page.is-dark-theme .year-load-more:hover,
.updates-page.is-dark-theme .year-load-more:focus-visible {
  border-color: rgba(154, 248, 251, 0.44);
  background: rgba(10, 31, 44, 0.9);
  color: rgba(232, 245, 248, 0.9);
}

.updates-page.is-dark-theme .state-line {
  background: linear-gradient(90deg, transparent, rgba(154, 248, 251, 0.78), transparent);
}

@media (max-width: 920px) {
  .updates-shell {
    width: min(100% - 28px, 1100px);
    padding: 30px 0 72px;
  }
}

@media (max-width: 720px) {
  .updates-summary-bar {
    padding-bottom: 24px;
  }

  .summary-inline {
    display: grid;
    justify-items: center;
    gap: 14px;
    padding: 12px 18px;
  }

  .summary-inline-item {
    flex-wrap: wrap;
    justify-content: center;
    gap: 6px 10px;
  }

  .summary-divider {
    width: 88px;
    height: 20px;
  }

  .timeline-feed {
    padding-left: 26px;
  }

  .timeline-feed::before {
    left: 4px;
  }

  .year-divider {
    padding-left: 18px;
  }

  .year-divider::before {
    left: -22px;
    width: 14px;
  }

  .timeline-entry {
    grid-template-columns: minmax(0, 1fr);
    gap: 12px;
  }

  .entry-marker {
    top: 18px;
    left: -25px;
  }

  .entry-stamp {
    display: flex;
    align-items: baseline;
    gap: 10px;
    padding-left: 0;
    padding-right: 0;
  }

  .entry-copy {
    padding-left: 0;
    padding-right: 0;
  }

  .entry-heading {
    flex-wrap: wrap;
  }

  .timeline-entry:hover .entry-copy,
  .timeline-entry:focus-visible .entry-copy {
    transform: translateY(-2px);
  }
}

@media (prefers-reduced-motion: reduce) {
  .timeline-year,
  .timeline-entry-wrap,
  .timeline-entry.is-latest .entry-marker {
    animation: none;
  }

  .timeline-year,
  .timeline-entry-wrap {
    opacity: 1;
    transform: none;
  }

  .entry-copy,
  .entry-marker {
    transition: none;
  }
}
</style>
