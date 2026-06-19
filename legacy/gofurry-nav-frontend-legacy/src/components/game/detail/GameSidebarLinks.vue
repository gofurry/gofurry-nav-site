<template>
  <div class="bg-orange-50 rounded-2xl p-4 shadow text-sm space-y-3">
    <div class="flex justify-between gap-x-2">
      <!-- Steam -->
      <button
          v-if="game?.appid"
          @click="goSteam"
          class="w-full py-2 rounded-lg bg-orange-400 text-white hover:bg-orange-400/70 transition"
      >
        {{t("game.detail.goToSteamPage")}}
      </button>
      <!-- 点评按钮 -->
      <button
          class="w-full py-2 rounded-lg bg-orange-200 text-orange-800 hover:bg-orange-300/80 transition"
          @click="showComment = true"
      >
        {{ t("game.detail.goComment") }}
      </button>
    </div>


    <!-- 资源 -->
    <div v-if="game?.resources?.length">
      <h3 class="font-semibold mb-2">{{t("game.detail.resources")}}</h3>
      <div class="flex flex-wrap gap-2 text-orange-700">
        <LinkTag
            v-for="(item, i) in game.resources"
            :key="item.key + i"
            :item="item"
        />
      </div>
    </div>

    <!-- 社群 -->
    <div v-if="safeGroups.length">
      <h3 class="font-semibold mb-2">{{t("game.detail.community")}}</h3>
      <SiteIconList :items="safeGroups" />
    </div>

    <!-- 相关链接 -->
    <div v-if="safeLinks.length">
      <h3 class="font-semibold mb-2">{{t("game.detail.relatedLinks")}}</h3>
      <SiteIconList :items="safeLinks" />
    </div>
  </div>

  <GameCommentDialog
      :visible="showComment"
      :game-name="game?.name || ''"
      @close="showComment = false"
  />
</template>

<script setup lang="ts">
import type { GameBaseInfoResponse, KvModel } from '@/types/game'
import LinkTag from '@/components/common/LinkTag.vue'
import SiteIconList from '@/components/common/SiteIconList.vue'
import GameCommentDialog from '@/components/game/detail/GameCommentDialog.vue'
import {computed, ref} from "vue";
import { i18n } from '@/main.ts'

const { t } = i18n.global

const showComment = ref(false)

const steamPrefix = import.meta.env.VITE_STEAM_APP_PREFIX_URL || ''

const props = defineProps<{
  game: GameBaseInfoResponse | null
}>()

const goSteam = () => {
  if (!props.game?.appid) return
  window.open(steamPrefix+`${props.game.appid}`, '_blank')
}

// 过滤数组
const safeGroups = computed<KvModel[]>(() =>
    (props.game?.groups ?? []).filter(
        (item): item is KvModel => !!item?.key && !!item?.value
    )
)

const safeLinks = computed<KvModel[]>(() =>
    (props.game?.links ?? []).filter(
        (item): item is KvModel => !!item?.key && !!item?.value
    )
)
</script>
