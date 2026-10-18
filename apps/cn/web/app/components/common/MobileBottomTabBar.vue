<template>
  <div class="mobile-bottom-tabs-root">
    <nav
      class="mobile-bottom-tabs"
      :class="{ 'mobile-bottom-tabs--visible': showBottomTabs }"
      :aria-label="t('navbar.expandNav')"
      :aria-hidden="!showBottomTabs"
      :inert="!showBottomTabs"
    >
      <NuxtLink
        :to="localePath('/')"
        class="mobile-bottom-tabs__item"
        :class="{ 'mobile-bottom-tabs__item--active': isHomeActive }"
        :aria-label="t('sidebar.nav')"
        :aria-current="isHomeActive ? 'page' : undefined"
        :tabindex="showBottomTabs ? 0 : -1"
      >
        <PhHouse :size="17" weight="regular" aria-hidden="true" />
        <span class="mobile-bottom-tabs__label">{{ t('sidebar.nav') }}</span>
      </NuxtLink>

      <NuxtLink
        :to="localePath('/games')"
        class="mobile-bottom-tabs__item"
        :class="{ 'mobile-bottom-tabs__item--active': isGamesActive }"
        :aria-label="t('sidebar.games')"
        :aria-current="isGamesActive ? 'page' : undefined"
        :tabindex="showBottomTabs ? 0 : -1"
      >
        <PhGameController :size="17" weight="regular" aria-hidden="true" />
        <span class="mobile-bottom-tabs__label">{{ t('sidebar.games') }}</span>
      </NuxtLink>

      <NuxtLink
        :to="localePath('/insights')"
        class="mobile-bottom-tabs__item"
        :class="{ 'mobile-bottom-tabs__item--active': isEcosystemActive }"
        :aria-label="t('sidebar.ecosystemShort')"
        :aria-current="isEcosystemActive ? 'page' : undefined"
        :tabindex="showBottomTabs ? 0 : -1"
      >
        <PhBinoculars :size="17" weight="regular" aria-hidden="true" />
        <span class="mobile-bottom-tabs__label">{{ t('sidebar.ecosystemShort') }}</span>
      </NuxtLink>

      <button
        type="button"
        class="mobile-bottom-tabs__item"
        :class="{ 'mobile-bottom-tabs__item--active': showModeModal }"
        :aria-label="t('navbar.mode')"
        :tabindex="showBottomTabs ? 0 : -1"
        @click="showModeModal = true"
      >
        <PhGearSix :size="17" weight="regular" aria-hidden="true" />
        <span class="mobile-bottom-tabs__label">{{ t('navbar.mode') }}</span>
      </button>
    </nav>

    <ModeSettingModal
      :show="showModeModal"
      :mode="mode"
      @cancel="showModeModal = false"
      @save="saveMode"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { PhBinoculars, PhGameController, PhGearSix, PhHouse } from '@phosphor-icons/vue'
import { useI18n } from 'vue-i18n'
import ModeSettingModal from '@/components/common/ModeSettingModal.vue'
import { readMode, subscribeModeChange, writeMode } from '@/utils/modeStorage'

const props = defineProps<{ navigationBusy?: boolean }>()
const emit = defineEmits<{ handoff: [value: boolean] }>()
const { t } = useI18n()
const route = useRoute()
const localePath = useLocalePath()
const isNarrowScreen = ref(false)
const isAwayFromTop = ref(false)
const showModeModal = ref(false)
const mode = ref('')
let mediaQuery: MediaQueryList | null = null
let stopModeSubscription: (() => void) | null = null
let frame = 0
let mounted = false
let geometryObserver: ResizeObserver | null = null

const normalizedPath = computed(() => route.path.replace(/^\/(zh|en)(?=\/|$)/, '') || '/')
const isHomeActive = computed(() => normalizedPath.value === '/')
const isGamesActive = computed(() => normalizedPath.value === '/games' || normalizedPath.value.startsWith('/games/'))
const isEcosystemActive = computed(() => normalizedPath.value === '/insights' || normalizedPath.value.startsWith('/insights/'))
const showBottomTabs = computed(() => isNarrowScreen.value && isAwayFromTop.value)

