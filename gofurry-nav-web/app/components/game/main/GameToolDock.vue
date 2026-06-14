<template>
  <div class="game-tool-dock">
    <Transition name="game-tool-panel-transition" mode="out-in">
      <RagPromptPanel
        v-if="activePanel === 'ask'"
        :title="t('game.tools.askTitle')"
        :description="t('game.tools.askDescription')"
        :placeholder="t('game.tools.askPlaceholder')"
        :submit-label="t('game.tools.askSubmit')"
        :templates="gamePromptTemplates"
        @ask="openArchivePrompt"
      />
    </Transition>

    <nav
      class="game-tool-rail"
      :aria-label="t('game.tools.label')"
    >
      <div class="game-tool-rail__primary">
        <button
          v-for="tool in tools"
          :key="tool.key"
          type="button"
          class="game-tool-button"
          :class="{ active: tool.panel && activePanel === tool.panel }"
          :title="tool.label"
          :aria-label="tool.label"
          @click="tool.action"
        >
          <span class="game-tool-icon-stack" aria-hidden="true">
            <img
              v-if="tool.image"
              class="game-tool-icon"
              :class="{ 'game-tool-icon--cover': tool.cover }"
              :src="tool.image"
              alt=""
              draggable="false"
            />
          </span>
        </button>
      </div>

      <a
        class="game-tool-feedback"
        href="https://github.com/gofurry/gofurry-nav-site/issues"
        target="_blank"
        rel="noopener noreferrer"
        :title="t('nav.tools.feedback')"
        :aria-label="t('nav.tools.feedback')"
      >
        <span class="game-tool-icon-stack" aria-hidden="true">
          <img class="game-tool-icon" :src="feedbackIconSrc" alt="" />
        </span>
        <span>{{ t('nav.tools.feedbackShort') }}</span>
      </a>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { i18n } from '@/main'
import RagPromptPanel from '@/components/common/RagPromptPanel.vue'
import { useThemeStore } from '@/stores/theme'
import askIconDark from '@/assets/svgs/ai-duotone-dark.svg'
import askIconLight from '@/assets/svgs/ai-duotone.svg'
import feedbackIconDark from '@/assets/svgs/ai-note-alt-1-duotone-dark.svg'
import feedbackIconLight from '@/assets/svgs/ai-note-alt-1-duotone.svg'

type RagPromptTemplate = {
  id: string
  title: string
  description: string
  prompt: string
}

const router = useRouter()
const { t } = i18n.global
const themeStore = useThemeStore()
const activePanel = ref<'ask' | null>(null)
const isDarkTheme = computed(() => themeStore.theme === 'dark')
const askIconSrc = computed(() => isDarkTheme.value ? askIconDark : askIconLight)
const feedbackIconSrc = computed(() => isDarkTheme.value ? feedbackIconDark : feedbackIconLight)

const gamePromptTemplates = computed<RagPromptTemplate[]>(() => [
  {
    id: 'recent',
    title: t('game.tools.prompts.recent.title'),
    description: t('game.tools.prompts.recent.description'),
    prompt: t('game.tools.prompts.recent.prompt'),
  },
  {
    id: 'similar',
    title: t('game.tools.prompts.similar.title'),
    description: t('game.tools.prompts.similar.description'),
    prompt: t('game.tools.prompts.similar.prompt'),
  },
  {
    id: 'beginner',
    title: t('game.tools.prompts.beginner.title'),
    description: t('game.tools.prompts.beginner.description'),
    prompt: t('game.tools.prompts.beginner.prompt'),
  },
])

const tools = computed(() => [
  {
    key: 'lottery',
    label: t('game.tools.lottery'),
    image: 'https://qcdn.go-furry.com/game/background/steam.jpg',
    cover: true,
    panel: null,
    action: () => router.push('/games/prize'),
  },
  {
    key: 'ask',
    label: t('game.tools.ask'),
    image: askIconSrc.value,
    cover: false,
    panel: 'ask' as const,
    action: () => {
      activePanel.value = activePanel.value === 'ask' ? null : 'ask'
    },
  },
])

function openArchivePrompt(prompt: string) {
  activePanel.value = null
  router.push({
    path: '/archive',
    query: {
      q: prompt,
      scene: 'games',
    },
  })
}
</script>
