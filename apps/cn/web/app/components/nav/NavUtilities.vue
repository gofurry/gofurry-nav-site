<template>
  <div ref="root" class="gf-nav__utilities pointer-events-none absolute inset-x-0 top-full hidden md:block"
    @keydown.esc.prevent.stop="closeSearch(true)">
    <nav class="gf-nav__utility-row flex h-11 items-center justify-center gap-0.5 lg:gap-1" :aria-label="t('nav.tools.label')">
      <NuxtLink :to="localePath('/games/prize')" class="gf-nav__icon-button pointer-events-auto inline-flex items-center justify-center"
        :title="t('game.tools.lottery')" :aria-label="t('game.tools.lottery')" @click="closeSearch(false)">
        <PhGift :size="20" weight="regular" aria-hidden="true" />
      </NuxtLink>
      <button ref="trigger" class="gf-nav__icon-button pointer-events-auto inline-flex items-center justify-center" type="button"
        :title="t('nav.tools.search')" :aria-label="t('nav.tools.search')" :aria-expanded="open" :aria-controls="panelId" @click="toggleSearch">
        <PhMagnifyingGlass :size="20" weight="regular" aria-hidden="true" />
      </button>
      <a href="https://github.com/gofurry/gofurry-nav-site/issues" target="_blank" rel="noopener noreferrer"
        class="gf-nav__icon-button pointer-events-auto inline-flex items-center justify-center"
        :title="t('nav.tools.feedback')" :aria-label="t('nav.tools.feedback')" @click="closeSearch(false)">
        <PhChatDots :size="20" weight="regular" aria-hidden="true" />
      </a>
    </nav>
    <Transition name="nav-utility-search" @before-leave="deactivatePanel">
      <section v-if="open" :id="panelId" ref="panel" class="gf-nav__search-panel pointer-events-auto absolute right-0 top-full mt-2 overflow-hidden"
        :aria-label="t('nav.tools.search')" :aria-busy="status === 'loading'">
        <div class="gf-nav__search-input flex items-center gap-2 p-3">
          <PhMagnifyingGlass :size="20" weight="regular" aria-hidden="true" />
          <input ref="input" v-model.trim="keyword" class="min-w-0 flex-1" type="search"
            :aria-label="t('nav.tools.search')" :placeholder="t('nav.tools.searchPlaceholder')" />
        </div>
        <p v-if="status === 'loading'" class="gf-nav__search-status px-3 py-4" role="status">{{ t('common.loading') }}</p>
        <p v-else-if="status === 'error'" class="gf-nav__search-status px-3 py-4" role="status">{{ t('nav.tools.searchUnavailable') }}</p>
        <div v-else-if="keyword && results.length" class="gf-nav__search-results overflow-y-auto overscroll-contain p-2">
          <button v-for="item in results" :key="item.id" type="button" class="flex w-full items-center gap-2 p-2 text-left" @click="openSite(item)">
            <ManagedAssetImage :object-key="item.icon" :alt="item.name" class="h-8 w-8 shrink-0 object-cover" loading="lazy" decoding="async" />
            <span class="min-w-0"><strong class="block truncate">{{ item.name }}</strong><small class="block truncate">{{ item.info }}</small></span>
          </button>
        </div>
        <p v-else-if="keyword" class="gf-nav__search-status px-3 py-4" role="status">{{ t('nav.tools.noSearchResults') }}</p>
      </section>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { PhChatDots, PhGift, PhMagnifyingGlass } from '@phosphor-icons/vue'
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import ManagedAssetImage from '@/components/common/ManagedAssetImage.vue'
import { getNavSiteDirectory } from '@/services/nav'
import type { Site } from '@/types/nav'
import { readDisplayMode, subscribeModeChange, type DisplayMode } from '@/utils/modeStorage'
import { recordRecentSite, toExternalUrl } from '@/utils/recentSites'

