<template>
  <div class="fixed right-4 top-24 z-40 hidden items-start gap-2 lg:flex">
    <section
      v-if="activePanel === 'search'"
      class="nav-tool-panel nav-tool-search"
      :aria-label="t('nav.tools.search')"
    >
      <div class="nav-tool-search__input">
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="m21 21-4.4-4.4M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15Z" />
        </svg>
        <input
          v-model.trim="keyword"
          type="search"
          :placeholder="t('nav.tools.searchPlaceholder')"
          @keydown.esc="closePanel"
        />
      </div>

      <div v-if="keyword && results.length" class="nav-tool-results">
        <button
          v-for="item in results"
          :key="item.id"
          type="button"
          @click="openSite(item)"
        >
          <img
            :src="`${logoPrefix ? `${logoPrefix}/` : ''}${item.icon || defaultLogo}`"
            alt=""
          />
          <span>
            <strong>{{ item.name }}</strong>
            <small>{{ item.info }}</small>
          </span>
        </button>
      </div>

      <div v-else-if="keyword" class="nav-tool-empty">
        {{ t('nav.tools.noSearchResults') }}
      </div>
    </section>

    <RagPromptPanel
      v-if="activePanel === 'ask'"
      :title="t('nav.tools.askTitle')"
      :description="t('nav.tools.askDescription')"
      :placeholder="t('nav.tools.askPlaceholder')"
      :submit-label="t('nav.tools.askSubmit')"
      :templates="navPromptTemplates"
      @ask="openArchivePrompt"
    />

    <nav class="flex flex-col gap-2" :aria-label="t('nav.tools.label')">
      <button
        v-for="tool in tools"
        :key="tool.key"
        type="button"
        class="nav-tool-button"
        :class="{ active: activePanel === tool.panel }"
        :title="tool.label"
        :aria-label="tool.label"
        @click="tool.action"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path v-if="tool.key === 'search'" d="m21 21-4.4-4.4M10.5 18a7.5 7.5 0 1 1 0-15 7.5 7.5 0 0 1 0 15Z" />
          <template v-else>
            <path d="M12 3.5 13.7 9l5.3 1.7-5.3 1.7L12 18l-1.7-5.6L5 10.7 10.3 9 12 3.5Z" />
            <path d="M18 15.5 18.8 18l2.2.8-2.2.7L18 22l-.8-2.5-2.2-.7 2.2-.8.8-2.5Z" />
          </template>
        </svg>
      </button>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Site } from '@/types/nav'
import { recordRecentSite, toExternalUrl } from '@/utils/recentSites'
import RagPromptPanel from '@/components/common/RagPromptPanel.vue'

type RagPromptTemplate = {
  id: string
  title: string
  description: string
  prompt: string
}

const props = defineProps<{
  items: Site[]
}>()

const router = useRouter()
const { t } = useI18n()
const logoPrefix = import.meta.env.VITE_SITE_LOGO_PREFIX_URL || ''
const defaultLogo = 'defaultLogo.svg'

const activePanel = ref<'search' | 'ask' | null>(null)
const keyword = ref('')

const results = computed(() => {
  const query = keyword.value.trim().toLowerCase()
  if (!query) {
    return []
  }
  return props.items
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
    action: () => {
      activePanel.value = activePanel.value === 'search' ? null : 'search'
    },
  },
  {
    key: 'ask',
    label: t('nav.tools.ask'),
    panel: 'ask' as const,
    action: () => {
      activePanel.value = activePanel.value === 'ask' ? null : 'ask'
    },
  },
])

function closePanel() {
  activePanel.value = null
  keyword.value = ''
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

onMounted(() => window.addEventListener('keydown', handleKeydown))
onUnmounted(() => window.removeEventListener('keydown', handleKeydown))
</script>

<style scoped>
.nav-tool-button {
  position: relative;
  display: grid;
  width: 2.75rem;
  height: 2.75rem;
  place-items: center;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.55);
  border-radius: 0.65rem;
  background: rgba(255, 255, 255, 0.7);
  box-shadow: 0 12px 32px rgba(76, 42, 18, 0.14);
  color: #334155;
  backdrop-filter: blur(18px);
  transition: background 180ms ease, border-color 180ms ease, color 180ms ease;
}

.nav-tool-button::before {
  content: '';
  position: absolute;
  inset: 0.25rem 0.5rem auto;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(251, 146, 60, 0.8), transparent);
}

.nav-tool-button:hover,
.nav-tool-button.active {
  border-color: rgba(253, 186, 116, 0.9);
  background: rgba(255, 255, 255, 0.9);
  color: #c2410c;
}

.nav-tool-button svg {
  width: 1.22rem;
  height: 1.22rem;
  fill: none;
  stroke: currentColor;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 1.9;
}

.nav-tool-panel {
  width: min(21rem, calc(100vw - 5rem));
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.56);
  border-radius: 0.75rem;
  background: rgba(255, 255, 255, 0.76);
  box-shadow: 0 18px 44px rgba(76, 42, 18, 0.16);
  color: #1f2937;
  backdrop-filter: blur(18px);
}

.nav-tool-search__input {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  padding: 0.72rem;
  border-bottom: 1px solid rgba(120, 113, 108, 0.14);
}

.nav-tool-search__input svg {
  width: 1rem;
  height: 1rem;
  flex: 0 0 auto;
  fill: none;
  stroke: #57534e;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 2;
}

.nav-tool-search__input input {
  min-width: 0;
  width: 100%;
  border: 0;
  background: transparent;
  font-size: 0.82rem;
  outline: none;
}

.nav-tool-results {
  max-height: 18rem;
  overflow: auto;
  padding: 0.45rem;
}

.nav-tool-results button {
  display: flex;
  width: 100%;
  gap: 0.62rem;
  align-items: center;
  border: 0;
  border-radius: 0.55rem;
  background: transparent;
  padding: 0.52rem;
  cursor: pointer;
  text-align: left;
  transition: background 160ms ease;
}

.nav-tool-results button:hover {
  background: rgba(254, 215, 170, 0.5);
}

.nav-tool-results img {
  width: 2rem;
  height: 2rem;
  flex: 0 0 auto;
  border-radius: 0.45rem;
  object-fit: cover;
  background: rgba(255, 237, 213, 0.9);
}

.nav-tool-results span {
  min-width: 0;
}

.nav-tool-results strong,
.nav-tool-results small {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.nav-tool-results strong {
  font-size: 0.78rem;
}

.nav-tool-results small {
  margin-top: 0.1rem;
  color: rgba(68, 64, 60, 0.65);
  font-size: 0.68rem;
}

.nav-tool-empty {
  padding: 1.2rem 0.9rem;
  color: rgba(68, 64, 60, 0.68);
  font-size: 0.78rem;
}
</style>
