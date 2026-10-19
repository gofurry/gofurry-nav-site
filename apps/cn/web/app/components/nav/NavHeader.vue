<template>
  <header
      ref="headerRef"
      class="nav-header"
  >
    <h1 class="sr-only">{{ homeHeading }}</h1>
    <HeroBackground :desktop-object-key="desktopObjectKey" :mobile-object-key="mobileObjectKey" />
    <div ref="searchRef" class="nav-header__search">
      <SearchBox />
    </div>

    <div
        ref="quickRef"
        v-if="showQuickAccess"
        class="nav-header__quick-access"
    >
      <NavQuickAccess
        :recent-sites="recentSites"
        :custom-sites="customSites"
        @visit-recent="visitRecentSite"
        @visit-custom="visitCustomSite"
        @manage="openCustomSitesModal"
      />
    </div>

    <div ref="hintRef" class="nav-header__scroll-hint" :class="{ 'nav-header__scroll-hint--hidden': !hintVisible }">
      <span class="nav-header__scroll-label">
        {{ t('navHeader.scrollHint') }}
      </span>
      <div class="nav-header__scroll-frame">
        <span class="nav-header__scroll-dot"></span>
      </div>
    </div>

    <transition name="quick-modal" @before-leave="deactivateCustomModal">
      <div
        v-if="showCustomSitesModal"
        class="quick-modal-backdrop"
        @click.self="closeCustomSitesModal()"
        @keydown="onCustomModalKey"
      >
        <div ref="customModal" class="quick-modal-panel" role="dialog" aria-modal="true" :aria-labelledby="customModalTitle" tabindex="-1">
          <div class="quick-modal-header">
            <div>
              <h3 :id="customModalTitle" class="quick-modal-title">{{ t('customSites.manageTitle') }}</h3>
              <p class="quick-modal-desc">{{ t('customSites.manageDescription', { count: customSites.length, max: MAX_CUSTOM_SITES }) }}</p>
            </div>
            <button class="quick-modal-close" type="button" :aria-label="t('common.cancel')" @click="closeCustomSitesModal()">
              <PhX :size="18" weight="regular" aria-hidden="true" />
            </button>
          </div>

          <form class="quick-modal-form" @submit.prevent="submitCustomSite">
            <div class="quick-modal-inline-field">
              <label class="sr-only" for="custom-site-name">{{ t('customSites.nameLabel') }}</label>
              <input
                id="custom-site-name"
                v-model.trim="customSiteForm.name"
                :placeholder="t('customSites.namePlaceholder')"
                type="text"
              />
            </div>

            <div class="quick-modal-inline-field quick-modal-inline-field-url">
              <label class="sr-only" for="custom-site-url">{{ t('customSites.urlLabel') }}</label>
              <input
                id="custom-site-url"
                v-model.trim="customSiteForm.url"
                :placeholder="t('customSites.urlPlaceholder')"
                type="text"
              />
            </div>

            <div class="quick-modal-inline-action">
              <button class="quick-modal-primary" type="submit" :disabled="customSites.length >= MAX_CUSTOM_SITES">
                {{ t('customSites.addSite') }}
              </button>
            </div>
          </form>
          <p v-if="customSiteError" class="quick-modal-error">{{ customSiteError }}</p>

          <p class="sr-only" role="status" aria-live="polite" aria-atomic="true">{{ customSortStatus }}</p>
          <div ref="customList" class="quick-modal-list">
            <div v-if="customSites.length" class="quick-modal-items" :class="{ 'quick-modal-items--sorting': draggingCustomSiteId }">
              <div
                v-for="item in customSites"
                :key="item.id"
                class="quick-modal-item"
                :data-custom-site-id="item.id"
                :class="{ 'quick-modal-item-dragging': draggingCustomSiteId === item.id }"
                :style="customRowStyle(item.id)"
              >
                <button class="quick-modal-grip" type="button" :aria-label="t('customSites.reorderSite', { name: item.name })"
                  :title="t('customSites.reorderHint')" @pointerdown="startCustomDrag($event, item.id)"
                  @pointermove="moveCustomDrag" @pointerup="finishCustomDrag" @pointercancel="cancelCustomDrag"
                  @lostpointercapture="cancelCustomDrag" @keydown="onCustomSortKey($event, item.id)">
                  <PhDotsSixVertical :size="20" weight="regular" aria-hidden="true" />
                </button>
                <div class="min-w-0 flex-1">
                  <p class="quick-modal-item-name">{{ item.name }}</p>
                  <p class="quick-modal-item-url">{{ item.url }}</p>
                </div>
                <button class="quick-modal-delete" type="button" :aria-label="t('customSites.deleteSite')" @click="deleteCustomSite(item.id)">
                  <PhX :size="18" weight="regular" aria-hidden="true" />
                </button>
              </div>
            </div>
            <p v-else class="quick-modal-empty">{{ t('customSites.emptyDescription') }}</p>
          </div>
        </div>
      </div>
    </transition>
  </header>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { PhDotsSixVertical, PhX } from '@phosphor-icons/vue'