const props = defineProps<{ blocked?: boolean }>()
const { t, locale } = useI18n()
const localePath = useLocalePath()
const route = useRoute()
const panelId = useId()
const root = ref<HTMLElement | null>(null)
const trigger = ref<HTMLButtonElement | null>(null)
const input = ref<HTMLInputElement | null>(null)
const panel = ref<HTMLElement | null>(null)
const open = ref(false)
const keyword = ref('')
const status = ref<'idle' | 'loading' | 'ready' | 'error'>('idle')
const displayMode = ref<DisplayMode>('sfw')
const directory = ref<Site[]>([])
const cache = new Map<string, Site[]>()
let controller: AbortController | null = null
let revision = 0
let mobileQuery: MediaQueryList | null = null
let stopModeSubscription: (() => void) | null = null

const results = computed(() => {
  const query = keyword.value.toLowerCase()
  if (!query || status.value !== 'ready') return []
  return directory.value
    .filter(item => displayMode.value === 'nsfw' || String(item.nsfw) !== '1')
    .filter(item => item.name?.toLowerCase().includes(query) || item.info?.toLowerCase().includes(query))
    .slice(0, 8)
})

function closeSearch(restoreFocus = false) {
  if (restoreFocus && open.value) trigger.value?.focus({ preventScroll: true })
  open.value = false
  keyword.value = ''
  revision++
  controller?.abort()
  controller = null
  directory.value = []
  status.value = 'idle'
}

async function toggleSearch() {
  if (open.value) { closeSearch(true); return }
  if (props.blocked || mobileQuery?.matches) return
  open.value = true
  const current = ++revision
  const lang = locale.value === 'en' ? 'en' : 'zh'
  const saved = cache.get(lang)
  directory.value = saved ?? []
  status.value = saved ? 'ready' : 'loading'
  await nextTick()
  if (!open.value || revision !== current) return
  input.value?.focus({ preventScroll: true })
  if (saved) return
  const request = new AbortController()
  controller = request
  try {
    const items = await getNavSiteDirectory(lang, request.signal)
    if (!open.value || current !== revision) return
    cache.set(lang, items)
    directory.value = items
    status.value = 'ready'
  } catch {
    if (current === revision && !request.signal.aborted) status.value = 'error'
  } finally {
    if (controller === request) controller = null
  }
}

function deactivatePanel(element: Element) {
  const target = element as HTMLElement
  target.inert = true
  target.setAttribute('aria-hidden', 'true')
}

function domainList(item: Site): string[] {
  if (Array.isArray(item.domain)) return item.domain
  if (typeof item.domain !== 'string' || !item.domain) return []
  try {
    const parsed = JSON.parse(item.domain)
    return Array.isArray(parsed?.domain) ? parsed.domain : [item.domain]
  } catch { return [item.domain] }
}

function openSite(item: Site) {
  const url = toExternalUrl(domainList(item)[0] ?? '')
  if (!url) return
  recordRecentSite({ id: item.id, name: item.name, url })
  window.open(url, '_blank', 'noopener,noreferrer')
  closeSearch(true)
}

function outsideInteraction(event: Event) {
  if (open.value && root.value && !event.composedPath().includes(root.value)) closeSearch(false)
}
function handleBreakpoint() { if (mobileQuery?.matches) closeSearch(false) }
watch(() => route.fullPath, () => closeSearch(false))
watch(locale, () => closeSearch(false), { flush: 'sync' })
watch(() => props.blocked, value => { if (value) closeSearch(false) }, { flush: 'sync' })
onMounted(() => {
  displayMode.value = readDisplayMode()
  stopModeSubscription = subscribeModeChange(change => { displayMode.value = change.displayMode })
  mobileQuery = matchMedia('(max-width: 767px)')
  mobileQuery.addEventListener('change', handleBreakpoint)
  document.addEventListener('pointerdown', outsideInteraction)
  document.addEventListener('focusin', outsideInteraction)
})
onUnmounted(() => {
  closeSearch(false)
  stopModeSubscription?.()
  mobileQuery?.removeEventListener('change', handleBreakpoint)
  document.removeEventListener('pointerdown', outsideInteraction)
  document.removeEventListener('focusin', outsideInteraction)
})
</script>
