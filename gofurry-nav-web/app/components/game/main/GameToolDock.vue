<template>
  <nav
    class="fixed right-4 top-24 z-40 hidden flex-col gap-2 lg:flex"
    :aria-label="t('game.tools.label')"
  >
    <button
      v-for="tool in tools"
      :key="tool.key"
      type="button"
      class="group relative grid size-11 place-items-center overflow-hidden rounded-lg border border-white/55 bg-white/70 text-slate-700 shadow-[0_12px_32px_rgba(76,42,18,0.14)] backdrop-blur-xl transition duration-200 hover:border-orange-200 hover:bg-white/[0.88] hover:text-orange-700"
      :title="tool.label"
      :aria-label="tool.label"
      @click="tool.action"
    >
      <span
        class="absolute inset-x-2 top-1 h-px bg-gradient-to-r from-transparent via-orange-200/80 to-transparent opacity-80"
        aria-hidden="true"
      />
      <img
        :src="tool.image"
        :alt="tool.label"
        class="size-full object-cover opacity-90 transition duration-200 group-hover:scale-105 group-hover:opacity-100"
        draggable="false"
      />
      <span class="sr-only">{{ tool.label }}</span>
    </button>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { i18n } from '@/main'

const router = useRouter()
const { t } = i18n.global

const tools = computed(() => [
  {
    key: 'lottery',
    label: t('game.tools.lottery'),
    image: 'https://qcdn.go-furry.com/game/background/steam.jpg',
    action: () => router.push('/games/prize'),
  },
])
</script>