import { useI18n } from 'vue-i18n'
import SearchBox from './SearchBox.vue'
import NavQuickAccess from './NavQuickAccess.vue'
import HeroBackground from './HeroBackground.vue'
import { loadRecentSites, RECENT_SITES_EVENT, type RecentSiteItem } from '@/utils/recentSites'
import {
  addCustomSite,
  CUSTOM_SITES_EVENT,
  loadCustomSites,
  MAX_CUSTOM_SITES,
  removeCustomSite,
  reorderCustomSites,
  visitCustomSite,
  type CustomSiteItem,
} from '@/utils/customSites'
import {
  readShowQuickAccess,
  subscribeNavHeaderSettingsChange,
} from '@/utils/navHeaderSettings'

defineProps<{
  desktopObjectKey?: string | null
  mobileObjectKey?: string | null
}>()

const { locale, t } = useI18n()
const homeHeading = computed(() => locale.value === 'en'
  ? 'GoFurry Navigation - Discover furry communities, art, fiction, games, tools, and site monitoring'
  : 'GoFurry 兽人控导航站 - 发现兽人社区、艺术、小说、游戏、工具与站点监测资源'
)
const recentSites = ref<RecentSiteItem[]>([])
const customSites = ref<CustomSiteItem[]>([])
const showQuickAccess = ref(true)
const showCustomSitesModal = ref(false)
const customSiteError = ref('')
const customSiteForm = ref({
  name: '',
  url: '',
})
const draggingCustomSiteId = ref<string | null>(null)
const route = useRoute()
const customModalTitle = useId()
const customModal = ref<HTMLElement | null>(null)
const customList = ref<HTMLElement | null>(null)
const customSortStatus = ref('')
const customOffsets = ref<Record<string, number>>({})
let customOpener: HTMLElement | null = null
let dragFrame: number | null = null
let drag: { id: string; pointer: number; handle: HTMLElement; y: number; start: number; scroll: number;
  rows: { id: string; top: number; height: number }[]; gap: number; order: string[] } | null = null

const headerRef = ref<HTMLElement | null>(null)
const searchRef = ref<HTMLElement | null>(null)
const quickRef = ref<HTMLElement | null>(null)
const hintRef = ref<HTMLElement | null>(null)
const hintVisible = ref(false)
let hintObserver: ResizeObserver | null = null
let hintFrame: number | null = null
function syncHint() {
  hintFrame = null
  if (window.innerWidth < 768 || !hintRef.value || !searchRef.value) { hintVisible.value = false; return }
  const hint = hintRef.value.getBoundingClientRect()
  const contentBottom = Math.max(searchRef.value.getBoundingClientRect().bottom, quickRef.value?.getBoundingClientRect().bottom ?? 0)
  hintVisible.value = hint.height > 0 && hint.top >= contentBottom + 24
}
function scheduleHint() {
  if (hintFrame === null) hintFrame = requestAnimationFrame(syncHint)
}
function observeHintLayout() {
  hintObserver?.disconnect()
  for (const element of [headerRef.value, searchRef.value, quickRef.value, hintRef.value]) {
    if (element) hintObserver?.observe(element)
  }
  scheduleHint()
}
watch(showQuickAccess, observeHintLayout, { flush: 'post' })

let stopNavHeaderSettingsSubscription: (() => void) | null = null

function syncRecentSites() {
  if (!import.meta.client) {
    return
  }

  recentSites.value = loadRecentSites().slice(0, 8)
}

function handleRecentSitesChange() {
  syncRecentSites()
}

function syncCustomSites() {
  if (!import.meta.client) {
    return
  }

  customSites.value = loadCustomSites()
}

function handleCustomSitesChange() {
  cancelCustomDrag()
  syncCustomSites()
}

function syncNavHeaderSettings() {
  showQuickAccess.value = readShowQuickAccess()
}

function visitRecentSite(site: RecentSiteItem) {
  window.open(site.url, '_blank')
}

