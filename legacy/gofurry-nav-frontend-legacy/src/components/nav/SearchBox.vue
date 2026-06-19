<template>
  <div
      ref="searchBoxRef"
      class="relative z-30 mt-4 flex w-full flex-col items-center"
  >
    <!-- 搜索类别 -->
    <div class="flex sm:gap-6 gap-4 mb-2">
      <div
          v-for="item in categories"
          :key="item"
          @click="selectedCategory = item"
          :class="['cursor-pointer px-3 py-1 rounded text-sm transition-all duration-200', selectedCategory === item
            ? 'bg-orange-400 text-white shadow-orange-400/30'
            : 'bg-gray-200/60 text-gray-700 hover:bg-gray-200 hover:shadow-md']"
      >
        {{ item }}
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
          @input="debouncedFetch"
          @focus="handleInputFocus"
          @blur="handleInputBlur"
          placeholder="搜索站点或内容..."
          class="w-full h-12 px-4 pr-10 rounded-lg ring-2 ring-black/60 bg-gray-300/60 focus:outline-none focus:ring-3 focus:ring-black/75 duration-300"
      />
      <img src="@/assets/svgs/search.svg"
           alt="search"
           class="absolute right-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-500
             hover:scale-110 transition-transform duration-200 cursor-pointer"
           @click="doSearch()"
      />

      <!-- 下拉建议框 -->
      <ul
          ref="dropdownRef"
          v-if="(keyword.trim() && dropdownVisible) || isLoading"
          :style="{ width: inputWidth + 'px', left: inputLeft + 'px', top: inputTop + 'px' }"
          class="fixed z-[999] max-h-60 overflow-y-auto overscroll-contain rounded-xl border border-gray-100 bg-white
            shadow-xl shadow-gray-100/80 transition-all duration-200 origin-top-left
            animate-fadeIn"
          @wheel.stop
          @touchmove.stop
      >
        <!-- 标题 -->
        <li class="px-4 py-2 text-xs text-gray-500 bg-gray-50 border-b border-gray-100">
          <template v-if="isLoading">{{ t('common.loading') }}</template>
          <template v-else>{{ t('searchBox.searchSuggest') }} ({{ suggestions.length }})</template>
        </li>

        <!-- 加载状态 -->
        <li v-if="isLoading" class="px-4 py-6 text-center text-gray-500">
          <div class="inline-block w-5 h-5 border-2 border-orange-400 border-t-transparent rounded-full animate-spin"></div>
        </li>

        <!-- 建议项 -->
        <template v-else-if="suggestions.length">
          <li
              v-for="(item, index) in suggestions"
              :key="index"
              @click="selectSuggestion(index)"
              @mouseenter="hoveredIndex = index"
              @mouseleave="hoveredIndex = -1"
              class="px-4 py-3 hover:bg-orange-50 cursor-pointer transition-colors duration-150
                text-gray-800 hover:text-orange-500 font-medium"
              :class="hoveredIndex === index ? 'bg-orange-50 text-orange-500' : ''"
          >
            <!-- 关键词高亮 -->
            <span v-html="highlightKeyword(item)"></span>
          </li>
        </template>

        <li v-else-if="keyword.trim()" class="px-4 py-3 text-gray-500 text-center">
          {{ t('searchBox.noSuggest') }}
        </li>
      </ul>
    </div>

    <!-- 搜索平台 -->
    <div
        class="mt-2 w-full max-w-[400px] md:max-w-[600px] grid grid-cols-2 md:flex flex-wrap gap-2 justify-center"
    >
      <div
          v-for="platform in platforms[selectedCategory]"
          :key="platform.name"
          @click="selectedPlatform = platform"
          :class="[
      'cursor-pointer px-2 py-1 rounded text-center text-xs whitespace-nowrap transition-all duration-200',
      selectedPlatform.name === platform.name
        ? 'bg-orange-400 text-white shadow-sm shadow-orange-400/30'
        : 'bg-gray-200/60 text-gray-700 hover:bg-gray-200 hover:shadow-sm'
    ]"
      >
        {{ platform.name }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onBeforeUnmount, computed } from 'vue'
import { getBaiduSuggestion, getBingSuggestion, getGoogleSuggestion, getBiliBiliSuggestion } from '@/utils/api/nav'
import { useI18n } from 'vue-i18n'
import { useLangStore } from '@/store/langStore'
import { setNavPageRevealLock } from '@/utils/navPageReveal'

const { t, locale } = useI18n()
const langStore = useLangStore()

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
  t('searchBox.platformCate.games'),
])

const selectedCategory = ref('')

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
  type: 'baidu' | 'bing' | 'google' | 'bilibili' | 'site'
  url?: string
}

