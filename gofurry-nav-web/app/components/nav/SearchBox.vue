<template>
  <div
      ref="searchBoxRef"
      class="search-box-shell relative isolate z-30 mt-2 flex w-full flex-col items-center md:mt-0"
  >
    <!-- 搜索类别 -->
    <div class="mb-3 w-full max-w-[620px] px-3 py-2.5">
      <div class="flex flex-wrap justify-center gap-2 sm:gap-3">
        <div
            v-for="item in categories"
            :key="item"
            @click="selectedCategory = item"
            :class="[
              'search-chip cursor-pointer rounded-xl px-3 py-1.5 text-sm font-medium transition-all duration-500',
              selectedCategory === item ? 'search-chip-active' : ''
            ]"
        >
          {{ item }}
        </div>
      </div>
    </div>

    <!-- 搜索框 -->
    <div class="relative z-30 w-full max-w-[500px]">
      <input
          ref="inputRef"
          type="text"
          v-model="keyword"
          @keydown.enter.prevent="handleEnterKey"
          @keydown.down.prevent="handleArrowDown"
          @keydown.up.prevent="handleArrowUp"
          @keydown.esc.prevent.stop="closeSearchSuggestions"
          @input="debouncedFetch"
          @focus="handleInputFocus"
          @blur="handleInputBlur"
          placeholder="搜索站点或内容..."
          class="search-input h-12 w-full rounded-xl px-4 pr-10 duration-500 focus:outline-none"
      />
      <img src="@/assets/svgs/search.svg"
           :alt="searchActionLabel"
           class="search-icon absolute right-3 top-1/2 h-5 w-5 -translate-y-1/2
             cursor-pointer transition-transform duration-500 hover:scale-110"
           @click="doSearch()"
      />

      <!-- 下拉建议框 -->
      <ul
          ref="dropdownRef"
          v-if="(keyword.trim() && dropdownVisible) || isLoading"
          class="search-suggestion-list absolute left-0 top-[calc(100%+0.5rem)] z-[999] w-full max-h-60 overflow-y-auto overscroll-contain rounded-2xl border border-white/12 bg-slate-950/72
            origin-top-left backdrop-blur-2xl shadow-2xl shadow-slate-950/35 ring-1 ring-white/8 transition-all duration-500
            animate-fadeIn"
          @wheel.stop
          @touchmove.stop
      >
        <!-- 标题 -->
        <li class="border-b border-white/10 bg-white/5 px-4 py-2 text-xs text-slate-300/90">
          <template v-if="isLoading">{{ t('common.loading') }}</template>
          <template v-else>{{ t('searchBox.searchSuggest') }} ({{ suggestions.length }})</template>
        </li>

        <!-- 加载状态 -->
        <li v-if="isLoading" class="px-4 py-6 text-center text-slate-300/80">
          <div class="inline-block h-5 w-5 animate-spin rounded-full border-2 border-white/30 border-t-cyan-300"></div>
        </li>

        <!-- 建议项 -->
        <template v-else-if="suggestions.length">
          <li
              v-for="(item, index) in suggestions"
              :key="index"
              @click="selectSuggestion(index)"
              @mouseenter="hoveredIndex = index"
              @mouseleave="hoveredIndex = -1"
              class="search-suggestion-item cursor-pointer px-4 py-3 text-sm font-medium text-slate-100/90"
              :class="hoveredIndex === index ? 'search-suggestion-item-active' : ''"
          >
            <!-- 关键词高亮 -->
            <span v-html="highlightKeyword(item)"></span>
          </li>
        </template>

        <li v-else-if="keyword.trim()" class="px-4 py-3 text-center text-slate-300/75">
          {{ t('searchBox.noSuggest') }}
        </li>
      </ul>
    </div>

    <!-- 搜索平台 -->
    <div class="mt-1 w-full max-w-[620px] px-3 py-2.5">
      <div class="grid grid-cols-2 justify-center gap-2 md:flex md:flex-wrap md:gap-2">
        <div
            v-for="platform in platforms[selectedCategory]"
            :key="platform.name"
            @click="selectedPlatform = platform"
            :class="[
              'search-chip cursor-pointer rounded-xl px-2.5 py-1.5 text-center text-xs font-medium whitespace-nowrap transition-all duration-500',
              selectedPlatform.name === platform.name ? 'search-chip-active' : ''
            ]"
        >
          {{ platform.name }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onBeforeUnmount, computed } from 'vue'
