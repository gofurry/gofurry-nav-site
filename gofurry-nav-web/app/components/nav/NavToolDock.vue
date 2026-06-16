<template>
  <div class="nav-tool-dock">
    <Transition name="nav-tool-panel-transition" mode="out-in">
      <section
        v-if="activePanel === 'search'"
        key="search"
        class="nav-tool-panel nav-tool-search"
        :aria-label="t('nav.tools.search')"
      >
        <div class="nav-tool-search__input">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="m21 21-4.4-4.4M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15Z" />
          </svg>
          <input
            ref="searchInputRef"
            v-model.trim="keyword"
            type="search"
            :placeholder="t('nav.tools.searchPlaceholder')"
            @keydown.esc.prevent.stop="closePanel"
          />
        </div>

        <Transition name="nav-tool-list-transition" mode="out-in">
          <div v-if="keyword && results.length" key="results" class="nav-tool-results">
            <button
              v-for="item in results"
              :key="item.id"
              type="button"
              @click="openSite(item)"
            >
              <img
                :src="siteLogoSrc(item)"
                :alt="item.name"
              />
              <span>
                <strong>{{ item.name }}</strong>
                <small>{{ item.info }}</small>
              </span>
            </button>
          </div>

          <div v-else-if="keyword" key="empty" class="nav-tool-empty">
            {{ t('nav.tools.noSearchResults') }}
          </div>
        </Transition>
      </section>

      <RagPromptPanel
        v-else-if="activePanel === 'ask'"
        key="ask"
        :title="t('nav.tools.askTitle')"
        :description="t('nav.tools.askDescription')"
        :placeholder="t('nav.tools.askPlaceholder')"
        :submit-label="t('nav.tools.askSubmit')"
        :templates="navPromptTemplates"
        @ask="openArchivePrompt"
      />
    </Transition>

    <nav class="nav-tool-rail" :aria-label="t('nav.tools.label')">
      <div class="nav-tool-rail__primary">
        <button
          v-for="tool in tools"
          :key="tool.key"
          type="button"
          class="nav-tool-button"
          :class="[{ active: activePanel === tool.panel }, `nav-tool-button--${tool.key}`]"
          :title="tool.label"
          :aria-label="tool.label"
          @click="tool.action"
        >
          <span class="nav-tool-icon-stack" aria-hidden="true">
            <img class="nav-tool-icon" :src="tool.icon" alt="" />
          </span>
        </button>
      </div>

      <a
        class="nav-tool-feedback"
        href="https://github.com/gofurry/gofurry-nav-site/issues"
        target="_blank"
        rel="noopener noreferrer"
        :title="t('nav.tools.feedback')"
        :aria-label="t('nav.tools.feedback')"
      >
        <span class="nav-tool-icon-stack" aria-hidden="true">
          <img class="nav-tool-icon" :src="feedbackIconSrc" alt="" />
        </span>
        <span>{{ t('nav.tools.feedbackShort') }}</span>
      </a>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Site } from '@/types/nav'
import { recordRecentSite, toExternalUrl } from '@/utils/recentSites'
import { useThemeStore } from '@/stores/theme'
import { getNavSiteDirectory } from '~/services/nav'
import { readDisplayMode, subscribeModeChange, type DisplayMode } from '@/utils/modeStorage'
import RagPromptPanel from '@/components/common/RagPromptPanel.vue'
import askIconDark from '@/assets/svgs/ai-duotone-dark.svg'
import askIconLight from '@/assets/svgs/ai-duotone.svg'
import feedbackIconDark from '@/assets/svgs/ai-note-alt-1-duotone-dark.svg'
import feedbackIconLight from '@/assets/svgs/ai-note-alt-1-duotone.svg'
import searchIconDark from '@/assets/svgs/search-white.svg'
import searchIconLight from '@/assets/svgs/search.svg'

type RagPromptTemplate = {
  id: string
  title: string
  description: string
  prompt: string
}

const router = useRouter()
const { t, locale } = useI18n()
const themeStore = useThemeStore()
const logoPrefix = import.meta.env.VITE_SITE_LOGO_PREFIX_URL || ''
const defaultLogo = 'defaultLogo.svg'

const activePanel = ref<'search' | 'ask' | null>(null)
const keyword = ref('')
const searchInputRef = ref<HTMLInputElement | null>(null)
const isDarkTheme = computed(() => themeStore.theme === 'dark')
const displayMode = ref<DisplayMode>('sfw')
const directoryItems = ref<Site[]>([])
const directoryLoaded = ref(false)
let stopModeSubscription: (() => void) | null = null
const askIconSrc = computed(() => isDarkTheme.value ? askIconDark : askIconLight)
const feedbackIconSrc = computed(() => isDarkTheme.value ? feedbackIconDark : feedbackIconLight)
const searchIconSrc = computed(() => isDarkTheme.value ? searchIconDark : searchIconLight)

