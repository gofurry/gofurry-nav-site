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
import askIconLight from '@/assets/svgs/ai-duotone.svg'
import feedbackIconLight from '@/assets/svgs/ai-note-alt-1-duotone.svg'

type RagPromptTemplate = {
  id: string
  title: string
  description: string
  prompt: string
}

const router = useRouter()
const { t } = i18n.global
const activePanel = ref<'ask' | null>(null)
const feedbackIconSrc = feedbackIconLight

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
    image: askIconLight,
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

<style scoped>
.game-tool-dock {
  position: fixed;
  top: 6rem;
  right: 1rem;
  bottom: 5.6rem;
  z-index: 40;
  display: none;
  align-items: flex-start;
  gap: 0.5rem;
  pointer-events: none;
}

@media (min-width: 1024px) {
  .game-tool-dock {
    display: flex;
  }
}

.game-tool-rail {
  display: flex;
  min-height: 100%;
  flex-direction: column;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  pointer-events: none;
}

.game-tool-rail__primary {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  pointer-events: auto;
}

.game-tool-button,
.game-tool-feedback {
  border-radius: 0.65rem;
}

.game-tool-button {
  position: relative;
  display: grid;
  width: 2.75rem;
  height: 2.75rem;
  place-items: center;
  overflow: hidden;
}

.game-tool-icon-stack {
  display: grid;
  width: 1.55rem;
  height: 1.55rem;
  place-items: center;
}

.game-tool-icon {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.game-tool-icon--cover {
  width: 100%;
  height: 100%;
  border-radius: 0.48rem;
  object-fit: cover;
}

.game-tool-button:has(.game-tool-icon--cover) .game-tool-icon-stack {
  width: 2.75rem;
  height: 2.75rem;
}

.game-tool-feedback {
  display: grid;
  width: 2.75rem;
  min-height: 4.3rem;
  place-items: center;
  gap: 0.26rem;
  font-size: 0.68rem;
  font-weight: 650;
  line-height: 1;
  text-decoration: none;
  pointer-events: auto;
}

.game-tool-feedback > span:last-child {
  max-width: 100%;
  overflow: hidden;
  padding: 0 0.12rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.game-tool-panel-transition-enter-active,
.game-tool-panel-transition-leave-active {
  transition:
    opacity 500ms ease,
    transform 500ms cubic-bezier(0.22, 1, 0.36, 1),
    filter 500ms ease;
}

.game-tool-panel-transition-enter-from,
.game-tool-panel-transition-leave-to {
  opacity: 0;
  transform: translateX(10px) scale(0.975);
  filter: blur(6px);
}
</style>
