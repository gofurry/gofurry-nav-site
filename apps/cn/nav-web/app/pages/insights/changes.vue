<template>
  <div class="insights-page insights-change-explorer" :data-domain="selectedDomain" :data-range="selectedRange">
    <main class="insights-container">
      <EcosystemNavigation />
      <h1 class="sr-only">{{ $t('insights.changeExplorer.title') }}</h1>

      <section class="insights-change-explorer-filters" :aria-label="$t('insights.changeExplorer.filters')">
        <div class="insights-change-explorer-filters__group">
          <span>{{ $t('insights.changeExplorer.domain') }}</span>
          <button
            v-for="domain in domains"
            :key="domain"
            type="button"
            :class="{ active: selectedDomain === domain }"
            :aria-pressed="selectedDomain === domain"
            :data-domain-filter="domain"
            @click="selectDomain(domain)"
          >
            {{ $t(`insights.changes.${domain}`) }}
          </button>
        </div>
        <div class="insights-change-explorer-filters__group">
          <span>{{ $t('insights.changeExplorer.range') }}</span>
          <button
            v-for="range in ranges"
            :key="range"
            type="button"
            :class="{ active: selectedRange === range }"
            :aria-pressed="selectedRange === range"
            :data-change-range="range"
            @click="selectRange(range)"
          >
            {{ $t(`insights.ranges.${range}`) }}
          </button>
        </div>
        <div class="insights-change-explorer-filters__group insights-change-explorer-filters__categories">
          <span>{{ $t('insights.changeExplorer.category') }}</span>
          <button
            type="button"
            :class="{ active: selectedCategory === '' }"
            :aria-pressed="selectedCategory === ''"
            data-category-filter="all"
            @click="selectCategory('')"
          >
            {{ $t('insights.changeExplorer.allCategories') }}
          </button>
          <button
            v-for="category in categories"
            :key="category"
            type="button"
            :class="{ active: selectedCategory === category }"
            :aria-pressed="selectedCategory === category"
            :data-category-filter="category"
            @click="selectCategory(category)"
          >
            {{ $t(`insights.changeExplorer.categories.${selectedDomain}.${category}`) }}
          </button>
        </div>
      </section>

      <InsightsChangeExplorerFeed
        :items="items"
        :next-cursor="nextCursor"
        :loading="loading"
        :unavailable="unavailable"
        @more="loadMore"
      />
    </main>
  </div>
</template>

<script setup lang="ts">
import EcosystemNavigation from '@/components/insights/EcosystemNavigation.vue'
import { computed, onMounted, ref, watch } from 'vue'
import type { LocationQueryRaw } from 'vue-router'
import { useI18n } from 'vue-i18n'
import InsightsChangeExplorerFeed from '@/components/insights/InsightsChangeExplorerFeed.vue'
import { getGameInsightChanges } from '@/services/game'
import { getNavInsightChanges } from '@/services/nav'
import type {
  GameInsightChangeCategory,
  InsightChangeCategory,
  InsightChangeExplorerPage,
  InsightChangeRange,
  InsightDomain,
  InsightExplorerChange,
  SiteInsightChangeCategory,
} from '@/types/insights'
import { buildInsightsSeo } from '@/utils/seo'

const domains = ['site', 'game'] as const satisfies readonly InsightDomain[]
const ranges = ['7d', '30d', '90d', 'all'] as const satisfies readonly InsightChangeRange[]
const siteCategories = ['capability', 'target', 'certificate'] as const satisfies readonly SiteInsightChangeCategory[]
const gameCategories = ['pricing_model', 'platform', 'release', 'price', 'discount'] as const satisfies readonly GameInsightChangeCategory[]
const route = useRoute()
const router = useRouter()
const { locale } = useI18n()
const initialDomain = normalizeDomain(queryValue(route.query.domain))
const initialRange = normalizeRange(queryValue(route.query.range))
const initialCategory = normalizeCategory(initialDomain, queryValue(route.query.category))

const snapshot = useAsyncData<InsightChangeExplorerPage | null>(
  `insights:changes:${initialDomain}:${initialRange}:${initialCategory || 'all'}`,
  async () => {
    try {
      return await fetchPage(initialDomain, initialRange, initialCategory, '')
    } catch {
      return null
    }
  },
  { default: () => null },
)

