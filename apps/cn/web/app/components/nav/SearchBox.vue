<template>
  <div
      ref="searchBoxRef"
      class="search-box-shell relative isolate z-30 mt-2 flex w-full flex-col items-center md:mt-0"
  >
    <!-- 搜索类别 -->
    <div class="search-category-row mb-3 w-full max-w-[500px] py-2.5">
      <div class="search-category-track overflow-x-auto">
      <div class="search-category-items mx-auto flex w-max flex-nowrap gap-2 md:gap-3">
        <button
            v-for="item in categories"
            type="button"
            :aria-pressed="selectedCategory === item"
            :key="item"
            @click="selectCategory(item, $event)"
            :class="[
              'search-chip shrink-0 cursor-pointer whitespace-nowrap px-3 py-1.5',
              selectedCategory === item ? 'search-chip-active' : ''
            ]"
        >
          {{ item }}
        </button>
      </div>
      </div>
    </div>

    <!-- 搜索框 -->
    <div class="relative z-30 w-full max-w-[500px]">
      <input
          ref="inputRef"
          type="text"
          :value="keyword"
          @keydown="handleKeyDown"
          @input="handleInput"
          @compositionstart="handleCompositionStart"
          @compositionend="handleCompositionEnd"
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
      <Transition name="search-suggestions" @leave="(_element, done) => done()">
      <ul
          ref="dropdownRef"
          v-if="(keyword.trim() && dropdownVisible) || isLoading"
          class="search-suggestion-list absolute left-0 top-[calc(100%+0.5rem)] z-[999] max-h-60 w-full origin-top-left overflow-y-auto overscroll-contain"
          @wheel.stop
          @touchmove.stop
      >
        <!-- 标题 -->
        <li class="search-suggestion-header px-4 py-2">
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
              class="search-suggestion-item cursor-pointer px-4 py-3"
              :class="hoveredIndex === index ? 'search-suggestion-item-active' : ''"
          >
            <!-- 关键词高亮 -->
            <span><template v-for="(segment, part) in highlightSegments(item)" :key="part"><span v-if="part % 2" class="search-highlight">{{ segment }}</span><template v-else>{{ segment }}</template></template></span>
          </li>
        </template>

        <li v-else-if="keyword.trim()" class="search-suggestion-empty px-4 py-3 text-center">
          {{ t('searchBox.noSuggest') }}
        </li>
      </ul>
      </Transition>
    </div>

    <div ref="platformPicker" class="search-platform-picker pointer-events-none relative mt-2 h-11 w-full max-w-[500px] md:hidden"
      @keydown.esc.stop.prevent="closePicker(true)" @focusout="handlePickerFocusOut">
      <div id="search-platform-options" role="group" :aria-label="platformTitle"
        class="absolute right-0 top-0 grid w-full gap-1.5" :style="{ gridTemplateColumns: `repeat(${pickerColumns}, minmax(0, 1fr))` }">
        <template v-for="(platform, index) in pickerPlatforms" :key="index">
          <button v-if="index === selectedSlot" ref="platformTrigger" type="button"
            class="search-platform-trigger search-platform-option pointer-events-auto inline-flex min-w-0 items-center justify-center gap-1 px-2"
            :style="{ gridColumn: index + 1, gridRow: 1 }"
            :aria-label="platform.name" :title="platform.name" aria-pressed="true"
            :aria-expanded="pickerExpanded" aria-controls="search-platform-options" @click="togglePicker">
            <span class="min-w-0 truncate">{{ platform.name }}</span>
            <PhCaretUp v-if="pickerExpanded" :size="12" class="shrink-0" aria-hidden="true" />
            <PhCaretDown v-else :size="12" class="shrink-0" aria-hidden="true" />
          </button>
          <Transition v-else name="search-platform-option" :css="!reducedMotion" @leave="(_element, done) => done()">
            <button v-if="pickerExpanded" type="button"
              class="search-platform-option pointer-events-auto min-w-0 truncate px-2"
              :style="{ gridColumn: index % pickerColumns + 1, gridRow: Math.floor(index / pickerColumns) + 1 }"
              :aria-label="platform.name" :title="platform.name" aria-pressed="false" @click="choosePlatform(index)">{{ platform.name }}</button>
          </Transition>
        </template>
      </div>
    </div>

    <!-- 搜索平台 -->
    <div class="search-platform-row mt-1 hidden w-full md:block max-w-[620px] px-3 py-2.5">
      <div class="grid grid-cols-2 justify-center gap-2 md:flex md:flex-wrap md:gap-2">
        <div
            v-for="platform in platforms[selectedCategory]"
            :key="platform.name"
            @click="choosePlatform(platformOrder.indexOf(platform.id), false)"
            :class="[
              'search-chip cursor-pointer px-2.5 py-1.5 text-center whitespace-nowrap',
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
import { PhCaretDown, PhCaretUp } from '@phosphor-icons/vue'
import { getSearchSuggestion } from '@/services/nav'
import { useI18n } from 'vue-i18n'
import { useReducedMotion } from '@/composables/useReducedMotion'
import { setNavPageRevealLock } from '@/utils/navPageReveal'

