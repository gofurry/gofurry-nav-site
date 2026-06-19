<template>
  <div class="space-y-6">

    <!-- 详细描述 -->
    <section v-if="game?.detailed_description">
<!--      <h3 class="text-base font-semibold text-gray-800 mb-2">-->
<!--        {{t("game.detail.detailedIntroduction")}}-->
<!--      </h3>-->

      <div
          class="prose prose-sm max-w-none
               prose-img:rounded-lg
               prose-a:text-orange-500
               prose-strong:text-gray-800"
          v-html="game.detailed_description"
      />
    </section>

    <!-- 游戏介绍 -->
<!--    <section v-if="game?.about_the_game">-->
<!--      <h3 class="text-base font-semibold text-gray-800 mb-2">-->
<!--        {{t("game.detail.aboutGame")}}-->
<!--      </h3>-->

<!--      <div-->
<!--          class="prose prose-sm max-w-none-->
<!--               prose-img:rounded-lg-->
<!--               prose-a:text-orange-500-->
<!--               prose-strong:text-gray-800"-->
<!--          v-html="game.about_the_game"-->
<!--      />-->
<!--    </section>-->

    <!-- 无内容占位 -->
    <div
        v-if="!game?.detailed_description && !game?.about_the_game"
        class="text-sm text-gray-500"
    >
      {{t("game.panel.none")}}
    </div>

  </div>
</template>

<script setup lang="ts">
import type { GameBaseInfoResponse } from '@/types/game'
import { i18n } from '@/main.ts'

const { t } = i18n.global

defineProps<{
  game: GameBaseInfoResponse | null
}>()
</script>
