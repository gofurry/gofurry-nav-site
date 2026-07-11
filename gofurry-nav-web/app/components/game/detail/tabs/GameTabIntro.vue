<template>
  <div class="game-detail-prose space-y-6">

    <section v-if="game?.about_the_game" class="space-y-2">
      <div
        class="max-w-none"
        v-html="game.about_the_game"
      />
    </section>

    <div
      v-if="!game?.about_the_game"
      class="game-detail-empty text-sm"
    >
      {{ t("game.panel.none") }}
    </div>

  </div>
</template>

<script setup lang="ts">
import type { GameBaseInfoResponse } from '@/types/game'
import { i18n } from '@/main'

const { t } = i18n.global

defineProps<{
  game: GameBaseInfoResponse | null
}>()
</script>
