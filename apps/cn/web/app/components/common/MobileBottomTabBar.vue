<template>
  <div class="mobile-bottom-tabs-root">
    <nav
      class="mobile-bottom-tabs"
      :class="{ 'mobile-bottom-tabs--visible': showBottomTabs }"
      :aria-label="t('navbar.expandNav')"
      :aria-hidden="!showBottomTabs"
    >
      <NuxtLink
        :to="localePath('/')"
        class="mobile-bottom-tabs__item"
        :class="{ 'mobile-bottom-tabs__item--active': isHomeActive }"
        :aria-label="t('sidebar.nav')"
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
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { PhBinoculars, PhGameController, PhGearSix, PhHouse } from '@phosphor-icons/vue'
import { useI18n } from 'vue-i18n'
import ModeSettingModal from '@/components/common/ModeSettingModal.vue'
import { readMode, subscribeModeChange, writeMode } from '@/utils/modeStorage'

const { t } = useI18n()
const route = useRoute()
const localePath = useLocalePath()
const isNarrowScreen = ref(false)
const isAwayFromTop = ref(false)
const showModeModal = ref(false)
const mode = ref('')
let mediaQuery: MediaQueryList | null = null
let stopModeSubscription: (() => void) | null = null
let scrollSyncTimer: ReturnType<typeof setInterval> | null = null

const normalizedPath = computed(() => route.path.replace(/^\/(zh|en)(?=\/|$)/, '') || '/')
const isHomeActive = computed(() => normalizedPath.value === '/')
const isGamesActive = computed(() => normalizedPath.value === '/games' || normalizedPath.value.startsWith('/games/'))
const isEcosystemActive = computed(() => normalizedPath.value === '/insights' || normalizedPath.value.startsWith('/insights/'))
const showBottomTabs = computed(() => isNarrowScreen.value && isAwayFromTop.value)

function updateScrollState() {
  isAwayFromTop.value = window.scrollY > 72
}

function updateScreenState() {
  isNarrowScreen.value = mediaQuery?.matches ?? window.innerWidth < 640
}

function saveMode(value: string) {
  showModeModal.value = false
  writeMode(value.trim().slice(0, 32))
}

function handleMediaChange() {
  updateScreenState()
}

onMounted(() => {
  mode.value = readMode()
  mediaQuery = window.matchMedia('(max-width: 639px)')
  updateScreenState()
  updateScrollState()

  mediaQuery.addEventListener('change', handleMediaChange)
  window.addEventListener('scroll', updateScrollState, { passive: true })
  document.addEventListener('scroll', updateScrollState, { passive: true, capture: true })
  window.addEventListener('resize', updateScreenState)
  scrollSyncTimer = setInterval(updateScrollState, 160)

  stopModeSubscription = subscribeModeChange(({ mode: nextMode }) => {
    mode.value = nextMode
  })
})

onUnmounted(() => {
  mediaQuery?.removeEventListener('change', handleMediaChange)
  window.removeEventListener('scroll', updateScrollState)
  document.removeEventListener('scroll', updateScrollState, { capture: true })
  window.removeEventListener('resize', updateScreenState)
  if (scrollSyncTimer) {
    clearInterval(scrollSyncTimer)
    scrollSyncTimer = null
  }
  stopModeSubscription?.()
})

watch(
  () => route.fullPath,
  () => {
    showModeModal.value = false
    if (import.meta.client) {
      updateScrollState()
    }
  }
)
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

@media (min-width: 640px) {
  .mobile-bottom-tabs {
    display: none;
  }
}
</style>
