<template>
  <div class="game-tool-dock">
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
import { computed } from 'vue'
import { i18n } from '@/main'
import { useThemeStore } from '@/stores/theme'
import feedbackIconDark from '@/assets/svgs/ai-note-alt-1-duotone-dark.svg'
import feedbackIconLight from '@/assets/svgs/ai-note-alt-1-duotone.svg'

const router = useRouter()
const { t } = i18n.global
const themeStore = useThemeStore()
const isDarkTheme = computed(() => themeStore.theme === 'dark')
const feedbackIconSrc = computed(() => isDarkTheme.value ? feedbackIconDark : feedbackIconLight)

const tools = computed(() => [
  {
    key: 'lottery',
    label: t('game.tools.lottery'),
    image: 'https://qcdn.go-furry.com/game/background/steam.jpg',
    cover: true,
    action: () => router.push('/games/prize'),
  }
])
</script>