const { t, locale } = useI18n()
const route = useRoute()
const reducedMotion = useReducedMotion()
const currentLang = computed(() => locale.value === 'en' ? 'en' : 'zh')
const searchActionLabel = computed(() => currentLang.value === 'en' ? 'Search GoFurry resources' : '搜索 GoFurry 资源')

// 搜索类别
const categories = computed(() => [
  t('searchBox.platformCate.search'),
  t('searchBox.platformCate.furry'),
  t('searchBox.platformCate.games'),
  t('searchBox.platformCate.art'),
  t('searchBox.platformCate.novel'),
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
  id: string
  name: string
  type: 'bing' | 'google' | 'duckduckgo' | 'bilibili' | 'site'
  url?: string
}

// 平台数据
const platforms = computed<Record<string, Platform[]>>(() => ({
  [t('searchBox.platformCate.search')]: [
    { id: 'bing', name: t('searchBox.platformName.bing'), type: 'bing' },
    { id: 'google', name: t('searchBox.platformName.google'), type: 'google' },
    { id: 'duckduckgo', name: t('searchBox.platformName.duckduckgo'), type: 'duckduckgo' },
    { id: 'bilibili', name: t('searchBox.platformName.bilibili'), type: 'bilibili', url: 'https://search.bilibili.com/all?keyword={kw}' },
    { id: 'xiaohongshu', name: t('searchBox.platformName.xiaohongshu'), type: 'site', url: 'https://www.xiaohongshu.com/search_result?keyword={kw}' },
    { id: 'zhihu', name: t('searchBox.platformName.zhihu'), type: 'site', url: 'https://www.zhihu.com/search?type=content&q={kw}' },
    { id: 'weibo', name: t('searchBox.platformName.weibo'), type: 'site', url: 'https://s.weibo.com/weibo?q={kw}' },
    { id: 'twitter', name: t('searchBox.platformName.twitter'), type: 'site', url: 'https://x.com/search?q={kw}&src=typed_query' },
  ],
  [t('searchBox.platformCate.furry')]: [
    { id: 'wikifur', name: t('searchBox.platformName.wikifur'), type: 'site', url: `https://${currentLang.value}.wikifur.com/wiki/{kw}` },
    { id: 'yiffParty', name: t('searchBox.platformName.yiffParty'), type: 'site', url: 'https://yiff-party.com/search/?tags={kw}' },
    { id: 'furaffinity', name: t('searchBox.platformName.furaffinity'), type: 'site', url: 'https://www.furaffinity.net/search/?q={kw}' },
    { id: 'e621', name: t('searchBox.platformName.e621'), type: 'site', url: 'https://e621.net/posts?tags={kw}' },
    { id: 'wilddream', name: t('searchBox.platformName.wilddream'), type: 'site', url: 'https://www.wilddream.net/Art/index/index?keyword={kw}' },
  ],
  [t('searchBox.platformCate.art')]: [
    { id: 'pixiv', name: t('searchBox.platformName.pixiv'), type: 'site', url: 'https://www.pixiv.net/tags/{kw}' },
    { id: 'deviantart', name: t('searchBox.platformName.deviantart'), type: 'site', url: 'https://www.deviantart.com/search?q={kw}' },
    { id: 'artstation', name: t('searchBox.platformName.artstation'), type: 'site', url: 'https://www.artstation.com/search?query={kw}' },
    { id: 'pinterest', name: t('searchBox.platformName.pinterest'), type: 'site', url: 'https://www.pinterest.com/search/pins/?q={kw}' },
    { id: 'zcool', name: t('searchBox.platformName.zcool'), type: 'site', url: 'https://www.zcool.com.cn/search/content?word={kw}' },
  ],
  [t('searchBox.platformCate.novel')]: [
    { id: 'furrynovel', name: t('searchBox.platformName.furrynovel'), type: 'site', url: 'https://furrynovel.com/zh/search?keyword={kw}' },
    { id: 'linpx', name: t('searchBox.platformName.linpx'), type: 'site', url: 'https://furrynovel.ink/search?word={kw}' },
  ],
  [t('searchBox.platformCate.games')]: [
    { id: 'itchIo', name: t('searchBox.platformName.itchIo'), type: 'site', url: 'https://itch.io/search?q={kw}' },
    { id: 'steam', name: t('searchBox.platformName.steam'), type: 'site', url: 'https://store.steampowered.com/search?term={kw}' },
    { id: 'epic', name: t('searchBox.platformName.epic'), type: 'site', url: 'https://store.epicgames.com/browse?q={kw}' },
  ],
}))

const selectedPlatform = ref<Platform>({ id: '', name: '', type: 'site' })
let timer: number | null = null
let suggestionAbortController: AbortController | null = null
let suggestionRequestId = 0
let composing = false
let scheduledKeyword: string | null = null

const inputRef = ref<HTMLInputElement | null>(null)
const searchBoxRef = ref<HTMLElement | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)
const dropdownVisible = ref(false)
const isInputFocused = ref(false)

