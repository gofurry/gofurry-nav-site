<template>
  <div class="fixed right-4 top-24 z-40 hidden items-start gap-2 lg:flex">
    <RagPromptPanel
      v-if="activePanel === 'ask'"
      :title="t('game.tools.askTitle')"
      :description="t('game.tools.askDescription')"
      :placeholder="t('game.tools.askPlaceholder')"
      :submit-label="t('game.tools.askSubmit')"
      :templates="gamePromptTemplates"
      @ask="openArchivePrompt"
    />

    <nav
      class="flex flex-col gap-2"
      :aria-label="t('game.tools.label')"
    >
      <button
        v-for="tool in tools"
        :key="tool.key"
        type="button"
        class="group relative grid size-11 place-items-center overflow-hidden rounded-lg border border-white/55 bg-white/70 text-slate-700 shadow-[0_12px_32px_rgba(76,42,18,0.14)] backdrop-blur-xl transition duration-200 hover:border-orange-200 hover:bg-white/[0.88] hover:text-orange-700"
        :class="{ 'border-orange-200 bg-white/[0.9] text-orange-700': tool.panel && activePanel === tool.panel }"
        :title="tool.label"
        :aria-label="tool.label"
        @click="tool.action"
      >
        <span
          class="absolute inset-x-2 top-1 h-px bg-gradient-to-r from-transparent via-orange-200/80 to-transparent opacity-80"
          aria-hidden="true"
        />
        <img
          v-if="tool.image"
          :src="tool.image"
          :alt="tool.label"
          class="size-full object-cover opacity-90 transition duration-200 group-hover:scale-105 group-hover:opacity-100"
          draggable="false"
        />
        <svg v-else viewBox="0 0 24 24" aria-hidden="true" class="size-5 fill-none stroke-current stroke-[1.9]">
          <path d="M12 3.5 13.7 9l5.3 1.7-5.3 1.7L12 18l-1.7-5.6L5 10.7 10.3 9 12 3.5Z" />
          <path d="M18 15.5 18.8 18l2.2.8-2.2.7L18 22l-.8-2.5-2.2-.7 2.2-.8.8-2.5Z" />
        </svg>
        <span class="sr-only">{{ tool.label }}</span>
      </button>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { i18n } from '@/main'
import RagPromptPanel from '@/components/common/RagPromptPanel.vue'

type RagPromptTemplate = {
  id: string
  title: string
  description: string
  prompt: string
}

const router = useRouter()
const { t } = i18n.global
const activePanel = ref<'ask' | null>(null)

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
    panel: null,
    action: () => router.push('/games/prize'),
  },
  {
    key: 'ask',
    label: t('game.tools.ask'),
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