const selectedDomain = computed<InsightDomain>(() => normalizeDomain(queryValue(route.query.domain)))
const selectedRange = computed<InsightChangeRange>(() => normalizeRange(queryValue(route.query.range)))
const selectedCategory = computed<InsightChangeCategory | ''>(() =>
  normalizeCategory(selectedDomain.value, queryValue(route.query.category)),
)
const categories = computed<readonly InsightChangeCategory[]>(() =>
  selectedDomain.value === 'site' ? siteCategories : gameCategories,
)
const items = ref<InsightExplorerChange[]>([])
const nextCursor = ref<string | null>(null)
const unavailable = ref(false)
const loading = ref(false)
const interactionReady = ref(false)
let requestSequence = 0

async function loadFirstPage() {
  const sequence = ++requestSequence
  items.value = []
  nextCursor.value = null
  unavailable.value = false
  loading.value = true
  try {
    const page = await fetchPage(selectedDomain.value, selectedRange.value, selectedCategory.value, '')
    if (sequence === requestSequence) {
      items.value = page.items
      nextCursor.value = page.next_cursor
    }
  } catch {
    if (sequence === requestSequence) unavailable.value = true
  } finally {
    if (sequence === requestSequence) loading.value = false
  }
}

async function loadMore() {
  if (!nextCursor.value || loading.value) return
  const sequence = ++requestSequence
  const cursor = nextCursor.value
  unavailable.value = false
  loading.value = true
  try {
    const page = await fetchPage(selectedDomain.value, selectedRange.value, selectedCategory.value, cursor)
    if (sequence === requestSequence) {
      items.value = [...items.value, ...page.items]
      nextCursor.value = page.next_cursor
    }
  } catch {
    if (sequence === requestSequence) unavailable.value = true
  } finally {
    if (sequence === requestSequence) loading.value = false
  }
}

function selectDomain(domain: InsightDomain) {
  if (domain === selectedDomain.value) return
  const query: LocationQueryRaw = { ...route.query, domain, range: selectedRange.value }
  delete query.category
  delete query.type
  delete query.cursor
  void router.push({ path: route.path, query, hash: route.hash })
}

function selectRange(range: InsightChangeRange) {
  if (range === selectedRange.value) return
  const query: LocationQueryRaw = { ...route.query, domain: selectedDomain.value, range }
  delete query.cursor
  void router.push({ path: route.path, query, hash: route.hash })
}

function selectCategory(category: InsightChangeCategory | '') {
  if (category === selectedCategory.value) return
  const query: LocationQueryRaw = { ...route.query, domain: selectedDomain.value, range: selectedRange.value }
  if (category) query.category = category
  else delete query.category
  delete query.type
  delete query.cursor
  void router.push({ path: route.path, query, hash: route.hash })
}

watch([selectedDomain, selectedRange, selectedCategory], () => {
  if (interactionReady.value) void loadFirstPage()
})

onMounted(async () => {
  if (queryValue(route.query.domain) !== selectedDomain.value || queryValue(route.query.range) !== selectedRange.value ||
    queryValue(route.query.category) !== selectedCategory.value || route.query.cursor !== undefined) {
    const query: LocationQueryRaw = { ...route.query, domain: selectedDomain.value, range: selectedRange.value }
    if (selectedCategory.value) query.category = selectedCategory.value
    else delete query.category
    delete query.cursor
    await router.replace({ path: route.path, query, hash: route.hash })
  }
  interactionReady.value = true
})

const { data } = await snapshot
if (data.value) {
  items.value = data.value.items
  nextCursor.value = data.value.next_cursor
} else {
  unavailable.value = true
}

const seo = computed(() => buildInsightsSeo('changes', locale.value))
useSeoMeta({
  title: () => seo.value.title,
  description: () => seo.value.description,
  ogTitle: () => seo.value.title,
  ogDescription: () => seo.value.description,
})

function fetchPage(domain: InsightDomain, range: InsightChangeRange, category: InsightChangeCategory | '', cursor: string) {
  const common = { range, cursor: cursor || undefined, limit: 20 }
  if (domain === 'site') {
    return getNavInsightChanges({
      ...common,
      category: category as SiteInsightChangeCategory | '',
    })
  }
  return getGameInsightChanges({
    ...common,
    category: category as GameInsightChangeCategory | '',
  })
}

function queryValue(value: unknown) {
  return typeof value === 'string' ? value : ''
}

function normalizeDomain(value: string): InsightDomain {
  return value === 'game' ? 'game' : 'site'
}

function normalizeRange(value: string): InsightChangeRange {
  return value === '7d' || value === '90d' || value === 'all' ? value : '30d'
}

function normalizeCategory(domain: InsightDomain, value: string): InsightChangeCategory | '' {
  const allowed: readonly InsightChangeCategory[] = domain === 'site' ? siteCategories : gameCategories
  return allowed.includes(value as InsightChangeCategory) ? value as InsightChangeCategory : ''
}
</script>
