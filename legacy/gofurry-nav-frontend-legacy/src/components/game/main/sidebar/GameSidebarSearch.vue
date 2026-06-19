<template>
  <div class="relative">
    <!-- 搜索框 -->
    <div class="relative">
      <img
          src="../../../../assets/svgs/search.svg"
          class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 opacity-60"
          alt="search"
      />
      <input
          v-model="keyword"
          type="text"
          :placeholder="t('game.search.simple')"
          class="w-full pl-9 pr-3 py-2
               rounded-lg bg-orange-50
               text-sm placeholder-gray-400
               focus:outline-none focus:ring-2 focus:ring-orange-200"
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
          class="absolute z-50 mt-2 bg-white/90 backdrop-blur-md rounded-lg shadow-lg p-3 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4 gap-3 2xl:grid-cols-5 pointer-events-auto"
          @mouseenter="hovering = true"
          @mouseleave="hovering = false"
      >
        <div
            v-for="item in results"
            :key="item.id"
            class="cursor-pointer rounded-lg overflow-hidden hover:bg-orange-100 transition p-1"
            @click="goToGame(item.id)"
        >
          <img
              :src="item.cover"
              class="w-full h-20 object-cover rounded-md mb-1"
              alt="game cover"
          />
          <p class="text-sm font-semibold truncate">{{ item.name }}</p>
          <p class="text-xs text-gray-500 truncate">{{ item.info }}</p>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { useRouter } from "vue-router";
import { getSearchSimple } from "@/utils/api/game.ts";
import type { SearchItemModel } from "@/types/game.ts";
import { useLangStore } from "@/store/langStore.ts";
import { i18n } from '@/main.ts'

const { t } = i18n.global

const router = useRouter();
const langStore = useLangStore();
const lang = ref(langStore.lang);

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
  router.push(`/games/${id}`);
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