import { getSearchSuggestion } from '@/services/nav'
import type { NavSearchSuggestionEngine } from '@/types/nav'
import { useI18n } from 'vue-i18n'
import { useLangStore } from '@/store/langStore'
import { setNavPageRevealLock } from '@/utils/navPageReveal'

const { t, locale } = useI18n()
const langStore = useLangStore()
const searchActionLabel = computed(() => locale.value === 'en' ? 'Search GoFurry resources' : '搜索 GoFurry 资源')

// 同步语言切换
watch(
    () => langStore.lang,
    async (newLang) => {
      locale.value = newLang
      await nextTick()
      // 切换语言后重新设置默认选中项
      resetSelection()
    },
    { immediate: true }
)

// 搜索类别
const categories = computed(() => [
  t('searchBox.platformCate.search'),
  t('searchBox.platformCate.furry'),
  t('searchBox.platformCate.games'),
  t('searchBox.platformCate.art'),
  t('searchBox.platformCate.novel'),
])

const selectedCategory = ref('')
type VisibleSuggestionEngine = Exclude<NavSearchSuggestionEngine, 'baidu'>

// 初始化默认选中类别
const initDefaultCategory = () => {
  const first = categories.value[0]
  if (!selectedCategory.value || !categories.value.includes(selectedCategory.value)) {
    selectedCategory.value = first ?? ''
  }
}


// 搜索关键词和建议
const keyword = ref('')
const suggestions = ref<string[]>([])
const hoveredIndex = ref(-1)
const isLoading = ref(false)

interface Platform {
  name: string
  type: VisibleSuggestionEngine | 'site'
  url?: string
}

// 平台数据
const platforms = computed<Record<string, Platform[]>>(() => ({
  [t('searchBox.platformCate.search')]: [
    { name: t('searchBox.platformName.bing'), type: 'bing' },
    { name: t('searchBox.platformName.google'), type: 'google' },
    { name: t('searchBox.platformName.duckduckgo'), type: 'duckduckgo' },
    { name: t('searchBox.platformName.bilibili'), type: 'bilibili', url: 'https://search.bilibili.com/all?keyword={kw}' },
    { name: t('searchBox.platformName.xiaohongshu'), type: 'site', url: 'https://www.xiaohongshu.com/search_result?keyword={kw}' },
    { name: t('searchBox.platformName.zhihu'), type: 'site', url: 'https://www.zhihu.com/search?type=content&q={kw}' },
    { name: t('searchBox.platformName.weibo'), type: 'site', url: 'https://s.weibo.com/weibo?q={kw}' },
    { name: t('searchBox.platformName.twitter'), type: 'site', url: 'https://x.com/search?q={kw}&src=typed_query' },
  ],
  [t('searchBox.platformCate.furry')]: [
    { name: t('searchBox.platformName.wikifur'), type: 'site', url: `https://${langStore.lang === 'zh' ? 'zh' : 'en'}.wikifur.com/wiki/{kw}` },
    { name: t('searchBox.platformName.yiffParty'), type: 'site', url: 'https://yiff-party.com/search/?tags={kw}' },
    { name: t('searchBox.platformName.furaffinity'), type: 'site', url: 'https://www.furaffinity.net/search/?q={kw}' },
    { name: t('searchBox.platformName.e621'), type: 'site', url: 'https://e621.net/posts?tags={kw}' },
    { name: t('searchBox.platformName.wilddream'), type: 'site', url: 'https://www.wilddream.net/Art/index/index?keyword={kw}' },
  ],
  [t('searchBox.platformCate.art')]: [
    { name: t('searchBox.platformName.pixiv'), type: 'site', url: 'https://www.pixiv.net/tags/{kw}' },
    { name: t('searchBox.platformName.deviantart'), type: 'site', url: 'https://www.deviantart.com/search?q={kw}' },
    { name: t('searchBox.platformName.artstation'), type: 'site', url: 'https://www.artstation.com/search?query={kw}' },
    { name: t('searchBox.platformName.pinterest'), type: 'site', url: 'https://www.pinterest.com/search/pins/?q={kw}' },
    { name: t('searchBox.platformName.zcool'), type: 'site', url: 'https://www.zcool.com.cn/search/content?word={kw}' },
  ],
  [t('searchBox.platformCate.novel')]: [
    { name: t('searchBox.platformName.furrynovel'), type: 'site', url: 'https://furrynovel.com/zh/search?keyword={kw}' },
    { name: t('searchBox.platformName.linpx'), type: 'site', url: 'https://furrynovel.ink/search?word={kw}' },
  ],
  [t('searchBox.platformCate.games')]: [
    { name: t('searchBox.platformName.itchIo'), type: 'site', url: 'https://itch.io/search?q={kw}' },
    { name: t('searchBox.platformName.steam'), type: 'site', url: 'https://store.steampowered.com/search?term={kw}' },
    { name: t('searchBox.platformName.epic'), type: 'site', url: 'https://store.epicgames.com/browse?q={kw}' },
  ],
}))

