<template>
  <div class="flex min-h-full w-full flex-1 flex-col">
    <div class="flex h-[5vh] items-center justify-between bg-orange-100 p-4 shadow-sm backdrop-blur-sm">
      <h2 class="flex items-center gap-2 text-lg font-semibold text-gray-800">
        <img src="@/assets/svgs/tv-dark.svg" alt="api" class="h-5 w-5" />
        {{ t('log.changelog') }}
      </h2>
    </div>

    <div class="flex min-h-0 flex-1">
      <aside class="hidden shrink-0 space-y-2 bg-orange-50 p-4 sm:block sm:w-32 md:w-48 lg:w-64">
        <div
          v-for="item in list"
          :key="item.create_time"
          class="cursor-pointer rounded-lg px-3 py-2 text-sm transition"
          :class="active?.title === item.title
            ? 'bg-orange-100 font-medium text-orange-700'
            : 'text-gray-600 hover:bg-orange-200'"
          @click="loadMarkdown(item)"
        >
          <div class="flex items-center justify-between">
            <span class="block w-full truncate">{{ item.title }}</span>
            <span
              v-if="item === list[0]"
              class="hidden rounded bg-orange-800 px-1.5 py-0.5 text-[10px] text-orange-50 md:block"
            >
              NEW
            </span>
          </div>
          <div class="mt-1 hidden text-xs text-gray-400 lg:block">{{ item.create_time }}</div>
        </div>
      </aside>

      <main class="flex-1 overflow-auto bg-orange-50 p-6 shadow">
        <div v-if="pending || loading" class="animate-pulse text-sm text-gray-400">
          {{ t('common.loading') }}
        </div>

        <div v-else-if="error" class="text-sm text-red-500">
          {{ loadFailedText }}
        </div>

        <MdPreview
          v-else
          class="custom-style"
          :editor-id="previewId"
          :model-value="state.text"
          :preview-theme="state.previewTheme"
          :code-theme="state.codeTheme"
        />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { MdPreview } from 'md-editor-v3'
import 'md-editor-v3/lib/preview.css'
import { getChangeLog } from '~/services/nav'
import type { changelogResp } from '~/types/nav'

interface UpdatesPageData {
  list: changelogResp[]
  active: changelogResp | null
  markdown: string
}

const { t } = useI18n()
const previewId = 'preview-only'

const state = reactive({
  text: '',
  codeTheme: 'github',
  previewTheme: 'vuepress',
})

const mdCache = new Map<string, string>()
const active = ref<changelogResp | null>(null)
const loading = ref(false)
const loadFailedText = 'Failed to load changelog.'

async function fetchMarkdown(url: string) {
  return await $fetch<string>('/api/v1/nav/site/changelog/content', {
    query: { url },
    responseType: 'text',
  })
}

const { data, pending, error } = await useAsyncData<UpdatesPageData>(
  'updates-page',
  async () => {
    const changelogList = await getChangeLog().catch(() => [])
    const firstItem = changelogList[0] ?? null
    const markdown = firstItem ? await fetchMarkdown(firstItem.url).catch(() => '# load fail\nfail to get changelog') : ''

    return {
      list: changelogList,
      active: firstItem,
      markdown,
    }
  },
  {
    default: () => ({
      list: [],
      active: null,
      markdown: '',
    }),
  }
)

const list = ref<changelogResp[]>(data.value?.list ?? [])
active.value = data.value?.active ?? null
state.text = data.value?.markdown ?? ''
if (active.value?.url && state.text) {
  mdCache.set(active.value.url, state.text)
}

useSeoMeta({
  title: () => 'GoFurry Updates',
  description: () => 'Latest changelog and updates from gofurry.',
  ogTitle: () => 'GoFurry Updates',
  ogDescription: () => 'Latest changelog and updates from gofurry.',
})

async function loadMarkdown(item: changelogResp) {
  if (active.value?.url === item.url) {
    return
  }

  active.value = item
  loading.value = true

  try {
    if (mdCache.has(item.url)) {
      state.text = mdCache.get(item.url) || ''
      return
    }

    const text = await fetchMarkdown(item.url)
    mdCache.set(item.url, text)
    state.text = text
  } catch (loadError) {
    console.error(loadError)
    state.text = '# load fail\nfail to get changelog'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.custom-style {
  max-width: 1240px;
  margin: 0 auto;
  background: none;
}
</style>
