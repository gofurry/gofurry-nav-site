<template>
  <Teleport to="body">
    <Transition name="preferences-modal" :css="!reducedMotion" @before-leave="deactivateLeaving"
        @enter="trackTransition" @leave="trackTransition"
        @enter-cancelled="cancelTransition?.()" @leave-cancelled="cancelTransition?.()">
    <div
        v-if="show"
        ref="backdrop"
        class="gf-modal-backdrop fixed inset-0 z-[120] flex items-center justify-center px-3 py-4 sm:px-4"
    >
      <div class="gf-modal gf-preferences-modal">
        <div class="gf-modal__header">
          <h2 class="gf-modal__title">
            {{ t("navbar.preferences") }}
          </h2>

          <div class="gf-modal__header-actions">
            <button
                type="button"
                class="gf-button gf-button--ghost"
                @click="cancel"
            >
              {{ t("common.cancel") }}
            </button>
            <button
                type="button"
                class="gf-button gf-button--primary"
                @click="save"
            >
              {{ t("common.save") }}
            </button>
          </div>
        </div>

        <div class="preferences-tabs" role="tablist" :aria-label="t('navbar.preferences')">
          <button v-for="(tab, index) in tabs" :id="`${panelId}-tab-${index}`" :key="tab" type="button" role="tab"
              :aria-selected="activeTab === index" :aria-controls="`${panelId}-panel-${index}`" :tabindex="activeTab === index ? 0 : -1"
              @click="selectTab(index)" @keydown="tabKeydown($event, index)">{{ tab }}</button>
        </div>
        <div ref="pages" class="preferences-pages" @scroll="syncTab">
          <div :id="`${panelId}-panel-0`" role="tabpanel" :aria-labelledby="`${panelId}-tab-0`" :inert="activeTab !== 0" class="preferences-page">
          <section class="gf-modal__section">
            <div class="gf-modal__copy">
              <label class="gf-modal__label" for="mode-setting-input">
                {{ t("navbar.displayMode") }}
              </label>
              <p class="gf-modal__help">{{ t("navbar.displayModeDesc") }}</p>
            </div>
            <input
                id="mode-setting-input"
                v-model="localMode"
                placeholder="nsfw"
                maxlength="32"
                class="gf-input"
            />
          </section>

          <section class="gf-modal__section gf-modal__section--inline">
            <div class="gf-modal__copy">
              <label class="gf-modal__label" for="quick-access-toggle">
                {{ t("navbar.quickAccess") }}
              </label>
              <p class="gf-modal__help">{{ t("navbar.quickAccessDesc") }}</p>
            </div>
            <button
                id="quick-access-toggle"
                type="button"
                class="preferences-toggle"
                :class="{ 'preferences-toggle--on': showQuickAccessLocal }"
                :aria-pressed="showQuickAccessLocal"
                :aria-label="t('navbar.quickAccess')"
                @click="showQuickAccessLocal = !showQuickAccessLocal"
            >
              <span></span>
            </button>
          </section>

          <HeroPreferencesEditor ref="heroEditor" :active="activeTab === 0" />

          </div>
          <div :id="`${panelId}-panel-1`" role="tabpanel" :aria-labelledby="`${panelId}-tab-1`" :inert="activeTab !== 1" class="preferences-page">
            <ResourceRoutingPreferencesEditor ref="routingEditor" />
          </div>
        </div>
      </div>
    </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, onUnmounted, ref, useId, watch } from 'vue'
import HeroPreferencesEditor from './HeroPreferencesEditor.vue'
import ResourceRoutingPreferencesEditor from './ResourceRoutingPreferencesEditor.vue'
import { i18n } from '@/main'
import { useReducedMotion } from '@/composables/useReducedMotion'
import {
  readShowQuickAccess,
  writeShowQuickAccess,
} from '@/utils/navHeaderSettings'

const { t } = i18n.global

const props = withDefaults(defineProps<{
  show: boolean
  mode: string
  returnFocus?: HTMLElement | null
}>(), { returnFocus: null })

const emit = defineEmits<{
  (e: 'save', value: string): void
  (e: 'cancel'): void
}>()

const panelId = useId()
const reducedMotion = useReducedMotion()
const backdrop = ref<HTMLElement | null>(null)
let opener: HTMLElement | null = null
let finishTransition: (() => void) | undefined
let cancelTransition: (() => void) | undefined