const results = computed(() => {
  const query = keyword.value.trim().toLowerCase()
  if (!query) {
    return []
  }
  return directoryItems.value
    .filter(item => displayMode.value === 'nsfw' || String(item.nsfw) !== '1')
    .filter(item => item.name?.toLowerCase().includes(query) || item.info?.toLowerCase().includes(query))
    .slice(0, 8)
})

const navPromptTemplates = computed<RagPromptTemplate[]>(() => [
  {
    id: 'discover',
    title: t('nav.tools.prompts.discover.title'),
    description: t('nav.tools.prompts.discover.description'),
    prompt: t('nav.tools.prompts.discover.prompt'),
  },
  {
    id: 'publish',
    title: t('nav.tools.prompts.publish.title'),
    description: t('nav.tools.prompts.publish.description'),
    prompt: t('nav.tools.prompts.publish.prompt'),
  },
  {
    id: 'alternative',
    title: t('nav.tools.prompts.alternative.title'),
    description: t('nav.tools.prompts.alternative.description'),
    prompt: t('nav.tools.prompts.alternative.prompt'),
  },
])

const tools = computed(() => [
  {
    key: 'search',
    label: t('nav.tools.search'),
    panel: 'search' as const,
    icon: searchIconSrc.value,
    action: () => {
      activePanel.value = activePanel.value === 'search' ? null : 'search'
    },
  },
  {
    key: 'ask',
    label: t('nav.tools.ask'),
    panel: 'ask' as const,
    icon: askIconSrc.value,
    action: () => {
      activePanel.value = activePanel.value === 'ask' ? null : 'ask'
    },
  },
])

watch(activePanel, async (panel) => {
  if (panel !== 'search') {
    return
  }

  await ensureDirectoryLoaded()
  await nextTick()
  searchInputRef.value?.focus()
})

watch(
  () => locale.value,
  async () => {
    directoryLoaded.value = false
    directoryItems.value = []
    if (activePanel.value === 'search') {
      await ensureDirectoryLoaded()
    }
  }
)

function closePanel() {
  activePanel.value = null
  keyword.value = ''
}

async function ensureDirectoryLoaded() {
  if (directoryLoaded.value) {
    return
  }

  try {
    directoryItems.value = await getNavSiteDirectory(locale.value === 'en' ? 'en' : 'zh')
    directoryLoaded.value = true
  } catch (error) {
    console.error('Failed to load site directory:', error)
    directoryItems.value = []
  }
}

function joinAssetUrl(prefix: string, path: string) {
  if (!prefix) {
    return path
  }

  return `${prefix.replace(/\/+$/, '')}/${path.replace(/^\/+/, '')}`
}

function withAssetVersion(url: string, version?: string | null) {
  const normalizedVersion = (version || '').trim()
  if (!normalizedVersion) {
    return url
  }
  const separator = url.includes('?') ? '&' : '?'
  return `${url}${separator}v=${encodeURIComponent(normalizedVersion)}`
}

function siteLogoSrc(item: Site) {
  const iconPath = item.icon || defaultLogo
  const assetURL = joinAssetUrl(logoPrefix, iconPath)
  if (!item.icon) {
    return assetURL
  }
  return withAssetVersion(assetURL, item.update_time)
}

function domainList(item: Site) {
  if (Array.isArray(item.domain)) {
    return item.domain
  }
  if (typeof item.domain !== 'string' || !item.domain) {
    return []
  }
  try {
    const parsed = JSON.parse(item.domain)
    return Array.isArray(parsed?.domain) ? parsed.domain : [item.domain]
  } catch {
    return [item.domain]
  }
}

function openSite(item: Site) {
  const url = toExternalUrl(domainList(item)[0])
  if (!url) {
    return
  }
  recordRecentSite({
    id: item.id,
    name: item.name,
    url,
  })
  window.open(url, '_blank')
  closePanel()
}

function openArchivePrompt(prompt: string) {
  closePanel()
  router.push({
    path: '/archive',
    query: {
      q: prompt,
      scene: 'nav',
    },
  })
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    closePanel()
  }
}

onMounted(() => {
  displayMode.value = readDisplayMode()
  stopModeSubscription = subscribeModeChange(({ displayMode: nextMode }) => {
    displayMode.value = nextMode
  })
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  stopModeSubscription?.()
  window.removeEventListener('keydown', handleKeydown)
})
</script>