const selectedPlatform = ref<Platform>({ name: '', type: 'site' })
let timer: number | null = null
let suggestionAbortController: AbortController | null = null
let suggestionRequestId = 0

// 重置默认选中逻辑
const resetSelection = () => {
  abortSuggestionRequest()
  initDefaultCategory()
  const defaultPlatform = platforms.value[selectedCategory.value]?.[0]
  selectedPlatform.value = defaultPlatform || { name: '', type: 'site' }
  suggestions.value = []
  hoveredIndex.value = -1
}

// 当语言或类别变化时重设默认平台
watch([categories, selectedCategory, platforms], () => {
  resetSelection()
}, { immediate: true })

const debounce = (fn: Function, delay = 600) => (...args: any[]) => {
  if (timer) clearTimeout(timer)
  timer = window.setTimeout(() => fn(...args), delay)
}

const inputRef = ref<HTMLInputElement | null>(null)
const searchBoxRef = ref<HTMLElement | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)
const dropdownVisible = ref(false)
const isInputFocused = ref(false)

const highlightKeyword = (item: string) => {
  if (!keyword.value.trim()) return item
  const escapedKeyword = keyword.value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return item.replace(
      new RegExp(`(${escapedKeyword})`, 'gi'),
      '<span class="text-orange-500 font-bold">$1</span>'
  )
}

const fetchSuggestions = async () => {
  const searchLabel = t('searchBox.platformCate.search')
  const requestKeyword = keyword.value.trim()
  if (!requestKeyword || selectedCategory.value !== searchLabel || !isSuggestionEngine(selectedPlatform.value.type)) {
    abortSuggestionRequest()
    suggestions.value = []
    dropdownVisible.value = false
    isLoading.value = false
    return
  }

  dropdownVisible.value = true
  isLoading.value = true
  abortSuggestionRequest()
  const controller = new AbortController()
  const requestId = ++suggestionRequestId
  suggestionAbortController = controller

  try {
    const response = await getSearchSuggestion(selectedPlatform.value.type, requestKeyword, controller.signal)
    if (requestId !== suggestionRequestId) return
    suggestions.value = response.suggestions
    if (response.suggestions.length > 0 && hoveredIndex.value === -1) hoveredIndex.value = 0
  } catch {
    if (controller.signal.aborted || requestId !== suggestionRequestId) return
    suggestions.value = []
  } finally {
    if (requestId === suggestionRequestId) {
      isLoading.value = false
    }
    if (suggestionAbortController === controller) {
      suggestionAbortController = null
    }
  }
}

