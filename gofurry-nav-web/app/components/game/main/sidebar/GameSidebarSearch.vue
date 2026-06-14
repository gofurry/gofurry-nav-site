<template>
  <div class="search-shell relative" :class="{ 'search-shell--dark': isDarkTheme }">
    <!-- 搜索框 -->
    <div class="relative">
      <img
          src="../../../../assets/svgs/search.svg"
          class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 opacity-60"
          alt=""
      />
      <input
          v-model="keyword"
          type="text"
          :placeholder="t('game.search.simple')"
          class="game-sidebar-search-input w-full rounded-lg border border-[rgba(126,92,58,0.17)] bg-[rgba(255,250,242,0.42)] py-2 pl-9 pr-3 text-sm text-stone-800 shadow-[0_4px_12px_rgba(91,62,28,0.03)] placeholder-stone-400 transition focus:border-[rgba(120,87,56,0.36)] focus:bg-[rgba(255,250,242,0.60)] focus:outline-none focus:ring-1 focus:ring-[rgba(120,87,56,0.10)] dark:border-slate-400/15 dark:bg-slate-950/35 dark:text-slate-100 dark:placeholder-slate-400 dark:shadow-none dark:focus:border-slate-300/40 dark:focus:bg-slate-900/55 dark:focus:ring-slate-300/20"
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
          @mouseenter="hovering = true"
          @mouseleave="hovering = false"
      >
        <div
            v-for="item in results"
            :key="item.id"
            class="search-result-card"
            @click="goToGame(item.id)"
        >
          <img
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
import { computed, ref, watch } from "vue";
import { getSearchSimple } from "@/utils/api/game";
import type { SearchItemModel } from "@/types/game";
import { useLangStore } from "@/store/langStore";
import { i18n } from '@/main'
import { useThemeStore } from '@/stores/theme'

const { t } = i18n.global

const router = useRouter();
const localePath = useLocalePath()
const langStore = useLangStore();
const themeStore = useThemeStore();
const lang = ref(langStore.lang);
const isDarkTheme = computed(() => themeStore.theme === 'dark')

const keyword = ref("");
const results = ref<SearchItemModel[]>([]);
const showResults = ref(false);
const hovering = ref(false);

let timer: number | null = null;
let blurTimer: number | null = null;

// 监听语言变化
watch(
    () => langStore.lang,
    (val) => {
      lang.value = val;
      if (keyword.value.trim()) fetchResults(keyword.value);
    }
);

// 防抖搜索
watch(keyword, (val) => {
  if (timer) clearTimeout(timer);

  if (!val.trim()) {
    results.value = [];
    showResults.value = false;
    return;
  }

  timer = window.setTimeout(() => {
    fetchResults(val.trim());
  }, 500);
});

async function fetchResults(val: string) {
  try {
    const res = await getSearchSimple(lang.value, val);
    results.value = res;
    showResults.value = res.length > 0;
  } catch (e) {
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
</script>

<style scoped>
.search-results-panel {
  pointer-events: auto;
  position: absolute;
  z-index: 50;
  margin-top: 0.5rem;
  display: grid;
  width: 100%;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.55rem;
  border: 1px solid rgba(126, 92, 58, 0.18);
  border-radius: 0.82rem;
  background: rgba(255, 250, 242, 0.94);
  padding: 0.62rem;
  box-shadow: 0 12px 30px rgba(91, 62, 28, 0.08);
  backdrop-filter: blur(8px);
}

.search-shell {
  container: game-search-shell / inline-size;
}

.game-sidebar-search-input:focus {
  border-color: rgba(120, 87, 56, 0.36) !important;
  box-shadow: 0 0 0 1px rgba(120, 87, 56, 0.10) !important;
}

.search-result-card {
  min-width: 0;
  cursor: pointer;
  overflow: hidden;
  border: 1px solid rgba(126, 92, 58, 0.10);
  border-radius: 0.68rem;
  background: rgba(255, 255, 255, 0.26);
  padding: 0.42rem;
  transition: background-color 180ms ease, border-color 180ms ease, transform 180ms ease;
}

.search-result-card:hover {
  border-color: rgba(180, 96, 24, 0.18);
  background: rgba(255, 244, 228, 0.54);
}

.search-result-cover {
  aspect-ratio: 460 / 215;
  width: 100%;
  border-radius: 0.48rem;
  object-fit: cover;
}

.search-result-title {
  margin-top: 0.38rem;
  overflow: hidden;
  color: rgba(28, 25, 23, 0.94);
  font-size: 0.82rem;
  font-weight: 750;
  line-height: 1.15;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-result-desc {
  margin-top: 0.18rem;
  overflow: hidden;
  color: rgba(87, 83, 78, 0.68);
  font-size: 0.72rem;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@container game-search-shell (min-width: 20rem) {
  .search-results-panel {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@container game-search-shell (min-width: 34rem) {
  .search-results-panel {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
}

:global(.dark) .search-results-panel,
:global(.games-search-page.games-page--dark) .search-results-panel,
.search-shell--dark .search-results-panel {
  border-color: rgba(148, 163, 184, 0.16);
  background: rgba(15, 23, 42, 0.96);
  box-shadow: 0 14px 34px rgba(0, 0, 0, 0.28);
}

:global(.dark) .game-sidebar-search-input:focus {
  border-color: rgba(203, 213, 225, 0.38) !important;
  box-shadow: 0 0 0 1px rgba(148, 163, 184, 0.18) !important;
}

:global(.games-page--dark) .game-sidebar-search-input:focus {
  border-color: rgba(203, 213, 225, 0.38) !important;
  box-shadow: 0 0 0 1px rgba(148, 163, 184, 0.18) !important;
}

:global(.dark) .search-result-card,
:global(.games-search-page.games-page--dark) .search-result-card,
.search-shell--dark .search-result-card {
  border-color: rgba(148, 163, 184, 0.20);
  background: rgba(30, 41, 59, 0.58);
}

:global(.dark) .search-result-card:hover,
:global(.games-search-page.games-page--dark) .search-result-card:hover,
.search-shell--dark .search-result-card:hover {
  border-color: rgba(203, 213, 225, 0.46);
  background: rgba(51, 65, 85, 0.72);
}

:global(.dark) .search-result-title,
:global(.games-search-page.games-page--dark) .search-result-title,
.search-shell--dark .search-result-title {
  color: rgba(248, 250, 252, 0.98);
  text-shadow: none;
}

:global(.dark) .search-result-desc,
:global(.games-search-page.games-page--dark) .search-result-desc,
.search-shell--dark .search-result-desc {
  color: rgba(226, 232, 240, 0.78);
  font-weight: 500;
  text-shadow: none;
}
</style>
