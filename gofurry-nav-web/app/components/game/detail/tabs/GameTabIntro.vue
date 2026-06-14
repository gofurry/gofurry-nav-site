<template>
  <div class="game-detail-prose space-y-6">

    <!-- 详细描述 -->
    <section v-if="game?.detailed_description">
      <div
          class="max-w-none"
          v-html="game.detailed_description"
      />
    </section>

    <!-- 无内容占位 -->
    <div
        v-if="!game?.detailed_description && !game?.about_the_game"
        class="game-detail-empty text-sm"
    >
      {{t("game.panel.none")}}
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