const debouncedFetch = debounce(fetchSuggestions, 600)

const doSearch = () => {
  const kw = encodeURIComponent(keyword.value.trim())
  if (!kw) return

  const mapping: Record<string, string> = {}
  Object.values(platforms.value).flat().forEach((p) => {
    if (p.url) mapping[p.name] = p.url.replace('{kw}', kw)
  })

  const searchLabel = t('searchBox.platformCate.search')
  const gameLabel = t('searchBox.platformCate.games')
  const siteLabels = new Set([
    t('searchBox.platformCate.furry'),
    t('searchBox.platformCate.art'),
    t('searchBox.platformCate.novel'),
    gameLabel,
  ])

  if (selectedCategory.value === searchLabel) {
    switch (selectedPlatform.value.type) {
      case 'bing': window.open(`https://www.bing.com/search?q=${kw}`, '_blank'); break
      case 'google': window.open(`https://www.google.com/search?q=${kw}`, '_blank'); break
      case 'duckduckgo': window.open(`https://duckduckgo.com/?q=${kw}`, '_blank'); break
      default:
        if (mapping[selectedPlatform.value.name])
          window.open(mapping[selectedPlatform.value.name], '_blank')
    }
  } else if (siteLabels.has(selectedCategory.value)) {
    if (mapping[selectedPlatform.value.name])
      window.open(mapping[selectedPlatform.value.name], '_blank')
  }
  dropdownVisible.value = false
}

const selectSuggestion = (index: number) => {
  if (suggestions.value[index]) {
    keyword.value = suggestions.value[index]
    doSearch()
  }
}

const handleArrowDown = () => {
  if (!dropdownVisible.value || isLoading.value) return
  if (!suggestions.value.length) return
  hoveredIndex.value = (hoveredIndex.value + 1) % suggestions.value.length
  scrollToSelectedItem()
}

const handleArrowUp = () => {
  if (!dropdownVisible.value || isLoading.value) return
  if (!suggestions.value.length) return
  hoveredIndex.value = hoveredIndex.value <= 0 ? suggestions.value.length - 1 : hoveredIndex.value - 1
  scrollToSelectedItem()
}

const handleEnterKey = () => {
  if (suggestions.value.length > 0 && hoveredIndex.value !== -1) {
    selectSuggestion(hoveredIndex.value)
  } else {
    doSearch()
  }
}

const scrollToSelectedItem = () => {
  nextTick(() => {
    const listItems = dropdownRef.value?.querySelectorAll('li:not(:first-child)') ?? []
    const selectedItem = listItems[hoveredIndex.value] as HTMLElement
    if (selectedItem) {
      const listContainer = selectedItem.parentElement
      if (listContainer) {
        const containerRect = listContainer.getBoundingClientRect()
        const itemRect = selectedItem.getBoundingClientRect()
        if (itemRect.bottom > containerRect.bottom)
          listContainer.scrollTop += itemRect.bottom - containerRect.bottom
        else if (itemRect.top < containerRect.top)
          listContainer.scrollTop -= containerRect.top - itemRect.top
      }
    }
  })
}

const handleInputFocus = () => {
  isInputFocused.value = true
  if (keyword.value.trim()) {
    dropdownVisible.value = true
    debouncedFetch()
  }
}

const handleInputBlur = () => {
  window.setTimeout(() => {
    const activeElement = document.activeElement
    isInputFocused.value = !!(
      activeElement &&
      searchBoxRef.value?.contains(activeElement)
    )
  }, 0)
}

const handleClickOutside = (e: MouseEvent) => {
  if (searchBoxRef.value && !searchBoxRef.value.contains(e.target as Node)) {
    closeSearchSuggestions()
  }
}

const closeSearchSuggestions = () => {
  abortSuggestionRequest()
  dropdownVisible.value = false
  hoveredIndex.value = -1
  isInputFocused.value = false
}

function isSuggestionEngine(type: Platform['type']): type is VisibleSuggestionEngine {
  return type === 'bing' || type === 'google' || type === 'bilibili' || type === 'duckduckgo'
}

