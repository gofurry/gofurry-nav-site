<template>
  <div ref="searchShellRef" class="search-shell relative" @focusin="onFocus" @focusout="onBlur" @keydown.esc.stop="dismissResults">
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
          :aria-busy="status === 'pending' || status === 'debouncing'"
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
          v-if="showResults && status !== 'idle'"
          :class="status === 'success' ? 'search-results-panel' : 'search-status-panel'"
          :data-state="status"
          :style="{ gridTemplateColumns: `repeat(${resultColumnCount}, minmax(0, 1fr))` }"
          @mouseenter="hovering = true"
          @mouseleave="onPanelLeave"
      >
        <div v-if="status !== 'success'" class="search-status-content col-span-full flex flex-wrap items-center gap-3 p-2" :role="status === 'error' ? 'alert' : 'status'">
          <p>{{ t(status === 'error' ? 'game.search.unavailable' : status === 'empty' ? 'game.search.noSuggestions' : 'game.search.loading') }}</p>
          <button v-if="status === 'error'" type="button" class="gf-button gf-button--surface" @click="retrySearch">{{ t('game.search.retry') }}</button>
        </div>
        <div
            v-for="item in status === 'success' ? results : []"
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
import { ApiError } from '@/types/api'

const { t, locale } = useI18n()

const router = useRouter();
const localePath = useLocalePath()
const lang = computed<'zh' | 'en'>(() => locale.value === 'en' ? 'en' : 'zh')

const keyword = ref("");
const results = ref<SearchItemModel[]>([]);
const showResults = ref(false);
const status = ref<'idle' | 'debouncing' | 'pending' | 'success' | 'empty' | 'error'>('idle');
const focused = ref(false);
const hovering = ref(false);
const searchShellRef = ref<HTMLElement | null>(null)
const resultColumnCount = ref(2)

let timer: number | null = null;
let blurTimer: number | null = null;
let searchController: AbortController | null = null;
let searchRequestToken = 0;
let resizeObserver: ResizeObserver | null = null;

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
    status.value = 'idle';
    searchRequestToken++;
    searchController?.abort();
    results.value = [];
    showResults.value = false;
    return;
  }

  results.value = [];
  status.value = 'debouncing';
  showResults.value = focused.value || hovering.value;
  timer = window.setTimeout(() => {
    timer = null;
    fetchResults(val.trim());
  }, 500);
});

async function fetchResults(val: string) {
  searchController?.abort();
  const controller = new AbortController();
  const currentToken = ++searchRequestToken;
  searchController = controller;
  const requestLang = lang.value;
  status.value = 'pending';

  try {
    const res = await getSearchSimple(lang.value, val, { signal: controller.signal });
    if (controller.signal.aborted || currentToken !== searchRequestToken || val !== keyword.value || requestLang !== lang.value) {
      return;
    }
    results.value = res;
    status.value = res.length > 0 ? 'success' : 'empty';
  } catch (e) {
    if (controller.signal.aborted || currentToken !== searchRequestToken || val !== keyword.value || requestLang !== lang.value) {
      return;
    }
    if (e instanceof ApiError || (e instanceof Error && e.name === 'FetchError')) {
      results.value = [];
      status.value = 'error';
      return;
    }
    throw e;
  }
}

function retrySearch() {
  if (status.value !== 'error' || !keyword.value.trim()) return;
  if (timer) clearTimeout(timer);
  timer = null;
  // The retry button leaves the DOM while pending; retain a real focus owner.
  searchShellRef.value?.querySelector('input')?.focus();
  return fetchResults(keyword.value);
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
  focused.value = true;
  showResults.value = !!keyword.value.trim();
  if (blurTimer) clearTimeout(blurTimer);
}

// 输入框失去焦点
function onBlur(event?: FocusEvent) {
  if (event?.relatedTarget instanceof Node && searchShellRef.value?.contains(event.relatedTarget)) return;
  focused.value = false;
  // 延迟隐藏
  blurTimer = window.setTimeout(() => {
    if (!hovering.value && !searchShellRef.value?.contains(document.activeElement)) showResults.value = false;
  }, 200);
}

function onPanelLeave() {
  hovering.value = false;
  if (!focused.value) onBlur();
}

function dismissResults() {
  showResults.value = false;
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
  searchRequestToken++;
  if (timer) clearTimeout(timer);
  if (blurTimer) clearTimeout(blurTimer);
  searchController?.abort();
  resizeObserver?.disconnect();
});
</script>

<style scoped>
.search-results-panel,
.search-status-panel {
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
