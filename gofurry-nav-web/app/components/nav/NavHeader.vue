<template>
  <header
      class="relative flex h-90 w-full flex-col items-center justify-start overflow-hidden px-4 pt-6 shadow-sm md:h-[100vh] md:px-6 md:pt-0"
  >
    <div
        v-if="bgImage"
        class="absolute inset-0 bg-cover bg-center transition-all duration-700"
        :style="{ backgroundImage: `url(${bgImage})` }"
    ></div>
    <div class="relative z-30 flex w-full justify-center md:absolute md:left-1/2 md:top-[76px] md:w-full md:max-w-[56rem] md:-translate-x-1/2 lg:top-[88px]">
      <SearchBox />
    </div>

    <div
        v-if="showQuickAccess"
        class="relative z-10 hidden w-full justify-center md:absolute md:left-1/2 md:top-[248px] md:flex md:w-full md:max-w-[56rem] md:-translate-x-1/2 lg:top-[268px]"
    >
      <NavQuickAccess
        :recent-sites="recentSites"
        :custom-sites="customSites"
        @visit-recent="visitRecentSite"
        @visit-custom="visitCustomSite"
        @manage="showCustomSitesModal = true"
      />
    </div>

    <div class="pointer-events-none absolute bottom-6 left-1/2 z-10 hidden -translate-x-1/2 items-center gap-2 text-white/85 md:flex md:flex-col">
      <span class="text-xs font-medium uppercase tracking-[0.28em]">
        {{ t('navHeader.scrollHint') }}
      </span>
      <div class="flex h-6 w-4 items-start justify-center rounded-full border border-white/35 bg-black/10 p-1 backdrop-blur-sm">
        <span class="h-2 w-2 animate-bounce rounded-full bg-white/90"></span>
      </div>
    </div>

    <transition name="quick-modal">
      <div
        v-if="showCustomSitesModal"
        class="fixed inset-0 z-60 hidden items-center justify-center bg-slate-950/52 px-4 backdrop-blur-sm md:flex"
        @click.self="closeCustomSitesModal"
      >
        <div class="quick-modal-panel">
          <div class="quick-modal-header">
            <div>
              <h3 class="quick-modal-title">{{ t('customSites.manageTitle') }}</h3>
              <p class="quick-modal-desc">{{ t('customSites.manageDescription', { count: customSites.length, max: MAX_CUSTOM_SITES }) }}</p>
            </div>
            <button class="quick-modal-close" type="button" :aria-label="t('common.cancel')" @click="closeCustomSitesModal">
              <span aria-hidden="true">×</span>
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

          <div class="quick-modal-list">
            <div v-if="customSites.length" class="quick-modal-items">
              <div
                v-for="item in customSites"
                :key="item.id"
                class="quick-modal-item"
                :class="{ 'quick-modal-item-dragging': draggingCustomSiteId === item.id }"
                draggable="true"
                @dragstart="handleCustomSiteDragStart(item.id)"
                @dragover.prevent
                @drop="handleCustomSiteDrop(item.id)"
                @dragend="handleCustomSiteDragEnd"
              >
                <div class="quick-modal-grip" aria-hidden="true">
                  <span></span>
                  <span></span>
                </div>
                <div class="min-w-0 flex-1">
                  <p class="quick-modal-item-name">{{ item.name }}</p>
                  <p class="quick-modal-item-url">{{ item.url }}</p>
                </div>
                <button class="quick-modal-delete" type="button" :aria-label="t('customSites.deleteSite')" @click="deleteCustomSite(item.id)">
                  <span aria-hidden="true">×</span>
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
import { onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import SearchBox from './SearchBox.vue'
import NavQuickAccess from './NavQuickAccess.vue'
import { getNavHomeBackgrounds } from '~/services/nav'
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
  CUSTOM_NAV_HEADER_BG_EVENT,
  loadRandomCustomNavHeaderBackground,
} from '@/utils/customNavHeaderBackground'
import {
  readShowQuickAccess,
  subscribeNavHeaderSettingsChange,
} from '@/utils/navHeaderSettings'

const props = defineProps<{
  desktopBgUrl?: string | null
  mobileBgUrl?: string | null
}>()

const { t } = useI18n()
const bgImage = ref<string | null>(null)
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

let fallbackBackgroundUpdater: (() => void) | null = null
let customBgObjectUrl: string | null = null
let stopNavHeaderSettingsSubscription: (() => void) | null = null