const pickerExpanded = ref(false)
const platformTrigger = ref<HTMLButtonElement[] | null>(null)
const platformPicker = ref<HTMLElement | null>(null)
const platformTitle = computed(() => currentLang.value === 'en' ? 'Search platform' : '选择搜索平台')
// Order uses stable provider IDs, never translated display names.
const platformOrder = ref<string[]>([])
const pickerColumns = computed(() => Math.min(4, platformOrder.value.length))
const selectedSlot = computed(() => pickerColumns.value - 1)
const pickerPlatforms = computed(() => platformOrder.value.map(id => platforms.value[selectedCategory.value]!.find(platform => platform.id === id)!))
let mobileQuery: MediaQueryList | undefined
let blurTimer: number | undefined

function selectCategory(category: string, event: MouseEvent) {
  closePicker()
  selectedCategory.value = category
  const button = event.currentTarget as HTMLElement
  const track = button.closest<HTMLElement>('.search-category-track')!
  const item = button.getBoundingClientRect(), bounds = track.getBoundingClientRect()
  // Only move the horizontal strip, never scroll the Hero/page to the focused button.
  if (item.left < bounds.left) track.scrollLeft -= bounds.left - item.left
  else if (item.right > bounds.right) track.scrollLeft += item.right - bounds.right
}

function togglePicker() {
  if (!mobileQuery?.matches) return
  if (pickerExpanded.value) { closePicker(true); return }
  window.clearTimeout(blurTimer)
  cancelSuggestions()
  isInputFocused.value = false
  pickerExpanded.value = true
}

function closePicker(restoreFocus = false) {
  pickerExpanded.value = false
  if (restoreFocus) platformTrigger.value?.[0]?.focus({ preventScroll: true })
}

function choosePlatform(index: number, restoreFocus = true) {
  const selected = pickerPlatforms.value[index]
  if (!selected) return
  const next = [...platformOrder.value]
  const current = next[selectedSlot.value]!
  next[selectedSlot.value] = selected.id
  next[index] = current
  platformOrder.value = next
  selectedPlatform.value = selected
  // The main slot remains the stable selected/focus anchor; the grid stays expanded.
  if (restoreFocus) platformTrigger.value?.[0]?.focus({ preventScroll: true })
}

function handlePickerFocusOut(event: FocusEvent) {
  if (!platformPicker.value?.contains(event.relatedTarget as Node | null)) closePicker()
}

function handleBreakpoint() {
  if (mobileQuery?.matches) return
  const ownedFocus = platformPicker.value?.contains(document.activeElement)
  closePicker()
  if (ownedFocus) searchBoxRef.value?.querySelector<HTMLButtonElement>('.search-category-row [aria-pressed="true"]')?.focus({ preventScroll: true })
}

watch(() => route.fullPath, () => closePicker())

// 重置默认选中逻辑
const resetSelection = () => {
  cancelSuggestions()
  closePicker()
  initDefaultCategory()
  const defaultPlatform = platforms.value[selectedCategory.value]?.[0]
  selectedPlatform.value = defaultPlatform || { id: '', name: '', type: 'site' }
  const order = platforms.value[selectedCategory.value]?.map(platform => platform.id) ?? []
  if (order.length) order.splice(Math.min(3, order.length - 1), 0, order.shift()!)
  platformOrder.value = order
  suggestions.value = []
  hoveredIndex.value = -1
}

// 当语言或类别变化时重设默认平台
watch([categories, selectedCategory, platforms], () => {
  resetSelection()
}, { immediate: true })