// 平台数据
const platforms = computed<Record<string, Platform[]>>(() => ({
  [t('searchBox.platformCate.search')]: [
    { name: t('searchBox.platformName.baidu'), type: 'baidu' },
    { name: t('searchBox.platformName.bing'), type: 'bing' },
    { name: t('searchBox.platformName.google'), type: 'google' },
    { name: t('searchBox.platformName.bilibili'), type: 'bilibili', url: 'https://search.bilibili.com/all?keyword={kw}' },
    { name: t('searchBox.platformName.xiaohongshu'), type: 'site', url: 'https://www.xiaohongshu.com/search_result?keyword={kw}' },
    { name: t('searchBox.platformName.zhihu'), type: 'site', url: 'https://www.zhihu.com/search?type=content&q={kw}' },
    { name: t('searchBox.platformName.weibo'), type: 'site', url: 'https://s.weibo.com/weibo?q={kw}' },
    { name: t('searchBox.platformName.twitter'), type: 'site', url: 'https://x.com/search?q={kw}&src=typed_query' },
  ],
  [t('searchBox.platformCate.games')]: [
    { name: t('searchBox.platformName.itchIo'), type: 'site', url: 'https://itch.io/search?q={kw}' },
    { name: t('searchBox.platformName.steam'), type: 'site', url: 'https://store.steampowered.com/search?term={kw}' },
  ],
}))

const selectedPlatform = ref<Platform>({ name: '', type: 'site' })

// 重置默认选中逻辑
const resetSelection = () => {
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

let timer: number | null = null
const debounce = (fn: Function, delay = 300) => (...args: any[]) => {
  if (timer) clearTimeout(timer)
  timer = window.setTimeout(() => fn(...args), delay)
}

const inputRef = ref<HTMLInputElement | null>(null)
const searchBoxRef = ref<HTMLElement | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)
const inputWidth = ref(0)
const inputLeft = ref(0)
const inputTop = ref(0)
const dropdownVisible = ref(false)
const isInputFocused = ref(false)

const updateInputPosition = () => {
  nextTick(() => {
    if (inputRef.value) {
      const rect = inputRef.value.getBoundingClientRect()
      inputWidth.value = rect.width
      inputLeft.value = rect.left
      inputTop.value = rect.bottom + 4
    }
  })
}

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
  if (!keyword.value.trim() || selectedCategory.value !== searchLabel) {
    suggestions.value = []
    dropdownVisible.value = false
    isLoading.value = false
    return
  }

  updateInputPosition()
  dropdownVisible.value = true
  isLoading.value = true

  try {
    let data: string[] = []
    switch (selectedPlatform.value.type) {
      case 'baidu': data = await getBaiduSuggestion(keyword.value); break
      case 'bing': data = await getBingSuggestion(keyword.value); break
      case 'google': data = await getGoogleSuggestion(keyword.value); break
      case 'bilibili': data = await getBiliBiliSuggestion(keyword.value); break
      default: data = []
    }
    suggestions.value = data
    if (data.length > 0 && hoveredIndex.value === -1) hoveredIndex.value = 0
  } catch {
    suggestions.value = []
  } finally {
    isLoading.value = false
  }
}

const debouncedFetch = debounce(fetchSuggestions, 300)

const doSearch = () => {
  const kw = encodeURIComponent(keyword.value.trim())
  if (!kw) return

  const mapping: Record<string, string> = {}
  Object.values(platforms.value).flat().forEach((p) => {
    if (p.url) mapping[p.name] = p.url.replace('{kw}', kw)
  })

  const searchLabel = t('searchBox.platformCate.search')
  const gameLabel = t('searchBox.platformCate.games')

  if (selectedCategory.value === searchLabel) {
    switch (selectedPlatform.value.type) {
      case 'baidu': window.open(`https://www.baidu.com/s?wd=${kw}`, '_blank'); break
      case 'bing': window.open(`https://www.bing.com/search?q=${kw}`, '_blank'); break
      case 'google': window.open(`https://www.google.com/search?q=${kw}`, '_blank'); break
      default:
        if (mapping[selectedPlatform.value.name])
          window.open(mapping[selectedPlatform.value.name], '_blank')
    }
  } else if (selectedCategory.value === gameLabel) {
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
    dropdownVisible.value = false
    hoveredIndex.value = -1
    isInputFocused.value = false
  }
}

watch(
    [dropdownVisible, isInputFocused],
    ([visible, focused]) => {
      setNavPageRevealLock('search-box', visible || focused)
    },
    { immediate: true }
)

onMounted(() => {
  window.addEventListener('resize', updateInputPosition)
  window.addEventListener('scroll', updateInputPosition, true)
  document.addEventListener('click', handleClickOutside)
  updateInputPosition()
  resetSelection() // 🟧 初始化默认选中项
})

onBeforeUnmount(() => {
  if (timer) clearTimeout(timer)
  window.removeEventListener('resize', updateInputPosition)
  window.removeEventListener('scroll', updateInputPosition, true)
  document.removeEventListener('click', handleClickOutside)
  setNavPageRevealLock('search-box', false)
})
</script>



<style scoped>
ul::-webkit-scrollbar {
  width: 6px;
}
ul::-webkit-scrollbar-thumb {
  background-color: rgba(0,0,0,0.2);
  border-radius: 3px;
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
  animation: fadeIn 0.2s ease-out forwards;
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
