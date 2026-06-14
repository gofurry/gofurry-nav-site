<template>
  <div
      class="games-search-page games-page relative isolate flex min-h-full w-full flex-col overflow-hidden"
  >
    <GoFurryGridBackground :fixed="false" palette="games" />
    <h1 class="sr-only">{{ searchPageSeo.heading }}</h1>
    <div class="relative z-10 mx-auto flex w-full max-w-[1720px] flex-1 flex-col gap-5 p-6">

      <div class="search-toolbar relative flex w-full items-center gap-3">
        <div class="flex-1">
          <GameSidebarSearch />
        </div>

        <button
            class="search-filter-button shrink-0"
            type="button"
            @click="showFilter = true"
        >
          {{t("game.search.advancedFilter")}}
        </button>
      </div>

      <section class="search-result-shell">
        <GameSearchResult
            :game-list="gameList"
            :current-page="query.pageNum"
            :total-pages="totalPages"
            :total="total"
            @page-change="onPageChange"
        />
      </section>

      <GameSearchFilter
          v-if="showFilter"
          :tag-groups="tagGroups"
          :query="query"
          @close="showFilter = false"
          @search="onSearch"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch, ref, reactive, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import GameSidebarSearch from '@/components/game/main/sidebar/GameSidebarSearch.vue'
import GameSearchFilter from '@/components/game/search/GameSearchFilter.vue'
import GameSearchResult from '@/components/game/search/GameSearchResult.vue'
import { searchGameAdvanced, getTagList } from '@/utils/api/game'
import { nowLocalDateTime } from '@/utils/util'
import type {
  SearchPageResponseItem,
  GameTagRecord,
  SearchPageQueryRequest
} from '@/types/game'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
import { useLangStore } from '@/store/langStore'
import { useThemeStore } from '@/stores/theme'
import { i18n } from '@/main'

const { t } = i18n.global
const { locale } = useI18n()

const langStore = useLangStore()
const themeStore = useThemeStore()
const lang = ref(langStore.lang)
const route = useRoute()
const router = useRouter()
const searchPageSeo = computed(() => locale.value === 'en'
  ? {
      heading: 'GoFurry Game Search',
      title: 'GoFurry Game Search - Find furry and anthro games',
      description: 'Search GoFurry game records by keyword, tag, score, release time, update time, and community review signals.'
    }
  : {
      heading: 'GoFurry 兽游搜索',
      title: 'GoFurry 兽游搜索 - 搜索兽人和拟人题材游戏',
      description: '按关键词、标签、评分、发布时间、更新时间和社区评价信号搜索 GoFurry 已收录的兽游资料。'
    }
)

useHead({
  meta: [
    { name: 'robots', content: 'noindex, follow' }
  ]
})

useSeoMeta({
  title: () => searchPageSeo.value.title,
  description: () => searchPageSeo.value.description,
  ogTitle: () => searchPageSeo.value.title,
  ogDescription: () => searchPageSeo.value.description,
})

type RouteQueryValue = string | null | Array<string | null>
type LocationQuery = Record<string, RouteQueryValue | undefined>
type LocationQueryRaw = Record<string, string | string[] | null | undefined>

const showFilter = ref(false)
const initialized = ref(false)

const gameList = ref<SearchPageResponseItem[]>([])
const total = ref(0)
const totalPages = ref(1)

const createDefaultQuery = (): SearchPageQueryRequest => ({
  pageNum: 1,
  pageSize: 20,
  content: '',
  pub_start_time: '2000-01-01 00:00:00',
  pub_end_time: nowLocalDateTime(),
  update_start_time: '2000-01-01 00:00:00',
  update_end_time: nowLocalDateTime(),
  score: false,
  remark_order: false,
  time_order: true,
  tag_list: []
})

const query = reactive<SearchPageQueryRequest>(createDefaultQuery())

const tagGroups = ref<GameTagRecord[]>([])

const getQueryValue = (value: LocationQuery[string] | undefined) => {
  if (Array.isArray(value)) {
    return value[0] ?? ''
  }

  return value ?? ''
}

const parsePositiveNumber = (value: LocationQuery[string] | undefined, fallback: number) => {
  const parsed = Number(getQueryValue(value))
  return Number.isFinite(parsed) && parsed > 0 ? Math.floor(parsed) : fallback
}

const parseBoolean = (value: LocationQuery[string] | undefined, fallback: boolean) => {
  const resolved = getQueryValue(value)
  if (resolved === 'true') return true
  if (resolved === 'false') return false
  return fallback
}

const parseTagList = (value: LocationQuery[string] | undefined) => {
  const resolved = getQueryValue(value)
  if (!resolved) {
    return []
  }

  return resolved
    .split(',')
    .map(item => Number(item))
    .filter(item => Number.isInteger(item) && item > 0)
}

