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
          class="search-input h-12 w-full px-4 pr-10"
      />
      <img src="@/assets/svgs/search.svg"
           :alt="searchActionLabel"
           class="search-icon"
           @click="doSearch()"
      />

      <!-- 下拉建议框 -->
      <ul
          ref="dropdownRef"
          v-if="(keyword.trim() && dropdownVisible) || isLoading"
          class="search-suggestion-list search-suggestion-list--entering absolute left-0 top-[calc(100%+0.5rem)] z-[999] max-h-60 w-full origin-top-left overflow-y-auto overscroll-contain"
          @wheel.stop
          @touchmove.stop
      >
        <!-- 标题 -->
        <li class="search-suggestion-header px-4 py-2 text-xs">
          <template v-if="isLoading">{{ t('common.loading') }}</template>
          <template v-else>{{ t('searchBox.searchSuggest') }} ({{ suggestions.length }})</template>
        </li>

        <!-- 加载状态 -->
        <li v-if="isLoading" class="search-suggestion-loading px-4 py-6 text-center">
          <div class="search-suggestion-spinner"></div>
        </li>

        <!-- 建议项 -->
        <template v-else-if="suggestions.length">
          <li
              v-for="(item, index) in suggestions"
              :key="index"
              @click="selectSuggestion(index)"
              @mouseenter="hoveredIndex = index"
              @mouseleave="hoveredIndex = -1"
              class="search-suggestion-item cursor-pointer px-4 py-3 text-sm font-medium"
              :class="hoveredIndex === index ? 'search-suggestion-item-active' : ''"
          >
            <!-- 关键词高亮 -->
            <span v-html="highlightKeyword(item)"></span>
          </li>
        </template>

        <li v-else-if="keyword.trim()" class="search-suggestion-empty px-4 py-3 text-center">
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
      '<span class="search-highlight">$1</span>'
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
