<template>
  <div class="game-detail-sidebar-card space-y-3 p-4 text-sm">
    <div class="flex justify-between gap-x-2">
      <!-- Steam -->
      <button
          v-if="game?.appid"
          @click="goSteam"
          class="game-detail-action game-detail-action--primary w-full py-2"
      >
        {{t("game.detail.goToSteamPage")}}
      </button>
      <!-- 点评按钮 -->
      <button
          class="game-detail-action game-detail-action--secondary w-full py-2"
          @click="showComment = true"
      >
        {{ t("game.detail.goComment") }}
      </button>
    </div>


    <!-- 资源 -->
    <div v-if="game?.resources?.length">
      <h3 class="game-detail-sidebar-title mb-2 font-semibold">{{t("game.detail.resources")}}</h3>
      <div class="flex flex-wrap gap-2">
        <LinkTag
            v-for="(item, i) in game.resources"
            :key="item.key + i"
            :item="item"
        />
      </div>
    </div>

    <!-- 社群 -->
    <div v-if="safeGroups.length">
      <h3 class="game-detail-sidebar-title mb-2 font-semibold">{{t("game.detail.community")}}</h3>
      <SiteIconList :items="safeGroups" />
    </div>

    <!-- 相关链接 -->
    <div v-if="safeLinks.length">
      <h3 class="game-detail-sidebar-title mb-2 font-semibold">{{t("game.detail.relatedLinks")}}</h3>
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
import { i18n } from '@/main'

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
