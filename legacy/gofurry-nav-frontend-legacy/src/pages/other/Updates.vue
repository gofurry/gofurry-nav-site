<template>
  <div class="flex min-h-full w-full flex-1 flex-col">
    <div
        class="h-[5vh] flex items-center justify-between p-4 bg-orange-100 backdrop-blur-sm shadow-sm"
    >
      <h2 class="text-lg font-semibold text-gray-800 flex items-center gap-2">
        <img src="@/assets/svgs/tv-dark.svg" alt="api" class="w-5 h-5" />
        {{ t("log.changelog") }}
      </h2>
    </div>

    <div class="flex flex-1 min-h-0">
      <aside class="hidden sm:block sm:w-32 md:w-48 lg:w-64 shrink-0 bg-orange-50 p-4 space-y-2">
        <div
            v-for="item in list"
            :key="item.create_time"
            @click="loadMarkdown(item)"
            class="cursor-pointer rounded-lg px-3 py-2 text-sm transition"
            :class="active?.title === item.title
            ? 'bg-orange-100 text-orange-700 font-medium'
            : 'hover:bg-orange-200 text-gray-600'"
        >
          <div class="flex items-center justify-between">
            <span class="truncate block w-full">{{ item.title }}</span>
            <span
                v-if="item === list[0]"
                class="hidden md:block text-[10px] px-1.5 py-0.5 rounded bg-orange-800 text-orange-50"
            >
              NEW
            </span>
          </div>
          <div class="hidden lg:block text-xs text-gray-400 mt-1">{{ item.create_time }}</div>
        </div>
      </aside>

      <main class="flex-1 bg-orange-50 shadow p-6 overflow-auto">
        <div v-if="loading" class="text-sm text-gray-400 animate-pulse">
          {{ t("common.loading") }}
        </div>

        <MdPreview
          class="custom-style"
          :editorId="id"
          :modelValue="state.text"
          :previewTheme="state.previewTheme"
          :codeTheme="state.codeTheme"
        />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ref, onMounted, reactive} from 'vue'
import axios from 'axios'
import { getChangeLog } from '@/utils/api/nav'
import type { changelogResp } from '@/types/nav'
import { i18n } from "@/main.ts";

const { t } = i18n.global

import { MdPreview} from 'md-editor-v3';
import 'md-editor-v3/lib/preview.css';

const id = 'preview-only';

const state = reactive({
  text: '',
  codeTheme: 'github',
  previewTheme: 'vuepress',
});

const list = ref<changelogResp[]>([])
const active = ref<changelogResp | null>(null)
const loading = ref(false)

const mdCache = new Map<string, string>()

async function loadMarkdown(item: changelogResp) {
  if (active.value?.url === item.url) return
  active.value = item
  loading.value = true

  try {
    if (mdCache.has(item.url)) {
      state.text = mdCache.get(item.url)!
      return
    }

    const res = await axios.get(item.url, { responseType: 'arraybuffer' })
    const text = new TextDecoder('utf-8').decode(res.data)
    mdCache.set(item.url, text)
    state.text = text
  } catch (err) {
    console.error(err)
    state.text = '# load fail\nfail to get changelog'
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  const data = await getChangeLog()
  list.value = data

  if (data.length > 0 && data[0] !== undefined) {
    await loadMarkdown(data[0])
  }
})
</script>

<style scoped>
.custom-style {
  max-width: 1240px;
  margin: 0 auto;
  background: none;
}
</style>
