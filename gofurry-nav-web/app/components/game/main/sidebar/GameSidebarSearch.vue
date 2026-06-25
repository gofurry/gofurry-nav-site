<template>
  <div ref="searchShellRef" class="search-shell relative">
    <!-- 搜索框 -->
    <div class="relative">
      <img
          src="../../../../assets/svgs/search.svg"
          class="game-sidebar-search-icon absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 opacity-60"
          alt=""
      />
      <input
          v-model="keyword"
          type="text"
          :placeholder="t('game.search.simple')"
          class="game-sidebar-search-input w-full rounded-lg py-2 pl-9 pr-3 text-sm transition focus:outline-none"
          @focus="onFocus"
          @blur="onBlur"
      />
    </div>

    <!-- 搜索结果提示框 -->
    <Transition
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="opacity-0 translate-y-1"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition duration-150 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 translate-y-1"
    >
      <div
          v-if="showResults && results.length > 0"
          class="search-results-panel"
          :style="{ gridTemplateColumns: `repeat(${resultColumnCount}, minmax(0, 1fr))` }"
          @mouseenter="hovering = true"
          @mouseleave="hovering = false"
      >
        <div
            v-for="item in results"
            :key="item.id"
            class="search-result-card"
            @click="goToGame(item.id)"
        >
          <SteamAssetImage
              :src="item.cover"
              class="search-result-cover"
              :alt="item.name"
          />
          <p class="search-result-title">{{ item.name }}</p>
          <p class="search-result-desc">{{ item.info }}</p>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { getSearchSimple } from "@/utils/api/game";
import type { SearchItemModel } from "@/types/game";
import SteamAssetImage from '@/components/common/SteamAssetImage.vue'
import { useI18n } from 'vue-i18n'

const { t, locale } = useI18n()

const router = useRouter();
const localePath = useLocalePath()
const lang = computed<'zh' | 'en'>(() => locale.value === 'en' ? 'en' : 'zh')

const keyword = ref("");
const results = ref<SearchItemModel[]>([]);
const showResults = ref(false);
const hovering = ref(false);
const searchShellRef = ref<HTMLElement | null>(null)
const resultColumnCount = ref(2)

let timer: number | null = null;
let blurTimer: number | null = null;
let searchController: AbortController | null = null;
let searchRequestToken = 0;
let resizeObserver: ResizeObserver | null = null;

const isAbortError = (error: unknown) =>
  error instanceof Error && error.name === 'AbortError'

// 监听语言变化
watch(
    lang,
    () => {
      if (keyword.value.trim()) fetchResults(keyword.value);
    }
);

// 防抖搜索
watch(keyword, (val) => {
  if (timer) clearTimeout(timer);

  if (!val.trim()) {
    searchController?.abort();
    results.value = [];
    showResults.value = false;
    return;
  }

  timer = window.setTimeout(() => {
    fetchResults(val.trim());
  }, 500);
});

async function fetchResults(val: string) {
  searchController?.abort();
  const controller = new AbortController();
  const currentToken = ++searchRequestToken;
  searchController = controller;

  try {
    const res = await getSearchSimple(lang.value, val, { signal: controller.signal });
    if (currentToken !== searchRequestToken) {
      return;
    }
    results.value = res;
    showResults.value = res.length > 0;
  } catch (e) {
    if (isAbortError(e)) {
      return;
    }
    console.error("搜索失败", e);
  }
}

// 点击跳转
function goToGame(id: string) {
  router.push(localePath(`/games/${id}`));
  keyword.value = "";
  results.value = [];
  showResults.value = false;
}

// 输入框获得焦点
function onFocus() {
  if (results.value.length > 0) showResults.value = true;
  if (blurTimer) clearTimeout(blurTimer);
}

// 输入框失去焦点
function onBlur() {
  // 延迟隐藏
  blurTimer = window.setTimeout(() => {
    if (!hovering.value) showResults.value = false;
  }, 200);
}

function syncResultColumns(width: number) {
  if (width >= 880) {
    resultColumnCount.value = 4
    return
  }

  if (width >= 620) {
    resultColumnCount.value = 3
    return
  }

  resultColumnCount.value = 2
}

onMounted(() => {
  if (!searchShellRef.value) {
    return
  }

  syncResultColumns(searchShellRef.value.clientWidth)

  resizeObserver = new ResizeObserver((entries) => {
    const entry = entries[0]
    if (!entry) {
      return
    }
    syncResultColumns(entry.contentRect.width)
  })

  resizeObserver.observe(searchShellRef.value)
})

onBeforeUnmount(() => {
  if (timer) clearTimeout(timer);
  if (blurTimer) clearTimeout(blurTimer);
  searchController?.abort();
  resizeObserver?.disconnect();
});
</script>

<style scoped>
.search-results-panel {
  pointer-events: auto;
  position: absolute;
  z-index: 50;
  margin-top: 0.5rem;
  display: grid;
  width: 100%;
  gap: 0.55rem;
}

:global(html.dark .game-sidebar-search-icon) {
  filter: brightness(0) invert(1);
}

.search-result-card {
  min-width: 0;
  cursor: pointer;
  overflow: hidden;
}

.search-result-cover {
  aspect-ratio: 460 / 215;
  width: 100%;
  object-fit: cover;
}

.search-result-title {
  margin-top: 0.38rem;
  overflow: hidden;
  font-size: 0.82rem;
  line-height: 1.15;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-result-desc {
  margin-top: 0.18rem;
  overflow: hidden;
  font-size: 0.72rem;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