// Read structural signals once per frame. A completed handoff only resets near
// the top, not when scroll direction reverses or the Hero re-enters the viewport.
function updateScrollState() {
  frame = 0
  if (!mounted) return
  const narrow = mediaQuery?.matches ?? window.innerWidth < 768
  const y = window.scrollY
  const header = document.querySelector<HTMLElement>('.gf-nav')
  const hero = isHomeActive.value ? document.querySelector<HTMLElement>('.nav-header') : null
  const headerEnd = header ? header.getBoundingClientRect().bottom + y : 76
  const heroEnd = hero ? hero.getBoundingClientRect().bottom + y : headerEnd
  const start = hero ? Math.max(headerEnd + 8, heroEnd - innerHeight / 2) : headerEnd + 8
  isNarrowScreen.value = narrow
  if (!narrow) {
    isAwayFromTop.value = false
    if (showModeModal.value) {
      showModeModal.value = false
      void nextTick(() => {
        if (mounted) document.querySelector<HTMLElement>('.gf-nav__brand')?.focus({ preventScroll: true })
      })
    }
  }
  else if (!props.navigationBusy && !showModeModal.value) {
    if (y <= 24) isAwayFromTop.value = false
    else if (y >= start) isAwayFromTop.value = true
  }
}
function scheduleSync() {
  if (mounted && !frame) frame = requestAnimationFrame(updateScrollState)
}
function observeGeometry() {
  geometryObserver?.disconnect()
  for (const element of [document.body, document.querySelector('.gf-nav'), document.querySelector('.nav-header')]) {
    if (element) geometryObserver?.observe(element)
  }
  scheduleSync()
}

function saveMode(value: string) {
  showModeModal.value = false
  writeMode(value.trim().slice(0, 32))
}

onMounted(() => {
  mounted = true
  mode.value = readMode()
  mediaQuery = window.matchMedia('(max-width: 767px)')
  mediaQuery.addEventListener('change', scheduleSync)
  window.addEventListener('scroll', scheduleSync, { passive: true })
  window.addEventListener('resize', scheduleSync)
  window.addEventListener('pageshow', scheduleSync)
  geometryObserver = new ResizeObserver(scheduleSync)
  observeGeometry()
  stopModeSubscription = subscribeModeChange(({ mode: nextMode }) => { mode.value = nextMode })
})
onUnmounted(() => {
  mounted = false
  mediaQuery?.removeEventListener('change', scheduleSync)
  window.removeEventListener('scroll', scheduleSync)
  window.removeEventListener('resize', scheduleSync)
  window.removeEventListener('pageshow', scheduleSync)
  geometryObserver?.disconnect()
  if (frame) cancelAnimationFrame(frame)
  stopModeSubscription?.()
})
watch(showBottomTabs, async visible => {
  const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null
  const moveFocus = previous?.closest(visible ? '.gf-nav' : '.mobile-bottom-tabs')
  emit('handoff', visible)
  await nextTick()
  if (mounted && moveFocus) document.querySelector<HTMLElement>(visible ? '.mobile-bottom-tabs__item' : '.gf-nav__brand')?.focus({ preventScroll: true })
})
watch([() => props.navigationBusy, showModeModal], scheduleSync)
watch(() => route.fullPath, () => {
  showModeModal.value = false
  scheduleSync()
})
// Same-page filters/hash navigation must not reset an already completed handoff.
watch(() => route.path, async () => {
  isAwayFromTop.value = false
  await nextTick()
  if (mounted) observeGeometry()
})
</script>

<style scoped>
.mobile-bottom-tabs {
  position: fixed;
  left: 50%;
  bottom: calc(1.125rem + env(safe-area-inset-bottom));
  z-index: 85;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 5px;
  pointer-events: none;
  transform: translate(-50%, 14px) scale(0.96);
}

.mobile-bottom-tabs--visible {
  pointer-events: auto;
  transform: translate(-50%, 0) scale(1);
}

.mobile-bottom-tabs__item {
  display: grid;
  width: 32px;
  height: 30px;
  place-items: center;
}

.mobile-bottom-tabs__item svg {
  width: 17px;
  height: 17px;
}

.mobile-bottom-tabs__label {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
}

@media (min-width: 768px) {
  .mobile-bottom-tabs {
    display: none;
  }
}
</style>