function abortSuggestionRequest() {
  suggestionRequestId++
  suggestionAbortController?.abort()
  suggestionAbortController = null
}

watch(
    [dropdownVisible, isInputFocused],
    ([visible, focused]) => {
      setNavPageRevealLock('search-box', visible || focused)
    },
    { immediate: true }
)

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  resetSelection() // 🟧 初始化默认选中项
})

onBeforeUnmount(() => {
  if (timer) clearTimeout(timer)
  abortSuggestionRequest()
  document.removeEventListener('click', handleClickOutside)
  setNavPageRevealLock('search-box', false)
})
</script>



<style scoped>
.search-box-shell::before {
  content: "";
  position: absolute;
  top: -1rem;
  bottom: -0.9rem;
  left: 50%;
  width: min(calc(100% - 2rem), 38rem);
  z-index: -1;
  transform: translateX(-50%);
  border-radius: 999px;
  background:
    radial-gradient(ellipse 66% 58% at 50% 50%, rgba(15, 23, 42, 0.48), rgba(15, 23, 42, 0.22) 48%, rgba(15, 23, 42, 0.08) 64%, transparent 82%);
  filter: blur(18px);
  opacity: 0.86;
  pointer-events: none;
  transition: opacity 500ms ease, background 500ms ease;
}

.search-chip {
  color: rgba(248, 250, 252, 0.92);
  background: rgba(15, 23, 42, 0.5);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.12),
    0 8px 24px rgba(2, 6, 23, 0.14);
  backdrop-filter: blur(14px) saturate(1.12);
  -webkit-backdrop-filter: blur(14px) saturate(1.12);
  transition:
    background 500ms ease,
    box-shadow 500ms ease,
    color 500ms ease;
}

.search-chip:hover {
  color: rgba(255, 255, 255, 0.98);
  background: rgba(15, 23, 42, 0.68);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.16),
    0 10px 28px rgba(2, 6, 23, 0.18);
}

.search-chip-active {
  color: rgba(15, 23, 42, 0.94);
  background: rgba(255, 248, 241, 0.92);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.65),
    0 12px 30px rgba(2, 6, 23, 0.2);
}

.search-chip-active:hover {
  color: rgba(15, 23, 42, 0.96);
  background: rgba(255, 255, 255, 0.96);
}

.search-input {
  color: rgba(15, 23, 42, 0.94);
  background: rgba(255, 255, 255, 0.78);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.86),
    0 12px 36px rgba(2, 6, 23, 0.22),
    0 0 0 1px rgba(255, 255, 255, 0.18);
  backdrop-filter: blur(18px) saturate(1.08);
  -webkit-backdrop-filter: blur(18px) saturate(1.08);
  transition:
    background 500ms ease,
    box-shadow 500ms ease,
    color 500ms ease;
}

.search-input::placeholder {
  color: rgba(71, 85, 105, 0.72);
}

.search-input:focus {
  background: rgba(255, 255, 255, 0.9);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    0 16px 42px rgba(2, 6, 23, 0.26),
    0 0 0 1px rgba(255, 255, 255, 0.34),
    0 0 0 4px rgba(15, 23, 42, 0.2);
}

.search-icon {
  opacity: 0.72;
  filter: drop-shadow(0 1px 1px rgba(255, 255, 255, 0.42));
}

.search-suggestion-list {
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.search-suggestion-list::-webkit-scrollbar {
  display: none;
  width: 0;
  height: 0;
}

.search-suggestion-item {
  transition:
    background 500ms ease,
    color 500ms ease;
}

.search-suggestion-item:hover,
.search-suggestion-item-active {
  color: rgba(255, 255, 255, 0.98);
  background: rgba(255, 255, 255, 0.08);
}

/* 淡入动画 */
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-8px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}
.animate-fadeIn {
  animation: fadeIn 0.5s ease-out forwards;
}

/* 加载动画 */
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
.animate-spin {
  animation: spin 1s linear infinite;
}
</style>