async function openCustomSitesModal() {
  customOpener = document.activeElement instanceof HTMLElement ? document.activeElement : null
  customSortStatus.value = ''
  showCustomSitesModal.value = true
  await nextTick()
  if (showCustomSitesModal.value) customModal.value?.querySelector<HTMLInputElement>('input')?.focus({ preventScroll: true })
}

function closeCustomSitesModal(restoreFocus = true) {
  cancelCustomDrag()
  if (restoreFocus && customOpener?.isConnected && customOpener.getClientRects().length) customOpener.focus({ preventScroll: true })
  showCustomSitesModal.value = false
  customSiteError.value = ''
  customSiteForm.value = {
    name: '',
    url: '',
  }
}

function deactivateCustomModal(element: Element) {
  const backdrop = element as HTMLElement
  backdrop.inert = true
  backdrop.setAttribute('aria-hidden', 'true')
}

function onCustomModalKey(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    event.preventDefault(); event.stopPropagation()
    if (drag) cancelCustomDrag()
    else closeCustomSitesModal()
  } else if (event.key === 'Tab') {
    cancelCustomDrag()
    const controls = [...customModal.value!.querySelectorAll<HTMLElement>('button:not(:disabled), input:not(:disabled), [tabindex="0"]')]
      .filter(el => el.getClientRects().length)
    const first = controls[0], last = controls.at(-1)
    if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last?.focus() }
    else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first?.focus() }
  }
}

function submitCustomSite() {
  cancelCustomDrag()
  if (!customSiteForm.value.name.trim()) {
    customSiteError.value = t('customSites.nameRequired')
    return
  }

  if (!customSiteForm.value.url.trim()) {
    customSiteError.value = t('customSites.urlRequired')
    return
  }

  const added = addCustomSite({
    name: customSiteForm.value.name,
    url: customSiteForm.value.url,
  })

  if (!added) {
    customSiteError.value = customSites.value.length >= MAX_CUSTOM_SITES
      ? t('customSites.maxReached')
      : t('customSites.urlInvalid')
    return
  }

  customSiteError.value = ''
  customSiteForm.value = {
    name: '',
    url: '',
  }
  syncCustomSites()
}

function deleteCustomSite(id: string) {
  cancelCustomDrag()
  removeCustomSite(id)
  syncCustomSites()
  void nextTick(() => customModal.value?.querySelector<HTMLInputElement>('input')?.focus({ preventScroll: true }))
}

function customRowStyle(id: string) {
  const offset = customOffsets.value[id]
  return offset === undefined ? undefined : { transform: `translateY(${offset}px)` }
}

function startCustomDrag(event: PointerEvent, id: string) {
  if (event.button !== 0 || !event.isPrimary || drag || !customList.value) return
  event.preventDefault()
  const handle = event.currentTarget as HTMLElement
  handle.focus({ preventScroll: true })
  const nodes = [...customList.value.querySelectorAll<HTMLElement>('.quick-modal-item')]
  const origin = nodes[0]!.getBoundingClientRect().top
  const rows = nodes.map(node => { const box = node.getBoundingClientRect(); return { id: node.dataset.customSiteId!, top: box.top - origin, height: box.height } })
  drag = { id, pointer: event.pointerId, handle, y: event.clientY, start: event.clientY, scroll: customList.value.scrollTop,
    rows, gap: parseFloat(getComputedStyle(nodes[0]!.parentElement!).rowGap) || 0, order: rows.map(row => row.id) }
  draggingCustomSiteId.value = id
  handle.setPointerCapture(event.pointerId)
  dragFrame = requestAnimationFrame(updateCustomDrag)
}

function moveCustomDrag(event: PointerEvent) {
  if (drag?.pointer === event.pointerId) { event.preventDefault(); drag.y = event.clientY }
}

function updateCustomDrag() {
  dragFrame = null
  if (!drag || !customList.value) return
  const bounds = customList.value.getBoundingClientRect()
  const edge = drag.y < bounds.top + 24 ? -6 : drag.y > bounds.bottom - 24 ? 6 : 0
  customList.value.scrollTop += edge
  const active = drag.rows.find(row => row.id === drag!.id)!
  const last = drag.rows.at(-1)!
  const movement = drag.y - drag.start + customList.value.scrollTop - drag.scroll
  const offset = Math.max(-active.top, Math.min(last.top + last.height - active.height - active.top,
    movement))
  const others = drag.rows.filter(row => row.id !== drag!.id)
  // Use pointer intent for ordering; clamping a taller row's paint at the end
  // must not prevent its center from crossing the final shorter neighbor.
  const center = active.top + movement + active.height / 2
  const index = others.filter(row => movement > 0 ? row.top + row.height / 2 <= center : row.top + row.height / 2 < center).length
  others.splice(index, 0, active)
  drag.order = others.map(row => row.id)
  let top = 0
  const offsets: Record<string, number> = {}
  for (const row of others) { offsets[row.id] = row.id === drag.id ? offset : top - row.top; top += row.height + drag.gap }
  customOffsets.value = offsets
  dragFrame = requestAnimationFrame(updateCustomDrag)
}

