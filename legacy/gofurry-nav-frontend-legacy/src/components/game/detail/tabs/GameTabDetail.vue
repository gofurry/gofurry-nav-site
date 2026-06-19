<template>
  <div class="space-y-6 text-sm text-gray-700">

    <!-- 基本信息 -->
    <section class="grid grid-cols-1 md:grid-cols-2 gap-x-6 gap-y-3">
      <div class="flex gap-2">
        <span class="text-gray-500 w-28 shrink-0">{{t("game.detail.infoCollectedTime")}}:</span>
        <span>{{ game?.create_time || t("game.panel.none") }}</span>
      </div>

      <div class="flex gap-2">
        <span class="text-gray-500 w-28 shrink-0">{{t("game.detail.infoUpdatedTime")}}:</span>
        <span>{{ game?.update_time || t("game.panel.none") }}</span>
      </div>

      <div class="flex gap-2">
        <span class="text-gray-500 w-28 shrink-0">{{t("game.detail.releaseDate")}}:</span>
        <span>{{ game?.release_date || t("game.panel.none") }}</span>
      </div>

      <div class="flex gap-2">
        <span class="text-gray-500 w-28 shrink-0">{{t("game.detail.supportedPlatforms")}}:</span>
        <span>{{ game?.platform || t("game.panel.none") }}</span>
      </div>

      <div class="flex gap-2">
        <span class="text-gray-500 w-28 shrink-0">{{t("game.detail.supportedLanguages")}}:</span>
        <span>{{ game?.supported_languages || t("game.panel.none") }}</span>
      </div>

      <div class="flex gap-2">
        <span class="text-gray-500 w-28 shrink-0">{{t("game.detail.ageRestriction")}}:</span>
        <span>{{ game?.required_age || t("game.panel.none") }}</span>
      </div>
    </section>

    <!-- 开发商 / 发行商 -->
    <section class="space-y-3">
      <div>
        <h4 class="font-bold text-gray-800 mb-1">{{t("game.detail.developer")}}</h4>
        <div class="flex flex-wrap gap-2">
          <span
              v-for="(d, i) in game?.developers || []"
              :key="i"
              class="px-2 py-0.5 text-xs rounded bg-orange-100 text-orange-700"
          >
            {{ d }}
          </span>
          <span v-if="!game?.developers?.length" class="text-gray-400 text-sm">
            {{t("game.panel.none")}}
          </span>
        </div>
      </div>

      <div>
        <h4 class="font-bold text-gray-800 mb-1">{{t("game.detail.publisher")}}</h4>
        <div class="flex flex-wrap gap-2">
          <span
              v-for="(p, i) in game?.publishers || []"
              :key="i"
              class="px-2 py-0.5 text-xs rounded bg-orange-100 text-orange-700"
          >
            {{ p }}
          </span>
          <span v-if="!game?.publishers?.length" class="text-gray-400 text-sm">
            {{t("game.panel.none")}}
          </span>
        </div>
      </div>
    </section>

    <!-- 价格 -->
    <section v-if="game?.price_list?.length" class="space-y-2">
      <h4 class="font-bold text-gray-800">{{t("game.detail.priceInfo")}}</h4>
      <div class="flex gap-x-1 sm:grid-cols-3 gap-2">
        <div
            v-for="(p, i) in game.price_list"
            :key="i"
            class="flex justify-center items-center px-3 py-1 rounded-lg bg-orange-100 text-orange-700"
        >
          <span class="font-medium">
            <strong>{{ countryMap[p.country] || p.country }}</strong>
            {{ p.price }}
          </span>
        </div>
      </div>
    </section>

    <!-- 官网 -->
    <section v-if="game?.website" class="space-y-1">
      <h4 class="font-bold text-gray-800">{{t("game.detail.officialWebsite")}}</h4>
      <div class="text-orange-500 hover:text-orange-400 hover:underline break-all">
        <a
            :href="game.website"
            target="_blank"
            class=""
        >
          {{ game.website }}
        </a>
      </div>
    </section>

    <!-- PC 配置 -->
    <section v-if="game?.pc_requirements" class="space-y-4">
      <h4 class="font-bold text-gray-800">{{t("game.detail.pcRequirements")}}</h4>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="bg-orange-50 rounded-xl p-4">
          <div
              v-html="game.pc_requirements.minimum || t('game.panel.none')"
              class="leading-relaxed"
          />
        </div>

        <div class="bg-orange-50 rounded-xl p-4">
          <div
              v-html="game.pc_requirements.recommended || t('game.panel.none')"
              class="leading-relaxed"
          />
        </div>
      </div>
    </section>

  </div>
</template>

<script setup lang="ts">
import type { GameBaseInfoResponse } from '@/types/game'
import { i18n } from '@/main.ts'

const { t } = i18n.global

defineProps<{
  game: GameBaseInfoResponse | null
}>()

const countryMap: Record<string, string> = {
  US: t("game.detail.globalRegion"),
  CN: t("game.detail.chinaRegion"),
  HK: t("game.detail.hongKongRegion")
}
</script>