const highlightSegments = (item: string) => {
  if (!keyword.value.trim()) return [item]
  const escapedKeyword = keyword.value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return item.split(new RegExp(`(${escapedKeyword})`, 'gi'))
}

const fetchSuggestions = async () => {
  const searchLabel = t('searchBox.platformCate.search')
  const requestKeyword = keyword.value.trim()
  if (pickerExpanded.value || composing || !requestKeyword || selectedCategory.value !== searchLabel) {
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
    const response = await getSearchSuggestion(requestKeyword, controller.signal)
    if (requestId !== suggestionRequestId) return
    if (response.state === 'unavailable') {
      suggestions.value = []
      dropdownVisible.value = false
      return
    }
    suggestions.value = response.suggestions
    if (response.suggestions.length > 0 && hoveredIndex.value === -1) hoveredIndex.value = 0
  } catch {
    if (controller.signal.aborted || requestId !== suggestionRequestId) return
    suggestions.value = []
    dropdownVisible.value = false
  } finally {
    if (requestId === suggestionRequestId) {
      isLoading.value = false
    }
    if (suggestionAbortController === controller) {
      suggestionAbortController = null
    }
  }
}

function scheduleSuggestions() {
  if (pickerExpanded.value || composing || scheduledKeyword === keyword.value) return
  cancelSuggestions()
  if (!keyword.value.trim() || selectedCategory.value !== t('searchBox.platformCate.search')) return
  scheduledKeyword = keyword.value
  timer = window.setTimeout(() => {
    timer = null
    void fetchSuggestions()
  }, 600)
}

function handleInput(event: Event) {
  closePicker()
  keyword.value = (event.currentTarget as HTMLInputElement).value
  if (!composing) scheduleSuggestions()
}

function handleCompositionStart() {
  closePicker()
  composing = true
  cancelSuggestions()
}

function handleCompositionEnd(event: CompositionEvent) {
  composing = false
  keyword.value = (event.currentTarget as HTMLInputElement).value
  // A trailing input event with the same final text must not restart the debounce.
  scheduleSuggestions()
}

function handleKeyDown(event: KeyboardEvent) {
  if (composing || event.isComposing) return
  switch (event.key) {
    case 'Enter': event.preventDefault(); handleEnterKey(); break
    case 'ArrowDown': event.preventDefault(); handleArrowDown(); break
    case 'ArrowUp': event.preventDefault(); handleArrowUp(); break
    case 'Escape': event.preventDefault(); event.stopPropagation(); closeSearchSuggestions(); break
  }
}

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
  closePicker()
  isInputFocused.value = true
  if (keyword.value.trim()) {
    scheduledKeyword = null
    scheduleSuggestions()
  }
}

const handleInputBlur = () => {
  window.clearTimeout(blurTimer)
  blurTimer = window.setTimeout(() => {
    const activeElement = document.activeElement
    isInputFocused.value = !!(
      !pickerExpanded.value && activeElement &&
      searchBoxRef.value?.contains(activeElement)
    )
  }, 0)
}

const handleClickOutside = (e: MouseEvent) => {
  if (platformPicker.value && !e.composedPath().includes(platformPicker.value)) closePicker()
  if (searchBoxRef.value && !searchBoxRef.value.contains(e.target as Node)) {
    closeSearchSuggestions()
  }
}

const closeSearchSuggestions = () => {
  cancelSuggestions()
  isInputFocused.value = false
}

function cancelSuggestions() {
  if (timer !== null) clearTimeout(timer)
  timer = null
  scheduledKeyword = null
  abortSuggestionRequest()
  suggestions.value = []
  dropdownVisible.value = false
  hoveredIndex.value = -1
  isLoading.value = false
}

function abortSuggestionRequest() {
  suggestionRequestId++
  suggestionAbortController?.abort()
  suggestionAbortController = null
}

watch(
    [dropdownVisible, isInputFocused, pickerExpanded],
    ([visible, focused, expanded]) => {
      setNavPageRevealLock('search-box', visible || focused || expanded)
    },
    { immediate: true, flush: 'sync' }
)

onMounted(() => {
  mobileQuery = window.matchMedia('(max-width: 767px)')
  mobileQuery.addEventListener('change', handleBreakpoint)
  document.addEventListener('click', handleClickOutside)
  resetSelection() // 🟧 初始化默认选中项
})

onBeforeUnmount(() => {
  closePicker()
  window.clearTimeout(blurTimer)
  mobileQuery?.removeEventListener('change', handleBreakpoint)
  cancelSuggestions()
  document.removeEventListener('click', handleClickOutside)
  setNavPageRevealLock('search-box', false)
})
</script>