// Await only this surface's CSS transitions, never an Editor's work. Observe
// Vue's to-class rather than racing its frames; no delayed-unmount timer owns show.
function trackTransition(element: Element, done: () => void) {
  cancelTransition?.()
  let active = true
  const observer = new MutationObserver(() => {
    if (!element.matches('.preferences-modal-enter-to, .preferences-modal-leave-to')) return
    observer.disconnect()
    const panel = element.querySelector('.gf-preferences-modal')
    const motion = element.getAnimations({ subtree: true }).filter(animation => {
      const target = (animation.effect as KeyframeEffect | null)?.target
      return animation instanceof CSSTransition && (target === element || target === panel)
        && ['opacity', 'transform'].includes(animation.transitionProperty)
    })
    void Promise.all(motion.map(animation => animation.finished.catch(() => {}))).then(finish)
  })
  const cancel = () => {
    active = false
    observer.disconnect()
    if (cancelTransition === cancel) { cancelTransition = undefined; finishTransition = undefined }
  }
  const finish = () => { if (active) { cancel(); done() } }
  cancelTransition = cancel
  finishTransition = finish
  if (reducedMotion.value) { finish(); return }
  observer.observe(element, { attributes: true, attributeFilter: ['class'] })
}
watch(reducedMotion, reduced => { if (reduced) finishTransition?.() }, { flush: 'sync' })

// Vue owns the leave lifetime. The closed surface must stop accepting input
// immediately, including while its final opacity frame is still on screen.
function deactivateLeaving(element: Element) {
  const root = element as HTMLElement
  if (root.contains(document.activeElement)) {
    if (opener?.isConnected && !opener.closest('[inert]')) opener.focus({ preventScroll: true })
    if (root.contains(document.activeElement)) (document.activeElement as HTMLElement | null)?.blur()
  }
  root.inert = true
  root.setAttribute('aria-hidden', 'true')
  clearTimeout(tabScrollTimer)
}
function cancel() {
  if (props.show) emit('cancel')
}
onBeforeUnmount(() => {
  cancelTransition?.()
  if (backdrop.value) deactivateLeaving(backdrop.value)
})
const activeTab = ref(0)
const pages = ref<HTMLElement | null>(null)
const tabs = computed(() => [t('navbar.homePreferences'), t('resourceRouting.title')])
function selectTab(index: number) {
  activeTab.value = index
  const element = pages.value
  if (element) element.scrollTo({ left: index * element.clientWidth, behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'instant' : 'smooth' })
}
let tabScrollTimer: ReturnType<typeof setTimeout> | undefined
function syncTab() {
  clearTimeout(tabScrollTimer)
  tabScrollTimer = setTimeout(() => {
    if (pages.value?.clientWidth) activeTab.value = Math.round(pages.value.scrollLeft / pages.value.clientWidth)
  }, 100)
}
onUnmounted(() => clearTimeout(tabScrollTimer))
function tabKeydown(event: KeyboardEvent, index: number) {
  const count = tabs.value.length
  const next = event.key === 'Home' ? 0 : event.key === 'End' ? count - 1
    : event.key === 'ArrowRight' ? (index + 1) % count : event.key === 'ArrowLeft' ? (index + count - 1) % count : undefined
  if (next === undefined) return
  event.preventDefault()
  selectTab(next)
  document.getElementById(`${panelId}-tab-${next}`)?.focus()
}
const localMode = ref('')
const routingEditor = ref<InstanceType<typeof ResourceRoutingPreferencesEditor> | null>(null)
const heroEditor = ref<InstanceType<typeof HeroPreferencesEditor> | null>(null)
const showQuickAccessLocal = ref(true)
function syncCustomBgState() {
  showQuickAccessLocal.value = readShowQuickAccess()
}

watch(
    () => props.mode,
    value => {
      localMode.value = value
    },
    { immediate: true }
)

watch(
    () => props.show,
    visible => {
      if (visible) {
        opener = props.returnFocus ?? (document.activeElement instanceof HTMLElement ? document.activeElement : null)
        activeTab.value = 0
        localMode.value = props.mode
        nextTick(() => { pages.value?.scrollTo({ left: 0, behavior: 'instant' }) })
        syncCustomBgState()
      }
    }
)

onMounted(() => {
  syncCustomBgState()
})

const save = async () => {
  if (!props.show) return
  if (heroEditor.value && !heroEditor.value.validate()) { selectTab(0); return }
  localMode.value = localMode.value.trim().slice(0, 32)

  if (heroEditor.value && !(await heroEditor.value.save())) { selectTab(0); return }

  writeShowQuickAccess(showQuickAccessLocal.value)
  routingEditor.value?.save()
  emit('save', localMode.value)
}
</script>

<style scoped>
.preferences-pages { display: flex; height: min(34rem, calc(100dvh - 11rem)); min-height: 0; overflow-x: auto; overflow-y: hidden; scroll-snap-type: x mandatory; scrollbar-width: none; overscroll-behavior-x: contain; }
.preferences-pages::-webkit-scrollbar { display: none; }
.preferences-page { flex: 0 0 100%; min-width: 0; padding: .8rem 1.3rem 1.2rem; overflow-y: auto; scroll-snap-align: start; overscroll-behavior-y: contain; }
</style>