function cancelCustomDrag() {
  const previous = drag
  drag = null
  if (dragFrame !== null) cancelAnimationFrame(dragFrame)
  dragFrame = null
  draggingCustomSiteId.value = null
  customOffsets.value = {}
  if (previous?.handle.hasPointerCapture(previous.pointer)) previous.handle.releasePointerCapture(previous.pointer)
}

function announceCustomSort(id: string, order: string[]) {
  customSortStatus.value = t('customSites.reordered', { name: customSites.value.find(item => item.id === id)?.name, position: order.indexOf(id) + 1, count: order.length })
}

function finishCustomDrag(event: PointerEvent) {
  if (!drag || drag.pointer !== event.pointerId) return
  // Include the final pointer position even when pointerup precedes the next RAF.
  drag.y = event.clientY
  if (dragFrame !== null) cancelAnimationFrame(dragFrame)
  updateCustomDrag()
  const { id, order, rows } = drag
  cancelCustomDrag()
  if (order.some((value, index) => value !== rows[index]!.id)) { reorderCustomSites(order); announceCustomSort(id, order) }
}

async function onCustomSortKey(event: KeyboardEvent, id: string) {
  const delta = event.key === 'ArrowUp' ? -1 : event.key === 'ArrowDown' ? 1 : 0
  if (!delta) return
  event.preventDefault()
  cancelCustomDrag()
  const order = customSites.value.map(item => item.id), from = order.indexOf(id), to = from + delta
  if (to < 0 || to >= order.length) return
  order.splice(from, 1); order.splice(to, 0, id)
  reorderCustomSites(order); announceCustomSort(id, order)
  await nextTick()
  const handle = customModal.value?.querySelector<HTMLElement>(`[data-custom-site-id="${CSS.escape(id)}"] .quick-modal-grip`)
  handle?.focus({ preventScroll: true })
  // Reveal only within the dialog list; keyboard sorting must never scroll Home.
  const row = handle?.closest('.quick-modal-item')?.getBoundingClientRect()
  const list = customList.value?.getBoundingClientRect()
  if (row && list && customList.value) {
    if (row.top < list.top) customList.value.scrollTop -= list.top - row.top
    else if (row.bottom > list.bottom) customList.value.scrollTop += row.bottom - list.bottom
  }
}

function customModalResize() {
  cancelCustomDrag()
  if (window.innerWidth < 768 && showCustomSitesModal.value) closeCustomSitesModal(false)
}
watch(() => route.fullPath, () => closeCustomSitesModal(false))

onMounted(() => {
  hintObserver = new ResizeObserver(scheduleHint)
  observeHintLayout()
  window.addEventListener('resize', scheduleHint)
  window.addEventListener('resize', customModalResize)
  window.addEventListener('blur', cancelCustomDrag)
  try {
    syncNavHeaderSettings()
    syncRecentSites()
    syncCustomSites()
    stopNavHeaderSettingsSubscription = subscribeNavHeaderSettingsChange(({ showQuickAccess: nextValue }) => {
      showQuickAccess.value = nextValue
    })
    window.addEventListener(RECENT_SITES_EVENT, handleRecentSitesChange)
    window.addEventListener(CUSTOM_SITES_EVENT, handleCustomSitesChange)
    window.addEventListener('storage', handleRecentSitesChange)
    window.addEventListener('storage', handleCustomSitesChange)
  } catch (err) {
    console.error('Get background image URL err:', err)
  }
})

onUnmounted(() => {
  cancelCustomDrag()
  window.removeEventListener('blur', cancelCustomDrag)
  window.removeEventListener('resize', customModalResize)
  hintObserver?.disconnect()
  window.removeEventListener('resize', scheduleHint)
  if (hintFrame !== null) cancelAnimationFrame(hintFrame)
  window.removeEventListener(RECENT_SITES_EVENT, handleRecentSitesChange)
  window.removeEventListener(CUSTOM_SITES_EVENT, handleCustomSitesChange)
  window.removeEventListener('storage', handleRecentSitesChange)
  window.removeEventListener('storage', handleCustomSitesChange)
  stopNavHeaderSettingsSubscription?.()
  stopNavHeaderSettingsSubscription = null
})
</script>