function revokeCustomBackgroundUrl() {
  if (customBgObjectUrl) {
    URL.revokeObjectURL(customBgObjectUrl)
    customBgObjectUrl = null
  }
}

async function applyBackground() {
  revokeCustomBackgroundUrl()

  try {
    const customBackground = await loadRandomCustomNavHeaderBackground()
    if (customBackground) {
      bgImage.value = customBackground
      customBgObjectUrl = customBackground
      return
    }
  } catch (error) {
    console.error('Load custom nav header background err:', error)
  }

  fallbackBackgroundUpdater?.()
}

function handleCustomBackgroundChange() {
  void applyBackground()
}

function handleResize() {
  if (customBgObjectUrl) {
    return
  }

  fallbackBackgroundUpdater?.()
}

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
  syncCustomSites()
}

function syncNavHeaderSettings() {
  showQuickAccess.value = readShowQuickAccess()
}

function visitRecentSite(site: RecentSiteItem) {
  window.open(site.url, '_blank')
}

function closeCustomSitesModal() {
  showCustomSitesModal.value = false
  customSiteError.value = ''
  customSiteForm.value = {
    name: '',
    url: '',
  }
}

function submitCustomSite() {
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
  removeCustomSite(id)
  syncCustomSites()
}

function handleCustomSiteDragStart(id: string) {
  draggingCustomSiteId.value = id
}

function handleCustomSiteDrop(targetId: string) {
  if (!draggingCustomSiteId.value || draggingCustomSiteId.value === targetId) {
    return
  }

  const nextSites = [...customSites.value]
  const fromIndex = nextSites.findIndex((item) => item.id === draggingCustomSiteId.value)
  const targetIndex = nextSites.findIndex((item) => item.id === targetId)
  if (fromIndex < 0 || targetIndex < 0) {
    draggingCustomSiteId.value = null
    return
  }

  const [moved] = nextSites.splice(fromIndex, 1)
  if (!moved) {
    draggingCustomSiteId.value = null
    return
  }
  nextSites.splice(targetIndex, 0, moved)
  reorderCustomSites(nextSites.map((item) => item.id))
  customSites.value = nextSites
  draggingCustomSiteId.value = null
}

function handleCustomSiteDragEnd() {
  draggingCustomSiteId.value = null
}

onMounted(async () => {
  try {
    fallbackBackgroundUpdater = () => {
      bgImage.value = window.innerWidth >= 768
        ? (props.desktopBgUrl ?? props.mobileBgUrl ?? null)
        : (props.mobileBgUrl ?? props.desktopBgUrl ?? null)
    }

    if (!props.desktopBgUrl && !props.mobileBgUrl) {
      const backgrounds = await getNavHomeBackgrounds()
      const resizedUrl = backgrounds.desktop
      const normalUrl = backgrounds.mobile

      fallbackBackgroundUpdater = () => {
        bgImage.value = window.innerWidth >= 768 ? resizedUrl : normalUrl
      }
    }

    await applyBackground()
    syncNavHeaderSettings()
    syncRecentSites()
    syncCustomSites()
    stopNavHeaderSettingsSubscription = subscribeNavHeaderSettingsChange(({ showQuickAccess: nextValue }) => {
      showQuickAccess.value = nextValue
    })
    window.addEventListener('resize', handleResize)
    window.addEventListener(CUSTOM_NAV_HEADER_BG_EVENT, handleCustomBackgroundChange)
    window.addEventListener(RECENT_SITES_EVENT, handleRecentSitesChange)
    window.addEventListener(CUSTOM_SITES_EVENT, handleCustomSitesChange)
    window.addEventListener('storage', handleRecentSitesChange)
    window.addEventListener('storage', handleCustomSitesChange)
  } catch (err) {
    console.error('Get background image URL err:', err)
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  window.removeEventListener(CUSTOM_NAV_HEADER_BG_EVENT, handleCustomBackgroundChange)
  window.removeEventListener(RECENT_SITES_EVENT, handleRecentSitesChange)
  window.removeEventListener(CUSTOM_SITES_EVENT, handleCustomSitesChange)
  window.removeEventListener('storage', handleRecentSitesChange)
  window.removeEventListener('storage', handleCustomSitesChange)
  stopNavHeaderSettingsSubscription?.()
  stopNavHeaderSettingsSubscription = null
  revokeCustomBackgroundUrl()
})
</script>

<style scoped>
header {
  transition: background 0.4s ease-in-out;
}

.quick-modal-panel {
  display: flex;
  max-height: min(36rem, calc(100vh - 4rem));
  width: min(100%, 36rem);
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 1rem;
  background: rgba(12, 17, 21, 0.84);
  box-shadow: 0 24px 60px rgba(7, 12, 16, 0.28);
  backdrop-filter: blur(22px);
}

.quick-modal-header,
.quick-modal-form,
.quick-modal-list {
  padding-inline: 1rem;
}

.quick-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding-top: 1rem;
}

