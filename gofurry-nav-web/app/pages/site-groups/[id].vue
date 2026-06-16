<template>
  <div class="games-page game-detail-page relative isolate min-h-screen overflow-hidden">
    <GoFurryGridBackground :fixed="false" palette="games" />

    <main class="site-group-content relative z-10 mx-auto w-full max-w-[1880px] px-4 pb-16 pt-6 sm:px-6 lg:pt-8 xl:px-8">
      <header class="site-group-header mb-6">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div class="flex min-w-0 flex-wrap items-baseline gap-x-3 gap-y-1">
            <h1 class="site-group-title truncate text-3xl font-semibold">
              {{ groupInfo?.name || route.params.id }}
            </h1>
            <span class="site-group-total text-sm font-semibold">
              {{ total }} {{ t('common.record') }}
            </span>
          </div>

          <button type="button" class="nav-group-toggle self-start sm:self-auto" @click="goHome">
            {{ t('archive.actions.backHome') }}
          </button>
        </div>

        <p class="site-group-summary mt-3 text-sm leading-6">
          {{ groupInfo?.info || fallbackDescription }}
        </p>
      </header>

      <section class="game-detail-tabs p-4 sm:p-5">
        <NavSiteGrid :sites="items" :ping-data="pingData" />

        <div v-if="state === 'missing'" class="game-detail-empty mt-8 rounded-xl px-4 py-3 text-sm">
          {{ missingText }}
        </div>

        <div v-else-if="!items.length" class="game-detail-empty mt-8 rounded-xl px-4 py-3 text-sm">
          {{ emptyText }}
        </div>

        <div v-if="hasMore" ref="sentinelRef" class="mt-10 flex justify-center">
          <button type="button" class="game-detail-load-more px-5 py-2 text-sm font-semibold" :disabled="isLoadingMore" @click="loadMore">
            {{ isLoadingMore ? t('common.loading') : t('common.loadMore') }}
          </button>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
import NavSiteGrid from '@/components/nav/NavSiteGrid.vue'
import { getNavHomePing, getNavSiteGroupPage } from '~/services/nav'
import type { Delay, NavSiteGroupPageResponse, Site } from '~/types/nav'

const route = useRoute()
const router = useRouter()
const localePath = useLocalePath()
const { t, locale } = useI18n()
const pageSize = 24
const defaultResponse: NavSiteGroupPageResponse = {
  schema_version: 1,
  generated_at: '',
  state: 'missing',
  page: 1,
  page_size: pageSize,
  total: 0,
  has_more: false,
  items: [],
}

const lang = computed(() => (locale.value === 'en' ? 'en' : 'zh'))
const sentinelRef = ref<HTMLElement | null>(null)
const items = ref<Site[]>([])
const page = ref(1)
const total = ref(0)
const hasMore = ref(false)
const state = ref<NavSiteGroupPageResponse['state']>('missing')
const groupInfo = ref<NavSiteGroupPageResponse['group'] | null>(null)
const isLoadingMore = ref(false)
const pingData = ref<Record<string, Delay>>({})
let observer: IntersectionObserver | null = null
let pingTimer: ReturnType<typeof setInterval> | null = null

const { data } = await useAsyncData<NavSiteGroupPageResponse>(
  () => `nav-site-group:${String(route.params.id)}:${lang.value}`,
  () => getNavSiteGroupPage(String(route.params.id), lang.value, 1, pageSize),
  {
    watch: [() => route.params.id, lang],
    default: () => defaultResponse,
  }
)

const groupPage = computed<NavSiteGroupPageResponse>(() => data.value ?? defaultResponse)
const fallbackDescription = computed(() => locale.value === 'en' ? 'Browse the complete site list for this group.' : '查看这个分组下的完整站点列表。')
const missingText = computed(() => groupPage.value.reason_messages?.[0] || (locale.value === 'en' ? 'This group cache is temporarily unavailable.' : '这个分组的缓存暂时不可用。'))
const emptyText = computed(() => locale.value === 'en' ? 'No sites available in this group yet.' : '这个分组下暂时还没有可展示的站点。')

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

async function refreshPingData() {
  try {
    const response = await getNavHomePing()
    parsePingData(response.ping)
  } catch {
    pingData.value = {}
  }
}

watch(
  groupPage,
  (value) => {
    items.value = [...value.items]
    page.value = value.page
    total.value = value.total
    hasMore.value = Boolean(value.has_more)
    state.value = value.state
    groupInfo.value = value.group ?? null
  },
  { immediate: true }
)

useHead(() => ({
  title: groupInfo.value?.name || String(route.params.id),
}))

async function loadMore() {
  if (isLoadingMore.value || !hasMore.value) {
    return
  }

  isLoadingMore.value = true
  try {
    const nextPage = page.value + 1
    const response = await getNavSiteGroupPage(String(route.params.id), lang.value, nextPage, pageSize)
    const existing = new Set(items.value.map(item => item.id))
    const nextItems = response.items.filter(item => !existing.has(item.id))
    items.value = [...items.value, ...nextItems]
    page.value = response.page
    total.value = response.total
    hasMore.value = response.has_more
    state.value = response.state
    groupInfo.value = response.group ?? groupInfo.value
  } finally {
    isLoadingMore.value = false
  }
}

function goHome() {
  void router.push(localePath('/'))
}

function setupObserver() {
  if (!import.meta.client || !sentinelRef.value) {
    return
  }

  observer?.disconnect()
  observer = new IntersectionObserver((entries) => {
    if (entries.some(entry => entry.isIntersecting)) {
      void loadMore()
    }
  }, { rootMargin: '240px 0px' })
  observer.observe(sentinelRef.value)
}

watch([sentinelRef, hasMore], () => {
  if (!hasMore.value) {
    observer?.disconnect()
    return
  }
  setupObserver()
})

onMounted(() => {
  setupObserver()
  void refreshPingData()
  pingTimer = setInterval(refreshPingData, 60000)
})

onUnmounted(() => {
  observer?.disconnect()
  if (pingTimer) {
    clearInterval(pingTimer)
  }
})
</script>