const applyRouteQuery = (routeQuery: LocationQuery) => {
  const defaults = createDefaultQuery()

  Object.assign(query, defaults, {
    pageNum: parsePositiveNumber(routeQuery.pageNum, defaults.pageNum),
    pageSize: parsePositiveNumber(routeQuery.pageSize, defaults.pageSize),
    content: getQueryValue(routeQuery.content),
    pub_start_time: getQueryValue(routeQuery.pubStartTime) || defaults.pub_start_time,
    pub_end_time: getQueryValue(routeQuery.pubEndTime) || defaults.pub_end_time,
    update_start_time: getQueryValue(routeQuery.updateStartTime) || defaults.update_start_time,
    update_end_time: getQueryValue(routeQuery.updateEndTime) || defaults.update_end_time,
    score: parseBoolean(routeQuery.score, defaults.score ?? false),
    remark_order: parseBoolean(routeQuery.remarkOrder, defaults.remark_order ?? false),
    time_order: parseBoolean(routeQuery.timeOrder, defaults.time_order ?? true),
    tag_list: parseTagList(routeQuery.tagList),
  })
}

const buildRouteQuery = (): LocationQueryRaw => {
  const defaults = createDefaultQuery()
  const nextQuery: LocationQueryRaw = {}

  if (query.pageNum !== defaults.pageNum) {
    nextQuery.pageNum = String(query.pageNum)
  }

  if (query.pageSize !== defaults.pageSize) {
    nextQuery.pageSize = String(query.pageSize)
  }

  if (query.content?.trim()) {
    nextQuery.content = query.content.trim()
  }

  if (query.pub_start_time && query.pub_start_time !== defaults.pub_start_time) {
    nextQuery.pubStartTime = query.pub_start_time
  }

  if (query.pub_end_time && query.pub_end_time !== defaults.pub_end_time) {
    nextQuery.pubEndTime = query.pub_end_time
  }

  if (query.update_start_time && query.update_start_time !== defaults.update_start_time) {
    nextQuery.updateStartTime = query.update_start_time
  }

  if (query.update_end_time && query.update_end_time !== defaults.update_end_time) {
    nextQuery.updateEndTime = query.update_end_time
  }

  if (query.score) {
    nextQuery.score = 'true'
  }

  if (query.remark_order) {
    nextQuery.remarkOrder = 'true'
  }

  if (query.time_order !== defaults.time_order) {
    nextQuery.timeOrder = String(Boolean(query.time_order))
  }

  if (query.tag_list?.length) {
    nextQuery.tagList = query.tag_list.join(',')
  }

  return nextQuery
}

const normalizeRouteQuery = (routeQuery: LocationQuery | LocationQueryRaw) => {
  const normalized: Record<string, string> = {}

  Object.keys(routeQuery)
    .sort()
    .forEach((key) => {
      const value = routeQuery[key]

      if (Array.isArray(value)) {
        normalized[key] = value.map(item => String(item ?? '')).join(',')
        return
      }

      if (value != null) {
        normalized[key] = String(value)
      }
    })

  return JSON.stringify(normalized)
}

const loadTags = async () => {
  tagGroups.value = await getTagList(lang.value)
}

const fetchData = async () => {
  const res = await searchGameAdvanced(query, lang.value)
  gameList.value = res.list ?? []
  total.value = res.total ?? 0
  totalPages.value = Math.max(
      1,
      Math.ceil(total.value / query.pageSize)
  )
}

const syncRouteWithQuery = async () => {
  const nextQuery = buildRouteQuery()

  if (normalizeRouteQuery(route.query) === normalizeRouteQuery(nextQuery)) {
    await fetchData()
    return
  }

  await router.replace({
    path: route.path,
    query: nextQuery,
  })
}

const onPageChange = async (page: number) => {
  query.pageNum = page
  await syncRouteWithQuery()
}

const onSearch = async () => {
  query.pageNum = 1
  showFilter.value = false
  await syncRouteWithQuery()
}

watch(
    () => langStore.lang,
    async (val) => {
      lang.value = val
      await loadTags()
      await fetchData()
    }
)

watch(
  () => route.query,
  async (nextQuery, previousQuery) => {
    if (!initialized.value) {
      return
    }

    if (normalizeRouteQuery(nextQuery) === normalizeRouteQuery(previousQuery)) {
      return
    }

    applyRouteQuery(nextQuery)
    await fetchData()
  },
  { deep: true }
)

onMounted(async () => {
  themeStore.initTheme()
  applyRouteQuery(route.query)
  await loadTags()
  initialized.value = true
  await fetchData()
})
</script>