.quick-modal-title {
  color: rgba(255, 248, 241, 0.96);
  font-size: 0.96rem;
  font-weight: 650;
}

.quick-modal-desc {
  color: rgba(233, 241, 246, 0.68);
  font-size: 0.76rem;
  line-height: 1.3;
  margin-top: 0.2rem;
}

.quick-modal-close {
  display: grid;
  place-items: center;
  width: 1.8rem;
  height: 1.8rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 999px;
  color: rgba(255, 248, 241, 0.88);
  background: rgba(255, 255, 255, 0.03);
}

.quick-modal-form {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.4fr) auto;
  gap: 0.55rem;
  align-items: center;
  padding-top: 0.85rem;
}

.quick-modal-inline-field {
  min-width: 0;
}

.quick-modal-inline-field input {
  width: 100%;
  height: 2.45rem;
  padding-inline: 0.8rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.8rem;
  color: rgba(255, 248, 241, 0.96);
  background: rgba(255, 255, 255, 0.045);
  outline: none;
}

.quick-modal-inline-field input:focus {
  border-color: rgba(151, 226, 255, 0.34);
}

.quick-modal-inline-action {
  display: flex;
}

.quick-modal-error {
  color: rgba(255, 173, 173, 0.94);
  font-size: 0.76rem;
  padding: 0.55rem 1rem 0;
}

.quick-modal-primary,
.quick-modal-delete {
  border-radius: 0.8rem;
  font-size: 0.82rem;
  transition: background 180ms ease, border-color 180ms ease, opacity 180ms ease;
}

.quick-modal-primary {
  height: 2.45rem;
  padding-inline: 0.95rem;
}

.quick-modal-primary {
  border: 1px solid rgba(151, 226, 255, 0.18);
  color: rgba(8, 15, 18, 0.94);
  background: rgba(151, 226, 255, 0.9);
}

.quick-modal-primary:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.quick-modal-list {
  padding-top: 0.85rem;
  padding-bottom: 1rem;
  overflow: auto;
}

.quick-modal-items {
  display: grid;
  gap: 0.5rem;
}

.quick-modal-item {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  min-height: 3rem;
  padding: 0.55rem 0.7rem;
  border-radius: 0.85rem;
  background: rgba(255, 255, 255, 0.045);
  cursor: grab;
}

.quick-modal-item-dragging {
  opacity: 0.48;
}

.quick-modal-grip {
  display: grid;
  gap: 0.18rem;
  flex-shrink: 0;
}

.quick-modal-grip span {
  width: 0.22rem;
  height: 0.22rem;
  border-radius: 999px;
  background: rgba(255, 248, 241, 0.42);
  box-shadow: 0 0.38rem 0 rgba(255, 248, 241, 0.42);
}

.quick-modal-item-name {
  color: rgba(255, 248, 241, 0.96);
  font-size: 0.82rem;
  font-weight: 600;
}

.quick-modal-item-url {
  margin-top: 0.12rem;
  color: rgba(233, 241, 246, 0.62);
  font-size: 0.72rem;
  word-break: break-all;
}

.quick-modal-delete {
  flex-shrink: 0;
  display: grid;
  place-items: center;
  width: 1.55rem;
  height: 1.55rem;
  border: 0;
  border-radius: 999px;
  color: rgba(255, 235, 235, 0.96);
  background: rgba(242, 78, 78, 0.95);
}

.quick-modal-empty {
  color: rgba(233, 241, 246, 0.62);
  font-size: 0.8rem;
  line-height: 1.5;
}

.quick-modal-enter-active,
.quick-modal-leave-active {
  transition: opacity 180ms ease;
}

.quick-modal-enter-active .quick-modal-panel,
.quick-modal-leave-active .quick-modal-panel {
  transition: transform 220ms cubic-bezier(0.22, 1, 0.36, 1), opacity 180ms ease;
}

.quick-modal-enter-from,
.quick-modal-leave-to {
  opacity: 0;
}

.quick-modal-enter-from .quick-modal-panel,
.quick-modal-leave-to .quick-modal-panel {
  opacity: 0;
  transform: translateY(8px) scale(0.985);
}
</style>
