<template>
  <header
      class="nav-header"
  >
    <h1 class="sr-only">{{ homeHeading }}</h1>
    <div
        v-if="bgImage"
        class="nav-header__background"
        :style="{ backgroundImage: `url(${bgImage})` }"
    ></div>
    <div class="nav-header__search">
      <SearchBox />
    </div>

    <div
        v-if="showQuickAccess"
        class="nav-header__quick-access"
    >
      <NavQuickAccess
        :recent-sites="recentSites"
        :custom-sites="customSites"
        @visit-recent="visitRecentSite"
        @visit-custom="visitCustomSite"
        @manage="showCustomSitesModal = true"
      />
    </div>

    <div class="nav-header__scroll-hint">
      <span class="nav-header__scroll-label">
        {{ t('navHeader.scrollHint') }}
      </span>
      <div class="nav-header__scroll-frame">
        <span class="nav-header__scroll-dot"></span>
      </div>
    </div>

    <transition name="quick-modal">
      <div
        v-if="showCustomSitesModal"
        class="quick-modal-backdrop"
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
import { computed, onMounted, onUnmounted, ref } from 'vue'
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

const { locale, t } = useI18n()
const homeHeading = computed(() => locale.value === 'en'
  ? 'GoFurry Navigation - Discover furry communities, art, fiction, games, tools, and site monitoring'
  : 'GoFurry 兽人控导航站 - 发现兽人社区、艺术、小说、游戏、工具与站点监测资源'
)
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
